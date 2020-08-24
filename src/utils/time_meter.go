package utils

type TimeMeter interface {
	BaseTimestamp() uint32
	CalcBaseTimestamp()
	RecTimestamp(timestamp uint32, typeID byte)
	SetPreTime()
	Alive() bool
}
