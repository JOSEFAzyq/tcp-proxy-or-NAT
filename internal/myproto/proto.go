package myproto

import (
	"bufio"
	"encoding/binary"
	"io"
	"net"
	"net/http"
)

const (
	MsgTypePing = iota
	MsgTypeHttpRequest
	MsgTypeHttpResponse
)

type HttpRequest struct {
	ReqId  string      `json:"reqId""`
	Host   string      `json:"host"`
	Method string      `json:"method"`
	Path   string      `json:"path"`
	Header http.Header `json:"header"`
	Body   []byte      `json:"body"`
	Code   int         `json:"code"`
}

type Msg struct {
	MsgType int
	Content string
}

// Receive
//
//	@Description: Receive 和 Send 方法旨在解决粘包问题,如果使用第三方程序则无需使用,比如gin.
//	@param conn
//	@return []byte
//	@return error
func Receive(conn net.Conn) ([]byte, error) {
	// 创建一个Reader
	rdr := bufio.NewReader(conn)

	lenBytes := make([]byte, 4)
	// 读取数据长度 这里用readFull很重要
	if _, err := io.ReadFull(rdr, lenBytes); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lenBytes)

	data := make([]byte, length)
	// 根据长度读取数据
	if _, err := io.ReadFull(rdr, data); err != nil {
		return nil, err
	}

	return data, nil
}

func Send(conn net.Conn, data []byte) error {
	length := int32(len(data))
	// 将数据长度写入一个4字节的切片
	lenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBytes, uint32(length))

	// 将长度和数据写入TCP流
	if _, err := conn.Write(lenBytes); err != nil {
		return err
	}
	if _, err := conn.Write(data); err != nil {
		return err
	}

	return nil
}
