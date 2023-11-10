package reply

type UnknownErrReply struct {
}

type ArgNumErrReply struct {
	Cmd string
}
type SyntaxErrReply struct{}

type WrongTypeErrReply struct{}

type ProtocolErrReply struct {
	Msg string
}

var wrongTypeErrBytes = []byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n")

// ToBytes marshals redis.Reply
func (r *WrongTypeErrReply) ToBytes() []byte {
	return wrongTypeErrBytes
}

func (r *WrongTypeErrReply) Error() string {
	return "WRONGTYPE Operation against a key holding the wrong kind of value"
}

func NewSyntaxErrReply() *SyntaxErrReply {
	return &SyntaxErrReply{}
}

var syntaxErrBytes = []byte("-Err syntax error\r\n")

func NewArgNumErrReply(cmd string) *ArgNumErrReply {
	return &ArgNumErrReply{Cmd: cmd}
}

func (r *ArgNumErrReply) Error() string {
	return "ERR wrong number of arguments for '" + r.Cmd + "' command"
}

func (r *ArgNumErrReply) ToBytes() []byte {
	return []byte("-ERR wrong number of arguments for '" + r.Cmd + "' command\r\n")
}

var unknownErrBytes = []byte("-Err unknown\r\n")

func (u *UnknownErrReply) Error() string {
	return "-Err unknown"
}

func (u *UnknownErrReply) ToBytes() []byte {
	return unknownErrBytes
}
func (s SyntaxErrReply) Error() string {
	return "Err syntax error"
}

func (s SyntaxErrReply) ToBytes() []byte {
	return syntaxErrBytes
}

// ToBytes marshals redis.Reply
func (r *ProtocolErrReply) ToBytes() []byte {
	return []byte("-ERR Protocol error: '" + r.Msg + "'\r\n")
}

func (r *ProtocolErrReply) Error() string {
	return "ERR Protocol error: '" + r.Msg
}
