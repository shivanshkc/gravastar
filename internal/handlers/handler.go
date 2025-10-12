package handlers

import (
	"encoding/json"
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
	var body struct {
		Position struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
		} `json:"position"`
	}

	if err := json.NewDecoder(io.LimitReader(r.Body, 100)).Decode(&body); err != nil {
		slog.ErrorContext(r.Context(), "error decoding body", "error", err)
		httputils.WriteErr(w, errutils.BadRequest().WithReasonStr("failed to decode request"))
		return
	}

	dot := physics.Dot{
		ID:       uuid.NewString(),
		Mass:     1,
		Radius:   1,
		Position: physics.Vec3{X: body.Position.X, Y: body.Position.Y, Z: 0},
		Velocity: physics.Vec3{},
		Color:    physics.NewRandVec3(),
	}

	h.Engine.AddDot(dot)

	// Return the full list of dots in the response so the client can sync their simulation immediately.
	dots := h.Engine.Read()
	// Add the ID of the newly created dot in the headers so the client can easily identify it.
	httputils.Write(w, http.StatusOK, map[string]string{"X-Dot-Id": dot.ID}, dots)
}

func (h *Handler) ListDots(w http.ResponseWriter, r *http.Request) {
	dots := h.Engine.Read()
	httputils.Write(w, http.StatusOK, nil, dots)
}
