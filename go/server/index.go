package server

import (
	"net/http"
	//	"fmt"
	"code.google.com/p/gorilla/appengine/sessions"
	"html/template"
)

//appengine enter point
func init() {
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/t/", TransportHandler)
	http.HandleFunc("/o/", OverrideHandler)
	http.HandleFunc("/i/", ImageProxyHandler)
	http.HandleFunc("/getapi/", GetApiHandler)
	http.HandleFunc("/oauth/", OAuthHandler)
}

var (
	templates = template.Must(template.ParseFiles(
		"template/index.html",
		"template/getapi.html",
		"template/oauth.html",
	))
)

//session
var (
	sessionsDStore = sessions.NewDatastoreStore("", []byte("kite-very-secret"))
	sessionsMStore = sessions.NewMemcacheStore("", []byte("kite-a-lot-secret"))
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "index.html", nil)
}
