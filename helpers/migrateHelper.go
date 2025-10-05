package helpers

import "swiki/model"

var (
	m = model.Migrate{
		Abbreviation: false,
		Links:        false,
		Pages:        true,
	}
)

func GetMigrateAbbreviation() bool {
	return m.Abbreviation
}

func GetMigrateLinks() bool {
	return m.Links
}

func GetMigratePages() bool {
	return m.Pages
}
