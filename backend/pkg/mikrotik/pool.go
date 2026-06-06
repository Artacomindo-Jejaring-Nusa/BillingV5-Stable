package mikrotik

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/go-routeros/routeros"
)

type CircuitState string

const (
	CircuitStateClosed   CircuitState = "closed"
	CircuitStateOpen     CircuitState = "open"
	CircuitStateHalfOpen CircuitState = "half_open"
)

type CircuitBreakerConfig struct {
	FailureThreshold int
	RecoveryTimeout  time.Duration
}

type ConnectionMetrics struct {
	TotalConnections    int64
	ActiveConnections   int64
	FailedConnections   int64
	RetryAttempts       int64
	CircuitBreakerTrips int64
}

type ConnectionHealth struct {
	Healthy             bool
	LastCheck           time.Time
	ConsecutiveFailures int
	TotalRequests       int64
	SuccessfulRequests  int64
}

type ServerPool struct {
	ServerKey        string
	Host             string
	Port             int
	Username         string
	Password         string
	Connections      []*PooledConnection
	CircuitState     CircuitState
	CircuitFailures  int
	LastFailureTime  time.Time
	ConnectionHealth ConnectionHealth
}

type PooledConnection struct {
	Client   *routeros.Client
	LastUsed time.Time
}

type MikrotikConnectionPool struct {
	mu             sync.Mutex
	maxConnections int
	timeout        time.Duration
	idleTimeout    time.Duration
	pools          map[string]*ServerPool
	cbConfig       CircuitBreakerConfig
	metrics        ConnectionMetrics
}

func NewMikrotikConnectionPool(maxConnections int, timeout, idleTimeout time.Duration) *MikrotikConnectionPool {
	return &MikrotikConnectionPool{
		maxConnections: maxConnections,
		timeout:        timeout,
		idleTimeout:    idleTimeout,
		pools:          make(map[string]*ServerPool),
		cbConfig: CircuitBreakerConfig{
			FailureThreshold: 3,
			RecoveryTimeout:  10 * time.Second,
		},
	}
}

// Global instance
var GlobalPool = NewMikrotikConnectionPool(15, 15*time.Second, 60*time.Second)

func (p *MikrotikConnectionPool) getPoolKey(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

func (p *MikrotikConnectionPool) initializeServerPool(key string, host string, port int, username, password string) *ServerPool {
	sp, exists := p.pools[key]
	if !exists {
		sp = &ServerPool{
			ServerKey:    key,
			Host:         host,
			Port:         port,
			Username:     username,
			Password:     password,
			CircuitState: CircuitStateClosed,
			ConnectionHealth: ConnectionHealth{
				Healthy: true,
			},
		}
		p.pools[key] = sp
	}
	return sp
}

func (p *MikrotikConnectionPool) checkCircuitBreaker(sp *ServerPool) bool {
	now := time.Now()
	if sp.CircuitState == CircuitStateClosed {
		return true
	}

	if sp.CircuitState == CircuitStateOpen {
		if now.Sub(sp.LastFailureTime) > p.cbConfig.RecoveryTimeout {
			sp.CircuitState = CircuitStateHalfOpen
			sp.CircuitFailures = 0
			return true
		}
		return false
	}

	// HALF_OPEN
	return true
}

func (p *MikrotikConnectionPool) recordSuccess(sp *ServerPool) {
	if sp.CircuitState == CircuitStateHalfOpen {
		sp.CircuitState = CircuitStateClosed
	}
	sp.CircuitFailures = 0
	sp.ConnectionHealth.Healthy = true
	sp.ConnectionHealth.ConsecutiveFailures = 0
	sp.ConnectionHealth.SuccessfulRequests++
}

func (p *MikrotikConnectionPool) recordFailure(sp *ServerPool, err error) {
	sp.CircuitFailures++
	sp.LastFailureTime = time.Now()
	sp.ConnectionHealth.Healthy = false
	sp.ConnectionHealth.ConsecutiveFailures++
	p.metrics.FailedConnections++

	if sp.CircuitState == CircuitStateClosed && sp.CircuitFailures >= p.cbConfig.FailureThreshold {
		sp.CircuitState = CircuitStateOpen
		p.metrics.CircuitBreakerTrips++
	} else if sp.CircuitState == CircuitStateHalfOpen {
		sp.CircuitState = CircuitStateOpen
	}
}

// dialConnection establishes a TCP connection with timeout and logs in
func (p *MikrotikConnectionPool) dialConnection(host string, port int, username, password string) (*routeros.Client, net.Conn, error) {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, p.timeout)
	if err != nil {
		return nil, nil, err
	}

	// Set deadline for login
	_ = conn.SetDeadline(time.Now().Add(p.timeout))

	client, err := routeros.NewClient(conn)
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}

	err = client.Login(username, password)
	if err != nil {
		client.Close()
		return nil, nil, err
	}

	// Reset deadline
	_ = conn.SetDeadline(time.Time{})
	return client, conn, nil
}

func (p *MikrotikConnectionPool) GetConnection(host string, port int, username, password string) (*routeros.Client, error) {
	key := p.getPoolKey(host, port)

	p.mu.Lock()
	sp := p.initializeServerPool(key, host, port, username, password)

	// Check Circuit Breaker
	if !p.checkCircuitBreaker(sp) {
		p.mu.Unlock()
		return nil, fmt.Errorf("service unavailable for server %s - circuit breaker is open", key)
	}

	p.metrics.TotalConnections++
	sp.ConnectionHealth.TotalRequests++

	// Clean up expired connections from this pool
	now := time.Now()
	var healthyConns []*PooledConnection
	for _, pc := range sp.Connections {
		if now.Sub(pc.LastUsed) > p.idleTimeout {
			// Close expired connection
			pc.Client.Close()
		} else {
			healthyConns = append(healthyConns, pc)
		}
	}
	sp.Connections = healthyConns

	// Try to reuse an existing connection (with a quick health check)
	if len(sp.Connections) > 0 {
		// Take the last connection (LIFO)
		pc := sp.Connections[len(sp.Connections)-1]
		sp.Connections = sp.Connections[:len(sp.Connections)-1]

		p.mu.Unlock()

		// Test connection health
		_, err := pc.Client.Run("/system/identity/print")
		if err == nil {
			p.mu.Lock()
			p.recordSuccess(sp)
			p.metrics.ActiveConnections++
			p.mu.Unlock()
			return pc.Client, nil
		}

		// Connection is dead, close it and record failure
		pc.Client.Close()
		p.mu.Lock()
		p.recordFailure(sp, err)
		p.mu.Unlock()

		// Try to establish a new one instead of failing
		p.mu.Lock()
	}

	// Establish a new connection
	p.mu.Unlock()
	client, _, err := p.dialConnection(host, port, username, password)
	if err != nil {
		p.mu.Lock()
		p.recordFailure(sp, err)
		p.mu.Unlock()
		return nil, err
	}

	p.mu.Lock()
	p.recordSuccess(sp)
	p.metrics.ActiveConnections++
	p.mu.Unlock()

	return client, nil
}

func (p *MikrotikConnectionPool) ReturnConnection(client *routeros.Client, host string, port int) {
	if client == nil {
		return
	}

	key := p.getPoolKey(host, port)

	p.mu.Lock()
	defer p.mu.Unlock()

	p.metrics.ActiveConnections = max(0, p.metrics.ActiveConnections-1)

	sp, exists := p.pools[key]
	if !exists {
		// No pool exists, just close it
		client.Close()
		return
	}

	// If the pool is not full, return it to the pool
	if len(sp.Connections) < p.maxConnections {
		sp.Connections = append(sp.Connections, &PooledConnection{
			Client:   client,
			LastUsed: time.Now(),
		})
	} else {
		// Otherwise, close it
		client.Close()
	}
}

func (p *MikrotikConnectionPool) ExecuteWithRetry(host string, port int, username, password string, operation func(*routeros.Client) (interface{}, error), maxRetries int, retryDelay time.Duration) (interface{}, error) {
	key := p.getPoolKey(host, port)
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		p.mu.Lock()
		sp := p.pools[key]
		if sp != nil && sp.CircuitState == CircuitStateOpen {
			now := time.Now()
			if now.Sub(sp.LastFailureTime) <= p.cbConfig.RecoveryTimeout {
				p.mu.Unlock()
				return nil, fmt.Errorf("circuit breaker open for %s, skipping attempt %d", key, attempt+1)
			}
		}
		p.mu.Unlock()

		if attempt > 0 {
			p.mu.Lock()
			p.metrics.RetryAttempts++
			p.mu.Unlock()
		}

		client, err := p.GetConnection(host, port, username, password)
		if err != nil {
			lastErr = err
			// Check if we should wait before retry
			if attempt < maxRetries-1 {
				// Exponential backoff with jitter
				backoff := retryDelay * (1 << attempt)
				jitter := time.Duration(rand.Float64() * float64(backoff) * 0.1)
				time.Sleep(backoff + jitter)
			}
			continue
		}

		result, opErr := operation(client)
		p.ReturnConnection(client, host, port)

		if opErr != nil {
			lastErr = opErr
			p.mu.Lock()
			sp = p.pools[key]
			if sp != nil {
				p.recordFailure(sp, opErr)
			}
			p.mu.Unlock()

			if attempt < maxRetries-1 {
				backoff := retryDelay * (1 << attempt)
				jitter := time.Duration(rand.Float64() * float64(backoff) * 0.1)
				time.Sleep(backoff + jitter)
			}
			continue
		}

		return result, nil
	}

	return nil, lastErr
}

func (p *MikrotikConnectionPool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, sp := range p.pools {
		for _, pc := range sp.Connections {
			pc.Client.Close()
		}
		sp.Connections = nil
	}
	p.metrics.ActiveConnections = 0
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
