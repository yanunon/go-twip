package server

import(
	"net/http"
	"fmt"
	"strings"
	"appengine"
	"appengine/urlfetch"
	"io/ioutil"
)

func TransportHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	httpClient := urlfetch.Client(c)
	c.Infof("Receive\n %+v\n\n", r)
	
	var forward_url string
	base_url := GetBaseUrl(r)
	
	forward_body_str := ""
	forward_header := r.Header
	if r.Body != nil {
		forward_body, _ := ioutil.ReadAll(r.Body)
		forward_body_str = string(forward_body)
	}
	forward_search := false
	if forward_search {
		forward_header["Host"] = []string{"search.twitter.com"}
		forward_url = strings.Replace(r.URL.RequestURI(), "/t", "http://search.twitter.com" ,-1)
	}else {
		forward_header["Host"] = []string{"api.twitter.com"}
		forward_header["Expect"] = []string{""}
		forward_header["Authorization"][0] = strings.Replace(forward_header["Authorization"][0], base_url + "/t", "http://api.twitter.com", -1)
		forward_url = strings.Replace(r.URL.RequestURI(), "/t", "http://api.twitter.com" ,-1)
	}
	forward_req, err := http.NewRequest(r.Method, forward_url, strings.NewReader(forward_body_str))
	if err != nil {
		c.Errorf("1%+v\n", err)
		return
	}
	forward_req.Header = forward_header
	c.Infof("Forward\n%+v\n\n", forward_req)
	resp, err := httpClient.Do(forward_req)
	if err == nil {
		body, _ := ioutil.ReadAll(resp.Body)
		body_str := string(body)
		c.Infof("Forward Body:%s", body_str)
		fmt.Fprint(w, body_str)
	}else {
		c.Errorf("Forward Body Error:%s", err)
	}
}
