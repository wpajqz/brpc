package brpc

import (
	"time"

	"github.com/silenceper/pool"
	"github.com/wpajqz/brpc/export"
)

var (
	defaultClient *Client
	interval      int64 = 60
)

type (
	Client struct {
		maxPayload      int
		initialCap      int
		maxCap          int
		idleTimeout     time.Duration
		clientPool      pool.Pool
		onOpen, onClose func()
		onError         func(error)
	}
)

func NewClient(server string, port int, opts ...Option) (*Client, error) {
	options := options{
		maxPayload: 10 * 1024 * 1024,
		initialCap: 10,
		maxCap:     30,
	}

	for _, o := range opts {
		o.apply(&options)
	}

	defaultClient = &Client{
		maxPayload:  options.maxPayload,
		initialCap:  options.initialCap,
		maxCap:      options.maxCap,
		idleTimeout: options.idleTimeout,
		onOpen:      options.onOpen,
		onClose:     options.onClose,
		onError:     options.onError,
	}

	p, err := defaultClient.newExportPool(server, port)
	if err != nil {
		return nil, err
	}

	defaultClient.clientPool = p

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
