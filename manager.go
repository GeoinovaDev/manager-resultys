package manager

import (
	"runtime"

	"git.resultys.com.br/lib/lower/convert/encode"
	"git.resultys.com.br/lib/lower/str"
	"git.resultys.com.br/motor/manager/web"
	"git.resultys.com.br/motor/models/token"
	"git.resultys.com.br/motor/service"
	"git.resultys.com.br/motor/webhook"
	"git.resultys.com.br/motor/worker"
)

// Manager ...
type Manager struct {
	Web     *web.Interface
	Worker  *worker.Worker
	webhook *webhook.Manager

	waitQueue *queue

	fnNew      func(*token.Token) interface{}
	fnCache    func(*token.Token) (interface{}, bool)
	fnResponse func(interface{}) interface{}
	fnFinish   func(*token.Token, interface{})
}

// New ...
func New(port int, timeout int, limit int) *Manager {
	m := &Manager{
		Web:       web.New(port),
		Worker:    worker.New(timeout),
		webhook:   webhook.New(limit),
		waitQueue: createQueue(),
	}

	m.Init()

	return m
}

// Capacity ...
func (m *Manager) Capacity(total int) *Manager {
	m.waitQueue.capacity(total)
	return m
}

// OnNew ...
func (m *Manager) OnNew(fn func(*token.Token) interface{}) *Manager {
	m.fnNew = fn

	return m
}

// OnResponse ...
func (m *Manager) OnResponse(fn func(interface{}) interface{}) *Manager {
	m.fnResponse = fn

	return m
}

// OnCache ...
func (m *Manager) OnCache(fn func(*token.Token) (interface{}, bool)) *Manager {
	m.fnCache = fn

	return m
}

// OnFinish ...
func (m *Manager) OnFinish(fn func(*token.Token, interface{})) *Manager {
	m.fnFinish = fn

	return m
}

// Init ...
func (m *Manager) Init() *Manager {
	m.Web.OnCreate(func(tken *token.Token) {
		unit := service.New(tken, m.fnNew(tken))
		if m.fnCache == nil {
			m.run(unit)
			return
		}

		cache, existCache := m.fnCache(tken)

		if existCache {
			unit.Item = cache
			m.sendResponse(unit)
		} else {
			m.run(unit)
		}
	})

	m.Web.OnRemove(func(id string) {
		m.waitQueue.removeByTokenID(id)
	})

	m.Web.OnStats(func() string {
		json := make(map[string]interface{})
		json["running"] = m.Worker.Running()
		json["waiting"] = m.waitQueue.items

		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		json["mem_alloc"] = mem.Alloc / 1024 / 1024
		json["mem_sys"] = mem.Sys / 1024 / 1024

		return encode.JSON(json)
	})

	m.Web.OnDebug(func() string {
		units := m.Worker.Running()

		return encode.JSON(units)
	})

	m.Web.OnReload(func() {
		m.Worker.Reload()
	})

	return m
}

func (m *Manager) run(unit *service.Unit) {
	if unit == nil {
		return
	}

	if m.waitQueue.push(unit) {
		return
	}

	m.Worker.Run(unit, func(unit *service.Unit) {
		if m.fnFinish != nil {
			m.fnFinish(unit.Token, unit.Item)
		}

		go m.run(m.waitQueue.pop())
	}, func(unit *service.Unit) {
		m.sendResponse(unit)
	}, func(unit *service.Unit) {
		m.sendResponse(unit)
	})
}

func (m *Manager) sendResponse(unit *service.Unit) {
	url := str.Format("{0}?id={1}", unit.Token.Webhook, unit.Token.TokenID.Hex())
	data := unit.Item
	if m.fnResponse != nil {
		data = m.fnResponse(data)
	}
	m.webhook.Trigger(url, data)
}

// Start ...
func (m *Manager) Start() *Manager {
	m.Worker.Load()
	m.Web.Start()

	return m
}
