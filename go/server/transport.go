package server

import (
	"appengine"
	"appengine/urlfetch"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func TransportHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	httpClient := urlfetch.Client(c)
	//c.Infof("Receive\n %+v\n\n", r)

	var forward_url string
	base_url := GetBaseUrl(r)

	forward_body_str := ""
	forward_header := make(http.Header)

	if r.Body != nil {
		forward_body, _ := ioutil.ReadAll(r.Body)
		forward_body_str = string(forward_body)
	}

	//c.Infof("Body:%s\n", forward_body_str)

	if strings.Index(r.URL.RequestURI(), "search") > -1 {
		forward_header.Set("Host", "search.twitter.com")
		forward_url = strings.Replace(r.URL.RequestURI(), "/t", "https://search.twitter.com", -1)
	} else {
		forward_header.Set("Host", "api.twitter.com")
		forward_url = strings.Replace(r.URL.RequestURI(), "/t", "https://api.twitter.com", -1)
	}
	forward_header.Set("Authorization", strings.Replace(r.Header.Get("Authorization"), base_url+"/t", "https://api.twitter.com", -1))
	forward_header.Set("Expect", r.Header.Get("Expect"))
	forward_header.Set("Content-Type", r.Header.Get("Content-Type"))
	forward_header.Set("User-Agent", r.Header.Get("User-Agent"))
	forward_header.Set("X-Forwarded-For", r.Header.Get("X-Forwarded-For"))

	forward_req, err := http.NewRequest(r.Method, forward_url, strings.NewReader(forward_body_str))
	if err != nil {
		c.Errorf("1%+v\n", err)
		return
	}
	forward_req.Header = forward_header
	//c.Infof("Forward\n%+v\n\n", forward_req)
	resp, err := httpClient.Do(forward_req)
	w.WriteHeader(resp.StatusCode)
	if err == nil {
		body, _ := ioutil.ReadAll(resp.Body)
		body_str := string(body)
		//c.Infof("%s\n", r.URL.RequestURI())
		if strings.Index(r.URL.RequestURI(), "oauth/authorize?oauth_token") > -1 {
			//c.Infof("Replace\n")
			body_str = strings.Replace(body_str, "<form action=\"https://api.twitter.com/oauth/authorize\"", "<form action=\""+base_url+"/t/oauth/authorize\"", -1)
			//body_str = strings.Replace(body_str, "<div id=\"signin_form\">", "<h1><strong style=\"color:red\">Warning!This page is proxied by twip and therefore you may leak your password to API proxy owner!</strong></h1><div id=\"signin_form\">", -1)
		}
		fmt.Fprint(w, body_str)
	} else {
		c.Errorf("Forward Body Error:%s", err)
	}
}
