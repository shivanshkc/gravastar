package connor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	defaultMaxConnections = 1000
	defaultPingInterval   = 30 * time.Second
)

// ErrMaxConnectionsReached means that more websocket connections cannot be accepted.
var ErrMaxConnectionsReached = errors.New("max connections reached")

// Manager for websocket connections.
type Manager struct {
	connections      map[*websocket.Conn]struct{}
	connectionsMutex *sync.RWMutex
	maxConnections   int
	pingInterval     time.Duration
	logger           *slog.Logger
}

// NewManager returns a new Manager instance.
//
// maxConn is the max number of connections that Manager will allow. Default is 1000.
//
// pingInterval is the rate at which the connection will be sent ping messages.
// Manager will also use this as a timeout to identify stale connections. If a connection does not reply with a pong
// message within twice the pingInterval, the connection will be terminated. Default is 30 seconds.
//
// logger is used for internal logging. Provide nil for no logging.
func NewManager(maxConn int, pingInterval time.Duration, logger *slog.Logger) *Manager {
	if maxConn <= 0 {
		maxConn = defaultMaxConnections
	}

	if pingInterval <= 0 {
		pingInterval = defaultPingInterval
	}

	// Noop logger.
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	logger.Info("connor.Manager config", "maxConnections", maxConn, "pingInterval", pingInterval)

	return &Manager{
		connections:      map[*websocket.Conn]struct{}{},
		connectionsMutex: &sync.RWMutex{},
		maxConnections:   maxConn,
		pingInterval:     pingInterval,
		logger:           logger,
	}
}

// Size of the connection map, or number of connections currently active.
func (m *Manager) Size() int {
	m.connectionsMutex.RLock()
	defer m.connectionsMutex.RUnlock()
	return len(m.connections)
}

// Register a new connection. It also sets up auto-cleanup of the connection.
func (m *Manager) Register(ctx context.Context, conn *websocket.Conn) error {
	if err := m.addConnectionSafe(conn); err != nil {
		return err
	}

	// Setup auto-disconnect as per given ping interval ===========================================
	readTimeout := m.pingInterval * 2

	if err := conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		m.logger.ErrorContext(ctx, "failed to set the very first read deadline, deleting connection",
			"remoteAddr", conn.RemoteAddr(), "error", err)
		m.deleteConnectionSafe(conn)
		return fmt.Errorf("could not set read deadline: %w", err)
	}

	conn.SetPongHandler(func(message string) error {
		m.logger.Debug("pong message received", "remoteAddr", conn.RemoteAddr())

		if err := conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
			m.logger.ErrorContext(ctx, "failed to set the next read deadline, deleting connection",
				"remoteAddr", conn.RemoteAddr(), "error", err)
			m.deleteConnectionSafe(conn)
		}

		return nil
	})

	// Setup periodic ping ========================================================================
	go func() {
		// Delete connection upon ping error.
		defer m.deleteConnectionSafe(conn)

		// Ticker for periodic ping.
		ticker := time.NewTicker(m.pingInterval)
		defer ticker.Stop()

		for {
			if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				m.logger.Error("failed to send ping message, deleting connection",
					"remoteAddr", conn.RemoteAddr(), "error", err)
				return
			}

			m.logger.Debug("ping message sent", "remoteAddr", conn.RemoteAddr())
			<-ticker.C
		}
	}()

	// Setup read loop. PongHandler won't work without it =========================================
	go func() {
		defer m.deleteConnectionSafe(conn)

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				m.logger.Error("failed to read message, deleting connection",
					"remoteAddr", conn.RemoteAddr(), "error", err)
				return
			}
		}
	}()

	return nil
}

// Broadcast a message to all active connections.
func (m *Manager) Broadcast(ctx context.Context, messageType int, data []byte) error {
	// Respect context deadline if given.
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(m.pingInterval)
		m.logger.ErrorContext(ctx, "context deadline not set, using ping interval as deadline",
			"deadline", deadline)
	} else {
		m.logger.InfoContext(ctx, "context deadline set", "deadline", deadline)
	}

	// Use PreparedMessage for efficiency.
	preparedMessage, err := websocket.NewPreparedMessage(messageType, data)
	if err != nil {
		return fmt.Errorf("failed to create prepared message: %w", err)
	}

	// Wait Group for concurrent message sending.
	wg := &sync.WaitGroup{}

	// Read lock.
	m.connectionsMutex.RLock()
	defer m.connectionsMutex.RUnlock()

	wg.Add(len(m.connections))

	for conn := range m.connections {
		// Capture loop var.
		conn := conn

		// Write without blocking.
		go func() {
			defer wg.Done()

			defer func() {
				// Reset write deadline. Otherwise, ping would fail.
				if err := conn.SetWriteDeadline(time.Time{}); err != nil {
					m.logger.Error("failed to reset write deadline", "remoteAddr", conn.RemoteAddr(), "error", err)
					return
				}
			}()

			if err := conn.SetWriteDeadline(deadline); err != nil {
				m.logger.Error("failed to set write deadline", "remoteAddr", conn.RemoteAddr(), "error", err)
				return
			}

			if err := conn.WritePreparedMessage(preparedMessage); err != nil {
				m.logger.Error("failed to write message", "remoteAddr", conn.RemoteAddr(), "error", err)
				return
			}
		}()
	}

	wg.Wait()
	return nil
}

func (m *Manager) addConnectionSafe(conn *websocket.Conn) error {
	m.connectionsMutex.Lock()
	defer m.connectionsMutex.Unlock()

	// Never exceed the connection limit.
	if len(m.connections) >= m.maxConnections {
		m.connectionsMutex.Unlock()
		return ErrMaxConnectionsReached
	}

	// Persist connection.
	m.connections[conn] = struct{}{}
	return nil
}

func (m *Manager) deleteConnectionSafe(conn *websocket.Conn) {
	m.connectionsMutex.Lock()
	defer m.connectionsMutex.Unlock()

	if conn == nil {
		return
	}

	// Cleanup.
	_ = conn.Close()
	delete(m.connections, conn)

	m.logger.Info("connection deleted", "totalConnections", len(m.connections))
}
