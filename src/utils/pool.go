package utils

import "sync"

type Pool struct {
	poolMap map[int]*sync.Pool
}

func NetPool(lens ...int) *Pool {
	p := &Pool{
		make(map[int]*sync.Pool),
	}
	for _, l := range lens {
		p.poolMap[l] = p.newPool(l)
	}
	return p
}

func (p *Pool) Get(l int) *[]byte {
	if _, ok := p.poolMap[l]; !ok {
		p.poolMap[l] = p.newPool(l)
	}
	res := p.poolMap[l].Get().(*[]byte)
	p.poolMap[l].Put(res)
	return res
}

func (p *Pool) Reset(lens ...int) {
	for _, l := range lens {
		bs := p.Get(l)
		for i := 0; i < l; i++ {
			(*bs)[i] = 0
		}
	}
}

func (p *Pool) ResetByBsp(bsps ...*[]byte) {
	for _, bsp := range bsps {
		if bsp != nil {
			p.Reset(len(*bsp))
		}
	}
}

func (p *Pool) newPool(l int) *sync.Pool {
	return &sync.Pool{
		New: func() interface{} {
			bs := make([]byte, l)
			return &bs
		},
	}
}
