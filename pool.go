package ldappool

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-ldap/ldap/v3"
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
	Index int
}

type Pool struct {
	connectionTimeout time.Duration
	wakeupInterval    time.Duration

	mu sync.Mutex

	addr            string
	bindCredentials *BindCredentials
	opts            []ldap.DialOpt

	conns []*PoolConnection

	availableConn chan *PoolConnection
	deadConn      chan *PoolConnection
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
	p.mu.Lock()
	pc.State = PoolConnectionAvailable
	p.mu.Unlock()

	p.availableConn <- pc
}

func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// errs := []error{}
	for _, c := range p.conns {
		if c == nil {
			continue
		}
		log.Printf("closing connection %d", c.Index)
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
	p.availableConn <- pc

	return pc, nil
}

func (p *Pool) init() error {
	p.mu.Lock()
	var err error

	for i := 0; i < connsCount; i++ {
		_, err = p.newConn(i)
		if err != nil {
			break
		}
	}

	if err != nil {
		p.mu.Unlock()
		return fmt.Errorf("Cannot open required connections (%+v)", err)
	}

	p.mu.Unlock()
	return nil
}

func (p *Pool) heartbeat(c *PoolConnection) error {
	closing := c.Conn.IsClosing()
	if closing {
		return fmt.Errorf("connection is closed or being closed")
	}

	_, err := c.Conn.Search(&ldap.SearchRequest{BaseDN: "", Scope: ldap.ScopeBaseObject, Filter: "(&)", Attributes: []string{"1.1"}})
	if err != nil {
		return fmt.Errorf("cannot heartbeat")
	}

	return nil
}

func (p *Pool) watcher(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case pc := <-p.deadConn:
			go func() {
				for {
					p.mu.Lock()
					if pc.State != PoolConnectionUnavailable {
						return
					}

					p.newConn(pc.Index)
					p.mu.Unlock()
					time.Sleep(p.wakeupInterval)
				}
			}()
		}
	}
}

func (p *Pool) pull() (*PoolConnection, error) {
	var pc *PoolConnection
waiting:
	select {
	case pc = <-p.availableConn:
	case <-time.After(p.connectionTimeout):
		return nil, fmt.Errorf("timeout while pulling connection")
	}

	err := p.heartbeat(pc)
	if err != nil {
		// Connection is probably dead, mark it has unavailable
		p.mu.Lock()
		pc.State = PoolConnectionUnavailable
		p.mu.Unlock()
		p.deadConn <- pc
		goto waiting
	}

	p.mu.Lock()
	pc.State = PoolConnectionBusy
	p.mu.Unlock()

	return pc, nil
}

type PoolOptions struct {
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
		addr:              po.URL,
		conns:             make([]*PoolConnection, connectionsCount),
		bindCredentials:   po.BindCredentials,
		availableConn:     make(chan *PoolConnection, connectionsCount),
		deadConn:          make(chan *PoolConnection),
		connectionTimeout: po.ConnectionTimeout,
		wakeupInterval:    po.WakeupInterval,
	}

	if pool.connectionTimeout == 0 {
		pool.connectionTimeout = 10 * time.Second
	}

	if pool.wakeupInterval == 0 {
		pool.wakeupInterval = 5 * time.Second
	}

	err := pool.init()
	if err != nil {
		return nil, err
	}

	go pool.watcher(ctx)

	log.Printf("LDAP pool initialized with %d connections. Wakeup interval set to %s, ConnectionTimeout set to %s.", connectionsCount, pool.wakeupInterval, pool.connectionTimeout)
	return pool, nil
}
