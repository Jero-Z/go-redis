package main

import (
	"fmt"
	config "go-redis/conf"
	"go-redis/lib/logger"
	"go-redis/resp/handler"
	"go-redis/tcp"
	"os"
)

const configFile string = "redis.conf"

var defaultProperties = &config.ServerProperties{
	Bind: "0.0.0.0",
	Port: 6379,
}

func fileExists(filename string) bool {
	stat, err := os.Stat(filename)
	return err == nil && stat.IsDir()
}


func main() {


	logger.Setup(&logger.Settings{
		Path:       "logs",
		Name:       "go-dis",
		Ext:        "log",
		TimeFormat: "2006-01-02",
	})
	if fileExists(configFile) {
		config.SetupConfig(configFile)
	} else {
		config.Properties = defaultProperties
	}

	err := tcp.ListenAndServeWithSignal(
		&tcp.Config{
			Address: fmt.Sprintf("%s:%d", config.Properties.Bind, config.Properties.Port),
		},
		handler.NewRespHandler())
	if err != nil {
		logger.Error(err)
	}
}
