package brpc

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/silenceper/pool"
	"github.com/wpajqz/brpc/export"
	"github.com/wpajqz/linker"
	"github.com/wpajqz/linker/plugins"
)

var interval int64 = 60

func (c *Client) newExportPool(server string, port int) (pool.Pool, error) {
	// ping请求的回调，出错的时候调用
	cb := RequestStatusCallback{
		Error: func(code int, msg string) {
			log.Printf("brpc error: %s\n", msg)
		},
	}

	// factory 创建连接的方法
	factory := func() (interface{}, error) {
		exportClient, err := export.NewClient(server, port, &ReadyStateCallback{Open: c.onOpen, Close: c.onClose, Error: func(err string) { c.onError(errors.New(err)) }})
		if err != nil {
			return nil, fmt.Errorf("brpc error: %s\n", err.Error())
		}

		exportClient.SetMaxPayload(c.maxPayload)
		exportClient.SetPluginForPacketSender([]linker.PacketPlugin{
			&plugins.Encryption{},
		})
		exportClient.SetPluginForPacketReceiver([]linker.PacketPlugin{
			&plugins.Decryption{},
		})

		go func(ec *export.Client) {
			ticker := time.NewTicker(time.Duration(interval) * time.Second)
			for {
				select {
				case <-ticker.C:
					err := ec.Ping(nil, cb)
					if err != nil {
						return
					}
				}
			}
		}(exportClient)

		return exportClient, nil
	}

	close := func(v interface{}) error {
		return v.(*export.Client).Close()
	}

	ping := func(v interface{}) error {
		return v.(*export.Client).Ping(nil, cb)
	}

	pc := &pool.Config{
		InitialCap: c.initialCap,
		MaxCap:     c.maxCap,
		Factory:    factory,
		Close:      close,
		Ping:       ping,
		//连接最大空闲时间，超过该时间的连接 将会关闭，可避免空闲时连接EOF，自动失效的问题
		IdleTimeout: c.idleTimeout,
	}

	return pool.NewChannelPool(pc)
}
