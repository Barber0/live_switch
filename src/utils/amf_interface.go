package utils

type AMFHandler interface {
	Encode(interface{}, AMFVersion) ([]byte, error)
	EncodeBatch(AMFVersion, ...interface{}) ([]byte, error)
	Decode([]byte, AMFVersion) (interface{}, error)
	DecodeBatch([]byte, AMFVersion) ([]interface{}, error)
	MetaDataReform([]byte, byte) ([]byte, error)
}

const (
	AMF0 AMFVersion = iota
	AMF3

	META_DATA_REFORM_FLAG_ADD = iota
	META_DATA_REFORM_FLAG_DEL
)

type (
	AMFVersion byte
	AMFObj     map[string]interface{}
)

var GetAMFHandler (func() AMFHandler)
