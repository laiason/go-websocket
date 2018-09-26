package main

import (
	"WebSocket/myws"
	"net/http"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		// 允许跨域
		CheckOrigin : func(request *http.Request) bool{
			return true
		},
	}
)

func wsHandler(response http.ResponseWriter, request *http.Request) {
	var (
		wsConn *websocket.Conn
		err error
		myconn *myws.Connection
	)

	// Upgrader为WebSocket协议
	if wsConn, err = upgrader.Upgrade(response, request, nil); err != nil{
		return
	}

	if myconn, err = myws.initConnection(wsConn); err != nil{
		goto ERR
	}

	for {
		if data, err = myconn.ReadMessage(); err != nil {
			goto ERR
		}

		if err = myconn.WriteMessage(data); err != nil {
			goto ERR
		}
	}

ERR :
	myconn.Close()
}

func main() {
	http.HandleFunc("/", wsHandler)
	http.ListenAndServe("0.0.0.0:9502", nil)
}