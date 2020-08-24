package utils

import (
	"bytes"
	"io"

	amf "github.com/Barber0/goamf-1"
)

var (
	globalAMFHandler AMFHandler = newAmfHandlerImpl()
)

func init() {
	GetAMFHandler = func() AMFHandler {
		return globalAMFHandler
	}
}

type amfHandlerImpl struct {
	encodeBuf *bytes.Buffer
	decodeBuf *bytes.Buffer
}

func (h *amfHandlerImpl) MetaDataReform(bts []byte, flag byte) ([]byte, error) {
	amfFlag := amf.ADD
	if flag == META_DATA_REFORM_FLAG_DEL {
		amfFlag = amf.DEL
	}
	return amf.MetaDataReform(bts, byte(amfFlag))
}

func newAmfHandlerImpl() *amfHandlerImpl {
	return &amfHandlerImpl{
		bytes.NewBuffer(nil),
		bytes.NewBuffer(nil),
	}
}

func (h *amfHandlerImpl) Encode(v interface{}, ver AMFVersion) ([]byte, error) {
	h.encodeBuf.Reset()
	encoder := amf.WriteValue
	if ver == AMF3 {
		encoder = amf.AMF3_WriteValue
	}
	_, err := encoder(h.encodeBuf, v)
	if err != nil {
		return nil, err
	}
	bs := make([]byte, len(h.encodeBuf.Bytes()))
	copy(bs, h.encodeBuf.Bytes())
	return bs, nil
}

func (h *amfHandlerImpl) EncodeBatch(ver AMFVersion, args ...interface{}) ([]byte, error) {
	h.encodeBuf.Reset()
	encoder := amf.WriteValue
	if ver == AMF3 {
		encoder = amf.AMF3_WriteValue
	}
	for _, v := range args {
		_, err := encoder(h.encodeBuf, v)
		if err != nil {
			return nil, err
		}
	}
	bs := make([]byte, len(h.encodeBuf.Bytes()))
	copy(bs, h.encodeBuf.Bytes())
	return bs, nil
}

func (h *amfHandlerImpl) Decode(bs []byte, ver AMFVersion) (interface{}, error) {
	h.decodeBuf.Reset()
	h.decodeBuf.Write(bs)
	decoder := amf.ReadValue
	if ver == AMF3 {
		decoder = amf.AMF3_ReadValue
	}
	v, err := decoder(h.decodeBuf)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (h *amfHandlerImpl) DecodeBatch(bs []byte, ver AMFVersion) ([]interface{}, error) {
	h.decodeBuf.Reset()
	h.decodeBuf.Write(bs)
	decoder := amf.ReadValue
	if ver == AMF3 {
		decoder = amf.AMF3_ReadValue
	}
	resArr := []interface{}{}
	for {
		v, err := decoder(h.decodeBuf)
		// p := new(interface{})
		// err := h.decoder.Decode(p)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		resArr = append(resArr, v)
	}
	return resArr, nil
}
