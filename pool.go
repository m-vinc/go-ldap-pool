package ldappool

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/m-vinc/ldap/v3"
)

const (
	connsCount = 10
)

type PoolConnectionState int

const (
	PoolConnectionAvailable PoolConnectionState = iota
	PoolConnectionBusy
	PoolConnectionUnavailable
)

type PoolConnection struct {
	*ldap.Conn
	State PoolConnectionState

	mx    sync.Mutex
	Index int
}

func (pc *PoolConnection) SetState(state PoolConnectionState) {
	pc.mx.Lock()
	defer pc.mx.Unlock()

	pc.State = state
}

type Pool struct {
	debug             bool
	connectionTimeout time.Duration

	addr            string
	bindCredentials *BindCredentials
	opts            []ldap.DialOpt

	conns     []*PoolConnection
	connsChan chan *PoolConnection
}

func (p *Pool) open() (*ldap.Conn, error) {
	conn, err := ldap.DialURL(p.addr, p.opts...)
	if err != nil {
		return nil, err
	}

	if p.bindCredentials != nil {
		err = conn.Bind(p.bindCredentials.Username, p.bindCredentials.Password)
		if err != nil {
			return nil, err
		}
	}

	return conn, nil
}

func (p *Pool) release(pc *PoolConnection) {

	if !pc.IsClosing() {
		p.connsChan <- pc
	}
}

func (p *Pool) Close() error {

	// errs := []error{}
	for _, c := range p.conns {
		if c == nil {
			continue
		}
		if p.debug {
			log.Printf("closing connection %d", c.Index)
		}
		c.Close()
	}

	return nil
}

func (p *Pool) newConn(i int) (*PoolConnection, error) {
	conn, err := p.open()
	if err != nil {
		return nil, err
	}

	pc := &PoolConnection{
		Conn:  conn,
		Index: i,
	}

	p.conns[i] = pc
	p.connsChan <- pc

	if p.debug {
		log.Printf("initializing working connection at index %d", i)
	}

	return pc, nil
}

func (p *Pool) heartbeat(c *PoolConnection) error {
	closing := c.Conn.IsClosing()
	if closing {
		if p.debug {
			log.Println("")
		}
		return fmt.Errorf("connection is closed or being closed")
	}

	_, err := c.Conn.Search(&ldap.SearchRequest{BaseDN: "", Scope: ldap.ScopeBaseObject, Filter: "(&)", Attributes: []string{"1.1"}})
	if err != nil {
		if p.debug {
			log.Printf("error while heartbeating connection %d - %+v\n", c.Index, err)
		}
		return fmt.Errorf("cannot heartbeat")
	}

	return nil
}

func (p *Pool) pull() (*PoolConnection, error) {
	var pc *PoolConnection
	select {
	case pc = <-p.connsChan:
	case <-time.After(p.connectionTimeout):
		if p.debug {
			log.Printf("connection %d timeout\n", pc.Index)
		}
		return nil, fmt.Errorf("timeout while pulling connection")
	}

	return pc, nil
}

func (p *Pool) watcher(ctx context.Context) {
	for i := range p.conns {
		go func(i int) {
			for {
				var err error
				conn := p.conns[i]

				if conn == nil {
					p.newConn(i)
					goto sleep
				}

				if conn.State == PoolConnectionBusy {
					goto sleep
				}

				err = p.heartbeat(conn)
				if err != nil {
					p.newConn(i)
					goto sleep
				}

				conn.SetState(PoolConnectionAvailable)
			sleep:
				time.Sleep(time.Second * 1)
			}
		}(i)
	}
}

type PoolOptions struct {
	Debug           bool
	URL             string
	BindCredentials *BindCredentials

	ConnectionsCount  int
	ConnectionTimeout time.Duration
	WakeupInterval    time.Duration
}

type BindCredentials struct {
	Username string
	Password string
}

func NewPool(ctx context.Context, po *PoolOptions) (*Pool, error) {
	connectionsCount := connsCount
	if po.ConnectionsCount != 0 {
		connectionsCount = po.ConnectionsCount
	}

	pool := &Pool{
		debug:             po.Debug,
		addr:              po.URL,
		conns:             make([]*PoolConnection, connectionsCount),
		bindCredentials:   po.BindCredentials,
		connsChan:         make(chan *PoolConnection, connectionsCount),
		connectionTimeout: po.ConnectionTimeout,
	}

	if pool.connectionTimeout == 0 {
		pool.connectionTimeout = 10 * time.Second
	}

	go pool.watcher(ctx)

	if pool.debug {
		log.Printf("LDAP pool initialized with %d connections. ConnectionTimeout set to %s.", connectionsCount, pool.connectionTimeout)
	}
	return pool, nil
}
