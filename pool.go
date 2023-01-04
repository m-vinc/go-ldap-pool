package ldappool

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/go-ldap/ldap/v3"
)

type Pool struct {
	debug bool
	// connectionTimeout time.Duration
	// wakeupInterval    time.Duration

	addr            string
	bindCredentials *BindCredentials
	opts            []ldap.DialOpt

	conns chan *ldap.Conn
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

func (p *Pool) release(conn *ldap.Conn) {
	p.conns <- conn
}

func (p *Pool) pull(ctx context.Context) (*ldap.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case conn := <-p.conns:
		if conn.IsClosing() {
			log.Println("connection closed, trying to re-open one connection")
			for {
				c, err := p.open()
				if err != nil {
					time.Sleep(5 * time.Second)
					continue
				}
				return c, nil
			}
		}
		return conn, nil
	}
}

func (p *Pool) Close() error {
	l := len(p.conns)
	for i := 0; i < l; i++ {
		c := <-p.conns
		if p.debug {
			log.Printf("closing %v", c)
		}
		c.Close()
	}
	return nil
}

func (p *Pool) init() error {
	errs := []error{}
	for i := 0; i < cap(p.conns); i++ {
		c, err := p.open()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		p.conns <- c
	}

	if len(errs) == cap(p.conns) {
		if p.debug {
			log.Printf("%+v", errs)
		}
		return errors.New("unable to initialize at most one connection")
	}

	return nil
}

type PoolOptions struct {
	Debug           bool
	URL             string
	BindCredentials *BindCredentials

	ConnectionsCount int
}

type BindCredentials struct {
	Username string
	Password string
}

func NewPool(po *PoolOptions) (*Pool, error) {
	connectionsCount := 5
	if po.ConnectionsCount != 0 {
		connectionsCount = po.ConnectionsCount
	}

	pool := &Pool{
		debug:           po.Debug,
		addr:            po.URL,
		conns:           make(chan *ldap.Conn, 5),
		bindCredentials: po.BindCredentials,
	}

	err := pool.init()
	if err != nil {
		return nil, err
	}

	if pool.debug {
		log.Printf("LDAP pool initialized with %d connections. Wakeup interval set to %s, ConnectionTimeout set to %s.", connectionsCount)
	}

	return pool, nil
}
