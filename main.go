package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"io/ioutil"
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
	expiresTime int64 `json:"expires_time"`
	data interface{} `json:"data"`
	wsConn *websocket.Conn `json:"ws_conn"`
}

// websocket响应消息格式
type wsResponse struct {
	Code int `json:"code"`
	MsgKey string `json:"msg_key"`
	Data interface{} `json:"data"`
	ErrorMsg string `json:"error_msg"`
}

// http响应消息格式
type httpResponse struct {
	Code int `json:"code"`
	Data interface{} `json:"data"`
	ErrorMsg string `json:"error_msg"`
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
		//	fmt.Println(key)
		// 保存websocket消息
		msgMap[key] = &msg

		wsData := wsResponse{Code:0, MsgKey:key, Data: struct {}{}, ErrorMsg:""}
		wsResp, _ := json.Marshal(wsData)
		//	fmt.Println(wsResp)
		if err = wsConn.WriteMessage(websocket.TextMessage, wsResp); err != nil {
			goto ERR
		}
	}

ERR :
	wsConn.Close()
}

// http处理函数
func httpHandler(response http.ResponseWriter, request *http.Request)  {
	// http返回结果
	var httpResponseResult = httpResponse{Code : 0, Data: struct {}{}, ErrorMsg : ""}

	// 异常处理：必须要先声明defer，否则不能捕获到panic异常
	defer func(result httpResponse, resp http.ResponseWriter){
		if err := recover(); err != nil{
			// 这里的err其实就是panic传入的内容
			result.Code = 1
			result.ErrorMsg = err.(string)
			rtn, _ := json.Marshal(result)
			resp.Write(rtn)
		}
	}(httpResponseResult, response)

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic("read body err：" + err.Error())
	}

	// 将消息转成json格式
	var msg interface{}
	if err := json.Unmarshal(body, &msg); err != nil {
		panic("json Unmarshal body err：" + err.Error())
	}

	// 将msg类型专线map类型
	requestData := msg.(map[string]interface{})
	mapKey := requestData["key"].(string)
	mapData:= requestData["data"]
	if err != nil {
		panic("json Marshal data err：" + err.Error())
	}

	if _, ok := msgMap[mapKey]; !ok {
		panic("not exist key：" + mapKey)
	}

	// 保存数据
	msgMap[mapKey].data = mapData
	// websocket链接
	wsConn := msgMap[mapKey].wsConn

	// 判断是否过期
	unixTime := time.Now().Unix()
	if unixTime > msgMap[mapKey].expiresTime {
		wsConn.Close()
		panic("websocket connect expired，key：" + err.Error())
	}

	// 发送websocket数据
	wsResp := wsResponse{Code : 0, MsgKey : mapKey, Data : mapData, ErrorMsg : ""}
	wsRtn, _ := json.Marshal(wsResp)
	if err := wsConn.WriteMessage(websocket.TextMessage, wsRtn); err != nil {
		wsConn.Close()
		panic("websocket send message err：" + err.Error())
	}

	httpRtn, _ := json.Marshal(httpResponseResult)
	response.Write(httpRtn)
}

func main() {
	addrs := []Server{{"0.0.0.0", 8080, wsHandler}, {"0.0.0.0", 8081, httpHandler}}

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