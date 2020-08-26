package rtmp

import (
	cmap "github.com/orcaman/concurrent-map"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"live/src/video"
)

type Publisher struct {
	id          string
	name        string
	cache       *PacketCache
	reader      streamReader
	consumerMap cmap.ConcurrentMap
}

type Consumer interface {
	GetStatus() bool
	SetStatus(v bool)
	Name() string
	Write(packet *video.Packet) error
	Close()
}

func newPublisher(name string) *Publisher {
	p := &Publisher{
		id:          uuid.NewV4().String(),
		name:        name,
		cache:       NewPacketCache(),
		consumerMap: cmap.New(),
	}
	return p
}

func GetPublisher(name string) *Publisher {
	if v, ok := publisherMap.Get(name); ok {
		return v.(*Publisher)
	}
	return nil
}

func (p *Publisher) AddConsumer(carr ...Consumer) {
	for _, c := range carr {
		cName := c.Name()
		p.consumerMap.Set(cName, c)
	}
}

func (p *Publisher) SetReader(r streamReader) {
	p.reader = r
}

func (p *Publisher) Start() {
	go func() {
		var pkt video.Packet

		for {
			if err := p.reader.Read(&pkt); err != nil {
				logrus.Error("read from rtmp failed, err: ", err)
				return
			}

			p.cache.Write(pkt)

			for item := range p.consumerMap.IterBuffered() {
				consumer := item.Val.(Consumer)
				if !consumer.GetStatus() {
					if err := p.cache.Send(consumer); err != nil {
						logrus.Errorf("send cache to consumer[%s] failed, err: %v", item.Key, err)
						p.consumerMap.Remove(item.Key)
						continue
					}
					consumer.SetStatus(true)
				} else {
					tmpPkt := pkt
					if err := consumer.Write(&tmpPkt); err != nil {
						logrus.Errorf("consumer[%s] write failed, err: %v", item.Key, err)
						p.consumerMap.Remove(item.Key)
					}
				}
			}
		}
	}()
}
