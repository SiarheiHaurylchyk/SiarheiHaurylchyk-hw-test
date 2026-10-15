package internalgrpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type Logger interface {
	Info(msg string)
}

func loggingInterceptor(logger Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)

		logger.Info(fmt.Sprintf("%s [%s] GRPC %s grpc %d %d %q",
			clientIP(ctx),
			start.Format("02/Jan/2006:15:04:05 -0700"),
			info.FullMethod,
			int(status.Code(err)),
			time.Since(start).Microseconds(),
			userAgent(ctx),
		))
		return resp, err
	}
}

func clientIP(ctx context.Context) string {
	requestPeer, ok := peer.FromContext(ctx)
	if !ok || requestPeer.Addr == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(requestPeer.Addr.String())
	if err != nil {
		return requestPeer.Addr.String()
	}
	return host
}

func userAgent(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get("user-agent")
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
