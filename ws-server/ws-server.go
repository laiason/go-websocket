package ws_server

import (
	"github.com/gorilla/websocket"
	"net/http"
)

var (
	upgrader = websocket.Upgrader{
		// 允许跨域
		CheckOrigin : func(request *http.Request) bool{
			return true
		},
	}
)

func WsHandler(response http.ResponseWriter, request *http.Request) {
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