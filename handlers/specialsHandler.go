package handlers

import (
	"encoding/json"
	"net/http"
	"swiki/helpers"
	"swiki/persistence"
	"time"
)

func SpecialsHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	t := time.Now()
	t2 := t.AddDate(0, 0, -10)
	t3 := t.AddDate(0, 0, -9)
	t4 := t.AddDate(0, 0, -8)
	t5 := t.AddDate(0, 0, -7)
	t6 := t.AddDate(0, 0, -6)
	t7 := t.AddDate(0, 0, -5)
	t8 := t.AddDate(0, 0, -4)
	t9 := t.AddDate(0, 0, -3)
	t10 := t.AddDate(0, 0, -1)

	date1 := helpers.ConvertIntToTwoDigitsString(t.Day()) + "-" + helpers.ConvertIntToTwoDigitsString(int(t.Month())) + "-" + helpers.ConvertIntToTwoDigitsString(t.Year())
	date2 := helpers.ConvertIntToTwoDigitsString(t2.Day()) + "-" + helpers.ConvertIntToTwoDigitsString(int(t2.Month())) + "-" + helpers.ConvertIntToTwoDigitsString(t2.Year())
	date3 := helpers.ConvertIntToTwoDigitsString(t3.Day()) + "-" + helpers.ConvertIntToTwoDigitsString(int(t3.Month())) + "-" + helpers.ConvertIntToTwoDigitsString(t3.Year())
	date4 := helpers.ConvertIntToTwoDigitsString(t4.Day()) + "-" + helpers.ConvertIntToTwoDigitsString(int(t4.Month())) + "-" + helpers.ConvertIntToTwoDigitsString(t4.Year())
	date5 := helpers.ConvertIntToTwoDigitsString(t5.Day()) + "-" + helpers.ConvertIntToTwoDigitsString(int(t5.Month())) + "-" + helpers.ConvertIntToTwoDigitsString(t5.Year())
	date6 := helpers.ConvertIntToTwoDigitsString(t6.Day()) + "-" + helpers.ConvertIntToTwoDigitsString(int(t6.Month())) + "-" + helpers.ConvertIntToTwoDigitsString(t6.Year())
	date7 := helpers.ConvertIntToTwoDigitsString(t7.Day()) + "-" + helpers.ConvertIntToTwoDigitsString(int(t7.Month())) + "-" + helpers.ConvertIntToTwoDigitsString(t7.Year())
	date8 := helpers.ConvertIntToTwoDigitsString(t8.Day()) + "-" + helpers.ConvertIntToTwoDigitsString(int(t8.Month())) + "-" + helpers.ConvertIntToTwoDigitsString(t8.Year())
	date9 := helpers.ConvertIntToTwoDigitsString(t9.Day()) + "-" + helpers.ConvertIntToTwoDigitsString(int(t9.Month())) + "-" + helpers.ConvertIntToTwoDigitsString(t9.Year())
	date10 := helpers.ConvertIntToTwoDigitsString(t10.Day()) + "-" + helpers.ConvertIntToTwoDigitsString(int(t10.Month())) + "-" + helpers.ConvertIntToTwoDigitsString(t10.Year())

	pages, err := persistence.GetPagesFromDates(date1, date2, date3, date4, date5, date6, date7, date8, date9, date10)

	response, err := json.Marshal(pages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helpers.WriteResponse(w, string(response))
}
