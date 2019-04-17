package bimux

// mockgen -source=./conn.go -destination=conn_mock.go -package=bimux
type Connection interface {
	Close()
	ReadPacket() (p []byte, err error)
	WritePacket(data []byte) error
}
