package rtmp

const (
	TYPE_ID_SET_CHUNK_SIZE     byte = 1
	TYPE_ID_ABORT_MSG          byte = 2
	TYPE_ID_ACK                byte = 3
	TYPE_ID_USER_CTRL_MSG      byte = 4
	TYPE_ID_WINDOW_ACK_SIZE    byte = 5
	TYPE_ID_SET_PEER_BANDWIDTH byte = 6

	TYPE_ID_AUDIO_MSG byte = 8
	TYPE_ID_VIDEO_MSG byte = 9

	TYPE_ID_DATA_MSG_AMF3 byte = 15
	TYPE_ID_DATA_MSG_AMF0 byte = 18

	TYPE_ID_SHARED_OBJ_MSG_AMF3 byte = 16
	TYPE_ID_CMD_MSG_AMF3        byte = 17
	TYPE_ID_SHARED_OBJ_MSG_AMF0 byte = 19
	TYPE_ID_CMD_MSG_AMF0        byte = 20
)

const (
	CSID_AUTO        = 0
	CSID_PRO_CTRL    = 2
	CSID_CMD         = 3
	CSID_AUDIO       = 4
	CSID_VIDEO       = 6
	CSID_CTRL_STREAM = 8
)

const (
	CMD_CONNECT       = "connect"
	CMD_CALL          = "call"
	CMD_CLOSE         = "close"
	CMD_CREATE_STREAM = "createStream"
	CMD_PUBLISH       = "publish"

	CMD_ONSTATUS = "onStatus"
)
