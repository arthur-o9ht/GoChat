package main

import (
	"flag"
	"github.com/astaxie/beego/session"
	"net/http"
)

var clientString = flag.String("addr", ":8090", "http service address")
var globalSessions *session.Manager

func init() {
	globalSessions, _ = session.NewManager("memory", `{"cookieName":"gosessionid","gclifetime":3600}`)
	go globalSessions.GC()
}

const userListKey = "chatUser_"

func main() {
	flag.Parse()
	http.HandleFunc("/", serverHome)
	http.HandleFunc("/login", loginHome)
	http.HandleFunc("/reg", regHome)
	http.HandleFunc("/chat", serveWs)

	go h.runHub()
	err := http.ListenAndServe(*clientString, nil)
	if err != nil {
		return
	}
}
