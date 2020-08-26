package rtmp

import "live/src/video"

const GOP_CACHE_CAP = 1

type PacketCache struct {
	gopPool   *GopCache
	videoPool *CachePool
	audioPool *CachePool
	metaPool  *CachePool
}

func NewPacketCache() *PacketCache {
	return &PacketCache{
		gopPool:   NewGopCache(GOP_CACHE_CAP),
		videoPool: &CachePool{},
		audioPool: &CachePool{},
		metaPool:  &CachePool{},
	}
}

type CachePool struct {
	full bool
	p    *video.Packet
}

func (cp *CachePool) Put(p *video.Packet) {
	cp.p = p
	cp.full = true
}

func (cp *CachePool) Send(w video.Writer) error {
	if !cp.full {
		return nil
	}
	return w.Write(cp.p)
}

func (pc *PacketCache) Write(p video.Packet) {
	switch p.DataType {
	case video.DATA_TYPE_META:
		pc.metaPool.Put(&p)
		return
	case video.DATA_TYPE_VIDEO:
		h, ok := p.Header.(video.VideoPacketHeader)
		if ok {
			if h.IsSeq() {
				pc.videoPool.Put(&p)
				return
			} else {
				return
			}
		}
	case video.DATA_TYPE_AUDIO:
		h, ok := p.Header.(video.AudioPacketHeader)
		if ok {
			if h.SoundFmt() == video.SOUND_FMT_AAC && h.AACPacketType() == video.AAC_PKT_TYPE_SEQHDR {
				pc.audioPool.Put(&p)
				return
			} else {
				return
			}
		}
	}
	pc.gopPool.Write(&p)
}

func (pc *PacketCache) Send(w video.Writer) error {
	if err := pc.metaPool.Send(w); err != nil {
		return err
	}

	if err := pc.videoPool.Send(w); err != nil {
		return err
	}

	if err := pc.audioPool.Send(w); err != nil {
		return err
	}

	if err := pc.gopPool.Send(w); err != nil {
		return err
	}

	return nil
}
