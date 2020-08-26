package main

import (
	"time"
	"fmt"
	"flag"
	//	"sync/atomic"

	"github.com/bilc/bimux"
)



func RpcServe(route int32 , req []byte, m bimux.Muxer) []byte{
		fmt.Println("hello",route, string(req))
		return append(req, []byte("-reverse call OK-")...)
}

func main() {
	flag.Parse()

	conn,_ := bimux.WsDial("localhost:18080", "/mux", RpcServe, nil ,nil)
	time.Sleep(time.Second*10)
	conn.Close()

}
