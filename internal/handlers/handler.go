package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/shivanshkc/gravastar/internal/utils/errutils"
	"github.com/shivanshkc/gravastar/internal/utils/httputils"
	"github.com/shivanshkc/gravastar/pkg/connor"
	"github.com/shivanshkc/gravastar/pkg/physics"
)

// Handler encapsulates all REST handlers.
//
// TODO: Close all connections gracefully on interruption.
// TODO: Dot Eviction Control
type Handler struct {
	engine   physics.GravityEngine
	upgrader *websocket.Upgrader
	manager  *connor.Manager
}

// NewHandler returns a new instance of the Handler.
func NewHandler(engine physics.GravityEngine, upgrader *websocket.Upgrader, manager *connor.Manager) *Handler {
	return &Handler{
		engine:   engine,
		upgrader: upgrader,
		manager:  manager,
	}
}

func (h *Handler) CreateDot(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID       string       `json:"id"`
		Position physics.Vec3 `json:"position"`
	}

	if err := json.NewDecoder(io.LimitReader(r.Body, 1024)).Decode(&body); err != nil {
		slog.ErrorContext(r.Context(), "error decoding body", "error", err)
		httputils.WriteErr(w, errutils.BadRequest().WithReasonStr("failed to decode request"))
		return
	}

	// Make sure the ID is a valid UUID v4.
	parsedID, err := uuid.Parse(body.ID)
	if err != nil {
		slog.ErrorContext(r.Context(), "error parsing id", "error", err)
		httputils.WriteErr(w, errutils.BadRequest().WithReasonStr("bad dot ID"))
		return
	}

	// Validate position.x.
	width, height := h.engine.Size()
	if body.Position.X < 0 || body.Position.X > float64(width) {
		slog.ErrorContext(r.Context(), "position x value out of bounds", "position", body.Position)
		httputils.WriteErr(w, errutils.BadRequest().WithReasonStr("position out of bounds"))
		return
	}

	// Validate position.y.
	if body.Position.Y < 0 || body.Position.Y > float64(height) {
		slog.ErrorContext(r.Context(), "position y value out of bounds", "position", body.Position)
		httputils.WriteErr(w, errutils.BadRequest().WithReasonStr("position out of bounds"))
		return
	}

	dot := physics.Dot{
		ID:       parsedID.String(),
		Mass:     1,
		Radius:   5,
		Position: physics.Vec3{X: body.Position.X, Y: body.Position.Y, Z: 0},
		Velocity: physics.Vec3{},
		Color:    physics.Vec3{X: 1, Y: 1, Z: 1},
	}

	if err := h.engine.AddDot(dot); err != nil {
		slog.ErrorContext(r.Context(), "error adding dot", "error", err)

		if errors.Is(err, physics.ErrDotAlreadyExists) {
			httputils.WriteErr(w, errutils.Conflict().WithReasonStr("dot already exists"))
		} else {
			httputils.WriteErr(w, err)
		}
		return
	}

	// Broadcast dot creation without blocking.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		// Message payload.
		message := websocketMessage{Event: dotCreated, Data: dot}
		messageBytes, err := json.Marshal(message)
		if err != nil {
			slog.ErrorContext(ctx, "error marshalling websocket message", "error", err)
			return
		}

		// Send to all connections.
		if err := h.manager.Broadcast(ctx, websocket.TextMessage, messageBytes); err != nil {
			slog.ErrorContext(ctx, "error broadcasting message", "error", err)
			return
		}
	}()

	httputils.Write(w, http.StatusNoContent, nil, nil)
}

func (h *Handler) ListDots(w http.ResponseWriter, r *http.Request) {
	dots := h.engine.Read()
	httputils.Write(w, http.StatusOK, nil, dots)
}

func (h *Handler) GetConn(w http.ResponseWriter, r *http.Request) {
	// Upgrade to websocket connection.
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.ErrorContext(r.Context(), "error upgrading connection", "error", err)
		httputils.WriteErr(w, errutils.InternalServerError().WithReasonStr("error upgrading connection"))
		return
	}

	slog.InfoContext(r.Context(), "successfully upgraded connection", "remoteAddr", conn.RemoteAddr())

	// Register connection.
	if err := h.manager.Register(r.Context(), conn); err != nil {
		slog.ErrorContext(r.Context(), "error registering connection", "error", err)
		if conn != nil {
			_ = conn.Close()
		}
		return
	}

	slog.InfoContext(r.Context(), "successfully registered connection",
		"remoteAddr", conn.RemoteAddr(), "totalConnections", h.manager.Size())
}
