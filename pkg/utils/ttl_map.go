package utils

import (
	"math"
	"sync"
	"time"
)

var PermanentTTL = time.Duration(math.MaxInt64)

type item struct {
	value      any
	lastAccess int64
	ttl        time.Duration
}

type TTLMap struct {
	m      map[string]*item
	l      sync.Mutex
	maxTTL time.Duration
}

func NewTTLMap(maxTTL time.Duration) (m *TTLMap) {
	m = &TTLMap{m: make(map[string]*item), maxTTL: maxTTL}
	go func() {
		for now := range time.Tick(time.Second) {
			m.l.Lock()
			for k, v := range m.m {
				if v.ttl == PermanentTTL {
					continue
				}
				if now.Unix()-v.lastAccess > int64(v.ttl/time.Second) {
					delete(m.m, k)
				}
			}
			m.l.Unlock()
		}
	}()
	return
}

func (m *TTLMap) Len() int {
	return len(m.m)
}

func (m *TTLMap) Put(k string, v any) {
	m.PutWithTTL(k, v, m.maxTTL)
}

func (m *TTLMap) PutWithTTL(k string, v any, ttl time.Duration) {
	m.l.Lock()
	it, ok := m.m[k]
	if !ok {
		it = &item{}
		m.m[k] = it
	}
	it.value = v
	it.lastAccess = time.Now().Unix()
	it.ttl = ttl
	m.l.Unlock()
}

func (m *TTLMap) Get(k string) (v any) {
	m.l.Lock()
	if it, ok := m.m[k]; ok {
		v = it.value
		it.lastAccess = time.Now().Unix()
	}
	m.l.Unlock()
	return
}

func (m *TTLMap) Peek(k string) (v any, ok bool) {
	m.l.Lock()
	if it, found := m.m[k]; found {
		ok = found
		v = it.value
	}
	m.l.Unlock()
	return
}
