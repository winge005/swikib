package model

type Statistic struct {
	Total       int         `json:"total"`
	SCategories []SCategory `json:"categories"`
}

type SCategory struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}
