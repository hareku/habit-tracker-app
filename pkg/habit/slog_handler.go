package habit

import (
	"context"
	"log/slog"

	"github.com/awslabs/aws-lambda-go-api-proxy/core"
)

// ContextValueLogHandler is a slog.Handler that adds the values from the context to the log record.
type ContextValueLogHandler struct {
	slog.Handler
}

func NewContextValueLogHandler(h slog.Handler) *ContextValueLogHandler {
	return &ContextValueLogHandler{Handler: h}
}

func (h *ContextValueLogHandler) Handle(ctx context.Context, r slog.Record) error {
	if v, ok := core.GetAPIGatewayContextFromContext(ctx); ok {
		r.AddAttrs(slog.Group("api_gateway"),
			slog.String("request_id", v.RequestID),
		)
	}
	return h.Handler.Handle(ctx, r)
}
