package database

import (
	"fmt"
	"go-redis/aof"
	config "go-redis/conf"
	"go-redis/interface/resp"
	"go-redis/lib/logger"
	"go-redis/resp/reply"
	"runtime/debug"
	"strconv"
	"strings"
)

type StandAloneDatabase struct {
	dbSet      []*DB
	aofHandler *aof.AofHandler
}

func NewStandaloneDatabase() *StandAloneDatabase {
	database := &StandAloneDatabase{}
	if config.Properties.Databases == 0 {
		config.Properties.Databases = 16
	}
	database.dbSet = make([]*DB, config.Properties.Databases)
	for i := range database.dbSet {
		singleDB := NewDB()
		singleDB.index = i
		database.dbSet[i] = singleDB
	}
	if config.Properties.AppendOnly {
		aofHandler, err := aof.NewAofHandler(database)
		if err != nil {
			panic(err)
		}
		database.aofHandler = aofHandler
		for _, db := range database.dbSet {
			// avoid closure
			singleDB := db
			singleDB.addAof = func(line CmdLine) {
				database.aofHandler.AddAof(singleDB.index, line)
			}
		}
	}

	return database
}

func (d *StandAloneDatabase) Exec(client resp.Connection, args [][]byte) (result resp.Reply) {

	defer func() {
		if err := recover(); err != nil {
			logger.Warn(fmt.Sprintf("error occurs: %v\n%s", err, string(debug.Stack())))
			result = &reply.UnknownErrReply{}
		}
	}()

	cmdName := strings.ToLower(string(args[0]))
	if cmdName == "select" {
		if len(args) != 2 {
			return reply.NewArgNumErrReply("select")
		}
		return execSelect(client, d, args[1:])
	}
	// normal commands
	dbIndex := client.GetDBIndex()
	if dbIndex >= len(d.dbSet) {
		return reply.NewErrReply("ERR DB index is out of range")
	}
	selectedDB := d.dbSet[dbIndex]
	return selectedDB.Exec(client, args)
}

func (d *StandAloneDatabase) Close() {

}

func (d *StandAloneDatabase) AfterClientClose(c resp.Connection) {

}
func execSelect(c resp.Connection, db *StandAloneDatabase, args [][]byte) resp.Reply {
	dbIndex, err := strconv.Atoi(string(args[0]))
	if err != nil {
		return reply.NewErrReply("ERR invalid DB index")
	}
	if dbIndex >= len(db.dbSet) {
		return reply.NewErrReply("ERR DB index is out of range")
	}
	c.SelectDB(dbIndex)
	return reply.NewOkReply()
}
