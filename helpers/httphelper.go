package helpers

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"swiki/model"
)

func EnableCors(w *http.ResponseWriter) {
	allowedHeaders := "Access-Control-Allow-Headers,Accept, Content-Type, Content-Length, Accept-Encoding, Authorization,X-CSRF-Token"
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
	(*w).Header().Set("Access-Control-Allow-Headers", allowedHeaders)
	(*w).Header().Add("Access-Control-Allow-Credentials", "true")
}

func WriteResponse(w http.ResponseWriter, response string) {
	_, _ = fmt.Fprintf(w, response)
}

func BooleanConvertor(value string) bool {
	if value == "0" {
		return false
	}
	return true
}

func BooleanToStringConvertor(value bool) string {
	if value == true {
		return "1"
	}
	return "0"
}

func GetRequestInfo(r *http.Request) model.RequestData {
	var requestData model.RequestData

	reqDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Fatal(err)
	}

	requestData.Body = string(reqDump)
	requestData.Method = r.Method
	requestData.Host = r.Host
	requestData.RequestURI = r.RequestURI

	spliitedRegDump := strings.Split(string(reqDump), "\r\n")

	lastline := spliitedRegDump[len(spliitedRegDump)-1]
	if len(lastline) > 0 {
		lastline = lastline[1:]
		lastline = lastline[0 : len(lastline)-1]
		lastlineParts := strings.Split(lastline, ",")
		kvMap := make(map[string]string)

		for _, kv := range lastlineParts {
			splitedKv := strings.Split(kv, ":")
			splitedKv[0] = trimQuotes(splitedKv[0])
			kvMap[splitedKv[0]] = trimQuotes(strings.ReplaceAll(splitedKv[1], "[k]", ","))
		}
		requestData.BodyParameters = kvMap
	}
	kvMap := make(map[string]string)

	for name, kv := range r.Header {
		if name == "Connection" {
			requestData.Connection = kv[0]
			continue
		}

		if name == "Content-Length" {
			v, err := strconv.Atoi(kv[0])
			if err != nil {
				log.Println(err.Error())
				continue
			}
			requestData.ContentLength = v
			continue
		}

		if name == "Origin" {
			requestData.Origin = kv[0]
			continue
		}

		kvMap[name] = kv[0]
		fmt.Println(name + " " + kv[0])
	}
	requestData.Headers = kvMap

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Println(err.Error())
	}
	kvMap = make(map[string]string)
	for name, kv := range params {
		kvMap[name] = kv[0]
	}

	requestData.QueryParams = kvMap
	return requestData
}

func trimQuotes(s string) string {
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func RemoveUrlEncoding(content string) string {
	content, err := url.QueryUnescape(content)

	if err != nil {
		println(err.Error())
		return content
	}
	return content
}

func Trim(value string) string {
	return strings.TrimSpace(value)
}
