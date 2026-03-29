package model

type Page struct {
	Id       int    `json:"id"`
	Category string `json:"category"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Created  string `json:"created"`
	Updated  string `json:"updated"`
}

type PageLocal struct {
	Id       int    `json:"id"`
	Category string `json:"category"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Created  string `json:"created"`
	Updated  string `json:"updated"`
	TursoId  int    `json:"tursoid"`
}
