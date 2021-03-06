package util

import "sync"
//参考: https://github.com/Golangltd/LollipopGo/blob/master/util/map.go#L11
type Map struct {
	sync.RWMutex
	m map[interface{}]interface{}
}

func (m *Map) init() {
	if m.m == nil {
		m.m = make(map[interface{}]interface{})
	}
}
func (m *Map) UnsafeGet(key interface{}) interface{} {
	if m.m == nil {
		return nil
	}
	return m.m[key]
}
func (m *Map) Get(key interface{}) interface{} {
	m.RLock()
	defer m.RUnlock()
	return m.UnsafeGet(key)
}
func (m *Map) UnsafeSet(key, value interface{}) {
	m.init()
	m.m[key] = value
}
func (m *Map) Set(key, value interface{}) {
	m.Lock()
	defer m.Unlock()
	m.UnsafeSet(key, value)
}
func (m *Map) TrySet(key interface{}, value interface{}) interface{} {
	m.Lock()
	defer m.Unlock()

	m.init()

	if v, ok := m.m[key]; ok {
		return v
	} else {
		m.m[key] = value
		return nil
	}
}
func (m *Map) UnsafeDel(key interface{}) {
	m.init()
	delete(m.m, key)
}

func (m *Map) Del(key interface{}) {
	m.Lock()
	defer m.Unlock()
	m.UnsafeDel(key)
}

func (m *Map) UnsafeLen() int {
	if m.m == nil {
		return 0
	} else {
		return len(m.m)
	}
}

func (m *Map) Len() int {
	m.RLock()
	defer m.RUnlock()
	return m.UnsafeLen()
}
func (m *Map) UnsafeRange(f func(interface{}, interface{})) {
	if m.m == nil {
		return
	}
	for k, v := range m.m {
		f(k, v)
	}
}
func (m *Map) RLockRange(f func(interface{}, interface{})) {
	m.RLock()
	defer m.RUnlock()
	m.UnsafeRange(f)
}

func (m *Map) LockRange(f func(interface{}, interface{})) {
	m.Lock()
	defer m.Unlock()
	m.UnsafeRange(f)
}

//并发安全读取
func (m *Map) SafeRLockRange(data map[string]interface{}) map[string]interface{} {
	m.RLock()
	defer m.RUnlock()
	// 枚举处理
	if m.m == nil {
		return nil
	}
	for k, v := range m.m {
		if k == nil {
			continue
		}
		data[k.(string)] = v
	}

	return data
}
