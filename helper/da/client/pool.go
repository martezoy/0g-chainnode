package client

import (
	"errors"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
)

type ConnectionPool interface {
	GetConnection() (*grpc.ClientConn, error)
	ReleaseConnection(*grpc.ClientConn)
	Close()
}

type connectionPoolImpl struct {
	address string
	maxSize int
	timeout time.Duration
	param   grpc.ConnectParams

	mu   sync.Mutex
	pool []*grpc.ClientConn
}

func NewConnectionPool(address string, maxSize int, timeout time.Duration) ConnectionPool {
	return &connectionPoolImpl{
		address: address,
		maxSize: maxSize,
		timeout: timeout,
		param: grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  1.0 * time.Second,
				Multiplier: 1.5,
				Jitter:     0.2,
				MaxDelay:   30 * time.Second,
			},
			MinConnectTimeout: 30 * time.Second,
		},
		pool: make([]*grpc.ClientConn, 0, maxSize),
	}
}

func (p *connectionPoolImpl) GetConnection() (*grpc.ClientConn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pool == nil {
		return nil, errors.New("connection pool is closed")
	}

	// Check if there's any available connection in the pool
	if len(p.pool) > 0 {
		conn := p.pool[0]
		p.pool = p.pool[1:]
		return conn, nil
	}

	// If the pool is empty, create a new connection
	conn, err := grpc.Dial(p.address, grpc.WithBlock(),
		grpc.WithConnectParams(p.param),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (p *connectionPoolImpl) ReleaseConnection(conn *grpc.ClientConn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pool != nil {
		// If the pool is full, close the connection
		if len(p.pool) >= p.maxSize {
			conn.Close()
			return
		}

		// Add the connection back to the pool
		p.pool = append(p.pool, conn)
	} else {
		conn.Close()
	}
}

func (p *connectionPoolImpl) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pool != nil {
		for _, conn := range p.pool {
			conn.Close()
		}

		p.pool = nil
	}
}
