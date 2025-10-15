package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/shivanshkc/gravastar/internal/utils/errutils"
	"github.com/shivanshkc/gravastar/internal/utils/httputils"
	"github.com/shivanshkc/gravastar/pkg/physics"
)

// Handler encapsulates all REST handlers.
type Handler struct {
	Engine physics.GravityEngine
}

func (h *Handler) CreateDot(w http.ResponseWriter, r *http.Request) {
	// TODO: Rate limiting.
	// TODO: Only allow use from the browser.
	// TODO: Validate color input.

	var body struct {
		ID       string       `json:"id"`
		Position physics.Vec3 `json:"position"`
		Color    physics.Vec3 `json:"color"`
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
	width, height := h.Engine.Size()
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
		Radius:   3,
		Position: physics.Vec3{X: body.Position.X, Y: body.Position.Y, Z: 0},
		Velocity: physics.Vec3{},
		Color:    body.Color,
	}

	if err := h.Engine.AddDot(dot); err != nil {
		slog.ErrorContext(r.Context(), "error adding dot", "error", err)

		if errors.Is(err, physics.ErrDotAlreadyExists) {
			httputils.WriteErr(w, errutils.Conflict().WithReasonStr("dot already exists"))
		} else {
			httputils.WriteErr(w, err)
		}
		return
	}

	httputils.Write(w, http.StatusNoContent, nil, nil)
}

func (h *Handler) ListDots(w http.ResponseWriter, r *http.Request) {
	dots := h.Engine.Read()
	httputils.Write(w, http.StatusOK, nil, dots)
}
