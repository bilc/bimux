package main

import (
	"time"
	"fmt"
	_ "net/http/pprof"

	"github.com/bilc/bimux"
)


func main() {

	connChan := make(chan bimux.Muxer, 10)
	go func() {
		for i := range connChan {
			ret,err :=i.Rpc(111, []byte("test"), time.Second)
			fmt.Println(string(ret), err)
		}
	} ()
	bimux.WsServe("0.0.0.0:18080", "/mux", connChan, nil , nil, nil)
}
