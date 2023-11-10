package aof

import (
	config "go-redis/conf"
	"go-redis/interface/database"
	"go-redis/lib/logger"
	"go-redis/lib/utils"
	"go-redis/resp/connection"
	"go-redis/resp/parse"
	"go-redis/resp/reply"
	"io"
	"os"
	"strconv"
	"sync"
)

type CommandLine = [][]byte
type payLoad struct {
	cmdLine CommandLine
	dbIndex int
}

const (
	aofQueueSize = 1 << 16
)

type AofHandler struct {
	database    database.Database
	aofChan     chan *payLoad
	aofFile     *os.File
	aofFileName string
	currentDb   int
	aofFinished chan struct{}
	pausingAof  sync.RWMutex
}

func NewAofHandler(database database.Database) (*AofHandler, error) {
	handler := &AofHandler{}
	handler.aofFileName = config.Properties.AppendFilename
	handler.database = database
	handler.LoadAof(0)
	aofFile, err := os.OpenFile(handler.aofFileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	handler.aofFile = aofFile
	handler.aofChan = make(chan *payLoad, aofQueueSize)
	handler.aofFinished = make(chan struct{})
	go func() {
		handler.handleAof()
	}()
	return handler, nil
}

func (handler *AofHandler) AddAof(dbIndex int, cmdLine CommandLine) {
	if config.Properties.AppendOnly && handler.aofChan != nil {
		handler.aofChan <- &payLoad{
			cmdLine: cmdLine,
			dbIndex: dbIndex,
		}
	}
}
func (handler *AofHandler) handleAof() {
	handler.currentDb = 0
	for p := range handler.aofChan {
		handler.pausingAof.RLock() // prevent other goroutines from pausing aof
		if p.dbIndex != handler.currentDb {
			// select db
			data := reply.NewMultiBulkReply(utils.ToCmdLine("SELECT", strconv.Itoa(p.dbIndex))).ToBytes()
			_, err := handler.aofFile.Write(data)
			if err != nil {
				logger.Warn(err)
				continue // skip this command
			}
			handler.currentDb = p.dbIndex
		}
		data := reply.NewMultiBulkReply(p.cmdLine).ToBytes()
		_, err := handler.aofFile.Write(data)
		if err != nil {
			logger.Warn(err)
		}
		handler.pausingAof.RUnlock()
	}
	handler.aofFinished <- struct{}{}
}
func (handler *AofHandler) LoadAof(maxBytes int) {
	aofChan := handler.aofChan
	handler.aofChan = nil
	defer func(aofChan chan *payLoad) {
		handler.aofChan = aofChan
	}(aofChan)

	file, err := os.Open(handler.aofFileName)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			return
		}
		logger.Warn(err)
		return
	}
	defer file.Close()

	var reader io.Reader
	if maxBytes > 0 {
		reader = io.LimitReader(file, int64(maxBytes))
	} else {
		reader = file
	}
	ch := parse.ParserStream(reader)

	fakeConn := &connection.FakeConn{} // only used for save dbIndex
	for p := range ch {
		if p.Err != nil {
			if p.Err == io.EOF {
				break
			}
			logger.Error("parse error: " + p.Err.Error())
			continue
		}
		if p.Data == nil {
			logger.Error("empty payload")
			continue
		}
		r, ok := p.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("require multi bulk reply")
			continue
		}
		ret := handler.database.Exec(fakeConn, r.Args)
		if reply.IsErrorReply(ret) {
			logger.Error("exec err", ret.ToBytes())
		}
	}
}
