package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"swiki/helpers"
	"swiki/model"
	"swiki/persistence"
	"swiki/persistencelocal"
	"swiki/search"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
)

func AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	if !persistencelocal.TryStartProgress("analyse") {
		http.Error(w, "Er loopt al een proces", http.StatusConflict)
		return
	}

	if err := persistencelocal.DeleteAnalysisState(); err != nil {
		persistencelocal.FinishProgress(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeAnalyzeMessage(w, "Analyse gestart")

	go runAnalysis(nil)
}

func AnalyzePauseHandler(w http.ResponseWriter, r *http.Request) {
	if !persistencelocal.RequestPause() {
		http.Error(w, "Er loopt geen analyse", http.StatusConflict)
		return
	}

	writeAnalyzeMessage(w, "Pauze aangevraagd")
}

func AnalyzeResumeHandler(w http.ResponseWriter, r *http.Request) {
	state, err := persistencelocal.LoadAnalysisState()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if state == nil || state.Completed {
		http.Error(w, "Geen gepauzeerde analyse gevonden", http.StatusNotFound)
		return
	}

	if !persistencelocal.TryStartProgress("analyse") {
		http.Error(w, "Er loopt al een proces", http.StatusConflict)
		return
	}

	writeAnalyzeMessage(w, "Analyse hervat")

	go runAnalysis(state)
}

func AnalyzeStateHandler(w http.ResponseWriter, r *http.Request) {
	info, err := persistencelocal.GetAnalysisStateInfo()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeAnalyzeJSON(w, info)
}

func AnalyzeResultHandler(w http.ResponseWriter, r *http.Request) {
	pairs, err := persistencelocal.GetAnalysisResults()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeAnalyzeJSON(w, pairs)
}

func AnalyzePageHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	page, err := persistencelocal.GetPage(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if page.Id == 0 {
		http.Error(w, "page not found", http.StatusNotFound)
		return
	}

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	page.Content = string(markdown.ToHTML([]byte(page.Content), nil, html.NewRenderer(opts)))

	imagesFound := helpers.ProcessImagesFromHtml(page.Content)
	if len(imagesFound) > 0 {
		replaceImageTags(imagesFound, &page.Content)
	}

	writeAnalyzeJSON(w, page)
}

func AnalyzeDeletePageHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	page, err := persistencelocal.GetPage(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if page.Id == 0 {
		http.Error(w, "page not found", http.StatusNotFound)
		return
	}

	if page.TursoId > 0 {
		if err := persistence.DeletePage(page.TursoId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		search.DeletePage(page.TursoId)
	}

	if err := persistencelocal.DeletePage(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := persistencelocal.RemoveAnalysisPairsForPage(id); err != nil {
		log.Printf("could not remove analysis pairs for page %d: %v", id, err)
	}

	writeAnalyzeMessage(w, "Verwijderd: "+strconv.Itoa(id))
}

func runAnalysis(resume *persistencelocal.AnalysisState) {
	defer func() {
		if rec := recover(); rec != nil {
			persistencelocal.FinishProgress(fmt.Errorf("%v", rec))
		}
	}()

	paused, err := persistencelocal.RunAnalysis(resume)
	if err != nil {
		persistencelocal.FinishProgress(err)
		return
	}
	if paused {
		persistencelocal.FinishPausedProgress()
		return
	}

	persistencelocal.FinishProgress(nil)
}

func writeAnalyzeMessage(w http.ResponseWriter, message string) {
	writeAnalyzeJSON(w, model.ResponseMessage{Message: message})
}

func writeAnalyzeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
