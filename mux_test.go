package bimux

import (
	//	"sync"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestNewMsg(t *testing.T) {
	m := new(muxer)
	m1 := m.newMsg(1, 1, nil)
	m2 := m.newMsg(2, 2, nil)
	assert.NotEqual(t, m1.Number, m2.Number)
	assert.Equal(t, m1.Number, uint64(1))
	assert.Equal(t, m2.Number, uint64(2))
}

func TestReadMsg(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	conn := NewMockConnection(ctrl)

	msg := &Message{Number: 1234}
	b, _ := proto.Marshal(msg)
	conn.EXPECT().ReadPacket().Return(b, nil)

	m := new(muxer)
	m.Connection = conn
	ret, err := m.readMsg()
	assert.Equal(t, msg.Number, ret.Number)
	assert.Equal(t, err, nil)

	msg = &Message{Number: 234}
	b, _ = proto.Marshal(msg)
	conn.EXPECT().ReadPacket().Return(b, nil)

	ret, err = m.readMsg()
	assert.Equal(t, msg.Number, ret.Number)
	assert.Equal(t, err, nil)
}

func TestWriteMsg(t *testing.T) {
}

func TestRegister(t *testing.T) {
	m := new(muxer)
	m.callerRsp = make(map[uint64]chan *Message)
	var no uint64
	for no = 0; no < 100; no++ {
		ch := m.register(no)
		msg := &Message{Number: no}
		m.notify(no, msg)
		ret := <-ch
		assert.Equal(t, no, ret.Number)
	}
	assert.True(t, len(m.callerRsp) == 100)
	for no = 0; no < 100; no++ {
		m.unRegister(no)
	}
	assert.True(t, len(m.callerRsp) == 0)
}

func TestWSRpc(t *testing.T) {
	http.HandleFunc("/mux", func(w http.ResponseWriter, r *http.Request) {
		NewWebSocketMuxer(w, r, func(route int32, req []byte, m Muxer) []byte {
			return append(req, []byte("ok")...)
		}, nil)
	})
	go func() {
		fmt.Println(http.ListenAndServe("127.0.0.1:12000", nil))
	}()

	time.Sleep(time.Second)
	m, err := Dial("ws://127.0.0.1:12000/mux", nil, nil, nil)
	assert.Nil(t, err)
	rsp, err := m.Rpc(1, []byte("1"), time.Second)
	assert.Nil(t, err)
	assert.Equal(t, rsp, []byte("1ok"))
}
