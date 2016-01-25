package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"text/template"
	"time"
	"strings"
)

const (
	writeWait      = 60 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
	MaxClient      = 4
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

/**
 *	hub 结构体构建
 */
type hub struct {
	connections map[string]*connection
	broadcast   chan []byte
	register    chan *connection
	unregister  chan *connection
	numbers     int
}

var h = hub{
	//广播数据
	broadcast: make(chan []byte),
	//注册接口
	register: make(chan *connection),
	//注销接口
	unregister: make(chan *connection),
	//连接存储
	connections: make(map[string]*connection),
	//链接总数
	numbers: 0,
}

/**
 *	连接 结构体构建
 */
type connection struct {
	//ws名称
	wsName string
	//ws链接
	ws *websocket.Conn
	//ws字符缓存
	msgBuf chan []byte
}

/**
 *	connection->持续读取
 *	将读取的值放入广播
 */
func (c *connection) readPump() {
	defer func() {
		h.unregister <- c
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		h.broadcast <- []byte(c.wsName+":"+string(message))
	}
}

/**
 *	connection->写
 *	真正写方法
 */
func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return c.ws.WriteMessage(mt, payload)
}

/**
 *	connection->持续接收广播
 */
func (c *connection) writePump() {
	ticker := time.NewTicker(60 * time.Second)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.msgBuf:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.TextMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

/**
 *	hub运行主体
 */
func (h *hub) runHub() {
	for {
		select {
		case c := <-h.register:
			if h.numbers < MaxClient {
				h.connections[c.wsName] = c
				h.numbers++
			} else {
				c.write(websocket.TextMessage, []byte("please try enter later"))
				c.write(websocket.CloseMessage, []byte{})
				close(c.msgBuf)
			}
		case c := <-h.unregister:
			if _, ok := h.connections[c.wsName]; ok {
				close(c.msgBuf)
				delete(h.connections, c.wsName)
				h.numbers--

			}
		case m := <-h.broadcast:
			msgStr := string(m)
			msgArr := strings.Split(msgStr, ":")
			for _,c := range h.connections {
				if msgArr[0] == c.wsName{
					m = []byte("你:"+string(msgArr[1]))
				}else{
					m = []byte(msgStr)
				}
				select {
				case c.msgBuf <- m:
				default:
					if c.msgBuf != nil {
						close(c.msgBuf)
						delete(h.connections, c.wsName)
					}
				}
			}
		}
		fmt.Println("client num:", h.numbers)
	}
}

/**
 *	端服务接口
 */
func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	sess, _ := globalSessions.SessionStart(w, r)
	defer sess.SessionRelease(w)
	userName := sess.Get("wsName").(string)
	c := &connection{msgBuf: make(chan []byte, 256), ws: ws, wsName: userName}
	h.register <- c
	go c.writePump()
	c.readPump()
}

//聊天室
var homeTemple = template.Must(template.ParseFiles("home.html"))

func serverHome(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	sess, _ := globalSessions.SessionStart(w, r)
	defer sess.SessionRelease(w)
	isLogin := sess.Get("isLogin")
	if isLogin != 1 {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTemple.Execute(w, r.Host)
}