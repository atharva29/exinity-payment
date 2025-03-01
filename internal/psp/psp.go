package psp

import (
	"fmt"
	"sync"
)

type PSP struct {
	mu       sync.RWMutex
	gateways map[string]IPSP
}

func Init(psp []IPSP) *PSP {
	p := make(map[string]IPSP)
	for _, v := range psp {
		p[v.GetName()] = v
	}
	return &PSP{gateways: p}
}

func (p *PSP) Get(name string) (IPSP, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	res, ok := p.gateways[name]
	if !ok {
		return nil, fmt.Errorf("gateway %s not found", name)
	}
	return res, nil
}
