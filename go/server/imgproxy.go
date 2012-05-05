package server

import (
	"net/http"
	"fmt"
	"strings"
	"math/rand"
	"io/ioutil"
	"appengine"
	"appengine/urlfetch"
)

func ImageProxyHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	httpClient := urlfetch.Client(c)
	
	url_parts := strings.Split(r.URL.Path,"/")
	img_url := strings.Join(url_parts[2:], "/")
	twimg_url := fmt.Sprintf("http://a%d.twimg.com/%s", rand.Int()%7 , img_url)
	resp, err := httpClient.Get(twimg_url)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	write_header := w.Header()
	resp_header := resp.Header
	for k, v := range resp_header {
		write_header.Set(k, v[0])
	}
	body, _ := ioutil.ReadAll(resp.Body)
	body_str := string(body)
	fmt.Fprint(w, body_str)
}
