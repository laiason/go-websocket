package ws_server

import (
	"net/http"
	"github.com/gorilla/websocket"
	"strconv"
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
		data []byte
		wsConn *websocket.Conn
		err error
	)

	// Upgrader为WebSocket协议
	if wsConn, err = upgrader.Upgrade(response, request, nil); err != nil{
		return
	}

	for {
		if _, data, err = wsConn.ReadMessage(); err != nil {
			goto ERR
		}

		if err = wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
			goto ERR
		}
	}

ERR :
	wsConn.Close()
}

func Start(ip string, port int) {
	http.HandleFunc("/ws", wsHandler)

	addr := ip + ":" + strconv.Itoa(port)
	http.ListenAndServe(addr, nil)
}