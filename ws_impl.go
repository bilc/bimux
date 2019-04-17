package bimux

/*
websocket implement
*/

import (
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

func NewWebSocketMuxer(
	w http.ResponseWriter,
	r *http.Request,
	rpcServeHook RpcServeFunc,
	onewayServeHook OnewayFunc,
) (Muxer, error) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return newMuxer(newWSConn(c), rpcServeHook, onewayServeHook, nil)
}

/*
 Connection with websocket addr(client)
*/
func Dial(
	addr string,
	rpcServeHook RpcServeFunc,
	onewayServeHook OnewayFunc,
	closeHook CloseFunc,
) (Muxer, error) {
	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		return nil, err
	}

	return newMuxer(newWSConn(c), rpcServeHook, onewayServeHook, closeHook)
}
