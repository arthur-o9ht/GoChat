package main

import (
	"fmt"
	"net/http"
	"text/template"
)

//登录界面和处理
var loginTemple = template.Must(template.ParseFiles("login.html"))

func loginHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method == "GET" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		loginTemple.Execute(w, r.Host)
	} else {
		r.ParseForm()
		var regUserData UserInfo
		regUserData.UserName = r.Form["username"][0]
		regUserData.PassWd = r.Form["password"][0]
		if res, err := checkUser(regUserData); err != nil {
			fmt.Println(err)
		} else {
			if res["PassWd"] == regUserData.PassWd {
				sess, _ := globalSessions.SessionStart(w, r)
				defer sess.SessionRelease(w)
				sess.Set("isLogin", 1)
				sess.Set("wsName", regUserData.UserName)
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				homeTemple.Execute(w, r.Host)
			} else {
				http.Error(w, "error user&pw", 502)
				return
			}
		}
	}
}
