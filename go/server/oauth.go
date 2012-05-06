package server

import (
	"appengine"
	"appengine/urlfetch"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	//	"code.google.com/p/gorilla/appengine/sessions"
	"github.com/garyburd/go-oauth/oauth"
)

type OAuthTemplateParam struct {
	GetType1 bool
	BaseUrl  string
}

const unreservedChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_.~"

var (
	valid_str_re          = regexp.MustCompile("[^a-zA-Z0-9]")
	oauth_token_re        = regexp.MustCompile("oauth_token=([0-9a-zA-Z]+)")
	authenticity_token_re = regexp.MustCompile("authenticity_token = '([0-9a-zA-Z]+)'")
	happy_callback_re     = regexp.MustCompile("\"(http.+?oauth_token.+?oauth_verifier[^\"]+)\"")
)

var (
	credentials = oauth.Credentials{
		Token:  OAUTH_KEY,
		Secret: OAUTH_SECRET,
	}
	oauthClient = oauth.Client{
		TemporaryCredentialRequestURI: "https://api.twitter.com/oauth/request_token",
		ResourceOwnerAuthorizationURI: "https://api.twitter.com/oauth/authenticate",
		TokenRequestURI:               "https://api.twitter.com/oauth/access_token",
		Credentials:                   credentials,
	}
)

func urlencode(val string) (ret string) {
	for i := 0; i < len(val); i++ {
		c := val[i]
		s := string(c)
		if strings.Index(unreservedChars, s) != -1 {
			ret += s
		} else {
			ret += fmt.Sprintf("%%%s%s", string("0123456789ADCDEF"[c>>4]), string("0123456789ABCDEF"[c&15]))
		}
	}
	return
}

func OAuthHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionsDStore.Get(r, "go-twip")
	c := appengine.NewContext(r)
	httpClient := urlfetch.Client(c)
	baseUrl := GetBaseUrl(r)

	oauthTemplateParam := OAuthTemplateParam{true, baseUrl}
	getType := r.FormValue("type")
	if getType == "2" {
		oauthTemplateParam.GetType1 = false
	}

	if r.Method == "POST" {
		url_suffix := r.FormValue("url_suffix")
		url_suffix = valid_str_re.ReplaceAllString(url_suffix, "")
		session.Values["url_suffix"] = url_suffix
		session.Save(r, w)
		callback_url := baseUrl + "/oauth/"
		tempCred, err := oauthClient.RequestTemporaryCredentials(httpClient, callback_url, nil)

		if err == nil {
			session.Values["oauth_token"] = tempCred.Token
			session.Values["oauth_token_secret"] = tempCred.Secret
			session.Save(r, w)
			oauthUrl := oauthClient.AuthorizationURL(tempCred, nil)
			if oauthTemplateParam.GetType1 {
				http.Redirect(w, r, oauthUrl, 302)
			} else {
				id := r.FormValue("username")
				passwd := r.FormValue("password")
				happy_callback_url, err := OAuthProxy(httpClient, id, passwd, oauthUrl)
				if err == nil {
					http.Redirect(w, r, happy_callback_url, 302)
				} else {
					fmt.Fprint(w, err)
					return
				}
			}
		} else {
			fmt.Fprint(w, err)
			return
		}
	} else if r.Method == "GET" {
		if r.FormValue("oauth_token") != "" && r.FormValue("oauth_verifier") != "" {
			tempCred := oauth.Credentials{
				//Token: session.Values["oauth_token"].(string),
				Token: r.FormValue("oauth_token"),
			}
			cred, vars, err := oauthClient.RequestToken(httpClient, &tempCred, r.FormValue("oauth_verifier"))
			if err == nil {
				url_suffix := session.Values["url_suffix"].(string)
				screen_name := vars["screen_name"][0]
				oauthUserData := OAuthUserData{
					ScreenName:       screen_name,
					OAuthToken:       cred.Token,
					OAuthTokenSecret: cred.Secret,
					UserId:           vars["user_id"][0],
					UrlSuffix:        url_suffix,
				}
				//c.Infof("%+v\n", vars)
				err := SetOAuthUserData(&c, &oauthUserData)
				if err == nil {
					redirect_url := baseUrl + "/getapi/?api=" + baseUrl + "/o/" + url_suffix
					http.Redirect(w, r, redirect_url, 302)
					return
				} else {
					fmt.Fprintln(w, err)
				}

			} else {
				fmt.Fprintln(w, err)
			}

			return
		}
	}
	templates.ExecuteTemplate(w, "oauth.html", oauthTemplateParam)
}

func OAuthProxy(httpClient *http.Client, id string, passwd string, oauthUrl string) (happy_callback_url string, err error) {
	resp, err := httpClient.Get(oauthUrl)
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	body_str := string(body)

	oauth_token := oauth_token_re.FindStringSubmatch(body_str)
	authenticity_token := authenticity_token_re.FindStringSubmatch(body_str)

	post_str := fmt.Sprintf("oauth_token=%s&authenticity_token=%s&session[username_or_email]=%s&session[password]=%s", urlencode(oauth_token[1]), urlencode(authenticity_token[1]), urlencode(id), urlencode(passwd))

	req, _ := http.NewRequest("POST", "https://api.twitter.com/oauth/authorize", strings.NewReader(post_str))
	cookies := resp.Cookies()
	for i := range cookies {
		req.AddCookie(cookies[i])
	}
	resp, err = httpClient.Do(req)
	if err != nil {
		return
	}

	body, _ = ioutil.ReadAll(resp.Body)
	body_str = string(body)

	happy_callback := happy_callback_re.FindStringSubmatch(body_str)
	if len(happy_callback) == 2 {
		happy_callback_url = happy_callback[1]
	}
	return
}
