package reply

import (
	"bytes"
	"go-redis/interface/resp"
	"strconv"
)

var (
	nullBulkReplyBytes = []byte("$-1")

	// CRLF is the line separator of redis serialization protocol
	CRLF = "\r\n"
)

type ErrorReply interface {
	Error() string
	ToBytes() []byte
}
type BulkReply struct {
	Arg []byte
}

func NewBulkReply(arg []byte) *BulkReply {
	return &BulkReply{Arg: arg}
}

type MultiBulkReply struct {
	Args [][]byte
}
type StatusReply struct {
	Status string
}

type IntReply struct {
	Code int64
}

func (i IntReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(i.Code, 10) + CRLF)
}

func NewIntReply(code int64) *IntReply {
	return &IntReply{Code: code}
}

func (r *StatusReply) ToBytes() []byte {
	return []byte("+" + r.Status + CRLF)
}

func NewStatusReply(status string) *StatusReply {
	return &StatusReply{Status: status}
}

func NewMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{Args: args}
}

func (m *MultiBulkReply) ToBytes() []byte {
	argLen := len(m.Args)
	var buf bytes.Buffer
	buf.WriteString("*" + strconv.Itoa(argLen) + CRLF)
	for _, arg := range m.Args {
		if arg == nil {
			buf.WriteString("$-1" + CRLF)
		} else {
			buf.WriteString("$" + strconv.Itoa(len(arg)) + CRLF + string(arg) + CRLF)
		}
	}
	return buf.Bytes()
}

func (b *BulkReply) ToBytes() []byte {
	if len(b.Arg) == 0 {
		return nullBulkReplyBytes
	}
	return []byte("$" + strconv.Itoa(len(b.Arg)) + CRLF + string(b.Arg) + CRLF)
}

type StandardErrReply struct {
	Status string
}

//NewErrReply  creates StandardErrReply
func NewErrReply(status string) *StandardErrReply {
	return &StandardErrReply{Status: status}
}

func (s *StandardErrReply) ToBytes() []byte {
	return []byte("-" + s.Status + CRLF)
}
func (s *StandardErrReply) Error() string {
	return s.Status
}
func IsErrorReply(reply resp.Reply) bool {
	return reply.ToBytes()[0] == '-'
}
