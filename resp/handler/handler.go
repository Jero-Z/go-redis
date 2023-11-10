package handler

import (
	"context"
	"go-redis/cluster"
	config "go-redis/conf"
	database2 "go-redis/database"
	"go-redis/interface/database"
	"go-redis/lib/logger"
	"go-redis/lib/sync/atomic"
	"go-redis/resp/connection"
	"go-redis/resp/parse"
	"go-redis/resp/reply"
	"io"
	"net"
	"strings"
	"sync"
)

var unknownErrReplyBytes = []byte("-ERR unknown\r\n")

type RespHandler struct {
	activeConn sync.Map
	db         database.Database
	closing    atomic.Boolean
}

func NewRespHandler() *RespHandler {
	var db database.Database
	if config.Properties.Self != "" &&
		len(config.Properties.Peers) > 0 {
		db = cluster.NewClusterDatabase()
	} else {
		db = database2.NewStandaloneDatabase()
	}
	return &RespHandler{
		db: db,
	}
}

func (r *RespHandler) closeClient(client *connection.Connection) {
	_ = client.Close()
	r.db.AfterClientClose(client)
	r.activeConn.Delete(client)
}

func (r *RespHandler) Handle(ctx context.Context, conn net.Conn) {
	if r.closing.Get() {
		_ = conn.Close()
	}
	client := connection.NewConnection(conn)
	r.activeConn.Store(client, struct {
	}{})
	ch := parse.ParserStream(conn)
	// 处理解析的消息
	for payload := range ch {
		if payload.Err != nil {
			// 关闭错误
			if payload.Err == io.EOF ||
				payload.Err == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				r.closeClient(client)
				logger.Info("connection closed:" + client.RemoteAddr().String())
				return
			}
			errReply := reply.NewErrReply(payload.Err.Error())
			err := client.Write(errReply.ToBytes()) // 回信错误
			if err != nil {
				r.closeClient(client)
				logger.Info("connection closed:" + client.RemoteAddr().String())
			}
			continue
		}
		// 处理正常流程
		if payload.Data == nil {
			logger.Error("empty payload")
			continue
		}
		res, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("require multi bulk reply")
			continue
		}
		result := r.db.Exec(client, res.Args)
		if result != nil {
			_ = client.Write(result.ToBytes())
		} else {
			_ = client.Write(unknownErrReplyBytes)
		}
	}
}

func (r *RespHandler) Close() error {
	logger.Info("handler shutting down")

	r.closing.Set(true)
	r.activeConn.Range(func(key, value any) bool {
		client := key.(*connection.Connection)
		_ = client.Close()
		return true
	})
	r.db.Close()
	return nil
}
