package manager

import (
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

	fnNew    func() interface{}
	fnCache  func(*token.Token) (interface{}, bool)
	fnFinish func(*token.Token, interface{})
}

// New ...
func New(port int, timeout int, limit int) *Manager {
	m := &Manager{
		Web:     web.New(port),
		Worker:  worker.New(timeout),
		webhook: webhook.New(limit),
	}

	m.Init()

	return m
}

// OnNew ...
func (m *Manager) OnNew(fn func() interface{}) *Manager {
	m.fnNew = fn

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
		unit := service.New(tken, m.fnNew())
		if m.fnCache == nil {
			go m.run(unit)
			return
		}

		cache, loaded := m.fnCache(tken)

		if !loaded {
			go m.run(unit)
		} else {
			unit.Item = cache
			go m.sendResponse(unit)
		}
	})

	m.Web.OnReload(func() {
		m.Worker.Reload()
	})

	return m
}

func (m *Manager) run(unit *service.Unit) {
	m.Worker.Run(unit, func(unit *service.Unit) {
		if m.fnFinish != nil {
			m.fnFinish(unit.Token, unit.Item)
		}
	}, func(unit *service.Unit) {
		m.sendResponse(unit)
	}, func(unit *service.Unit) {
		m.sendResponse(unit)
	})
}

func (m *Manager) sendResponse(unit *service.Unit) {
	url := str.Format("{0}?id={1}", unit.Token.Webhook, unit.Token.TokenID.Hex())
	m.webhook.Trigger(url, unit.Item)
}

// Start ...
func (m *Manager) Start() *Manager {
	m.Worker.Load()
	m.Web.Start()

	return m
}
