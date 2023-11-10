package database

import (
	"go-redis/interface/resp"
	"go-redis/resp/reply"
)

func Ping(db *DB, args [][]byte) resp.Reply {
	return &reply.PongReply{}
	//return reply.NewStatusReply(string(args[0]))
}
func init() {
	RegisterCommand("ping", Ping, 1)
}
