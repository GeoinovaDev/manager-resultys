package web

import (
	"strconv"

	"git.resultys.com.br/lib/lower/convert/decode"
	"git.resultys.com.br/lib/lower/net"
	"git.resultys.com.br/lib/lower/server"
	"git.resultys.com.br/motor/models/token"
)

// Interface struct
type Interface struct {
	port int

	fnIndex  func() string
	fnCreate func(*token.Token)
	fnRemove func(string)
	fnReload func()
	fnDebug  func() string
	fnStats  func() string
}

// New ...
func New(port int) *Interface {
	return &Interface{port: port}
}

// SetPort ...
func (in *Interface) SetPort(port int) *Interface {
	in.port = port

	return in
}

// OnIndex ...
func (in *Interface) OnIndex(fn func() string) *Interface {
	in.fnIndex = fn

	return in
}

// OnCreate ...
func (in *Interface) OnCreate(fn func(*token.Token)) *Interface {
	in.fnCreate = fn

	return in
}

// OnRemove ...
func (in *Interface) OnRemove(fn func(string)) *Interface {
	in.fnRemove = fn

	return in
}

// OnReload ...
func (in *Interface) OnReload(fn func()) *Interface {
	in.fnReload = fn

	return in
}

// OnDebug ...
func (in *Interface) OnDebug(fn func() string) *Interface {
	in.fnDebug = fn

	return in
}

// OnStats ...
func (in *Interface) OnStats(fn func() string) *Interface {
	in.fnStats = fn

	return in
}

// Start ...
func (in *Interface) Start() {
	server.OnGet("/", func(qs server.QueryString) string {
		return in.fnIndex()
	})

	server.OnPost("/create", func(qs server.QueryString, data string) string {
		token := token.New()
		id := token.TokenID

		if len(data) > 0 {
			decode.JSON(data, token)
			token.TokenID = id
			go in.fnCreate(token)
		}

		return net.Success(token)
	})

	server.OnGet("/remove", func(qs server.QueryString) string {
		in.fnRemove(qs.Get("id"))

		return net.Success(nil)
	})

	server.OnGet("/reload", func(qs server.QueryString) string {
		in.fnReload()

		return "ok"
	})

	server.OnGet("/stats", func(qs server.QueryString) string {
		return in.fnStats()
	})

	server.OnGet("/debug", func(qs server.QueryString) string {
		return in.fnDebug()
	})

	server.Port = ":" + strconv.Itoa(in.port)
	server.Start()
}
