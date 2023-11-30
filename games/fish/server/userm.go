package fish

import (
	"sync"
)

var _userM *userm

type userm struct {
	// 用户登录后都放在这个map,禁止遍历
	us map[string]*user
	// 用来锁us
	mu sync.RWMutex
}

func initUserM() {
	_userM = &userm{}
	_userM.us = make(map[string]*user)
}

func (m *userm) addUser(u *user) {
	m.mu.Lock()
	m.us[u.userInfo.Mid] = u
	m.mu.Unlock()
}

func (m *userm) delUser(mid string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.us, mid)
}

func (m *userm) getUser(mid string) (*user, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.us[mid]
	return u, ok
}
