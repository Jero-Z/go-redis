package tcp

import (
	"bufio"
	"context"
	"go-redis/lib/logger"
	"go-redis/lib/sync/atomic"
	"go-redis/lib/sync/wait"
	"io"
	"net"
	"sync"
	"time"
)

type EchoHandler struct {
	activeConn sync.Map
	closing    atomic.Boolean
}

func NewEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

type EchoClient struct {
	Conn    net.Conn
	Waiting wait.Wait
}

func (e *EchoClient) Close() error {
	e.Waiting.WaitWithTimeout(time.Second * 10)

	_ = e.Conn.Close()
	return nil
}

func (handler *EchoHandler) Handle(ctx context.Context, conn net.Conn) {

	if handler.closing.Get() {
		_ = conn.Close()
	}
	client := &EchoClient{Conn: conn}
	handler.activeConn.Store(client, struct{}{})
	reader := bufio.NewReader(conn)
	for {
		readString, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logger.Info("Connection close")
				handler.activeConn.Delete(client)
			} else {
				logger.Warn(err)
			}
			return
		}
		// 给每个client 进行业务运行标示
		client.Waiting.Add(1)

		b := []byte(readString)
		// 回复给client
		_, _ = conn.Write(b)
		client.Waiting.Done()
	}
}

func (handler *EchoHandler) Close() error {
	logger.Info("handler shutting down ")

	handler.closing.Set(true)
	handler.activeConn.Range(func(key, value any) bool {
		client := key.(*EchoClient)
		_ = client.Conn.Close()
		return true
	})
	return nil
}
