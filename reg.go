package main

import (
	"fmt"
	"net/http"
	"text/template"
)

//注册界面和处理
var regTemple = template.Must(template.ParseFiles("reg.html"))

func regHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/reg" {
		http.Error(w, "Not found", 404)
		return
	}

	if r.Method == "GET" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		regTemple.Execute(w, r.Host)
	} else {
		r.ParseForm()
		var regUserData UserInfo
		regUserData.UserName = r.Form["username"][0]
		regUserData.PassWd = r.Form["password"][0]

		if res, err := regUser(regUserData); res != true {
			fmt.Println(err)
		} else {

			sess, _ := globalSessions.SessionStart(w, r)
			defer sess.SessionRelease(w)
			sess.Set("isLogin", 1)
			sess.Set("wsName", regUserData.UserName)
			fmt.Println(regUserData)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			homeTemple.Execute(w, r.Host)
		}
	}
}
