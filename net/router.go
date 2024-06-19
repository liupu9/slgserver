package net

import (
	"strings"

	"github.com/llr104/slgserver/log"
	"go.uber.org/zap"
)

// HandlerFunc 消息处理
type HandlerFunc func(req *WsMsgReq, rsp *WsMsgRsp)

// MiddlewareFunc 中间件
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// Group 路由分组（固定前缀）
type Group struct {
	prefix     string                      // 前缀
	hMap       map[string]HandlerFunc      // 消息处理器
	hMapMidd   map[string][]MiddlewareFunc // 中间件处理器
	middleware []MiddlewareFunc            // 中间件数组
}

func (this *Group) AddRouter(name string, handlerFunc HandlerFunc, middleware ...MiddlewareFunc) {
	this.hMap[name] = handlerFunc
	this.hMapMidd[name] = middleware
}

func (this *Group) Use(middleware ...MiddlewareFunc) *Group {
	this.middleware = append(this.middleware, middleware...)
	return this
}

func (this *Group) applyMiddleware(name string) HandlerFunc {

	h, ok := this.hMap[name]
	if ok == false {
		//通配符
		h, ok = this.hMap["*"]
	}

	if ok {
		for i := len(this.middleware) - 1; i >= 0; i-- {
			h = this.middleware[i](h)
		}

		for i := len(this.hMapMidd[name]) - 1; i >= 0; i-- {
			h = this.hMapMidd[name][i](h)
		}
	}

	return h
}

func (this *Group) exec(name string, req *WsMsgReq, rsp *WsMsgRsp) {
	h := this.applyMiddleware(name)
	if h == nil {
		log.DefaultLog.Warn("Group has not",
			zap.String("msgName", req.Body.Name))
	} else {
		h(req, rsp)
	}
}

// Router 路由，包含多个路由分组
type Router struct {
	groups []*Group
}

func (this *Router) Group(prefix string) *Group {
	g := &Group{
		prefix:   prefix,
		hMap:     make(map[string]HandlerFunc),
		hMapMidd: make(map[string][]MiddlewareFunc),
	}

	this.groups = append(this.groups, g)
	return g
}

func (this *Router) Run(req *WsMsgReq, rsp *WsMsgRsp) {
	name := req.Body.Name
	msgName := name
	sArr := strings.Split(name, ".")
	prefix := ""
	if len(sArr) == 2 {
		prefix = sArr[0]
		msgName = sArr[1]
	}

	for _, g := range this.groups {
		if g.prefix == prefix {
			g.exec(msgName, req, rsp)
		} else if g.prefix == "*" {
			g.exec(msgName, req, rsp)
		}
	}
}
