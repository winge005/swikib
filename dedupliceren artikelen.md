Ja. Ik zou dit niet oplossen met alleen een SQL-query, maar met een combinatie van SQLite + Go. Je hebt eigenlijk drie soorten “dubbel”:

exact dezelfde tekst;
grotendeels dezelfde tekst;
dezelfde informatie, maar anders geformuleerd.

Voor jouw doel — artikelen die ongeveer hetzelfde behandelen — zou ik beginnen met FTS5 + een similarity-score in Go. Als je later ook semantische overeenkomsten wilt herkennen, kun je embeddings toevoegen.

1. Eerst zou ik je tabel iets aanpassen

Je id is nu geen primary key. Dat zou ik sowieso veranderen:

CREATE TABLE IF NOT EXISTS pages (
id INTEGER PRIMARY KEY,
category TEXT,
title TEXT,
content TEXT,
created TEXT,
updated TEXT
);

In SQLite hebben varchar(200) en varchar(255) trouwens nauwelijks voordeel boven TEXT; SQLite handhaaft die lengtelimieten niet automatisch.

2. Gebruik SQLite FTS5 om vergelijkbare artikelen te vinden

SQLite heeft hiervoor een ingebouwde full-text search engine: FTS5. Daarmee kun je eerst een kleine groep kandidaten vinden in plaats van ieder artikel met ieder ander artikel te vergelijken.

Je kunt bijvoorbeeld een index maken:

CREATE VIRTUAL TABLE pages_fts USING fts5(
title,
content,
content='pages',
content_rowid='id'
);

Voor bestaande artikelen moet je daarna één keer de index vullen:

INSERT INTO pages_fts(rowid, title, content)
SELECT id, title, content
FROM pages;

En daarna triggers maken zodat de index automatisch actueel blijft:

CREATE TRIGGER pages_ai AFTER INSERT ON pages BEGIN
INSERT INTO pages_fts(rowid, title, content)
VALUES (new.id, new.title, new.content);
END;

CREATE TRIGGER pages_ad AFTER DELETE ON pages BEGIN
INSERT INTO pages_fts(pages_fts, rowid, title, content)
VALUES ('delete', old.id, old.title, old.content);
END;

CREATE TRIGGER pages_au AFTER UPDATE ON pages BEGIN
INSERT INTO pages_fts(pages_fts, rowid, title, content)
VALUES ('delete', old.id, old.title, old.content);

    INSERT INTO pages_fts(rowid, title, content)
    VALUES (new.id, new.title, new.content);
END;

Dit is ook de aanpak die SQLite zelf aanbeveelt voor een FTS5-index boven een gewone tabel.

Maar FTS5 alleen zegt nog niet echt:

"artikel A en artikel B zijn voor 83% hetzelfde."

Daarom zou ik FTS gebruiken voor kandidaatselectie en daarna Go laten bepalen hoe sterk de artikelen daadwerkelijk overeenkomen.

3. In Go kun je bijvoorbeeld Jaccard similarity gebruiken

Een relatief eenvoudige methode is om woorden of groepjes woorden met elkaar te vergelijken.

Stel:

Artikel A:
Rusland viel Oekraïne binnen in februari 2022.

Artikel B:
In februari 2022 begon Rusland de invasie van Oekraïne.

Ze zijn niet letterlijk hetzelfde, maar bevatten veel dezelfde belangrijke woorden.

Een eenvoudige Go-versie:

package similarity

import (
"regexp"
"strings"
)

var nonWord = regexp.MustCompile(`[^\p{L}\p{N}]+`)

func tokenize(s string) map[string]struct{} {
s = strings.ToLower(s)
s = nonWord.ReplaceAllString(s, " ")

	words := strings.Fields(s)

	result := make(map[string]struct{})

	for _, word := range words {
		if len(word) < 3 {
			continue
		}

		result[word] = struct{}{}
	}

	return result
}

func Jaccard(a, b string) float64 {
setA := tokenize(a)
setB := tokenize(b)

	if len(setA) == 0 && len(setB) == 0 {
		return 1
	}

	intersection := 0

	for word := range setA {
		if _, exists := setB[word]; exists {
			intersection++
		}
	}

	union := len(setA) + len(setB) - intersection

	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

Dan:

score := similarity.Jaccard(page1.Content, page2.Content)

fmt.Printf("%.2f\n", score)

Bijvoorbeeld:

0.91 = bijna zeker duplicaat
0.78 = heel veel overlap
0.55 = gerelateerd onderwerp
0.22 = waarschijnlijk verschillend

De precieze grenzen moet je op je eigen content afstellen.

Ik zou bijvoorbeeld beginnen met:

switch {
case score >= 0.85:
// waarschijnlijk duplicaat
case score >= 0.65:
// sterke overlap, handmatig controleren
case score >= 0.45:
// gerelateerd
default:
// waarschijnlijk ander artikel
}
4. Nog beter: vergelijk groepjes woorden

Losse woorden geven soms verkeerde matches.

Deze twee artikelen:

Donald Trump wint verkiezingen in Verenigde Staten

en:

Geschiedenis van verkiezingen in de Verenigde Staten

hebben bijvoorbeeld aardig wat woorden gemeenschappelijk terwijl de inhoud heel anders kan zijn.

Daarom zou ik liever word shingles gebruiken.

Bijvoorbeeld 3-grams.

Van:

Rusland viel Oekraïne binnen in februari 2022

maak je:

rusland viel oekraïne
viel oekraïne binnen
oekraïne binnen februari
binnen februari 2022

Daarna bereken je weer Jaccard.

Dat werkt verrassend goed voor teksten die gedeeltelijk gekopieerd, aangepast of herschreven zijn.

Bijvoorbeeld:

func shingles(s string, size int) map[string]struct{} {
s = strings.ToLower(s)
s = nonWord.ReplaceAllString(s, " ")

	words := strings.Fields(s)

	result := make(map[string]struct{})

	if len(words) < size {
		return result
	}

	for i := 0; i <= len(words)-size; i++ {
		shingle := strings.Join(words[i:i+size], " ")
		result[shingle] = struct{}{}
	}

	return result
}

Met een helper:

func JaccardSets(a, b map[string]struct{}) float64 {
if len(a) == 0 && len(b) == 0 {
return 1
}

	intersection := 0

	for value := range a {
		if _, ok := b[value]; ok {
			intersection++
		}
	}

	union := len(a) + len(b) - intersection

	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

En:

a := shingles(page1.Content, 3)
b := shingles(page2.Content, 3)

score := JaccardSets(a, b)

Voor jouw use-case vind ik dit al een stuk interessanter.

Maar er is één probleem

Stel je hebt:

Artikel A
Napoleon verloor uiteindelijk zijn macht nadat zijn invasie
van Rusland mislukte.
Artikel B
De Franse keizer raakte ernstig verzwakt door de desastreuze
veldtocht richting Moskou.

Voor een mens gaan deze teksten vrijwel over hetzelfde.

Een woordenvergelijking ziet echter:

Napoleon != Franse keizer
invasie Rusland != veldtocht Moskou
verloor macht != ernstig verzwakt

De similarity-score kan daardoor heel laag zijn.

Daar komen embeddings om de hoek kijken.

5. Embeddings zijn waarschijnlijk uiteindelijk de beste oplossing voor jou

Een embedding zet een tekst om in bijvoorbeeld zoiets:

[
0.02343,
-0.98234,
0.41232,
0.00342,
...
]

Niet de woorden zelf worden vergeleken, maar grofweg de betekenis van de tekst.

Daardoor kunnen:

Napoleon's invasion of Russia was disastrous.

en

The French emperor suffered catastrophic losses
during his Moscow campaign.

een hoge similarity krijgen.

Je berekent vervolgens de cosine similarity tussen twee vectors:

func CosineSimilarity(a, b []float64) float64 {
if len(a) != len(b) {
return 0
}

	var dot float64
	var normA float64
	var normB float64

	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

Dan krijg je bijvoorbeeld:

Page 12 → Page 81: 0.94
Page 12 → Page 43: 0.88
Page 12 → Page 99: 0.52

Waarbij iets als:

> 0.92   zeer waarschijnlijk dezelfde informatie
0.85–0.92 sterk gerelateerd
0.70–0.85 zelfde onderwerp
< 0.70   waarschijnlijk voldoende verschillend

slechts startwaarden zijn. De echte grenzen hangen sterk af van het embeddingmodel en jouw artikelen.

6. Ik zou voor jouw applicatie deze pipeline bouwen

Dat wordt ongeveer:

                 nieuw artikel
                       │
                       ▼
              ┌────────────────┐
              │ normaliseren   │
              └───────┬────────┘
                      │
             ┌────────▼────────┐
             │ SHA-256 hash    │
             └────────┬────────┘
                      │
               exact gelijk?
              YES ────┼──── NO
               │      │
               ▼      ▼
           duplicate  FTS5
                      │
                      ▼
               10-20 kandidaten
                      │
                      ▼
              Jaccard / shingles
                      │
                      ▼
                 embedding
                      │
                      ▼
               cosine similarity
                      │
        ┌─────────────┼─────────────┐
        ▼             ▼             ▼
      >0.92        0.85-0.92       <0.85
    duplicate       review          OK

Je hoeft overigens niet direct alle stappen te implementeren.

Voor een eerste versie zou ik doen:

1. Exacte hash
2. FTS5
3. shingle similarity

Dat is snel, goedkoop en draait volledig lokaal.

Daarna kun je embeddings toevoegen.

7. Exacte duplicaten zou ik sowieso apart afvangen

Voeg bijvoorbeeld toe:

ALTER TABLE pages ADD COLUMN content_hash TEXT;

Bij het opslaan:

func hashContent(content string) string {
content = strings.ToLower(strings.TrimSpace(content))

	sum := sha256.Sum256([]byte(content))

	return hex.EncodeToString(sum[:])
}

Daarna:

SELECT id, title
FROM pages
WHERE content_hash = ?
LIMIT 1;

Dat kost praktisch niets en voorkomt dat je dure similarity-berekeningen uitvoert voor letterlijke duplicaten.

8. Titel en content zou ik niet even zwaar laten meetellen

Dit is ook belangrijk.

Stel:

Titel:
Waarom viel het Romeinse Rijk?

Content:
3000 woorden

De titel vertelt relatief veel over het onderwerp.

Ik zou daarom bijvoorbeeld berekenen:

titleSimilarity := Jaccard(a.Title, b.Title)
contentSimilarity := Jaccard(a.Content, b.Content)

score :=
titleSimilarity*0.35 +
contentSimilarity*0.65

Of met embeddings:

score :=
titleCosine*0.30 +
contentCosine*0.70

Je kunt category ook gebruiken als bonus/filter:

if a.Category == b.Category {
score += 0.05
}

Hoewel ik category liever gebruik om kandidaten te selecteren dan om blind de similarity-score te verhogen.

9. Als je bijvoorbeeld 1.000 artikelen hebt

Dan zou ik niet:

for _, a := range pages {
for _, b := range pages {
compare(a, b)
}
}

doen.

Dat zijn grofweg:

1.000 × 1.000
= 1.000.000 vergelijkingen

En bij 100.000 artikelen:

10.000.000.000

vergelijkingen.

Gebruik FTS5 daarom als eerste selectie:

100.000 artikelen
↓
FTS zoekopdracht
↓
25 kandidaten
↓
similarity
↓
5 relevante kandidaten

Dat schaalt veel beter.

Hoe ik het concreet in jouw Go-app zou ontwerpen

Bij het opslaan van een artikel zou je bijvoorbeeld een functie hebben:

type SimilarPage struct {
ID      int64
Title   string
Score   float64
}

func FindSimilarPages(
ctx context.Context,
db *sql.DB,
page Page,
) ([]SimilarPage, error) {
// 1. Exacte hash controleren

	// 2. FTS kandidaten zoeken

	// 3. Titel similarity

	// 4. Content/shingle similarity

	// 5. Scores combineren

	// 6. Sorteren

	// 7. Alleen bijvoorbeeld score > 0.65 teruggeven

	return results, nil
}

Je UI kan vervolgens bij een nieuw artikel laten zien:

⚠ Mogelijk bestaande informatie

93%  "De val van het Romeinse Rijk"
81%  "Waarom Rome uiteindelijk instortte"
67%  "Het West-Romeinse Rijk in de vijfde eeuw"

En dan kun je beslissen:

[Artikel toch toevoegen]

[Open bestaand artikel]

[Informatie samenvoegen]

Dat zou voor een CMS/kennisbank een erg nette oplossing zijn.

Mijn advies voor jouw situatie

Als je momenteel een normale hoeveelheid artikelen hebt — zeg enkele honderden tot enkele tienduizenden — zou ik FTS5 + 3-word shingles/Jaccard eerst bouwen. FTS5 is specifiek bedoeld om grote hoeveelheden tekst efficiënt te doorzoeken en ondersteunt ook rankingfuncties zoals BM25.

Wil je daadwerkelijk herkennen dat twee artikelen dezelfde feiten/onderwerpen behandelen terwijl ze compleet anders zijn geschreven, dan zou ik daarbovenop embeddings gebruiken. Dat is precies het punt waarop klassieke string similarity tekortschiet.

Als je wilt, kan ik ook meteen een complete werkende Go 1.26 implementatie voor je maken met database/sql + SQLite, inclusief migration, FTS5-index, triggers en een FindSimilarPages() functie die bijvoorbeeld de 10 meest vergelijkbare artikelen met een score van 0–100% retourneert.