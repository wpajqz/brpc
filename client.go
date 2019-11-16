package brpc

import (
	"sync"
	"time"

	"github.com/silenceper/pool"
	"github.com/wpajqz/go-sdk/export"
)

const (
	maxPayload = 10 * 1024 * 1024
)

var (
	defaultClient *Client
	once                = &sync.Once{}
	interval      int64 = 60
)

type (
	Client struct {
		initialCap      int
		maxCap          int
		idleTimeout     time.Duration
		clientPool      pool.Pool
		onOpen, onClose func()
		onError         func(error)
	}

	Options struct {
		InitialCap      int
		MaxCap          int
		IdleTimeout     time.Duration
		OnOpen, OnClose func()
		OnError         func(error)
	}
)

func NewClient(server string, port int, options Options) (*Client, error) {
	var err error
	once.Do(func() {
		defaultClient = &Client{
			initialCap:  options.InitialCap,
			maxCap:      options.MaxCap,
			idleTimeout: options.IdleTimeout,
			onOpen:      options.OnOpen,
			onClose:     options.OnClose,
			onError:     options.OnError,
		}

		var p pool.Pool
		p, err = defaultClient.newExportPool(server, port)
		if err != nil {
			return
		}

		defaultClient.clientPool = p
	})

	return defaultClient, err
}

func Session() (*export.Client, error) {
	return defaultClient.Session()
}

func (c *Client) Session() (*export.Client, error) {
	v, err := c.clientPool.Get()
	if err != nil {
		return nil, err
	}
	defer c.clientPool.Put(v)

	return v.(*export.Client), nil
}
