package manager

import (
	"sync"

	"git.resultys.com.br/motor/service"
)

type queue struct {
	items   []*service.Unit
	mx      *sync.Mutex
	max     int
	running int
}

func createQueue() *queue {
	return &queue{
		mx:      &sync.Mutex{},
		items:   []*service.Unit{},
		max:     0,
		running: 0,
	}
}

func (q *queue) push(unit *service.Unit) bool {
	if q.max == 0 {
		return false
	}

	q.lock()
	if q.max > 0 && q.running == q.max {
		unit.SetStatus("Queued")
		q.push_unlock(unit)
		q.unlock()
		return true
	}
	q.running++
	q.unlock()

	return false
}

func (q *queue) exist(unit *service.Unit) bool {
	if q.max == 0 {
		return false
	}

	q.lock()
	defer q.unlock()

	for i := 0; i < len(q.items); i++ {
		if q.items[i].ID == unit.ID {
			return true
		}
	}

	return false
}

func (q *queue) removeByTokenID(id string) {
	if q.max == 0 {
		return
	}

	q.lock()
	defer q.unlock()

	for i := 0; i < len(q.items); i++ {
		if q.items[i].Token.TokenID.Hex() == id {
			q.items = append(q.items[:i], q.items[i+1:]...)
			break
		}
	}
}

func (q *queue) remove(unit *service.Unit) {
	if q.max == 0 {
		return
	}

	q.lock()
	defer q.unlock()

	for i := 0; i < len(q.items); i++ {
		if q.items[i].ID == unit.ID {
			q.items = append(q.items[:i], q.items[i+1:]...)
			break
		}
	}
}

func (q *queue) pop() *service.Unit {
	if q.max == 0 {
		return nil
	}

	q.lock()
	defer q.unlock()

	if q.running > 0 {
		q.running--
	}

	return q.pop_unlock()
}

func (q *queue) capacity(max int) *queue {
	q.max = max

	return q
}

func (q *queue) push_unlock(unit *service.Unit) {
	q.items = append(q.items, unit)
}

func (q *queue) pop_unlock() *service.Unit {
	if len(q.items) > 0 {
		unit := q.items[0]
		q.items = append(q.items[:0], q.items[1:]...)
		return unit
	}

	return nil
}

func (q *queue) lock() {
	q.mx.Lock()
}

func (q *queue) unlock() {
	q.mx.Unlock()
}
