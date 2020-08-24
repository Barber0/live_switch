package rtmp

import (
	"fmt"
	"live/src/utils"
	"net"
	"runtime/debug"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/sirupsen/logrus"
)

var publisherMap = cmap.New()

func NewRTMPServer(addr string) {
	defer utils.HandlePanic(func(err error) {
		logrus.Error("handle panic, err: ", err)
		fmt.Println(string(debug.Stack()))
	})
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		logrus.Error("resolve tcp addr failed, err: ", err)
		return
	}
	lis, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		logrus.Error("listen tcp failed, err: ", err)
	}
	for {
		tcpConn, err := lis.Accept()
		if err != nil {
			logrus.Error("listener accept failed, err: ", err)
			continue
		}
		conn := newConn(tcpConn)

		// go handleConn(conn)
		handleConn(conn)
	}
}

func handleConn(conn *connection, carr ...Consumer) {
	defer utils.HandlePanic(func(err error) {
		logrus.Error("conn panic, err: ", err)
		conn.Close()
	})
	utils.CheckErr(conn.handshake())

	for !conn.done {
		msg, err := conn.readMsg()
		utils.CheckErr(err)

		conn.handleCtrlMsg(msg)
		utils.CheckErr(conn.ack(msg.length))

		switch msg.typeId {
		case TYPE_ID_CMD_MSG_AMF0, TYPE_ID_CMD_MSG_AMF3:
			utils.CheckErr(conn.handleCmdMsg(msg))
		}
	}

	var publisher *Publisher
	if tmpPub, ok := publisherMap.Get(conn.getPublisherName()); ok {
		publisher = tmpPub.(*Publisher)
		publisher.SetReader(NewVideoReader(conn))
		publisher.AddConsumer(carr...)
		publisher.Start()
		fmt.Println(publisher.name)
	}
}
