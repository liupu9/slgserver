package net

import (
	"context"
	"time"
)

// ReqBody 请求内容
type ReqBody struct {
	Seq   int64       `json:"seq"`   // 请求序列号
	Name  string      `json:"name"`  // 请求协议名
	Msg   interface{} `json:"msg"`   // 请求内容
	Proxy string      `json:"proxy"` // 代理转发
}

// RspBody 响应内容
type RspBody struct {
	Seq  int64       `json:"seq"`  // 响应序列号
	Name string      `json:"name"` // 响应协议名
	Code int         `json:"code"` // 响应编码
	Msg  interface{} `json:"msg"`  // 响应内容
}

type WsMsgReq struct {
	Body *ReqBody
	Conn WSConn
}

type WsMsgRsp struct {
	Body *RspBody
}

const HandshakeMsg = "handshake"
const HeartbeatMsg = "heartbeat"

type Handshake struct {
	Key string `json:"key"`
}

type Heartbeat struct {
	CTime int64 `json:"ctime"`
	STime int64 `json:"stime"`
}

// WSConn websocket连接处理
type WSConn interface {
	SetProperty(key string, value interface{})
	GetProperty(key string) (interface{}, error)
	RemoveProperty(key string)
	Addr() string
	Push(name string, data interface{})
}

type syncCtx struct {
	ctx     context.Context
	cancel  context.CancelFunc
	outChan chan *RspBody
}

func newSyncCtx() *syncCtx {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	return &syncCtx{ctx: ctx, cancel: cancel, outChan: make(chan *RspBody)}
}

func (this *syncCtx) wait() *RspBody {
	defer this.cancel()
	select {
	case data := <-this.outChan:
		return data
	case <-this.ctx.Done():
		return nil
	}
}
