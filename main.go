package main

import (
	"live/src/protocol/httpflv"
	rtmp "live/src/protocol/rtmp"
	"net/http"
	"sync"
)

func main() {
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		svr := httpflv.NewHTTPFLVServer()
		http.ListenAndServe(":50001", svr)
	}()
	go func() {
		defer wg.Done()
		rtmp.NewRTMPServer("127.0.0.1:1935")
	}()

	wg.Wait()
}

//http://127.0.0.1:50001/live/alpha



//http://127.0.0.1:7001/live/movie.flv
