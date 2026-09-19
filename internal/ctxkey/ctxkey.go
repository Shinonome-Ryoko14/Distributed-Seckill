// Package ctxkey 定义 context 中跨层传递的关键键（trace_id、user_id）。
package ctxkey

import "context"

type key int

const (
	traceIDKey key = iota
	userIDKey
)

// WithTraceID 把 trace_id 注入 ctx，供跨层日志与下游调用透传。
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// TraceID 读取 ctx 中的 trace_id。
func TraceID(ctx context.Context) (string, bool) {
	traceID, ok := ctx.Value(traceIDKey).(string)
	return traceID, ok
}

// WithUserID 把已鉴权的用户 ID 注入 ctx，供 service 层使用。
func WithUserID(ctx context.Context, userID uint64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserID 读取 ctx 中已鉴权的用户 ID。
func UserID(ctx context.Context) (uint64, bool) {
	userID, ok := ctx.Value(userIDKey).(uint64)
	return userID, ok
}
