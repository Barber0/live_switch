package httpflv

import (
	"github.com/gorilla/mux"
	"live/src/protocol/rtmp"
	"net/http"
)

func NewHTTPFLVServer() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/{type}/{name}", StreamHandler)
	return r
}

func StreamHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pubType := vars["type"]
	pubName := vars["name"]
	pub := rtmp.GetPublisher(pubType + "/" + pubName)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	flvWriter := NewFLVWriter(w)
	pub.AddConsumer(flvWriter)
	flvWriter.Wait()
}
