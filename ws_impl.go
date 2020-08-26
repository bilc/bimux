package bimux

/*
websocket implement
*/

import (
	"net/url"
	"net/http"

	"github.com/gorilla/websocket"
)

type wsConn struct {
	con *websocket.Conn
}

func newWSConn(c *websocket.Conn) Connection {
	return &wsConn{con: c}
}

func (c *wsConn) Close() {
	c.con.Close()
}

func (c *wsConn) ReadPacket() (p []byte, err error) {
	_, message, err := c.con.ReadMessage()
	return message, err
}

func (c *wsConn) WritePacket(data []byte) error {
	return c.con.WriteMessage(websocket.BinaryMessage, data)
}

/*
 Connection with web(server)
*/
var upgrader = websocket.Upgrader{} // use default options


func WsServe(addr, uri string,
	connections chan Muxer,
	rpcServeHook RpcServeFunc,
	onewayServeHook OnewayFunc,
	closeHook CloseFunc,
)  error{
	
	http.HandleFunc(uri, func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return 
		}
		muxer, _ := newMuxer(newWSConn(c), rpcServeHook, onewayServeHook, nil)
		connections <- muxer
	} )
	return http.ListenAndServe(addr, nil)
}

func WsDial(addr, uri string,
	rpcServeHook RpcServeFunc,
	onewayServeHook OnewayFunc,
	closeHook CloseFunc,
) (Muxer, error){
	
	u := url.URL{Scheme: "ws", Host: addr, Path: uri}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	return newMuxer(newWSConn(c), rpcServeHook, onewayServeHook, closeHook)
}
