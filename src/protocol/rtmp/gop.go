package rtmp

import (
	"fmt"
	"live/src/video"
)

const MAX_GOP_CAP = 1024

var ErrGopTooBig = fmt.Errorf("gop too big")

type gopArray struct {
	idx     int
	packets []*video.Packet
}

func newArray() *gopArray {
	arr := &gopArray{
		idx:     0,
		packets: make([]*video.Packet, 0, MAX_GOP_CAP),
	}
	return arr
}

func (arr *gopArray) write(p *video.Packet) error {
	if arr.idx > MAX_GOP_CAP {
		return ErrGopTooBig
	}
	arr.packets = append(arr.packets, p)
	return nil
}

func (arr *gopArray) reset() {
	arr.idx = 0
	arr.packets = arr.packets[:0]
}

func (arr *gopArray) send(w video.Writer) (err error) {
	for i := 0; i < arr.idx; i++ {
		pkt := arr.packets[i]
		if err = w.Write(pkt); err != nil {
			return
		}
	}
	return
}

type GopCache struct {
	gops     []*gopArray
	nextIdx  int
	length   int
	capacity int
	init     bool
}

func NewGopCache(cap int) *GopCache {
	gc := &GopCache{
		capacity: cap,
		gops:     make([]*gopArray, cap),
	}
	return gc
}

func (gc *GopCache) Write(p *video.Packet) {
	var ok bool
	switch p.DataType {
	case video.DATA_TYPE_VIDEO:
		h := p.Header.(video.VideoPacketHeader)
		if h.IsKeyFrame() && h.IsSeq() {
			ok = true
		}
	}
	if ok || gc.init {
		gc.init = true
		gc.write(p, ok)
	}
}

func (gc *GopCache) write(p *video.Packet, newArr bool) error {
	var arr *gopArray
	nextIdx := (gc.nextIdx + 1) % gc.capacity
	if newArr {
		arr = gc.gops[gc.nextIdx]
		if arr == nil {
			arr = newArray()
			gc.length++
			gc.gops[gc.nextIdx] = arr
		} else {
			arr.reset()
		}
		gc.nextIdx = nextIdx
	} else {
		arr = gc.gops[nextIdx]
	}
	return arr.write(p)
}

func (gc *GopCache) Send(w video.Writer) error {
	return gc.send(w)
}

func (gc *GopCache) send(w video.Writer) (err error) {
	pos := (gc.nextIdx + 1) % gc.capacity
	start := pos - gc.length + 1
	for i := 0; i < gc.length; i++ {
		idx := start + i
		if idx < 0 {
			idx += gc.capacity
		}
		gopArr := gc.gops[idx]
		if err = gopArr.send(w); err != nil {
			return
		}
	}
	return
}
