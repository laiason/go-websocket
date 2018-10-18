package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Server struct {
	ip string
	port int
	handler func(response http.ResponseWriter, request *http.Request)
}

type wsMsg struct {
	expiresTime int64
	data []byte
	wsConn *websocket.Conn
}

type httpMsg struct {
	key string
	data []byte
}

const VALID_TIME int64 = 3600

var msgMap = make(map[string] *wsMsg)

// websocket处理函数
func wsHandler(response http.ResponseWriter, request *http.Request) {
	var (
		upgrader = websocket.Upgrader{
			// 允许跨域
			CheckOrigin : func(request *http.Request) bool{
				return true
			},
		}
	)

	var (
		wsConn *websocket.Conn
		data []byte
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

		key := string(data[:])
		expiresTime := time.Now().Unix() + VALID_TIME

		msg := wsMsg{wsConn : wsConn, expiresTime : expiresTime, data : []byte{}}
		msgMap[key] = &msg
		fmt.Println(key)
	}

ERR :
	wsConn.Close()
}

// http处理函数
func httpHandler(response http.ResponseWriter, request *http.Request)  {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Fatal("read body err：", err)
		return
	}


	//req_data := string(body[:])
	//fmt.Println(req_data)

	var msg interface{}
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Fatal("json Unmarshal body err：", err)
		return
	}

	requestData := msg.(map[string]interface{})
	mapKey := requestData["key"].(string)

	mapData, err := json.Marshal(requestData["data"])
	if err != nil {
		log.Fatal("json Marshal data err：", err)
		return
	}

	fmt.Println(mapKey)
	fmt.Println(mapData)

	if _, ok := msgMap[mapKey]; !ok {
		log.Fatal("not exist key：", mapKey)
		return
	}
	// 保存数据
	msgMap[mapKey].data = mapData

	// websocket链接
	wsConn := msgMap[mapKey].wsConn

	// 判断是否过期
	unixTime := time.Now().Unix()
	if unixTime > msgMap[mapKey].expiresTime {
		log.Fatal("websocket connect expired，key：", mapKey)
		goto ERR
	}

	// 发送websocket数据
	if err := wsConn.WriteMessage(websocket.TextMessage, mapData); err != nil {
		log.Fatal("websocket send message err：", err)
		goto ERR
	}

ERR :
	wsConn.Close()
}

func main() {
	addrs := []Server{{"127.0.0.1", 8080, wsHandler}, {"127.0.0.1", 8081, httpHandler}}

	startServer := func (server Server)  {
		mux := http.NewServeMux()
		mux.HandleFunc("/", server.handler)
		host := server.ip + ":" + strconv.Itoa(server.port)
		err := http.ListenAndServe(host, mux)
		if err != nil {
			str := "ListenAndServe " + host + "Error："
			log.Fatal(str, err)
		}
	}

	for _, v := range addrs {
		// 每个端口启动一个goroutine
		go startServer(v)
	}

	// 阻塞
	select {}
}


