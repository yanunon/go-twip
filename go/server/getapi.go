package server

import (
	"net/http"
	"strings"

//	"text/template"
)

type GetApiParam struct {
	GetApi   string
	BaseUrl  string
	ImageUrl string
}

func GetApiHandler(w http.ResponseWriter, r *http.Request) {
	get_api := r.FormValue("api")
	//get_api := "http://localhost/o/"
	img_api := ""
	baseUrl := GetBaseUrl(r)
	if get_api != "" {
		img_api = strings.Replace(get_api, "/o/", "/i/", -1)
	}
	var getApiParam = GetApiParam{get_api, baseUrl, img_api}
	templates.ExecuteTemplate(w, "getapi.html", getApiParam)
}
