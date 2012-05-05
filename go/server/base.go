package server

import (
    "net/http"
    "regexp"
    "strings"
    "errors"
    "appengine"
    "appengine/datastore"
	"appengine/memcache"
    "encoding/base64"
	"encoding/json"
    "fmt"
)

var (
	expand_tco_xml_re = regexp.MustCompile("<url>([\\w\\.:/]+?)</url>\\s+?<display_url>[\\w\\.:/]+?</display_url>\\s+?<expanded_url>([\\w\\.:/]+?)</expanded_url>")
	expand_tco_json_re = regexp.MustCompile("\"url\":\"([^\"]+?)\",\"indices\":.+?,\"expanded_url\":\"([^\"]+?)\"")
	parse_profile_img_json_re = regexp.MustCompile("\"(https?:\\\\/\\\\/[\\w]+?\\.twimg\\.com)([^\"]+?)\"")
	parse_profile_img_xml_re = regexp.MustCompile(">(https?://[\\w]+?\\.twimg\\.com)([^<]+?)<")
)

type OAuthUserData struct {
	ScreenName string
	OAuthTokenSecret string
	OAuthToken string
	UserId string
	UrlSuffix string
}

func GetMemOAuthUserData(c *appengine.Context, key string) (oauthUserData *OAuthUserData, err error){
	item, err := memcache.Get(*c, key)
	if err == nil {
		var o OAuthUserData
		err = json.Unmarshal(item.Value, &o)
		if err == nil {
			oauthUserData = &o
		}
	}
	return
}

func SetMemOAuthUserData(c *appengine.Context, oauthUserData *OAuthUserData) (err error){
	key := "OAuthUserData-" + oauthUserData.ScreenName
	value, err:= json.Marshal(oauthUserData)
	item := &memcache.Item{
		Key: key,
		Value: value,
	}
	err = memcache.Set(*c, item)
	return
}

func SetOAuthUserData(c *appengine.Context, oauthUserData *OAuthUserData) (err error){
	screen_name := oauthUserData.ScreenName
	q := datastore.NewQuery("OAuthUserData").Filter("ScreenName =", screen_name)
	for t := q.Run(*c); ; {
		var x OAuthUserData
		key, err := t.Next(&x)
		if err == datastore.Done {
			break
		}
		datastore.Delete(*c, key)
	}
	_, err = datastore.Put(*c, datastore.NewIncompleteKey(*c, "OAuthUserData", nil), oauthUserData)
	err = SetMemOAuthUserData(c, oauthUserData)
	return
}

func GetOAuthUserData(c *appengine.Context, screen_name string) (oauthUserData *OAuthUserData, err error){
	oauthUserData, err = GetMemOAuthUserData(c, "OAuthUserData-" + screen_name)
	if err != nil {
		var o OAuthUserData
		q := datastore.NewQuery("OAuthUserData").Filter("ScreenName =", screen_name)
		t := q.Run(*c)
		_, err = t.Next(&o)
		if err == nil {
			oauthUserData = &o
			SetMemOAuthUserData(c, oauthUserData)
		}
	}else {
		(*c).Infof("%s\n","Memcache hit!")
	}
	return
}

func GetBasicAuth(r *http.Request)(user, passwd string, err error){
	auth_str := r.Header.Get("Authorization")
	if auth_str == "" {
		err = errors.New("No Authorization Data")
		return
	}
	auths := strings.Split(auth_str, " ")
	if strings.ToLower(auths[0]) != "basic" {
		err = errors.New("Not Basic Authorization")
		return
	}
	auths_bytes, err := base64.StdEncoding.DecodeString(auths[1])
	if err != nil {
		return
	}
	user_pass := strings.Split(string(auths_bytes), ":")
	user = user_pass[0]
	passwd = user_pass[1]
	return
}


func ExpandTCO(body, file_type string) (body_expanded string){
	//body_expanded = expand_tco_xml_re.ReplaceAllString(body, "<url>$3</url>$2<expanded_url>$3</expanded_url>")
	body_expanded = body
	var expand_tco_re *regexp.Regexp
	if file_type == "xml" {
		expand_tco_re = expand_tco_xml_re
	}else{
		expand_tco_re = expand_tco_json_re
	}
	find_pair := expand_tco_re.FindAllStringSubmatch(body_expanded, -1)
	for i := range find_pair {
		body_expanded = strings.Replace(body_expanded, find_pair[i][1], find_pair[i][2], -1)
	}
	return
}

func ParseProfileImageUrl(body, file_type string, r *http.Request) (body_parsed string) {
	if file_type == "xml"{
		replace_str := fmt.Sprintf(">%s://%s/i$2<", r.URL.Scheme, r.URL.Host)
		body_parsed = parse_profile_img_xml_re.ReplaceAllString(body, replace_str)
	}else{
		replace_str := fmt.Sprintf("\"%s:\\/\\/%s\\/i$2\"", r.URL.Scheme, r.URL.Host)
		body_parsed = parse_profile_img_json_re.ReplaceAllString(body, replace_str)
	}
	return
}

func GetBaseUrl(r *http.Request) (baseUrl string) {
	if r.URL.Scheme == "" {
		baseUrl = "http://localhost:8080"
	}else{
		baseUrl = r.URL.Scheme + "://" + r.URL.Host
	}
	return
}
