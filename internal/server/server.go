package server

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"tcpproxy/internal"
	"tcpproxy/internal/myproto"
	"time"
)

var ClientConn *websocket.Conn //一次只能保持一个连接

var clientRequest map[string]chan myproto.HttpRequest

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func StartServer() {
	clientRequest = map[string]chan myproto.HttpRequest{}
	r := gin.Default()
	r.GET("/connect", handleConnect)
	r.GET("/proxy/*path", handleProxy)
	fmt.Println("启动socket以及tcp服务")
	err := r.Run("0.0.0.0:" + internal.MyConfig.Server.Port)
	if err != nil {
		fmt.Println("挂了")
		return
	}
}

func handleConnect(c *gin.Context) {
	// 升级成 websocket 连接
	ws, err := upgrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("客户端连接成功:", ws.RemoteAddr())
	ClientConn = ws
	// 完成时关闭连接释放资源
	defer ws.Close()
	go func() {
		// 监听连接“完成”事件，其实也可以说丢失事件
		<-c.Done()
		// 这里也可以做用户在线/下线功能
		fmt.Println("客户端断开", ws.RemoteAddr())
	}()
	for {
		// 读取客户端发送过来的消息，如果没发就会一直阻塞住
		_, message, err := ws.ReadMessage()
		if err != nil {
			fmt.Println("read error", err.Error())
			break
		}

		var msg myproto.Msg
		err = json.Unmarshal(message, &msg)
		if err != nil {
			fmt.Println("解析失败,未知消息:" + string(message))
			return
		}
		switch msg.MsgType {
		case myproto.MsgTypePing:
			break
		case myproto.MsgTypeHttpRequest:
			fmt.Println("收到请求,开始解析")
			var httpRequest myproto.HttpRequest
			httpRequestString := msg.Content
			err = json.Unmarshal([]byte(httpRequestString), &httpRequest)
			if err != nil {
				fmt.Println("http解析失败 ", err.Error())
				return
			}
			fmt.Println("完成解析 chan 通信", httpRequest.ReqId)
			clientRequest[httpRequest.ReqId] <- httpRequest

		}
	}
}

func handleProxy(c *gin.Context) {
	fmt.Println("收到http请求,准备进行tcp转发")
	r := c.Request
	bodyBytes, _ := io.ReadAll(r.Body)
	httpRequest := &myproto.HttpRequest{
		ReqId:  time.Now().String(),
		Method: r.Method,
		Path:   c.Param("path"),
		Host:   internal.MyConfig.Server.Proxy,
		Header: r.Header,
		Body:   bodyBytes,
	}
	requestJson, err := json.Marshal(httpRequest)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	msg := myproto.Msg{
		MsgType: myproto.MsgTypeHttpRequest,
		Content: string(requestJson),
	}
	msgJson, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("向客户端发送tcp消息")
	err = ClientConn.WriteMessage(websocket.BinaryMessage, msgJson)

	if err != nil {
		fmt.Println("发送成功", err.Error())
		return
	}
	fmt.Println("发送成功")
	// 要阻塞,指到客户端返回
	clientRequest[httpRequest.ReqId] = make(chan myproto.HttpRequest, 1)

	// 阻塞
	fmt.Println("阻塞,等待客户端返回")
	resp := <-clientRequest[httpRequest.ReqId]
	fmt.Println("客户端返回,开始拼装响应体")
	for k, v := range resp.Header {
		if w, ok := c.Writer.(gin.ResponseWriter); ok {
			w.Header()[k] = v
		}
	}
	c.Data(resp.Code, resp.Header.Get("Content-Type"), resp.Body)
	fmt.Println("完成响应,删除 chan ")
	delete(clientRequest, httpRequest.ReqId)

}
