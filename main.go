package main

import rtmp "live/src/protocol/rtmp"

func main() {
	rtmp.NewRTMPServer("127.0.0.1:1935")
}
