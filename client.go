package noeq

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"sync"
)

var (
	ErrNoAddrs      = errors.New("noeq: no addresses provided")
	ErrInvalidToken = errors.New("noeq: token > 255 bytes in length")
)

type Client struct {
	mu    sync.Mutex
	cn    net.Conn
	addrs []string
	token string
}

func New(token string, addrs ...string) (*Client, error) {
	if len(addrs) == 0 {
		return nil, ErrNoAddrs
	}

	if len(token) > 255 {
		return nil, ErrInvalidToken
	}

	return &Client{token: token, addrs: addrs}, nil
}

func (c *Client) connect() (err error) {
	n := rand.Intn(len(c.addrs))
	c.cn, err = net.Dial("tcp", c.addrs[n])
	if err != nil {
		return
	}

	return c.auth()
}

func (c *Client) auth() (err error) {
	if c.token != "" {
		l := fmt.Sprintf("\000%s", []byte{byte(len(c.token))})
		_, err = io.WriteString(c.cn, l+c.token)
		return
	}

	return
}

func (c *Client) Gen(n uint8) (ids []uint64, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	defer func() {
		if err != nil {
			c.cn = nil
		}
	}()

	if c.cn == nil {
		err = c.connect()
		if err != nil {
			return
		}
	}

	_, err = c.cn.Write([]byte{n})
	if err != nil {
		return
	}

	ids = make([]uint64, n)
	err = binary.Read(c.cn, binary.BigEndian, &ids)
	return
}

func (c *Client) GenOne() (uint64, error) {
	ids, err := c.Gen(1)
	if len(ids) == 0 {
		return 0, err
	}
	return ids[0], nil
}
