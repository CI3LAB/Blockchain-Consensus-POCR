package process

import (
	"sync"
)

// the execute op num now in state
type ExecuteOpNum struct {
	num    int
	locker *sync.RWMutex
}

func NewExecuteOpNum() *ExecuteOpNum {
	return &ExecuteOpNum{
		num:    0,
		locker: new(sync.RWMutex),
	}
}

func (n *ExecuteOpNum) Get() int {
	return n.num
}

func (n *ExecuteOpNum) Inc() {
	n.num = n.num + 1
}

func (n *ExecuteOpNum) Dec() {
	n.Lock()
	n.num = n.num - 1
	n.UnLock()
}

func (n *ExecuteOpNum) Zero() {
	n.Lock()
	n.num = 0
	n.UnLock()
}

func (n *ExecuteOpNum) Lock() {
	n.locker.Lock()
}

func (n *ExecuteOpNum) UnLock() {
	n.locker.Unlock()
}
