package server

import (
	"net/http"
	"net/url"
	"fmt"
	"strings"
	"io/ioutil"
	"appengine"
	"appengine/urlfetch"
	"github.com/garyburd/go-oauth/oauth"
)



func OverrideHandler(w http.ResponseWriter, r *http.Request) {
	//baseUrl := GetBaseUrl(r)
	c := appengine.NewContext(r)
	httpClient := urlfetch.Client(c)
	//c.Infof("Request:%+v\n", r)
	urlParts := strings.Split(r.URL.Path,"/")
	urlSuffix := urlParts[2]

	file_type_idx := strings.LastIndex(r.URL.Path, ".")
	file_type := ""
	if file_type_idx > -1 {
		file_type = r.URL.Path[file_type_idx + 1:]
	}

	screen_name, _, err := GetBasicAuth(r)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	//c.Infof("%+v\n", screen_name)
	oauthUserData, err := GetOAuthUserData(&c, screen_name)

	if oauthUserData.UrlSuffix != urlSuffix {
		fmt.Fprintf(w, "Error: [urlsuffix]:%s not match user [screen_name]%s", urlSuffix, screen_name)
		return
	}
	cred := oauth.Credentials{
		Token: oauthUserData.OAuthToken,
		Secret: oauthUserData.OAuthTokenSecret,
	}
	//c.Infof("%+v\n", cred)
	forwardHeader := r.Header
	forwardUrl := strings.Join(urlParts[3:], "/")
	forwardUrl = "https://api.twitter.com/1/" + forwardUrl
	params := make(url.Values)
	r.ParseForm()
	if r.Form != nil {
		params = r.Form
	}
	params["include_entities"] = []string{"true"} //force it true
	if params.Get("since_id") == "-1" {
		params.Del("since_id")
	}
	oauthClient.SignParam(&cred, r.Method, forwardUrl, params)
	forwardUrl = forwardUrl + "?" + params.Encode()
	forwardBody := ""
	if r.Body != nil {
		forwardBodys, _ := ioutil.ReadAll(r.Body) 
		forwardBody = string(forwardBodys)
		c.Infof("Forward Body:%s\n", forwardBody)
	}
	forwardReq, _ := http.NewRequest(r.Method, forwardUrl, strings.NewReader(forwardBody))
	forwardReq.Header = forwardHeader
	resp, err := httpClient.Do(forwardReq)
	if err != nil {
		c.Errorf("%+v\n", err)
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	body_str := string(body)
	body_str = ExpandTCO(body_str, file_type)
	if r.Header.Get("User-Agent") != "Twigee" {
		body_str = ParseProfileImageUrl(body_str, file_type, r)
	}
	if resp.StatusCode != 200 {
		c.Infof("StatusCode:%d   Body:%s\n", resp.StatusCode, body_str)
	}
	w.WriteHeader(resp.StatusCode)
	fmt.Fprint(w, body_str)
}
