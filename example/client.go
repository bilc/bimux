package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	//	"sync/atomic"
	"time"

	".."
)

var addr = flag.String("addr", "localhost:18080", "http service address")

//var count int64 = 0

func main() {
	flag.Parse()

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/mux"}
	mux, _ := bimux.Dial(u.String(), nil, nil, nil)

	http.HandleFunc("/mux", func(w http.ResponseWriter, r *http.Request) {
		//atomic.AddInt64(&count, 1)
		ret, err := mux.Rpc(1, nil, time.Second)
		if err == nil {
			w.Write(ret)
		} else {
			w.WriteHeader(505)
			w.Write([]byte(err.Error()))
		}
	})
	fmt.Println(http.ListenAndServe("127.0.0.1:12000", nil))
}
