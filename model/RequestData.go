package model

type RequestData struct {
	Method         string
	Host           string
	RequestURI     string
	ContentLength  int
	Origin         string
	Connection     string
	QueryParams    map[string]string
	PathParams     map[string]string
	BodyParameters map[string]string
	Body           string
	Headers        map[string]string
}
