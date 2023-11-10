package reply

type PongReply struct {
}
type OkReply struct {
}

type NullBulkReply struct {
}
type EmptyMultiBulkReply struct {
}
type NoReply struct {
}

var pongBytes = []byte("+PONG\n\r")
var okBytes = []byte("+OK\n\r")
var nullBulkBytes = []byte("$-1\r\n")
var emptyMultiBulkBytes = []byte("*0\r\n")
var noBytes = []byte("")

func NewPongReply() *PongReply {
	return &PongReply{}
}
func NewOkReply() *OkReply {
	return &OkReply{}
}
func NewNullBulkReply() *NullBulkReply {
	return &NullBulkReply{}
}
func NewEmptyMultiBulkReply() *EmptyMultiBulkReply {
	return &EmptyMultiBulkReply{}
}
func NewNoReply() *NoReply {
	return &NoReply{}
}

func (p *PongReply) ToBytes() []byte {
	return pongBytes
}
func (o *OkReply) ToBytes() []byte {
	return okBytes
}
func (n *NullBulkReply) ToBytes() []byte {
	return nullBulkBytes
}

func (e EmptyMultiBulkReply) ToBytes() []byte {
	return emptyMultiBulkBytes
}

func (n *NoReply) ToBytes() []byte {
	return noBytes
}
