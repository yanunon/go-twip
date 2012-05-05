package server

import (
	"net/http"
	"fmt"
	"appengine"
//	"io/ioutil"
	//"regexp"
//	"strings"
)

func TestHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	o, err := GetOAuthUserData(&c, "yanunon")
	err = SetMemOAuthUserData(&c, o)
	if err != nil {
		fmt.Fprintf(w, "Set Mem:%v", err)
		return
	}
	o2, err := GetMemOAuthUserData(&c, "OAuthUserData-yanunon")
	if err != nil {
		fmt.Fprintf(w, "Get Mem:%v", err)
	}
	fmt.Fprintf(w, "%+v\n", o2)

}
