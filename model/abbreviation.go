package model

type Abbreviation struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AbbreviationLocal struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TursoId     int    `json:"tursoid"`
}
