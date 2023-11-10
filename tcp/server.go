package tcp

import (
	"context"
	"go-redis/interface/tcp"
	"go-redis/lib/logger"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Config struct {
	Address string
}

func ListenAndServeWithSignal(conf *Config, handler tcp.Handler) error {

	listener, err := net.Listen("tcp", conf.Address)
	closeChan := make(chan struct{})
	sigChan := make(chan os.Signal)
	// 接收OS 转发的信号
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sigChan
		//将关闭信号向下转发
		switch s {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()
	if err != nil {
		return err
	}
	logger.Info("start listen")
	ListenAndServe(listener, handler, closeChan)
	return nil
}

func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	// 监听closeChan
	go func() {
		// 没有缓存区的chan 则会阻塞 goroutine 进行挂起
		<-closeChan
		logger.Info("shutting down")
		// 保证所有的链接都在退出时关闭
		_ = listener.Close()
		_ = handler.Close()
	}()
	// 保证所有的链接都在退出时关闭
	defer func() {
		_ = listener.Close()
		_ = handler.Close()
	}()
	ctx := context.Background()
	// 保证所有链接都处理完毕后再关闭链接
	var waitDone sync.WaitGroup

	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		logger.Info("accept link")
		waitDone.Add(1)
		go func() {
			defer func() {
				waitDone.Done()
			}()
			handler.Handle(ctx, conn)

		}()
	}
	waitDone.Wait()
}
