package persistence

import (
	"log"
	"sort"
	"strings"
	"swiki/model"
	"swiki/similarity"
)

const similarScoreThreshold = 0.45

var pagesFTSLastError string

var pagesFTS = `CREATE VIRTUAL TABLE IF NOT EXISTS pages_fts USING fts5(
	title,
	content,
	content='pages',
	content_rowid='id'
)`

var pagesFTSRebuild = `INSERT INTO pages_fts(pages_fts) VALUES('rebuild')`

var pagesFTSTriggers = []string{
	`CREATE TRIGGER IF NOT EXISTS pages_ai AFTER INSERT ON pages BEGIN
		INSERT INTO pages_fts(rowid, title, content)
		VALUES (new.id, new.title, new.content);
	END`,
	`CREATE TRIGGER IF NOT EXISTS pages_ad AFTER DELETE ON pages BEGIN
		INSERT INTO pages_fts(pages_fts, rowid, title, content)
		VALUES ('delete', old.id, old.title, old.content);
	END`,
	`CREATE TRIGGER IF NOT EXISTS pages_au AFTER UPDATE ON pages BEGIN
		INSERT INTO pages_fts(pages_fts, rowid, title, content)
		VALUES ('delete', old.id, old.title, old.content);
		INSERT INTO pages_fts(rowid, title, content)
		VALUES (new.id, new.title, new.content);
	END`,
}

var pagesFTSDropTriggers = []string{
	`DROP TRIGGER IF EXISTS pages_ai`,
	`DROP TRIGGER IF EXISTS pages_ad`,
	`DROP TRIGGER IF EXISTS pages_au`,
}

type SimilarPage struct {
	Id       int     `json:"id"`
	Title    string  `json:"title"`
	Category string  `json:"category"`
	Score    float64 `json:"score"`
}

type PagesFTSStatus struct {
	TableExists  bool   `json:"table_exists"`
	PagesCount   int    `json:"pages_count"`
	ProbeTerm    string `json:"probe_term"`
	ProbeMatches int    `json:"probe_matches"`
	LastError    string `json:"last_error"`
}

type pageCandidate struct {
	id       int
	title    string
	category string
	content  string
}

func setPagesFTSError(format string, err error) {
	pagesFTSLastError = err.Error()
	log.Printf(format, err)
}

func EnsurePagesFTS() {
	if db == nil {
		return
	}

	for _, stmt := range pagesFTSDropTriggers {
		if _, err := db.Exec(stmt); err != nil {
			setPagesFTSError("Could not drop FTS trigger: %v", err)
			return
		}
	}

	if _, err := db.Exec(dropPagesFTS); err != nil {
		setPagesFTSError("Could not drop pages_fts: %v", err)
		return
	}

	if _, err := db.Exec(pagesFTS); err != nil {
		setPagesFTSError("Could not create pages_fts (FTS5 unavailable?): %v", err)
		return
	}

	if _, err := db.Exec(pagesFTSRebuild); err != nil {
		setPagesFTSError("Could not rebuild pages_fts: %v", err)
		return
	}

	for _, stmt := range pagesFTSTriggers {
		if _, err := db.Exec(stmt); err != nil {
			setPagesFTSError("Could not create FTS trigger: %v", err)
			return
		}
	}

	var pagesCount int
	if err := db.QueryRow("SELECT count(*) FROM pages").Scan(&pagesCount); err == nil {
		log.Printf("pages_fts ready (%d pages)", pagesCount)
	}

	pagesFTSLastError = ""
}

func FindSimilarPages(page model.Page, limit int) ([]SimilarPage, error) {
	if limit <= 0 {
		limit = 10
	}

	candidates := make([]pageCandidate, 0)
	category := strings.ToLower(strings.TrimSpace(page.Category))

	const candidateLimit = 20

	matchQuery := similarity.BuildMatchQuery(page.Title+" "+page.Content, 24)
	if matchQuery != "" {
		ftsMatches, err := ftsCandidates(matchQuery, category, candidateLimit)
		if err != nil {
			log.Printf("similar: FTS search failed, using LIKE fallback: %v", err)
		} else {
			candidates = ftsMatches
		}
	}

	if len(candidates) == 0 {
		likeMatches, err := likeCandidates(category, page, candidateLimit)
		if err != nil {
			return nil, err
		}
		candidates = likeMatches
	}

	results := scoreCandidates(page, candidates, limit)
	log.Printf("similar: %q -> %d candidates, %d above threshold", page.Title, len(candidates), len(results))

	return results, nil
}

func ftsCandidates(matchQuery, category string, limit int) ([]pageCandidate, error) {
	query := `
		SELECT p.id, p.title, p.category, p.content
		FROM pages_fts
		JOIN pages p ON p.id = pages_fts.rowid
		WHERE pages_fts MATCH ?`
	args := []any{matchQuery}

	if category != "" {
		query += " AND p.category = ?"
		args = append(args, category)
	}

	query += " ORDER BY rank LIMIT ?"
	args = append(args, limit)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := make([]pageCandidate, 0)
	for rows.Next() {
		var c pageCandidate
		if err := rows.Scan(&c.id, &c.title, &c.category, &c.content); err != nil {
			return nil, err
		}
		candidates = append(candidates, c)
	}
	return candidates, rows.Err()
}

func likeCandidates(category string, page model.Page, limit int) ([]pageCandidate, error) {
	terms := similarity.TopTerms(page.Title+" "+page.Content, 6)
	if len(terms) == 0 {
		return nil, nil
	}

	clauses := make([]string, 0, len(terms)*2)
	args := make([]any, 0, len(terms)*2+2)
	for _, term := range terms {
		clauses = append(clauses, "title LIKE ?", "content LIKE ?")
		pattern := "%" + term + "%"
		args = append(args, pattern, pattern)
	}

	query := "SELECT id, title, category, content FROM pages WHERE "
	if category != "" {
		query += "category = ? AND "
		args = append([]any{category}, args...)
	}
	query += "(" + strings.Join(clauses, " OR ") + ") LIMIT ?"
	args = append(args, limit)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := make([]pageCandidate, 0)
	for rows.Next() {
		var c pageCandidate
		if err := rows.Scan(&c.id, &c.title, &c.category, &c.content); err != nil {
			return nil, err
		}
		candidates = append(candidates, c)
	}
	return candidates, rows.Err()
}

func scoreCandidates(page model.Page, candidates []pageCandidate, limit int) []SimilarPage {
	results := make([]SimilarPage, 0, len(candidates))

	for _, candidate := range candidates {
		if page.Id != 0 && candidate.id == page.Id {
			continue
		}

		score := similarity.Score(
			page.Title, page.Content,
			candidate.title, candidate.content,
			page.Category != "" && strings.EqualFold(page.Category, candidate.category),
		)

		if score < similarScoreThreshold {
			continue
		}

		results = append(results, SimilarPage{
			Id:       candidate.id,
			Title:    candidate.title,
			Category: candidate.category,
			Score:    score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Id < results[j].Id
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results
}

func GetPagesFTSStatus() PagesFTSStatus {
	status := PagesFTSStatus{LastError: pagesFTSLastError}

	if db == nil {
		status.LastError = "database not initialized"
		return status
	}

	var tableCount int
	if err := db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='pages_fts'").Scan(&tableCount); err != nil {
		status.LastError = err.Error()
		return status
	}
	status.TableExists = tableCount > 0

	if err := db.QueryRow("SELECT count(*) FROM pages").Scan(&status.PagesCount); err != nil {
		status.LastError = err.Error()
		return status
	}

	if !status.TableExists {
		return status
	}

	var title, content string
	if err := db.QueryRow("SELECT title, content FROM pages LIMIT 1").Scan(&title, &content); err != nil {
		return status
	}

	terms := similarity.TopTerms(title+" "+content, 1)
	if len(terms) == 0 {
		return status
	}

	status.ProbeTerm = terms[0]
	probe := `"` + strings.ReplaceAll(terms[0], `"`, `""`) + `"`
	if err := db.QueryRow("SELECT count(*) FROM pages_fts WHERE pages_fts MATCH ?", probe).Scan(&status.ProbeMatches); err != nil {
		status.LastError = err.Error()
	}

	return status
}
