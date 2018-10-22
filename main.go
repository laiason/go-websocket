package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"time"
)

// websocket和http服务IP、端口、回调处理函数配置
type Server struct {
	ip string
	port int
	handler func(response http.ResponseWriter, request *http.Request)
}

// websocket消息格式
type wsMsg struct {
	expiresTime int64
	data []byte
	wsConn *websocket.Conn
}

// http消息格式
type httpResponse struct {
	result int
	msg string
}

// websocket消息有效时长
const VALID_TIME int64 = 3600

// 存储websocket消息map
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

		// byte类型转string，消息的key值
		key := string(data[:])
		// 消息过期时间
		expiresTime := time.Now().Unix() + VALID_TIME
		msg := wsMsg{wsConn : wsConn, expiresTime : expiresTime, data : []byte{}}

		// 保存websocket消息
		msgMap[key] = &msg

		fmt.Println(key)
	}

ERR :
	wsConn.Close()
}

// http处理函数
func httpHandler(response http.ResponseWriter, request *http.Request)  {
	//fmt.Println(111)
	//// http返回结果
	var httpResponseResult = httpResponse{result : 0, msg : ""}
	//
	//body, err := ioutil.ReadAll(request.Body)
	//if err != nil {
	//	log.Fatal("read body err：", err)
	//	httpResponseResult.result = 1
	//	httpResponseResult.msg = "read body err：" + err.Error()
	//
	//	return
	//}
	//
	//// 将消息转成json格式
	//var msg interface{}
	//if err := json.Unmarshal(body, &msg); err != nil {
	//	log.Fatal("json Unmarshal body err：", err)
	//	return
	//}
	//
	//// 将msg类型专线map类型
	//requestData := msg.(map[string]interface{})
	//mapKey := requestData["key"].(string)
	//mapData, err := json.Marshal(requestData["data"])
	//if err != nil {
	//	log.Fatal("json Marshal data err：", err)
	//	return
	//}
	//
	//fmt.Println(mapKey)
	//fmt.Println(mapData)
	//
	//if _, ok := msgMap[mapKey]; !ok {
	//	log.Fatal("not exist key：", mapKey)
	//	return
	//}
	//// 保存数据
	//msgMap[mapKey].data = mapData
	//
	//// websocket链接
	//wsConn := msgMap[mapKey].wsConn
	//
	//// 判断是否过期
	//unixTime := time.Now().Unix()
	//if unixTime > msgMap[mapKey].expiresTime {
	//	log.Fatal("websocket connect expired，key：", mapKey)
	//	wsConn.Close()
	//	return
	//}
	//
	//// 发送websocket数据
	//if err := wsConn.WriteMessage(websocket.TextMessage, mapData); err != nil {
	//	log.Fatal("websocket send message err：", err)
	//	wsConn.Close()
	//	return
	//}
	//response.Write([]byte("Hello, world!"))
	rtn, _ := json.Marshal(httpResponseResult)
	fmt.Println(rtn)
	response.Write(rtn)
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


