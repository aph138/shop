package logger

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type logger struct {
	logger *slog.Logger
}

func NewLogger(l *slog.Logger) *logger {
	return &logger{
		logger: l,
	}
}
func (l *logger) UnaryServerLogger(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	s := time.Now()
	res, err := handler(ctx, req)
	duration := time.Since(s)
	var msg string
	if err != nil {
		msg = fmt.Sprintf("method:%s ,duration:%s ,err:%s ,code:%s",
			info.FullMethod,
			duration.String(),
			err.Error(),
			status.Code(err))
	} else {
		msg = fmt.Sprintf("method:%s ,duration:%s", info.FullMethod, duration.String())
	}
	l.logger.Info(msg)
	return res, err
}
