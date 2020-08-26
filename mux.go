package bimux

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
)

type RpcServeFunc func(route int32, req []byte, m Muxer) []byte
type OnewayFunc func(route int32, req []byte, m Muxer)
type CloseFunc func(Muxer)

var ErrTimeout error = errors.New("timeout")

type Muxer interface {
	Send(route int32, data []byte) error
	Rpc(route int32, req []byte, timeout time.Duration) (rsp []byte, err error)
	SetID(id string)
	GetID() string

	//raw read and write
	ReadPacket() (data []byte, err error)
	WritePacket(data []byte) error

	Wait() error
	Close()
}

// mockgen -source=./mux.go -destination=mux_mock.go -package=bimux

type muxer struct {
        Connection
	id string

	writeMutex sync.Mutex
	readMutex  sync.Mutex

	callerRsp      map[uint64]chan *Message
	callerRspMutex sync.Mutex

	rpcServeHook    RpcServeFunc
	onewayServeHook OnewayFunc
	closeHook       CloseFunc

	isClose bool

	wg sync.WaitGroup

	number   uint64
	existErr error
}

func newMuxer(conn Connection, rpcServeHook RpcServeFunc, onewayServeHook OnewayFunc, closeHook CloseFunc) (Muxer, error) {
	m := &muxer{
		Connection:      conn,
		callerRsp: make(map[uint64]chan *Message),

		rpcServeHook:    rpcServeHook,
		onewayServeHook: onewayServeHook,
		closeHook:       closeHook,
	}

	m.wg.Add(1)
	go m.loop()

	return m, nil
}

func (m *muxer) SetID(id string) {
	m.id = id
}

func (m *muxer) GetID() string{
	return m.id
}

/*
 Send Message
*/
func (m *muxer) Send(route int32, data []byte) error {
	return m.writeMsg(m.newMsg(Flag_oneway, route, data))
}

/*
 Rpc Send Message
*/
func (m *muxer) Rpc(route int32, req []byte, timeout time.Duration) (rsp []byte, err error) {
	// in case response will be handled before Ask get lock, create channel first and send request.
	reqMsg := m.newMsg(Flag_request, route, req)
	waitChan := m.register(reqMsg.Number)
	defer m.unRegister(reqMsg.Number)

	if err := m.writeMsg(reqMsg); err != nil {
		return nil, err
	}

	select {
	case answer := <-waitChan:
		return answer.Data, nil
	case <-time.After(timeout):
		return nil, ErrTimeout
	}
}

func (m *muxer) Wait() error {
	m.wg.Wait()
	return m.existErr
}

func (m *muxer) Close() {
	m.isClose = true
	m.Connection.Close()
}

func (m *muxer) readMsg() (*Message, error) {
	m.readMutex.Lock()
	defer m.readMutex.Unlock()
	pack, err := m.Connection.ReadPacket()
	if err != nil {
		return nil, err
	}
	var msg Message
	err = proto.Unmarshal(pack, &msg)
	return &msg, err
}

func (m *muxer) newMsg(flag Flag, route int32, data []byte) *Message {
	return &Message{
		Number: atomic.AddUint64(&m.number, uint64(1)),
		Flag:   flag,
		Route:  route,
		Data:   data,
	}
}

func (m *muxer) responseMsg(reqMsg *Message, route int32, data []byte) *Message {
	return &Message{
		Number: reqMsg.Number,
		Flag:   Flag_response,
		Route:  route,
		Data:   data,
	}
}

func (m *muxer) writeMsg(msg *Message) error {
	m.writeMutex.Lock()
	defer m.writeMutex.Unlock()
	b, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	err = m.Connection.WritePacket(b)
	return err
}

func (m *muxer) register(number uint64) chan *Message {
	waitChan := make(chan *Message, 1)
	m.callerRspMutex.Lock()
	m.callerRsp[number] = waitChan
	m.callerRspMutex.Unlock()
	return waitChan
}

func (m *muxer) unRegister(number uint64) {
	m.callerRspMutex.Lock()
	close(m.callerRsp[number])
	delete(m.callerRsp, number)
	m.callerRspMutex.Unlock()
}

func (m *muxer) notify(number uint64, msg *Message) {
	m.callerRspMutex.Lock()
	if waitChan, ok := m.callerRsp[number]; ok {
		waitChan <- msg
	}
	m.callerRspMutex.Unlock()
}

func (m *muxer) loop() {
	defer m.wg.Done()
	for {
		msg, err := m.readMsg()
		if err != nil {
			if !m.isClose && m.closeHook != nil {
				m.closeHook(m)
			}
			return
		}

		switch msg.Flag {
		case Flag_response:
			m.notify(msg.Number, msg)

		case Flag_request:
			if m.rpcServeHook != nil {
				go func(tmp *Message) {
					rsp := m.rpcServeHook(tmp.Route, tmp.Data, m)
					rspMsg := m.responseMsg(tmp, tmp.Route, rsp)
					m.writeMsg(rspMsg)
				}(msg)
			}

		case Flag_oneway:
			if m.onewayServeHook != nil {
				go func(tmp *Message) {
					m.onewayServeHook(tmp.Route, tmp.Data, m)
				}(msg)
			}
		}
	}
}
