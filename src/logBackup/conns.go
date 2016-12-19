package logBackup

import "sync"

type ConnManager struct {
	*sync.WaitGroup
	Counter int
}

func NewConnManager() *ConnManager {
	cm := &ConnManager{}
	cm.WaitGroup = &sync.WaitGroup{}
	return cm
}

func (cm *ConnManager) Add(delta int) {
	cm.Counter += delta
	cm.WaitGroup.Add(delta)
}

func (cm *ConnManager) Done() {
	cm.Counter--
	cm.WaitGroup.Done()
}

type ConnState struct {

}