package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"tcpproxy/internal"
	"tcpproxy/internal/myproto"
)

func RunClient() {
	ConnectServer()
	//Proxy()
}

var conn *websocket.Conn

func ConnectServer() {
	fmt.Println("连接server!")
	url := "ws://" + internal.MyConfig.Server.Host + ":" + internal.MyConfig.Server.Port + "/connect" //服务器地址
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	conn = ws
	if err != nil {
		log.Fatal(err)
	}
	//go func() {
	//	for {
	//		err := ws.WriteMessage(websocket.BinaryMessage, []byte("ping"))
	//		if err != nil {
	//			log.Fatal(err)
	//		}
	//		time.Sleep(time.Second * 2)
	//	}
	//}()

	for {
		_, data, err := ws.ReadMessage()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("收到服务端消息,长度:", len(string(data)))
		Proxy(data)
	}

}

func Proxy(dump []byte) {
	fmt.Println("获取server传输来的数据,进行转发!")

	var msg myproto.Msg
	err := json.Unmarshal(dump, &msg)
	if err != nil {
		return
	}
	var httpRequest myproto.HttpRequest
	httpRequestString := msg.Content
	err = json.Unmarshal([]byte(httpRequestString), &httpRequest)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("开始转发,拼装请求")
	req, _ := http.NewRequest(httpRequest.Method, httpRequest.Host+httpRequest.Path, bytes.NewReader(httpRequest.Body))
	req.Header = httpRequest.Header

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("请求失败:", err.Error())
		return
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("响应解析失败:", err.Error())
		return
	}
	httpRequest.Body = resBody
	httpRequest.Header = res.Header
	httpRequest.Code = res.StatusCode
	jsonStr, _ := json.Marshal(httpRequest)
	responseMsg := myproto.Msg{
		MsgType: 1,
		Content: string(jsonStr),
	}
	msgJson, _ := json.Marshal(responseMsg)
	fmt.Println("回写数据给服务端")
	conn.WriteMessage(websocket.BinaryMessage, msgJson)
	return

}
