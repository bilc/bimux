package main

import (
	"flag"
	"fmt"
	"net/http"

	_ "net/http/pprof"

	".."
)

var addr = flag.String("addr", "0.0.0.0:18080", "http service address")

func handler(w http.ResponseWriter, r *http.Request) {
	muxer, _ := bimux.NewWebSocketMuxer(w, r,
		func(route int32, req []byte, m bimux.Muxer) []byte {
			return []byte("helloworld")
		}, nil)

	fmt.Println("server conn over ", muxer.Wait())
}

func main() {
	flag.Parse()
	http.HandleFunc("/mux", handler)
	fmt.Println(http.ListenAndServe(*addr, nil))
}
