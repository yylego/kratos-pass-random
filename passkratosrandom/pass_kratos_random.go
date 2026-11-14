// Package passkratosrandom: Probabilistic request blocking middleware with chaos testing support
// Provides random pass-through rate with configurable odds
// Good fit in chaos engineering and fault injection, to check how clients handle failures
// Enables blocking specific request ratio within designated route scope
//
// passkratosrandom: 概率性请求阻断中间件，支持混沌测试
// 提供可配置概率的随机放行率控制
// 适用于混沌工程和故障注入，检验调用方怎么应对失败
// 可在指定路由范围内阻断一定百分比的请求
package passkratosrandom

import (
	"context"
	"log/slog"
	"math/rand/v2"

	"github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/middleware/selector"
	"github.com/yylego/kratos-auth/authkratos"
)

// Config holds the random pass configuration: the route scope and the share of requests to let through
//
// Config 保存随机放行的配置：路由范围和要放行的请求比例
type Config struct {
	routeScope *authkratos.RouteScope
	passRate   float64
}

// NewConfig creates a random pass config
// passRate is the share of requests in the scope to let through: 0.8 lets 80% through, 1 (and above) lets each one through, 0 (and below) blocks each one
//
// NewConfig 创建随机放行配置
// passRate 是范围内要放行的请求比例：0.8 放行 80%，大于等于 1 全部放行，小于等于 0 全部拦下
func NewConfig(routeScope *authkratos.RouteScope, passRate float64) *Config {
	return &Config{
		routeScope: routeScope,
		passRate:   passRate,
	}
}

// NewMiddleware creates middleware that blocks a share of requests in the scope with 503 RANDOM_RATE_UNAVAILABLE
// Requests that pass, and requests outside the scope, go on to the next handle func
// Each pass and block outcome goes to applog at debug level, so the level of applog decides if it shows
//
// NewMiddleware 创建中间件：范围内按比例拦下一部分请求，返回 503 RANDOM_RATE_UNAVAILABLE
// 放行的请求和范围外的请求照常交给处理函数
// 每次放行和拦截结果以 debug 级别写进 applog，显示与否由 applog 的日志级别决定
func NewMiddleware(cfg *Config, applog *slog.Logger) middleware.Middleware {
	applog.Info(
		"pass-kratos-random: new middleware",
		"side", cfg.routeScope.Side,
		"operations", len(cfg.routeScope.OperationSet),
		"pass-rate", cfg.passRate,
	)
	applog.Debug("pass-kratos-random: new middleware route-scope", "route-scope", cfg.routeScope)
	return selector.Server(blockRequest).Match(matchFunc(cfg, applog)).Build()
}

func matchFunc(cfg *Config, applog *slog.Logger) selector.MatchFunc {
	return func(ctx context.Context, operation string) bool {
		if !cfg.routeScope.Match(operation) {
			applog.Debug("pass-kratos-random: match skip -- outside the route scope", "operation", operation, "side", cfg.routeScope.Side)
			return false
		}
		// passRate=0.6 lets 60% through; at 1 (and above) each one passes, at 0 (and below) none passes
		// passRate=0.6 表示 60% 的概率放行；大于等于 1 全部放行，小于等于 0 全部拦下
		pass := rand.Float64() < cfg.passRate
		applog.Debug("pass-kratos-random: random pass", "operation", operation, "side", cfg.routeScope.Side, "pass", pass)
		// A match sends the request into blockRequest, so the match is the inverse of the pass
		// 匹配上的请求会进入 blockRequest，所以匹配结果是放行结果取反
		return !pass
	}
}

// blockRequest answers each request it gets with 503, the requests matchFunc picks to block
// blockRequest 对收到的每个请求都返回 503，这些是 matchFunc 选中要拦下的请求
func blockRequest(handleFunc middleware.Handler) middleware.Handler {
	return func(ctx context.Context, req any) (any, error) {
		return nil, errors.ServiceUnavailable("RANDOM_RATE_UNAVAILABLE", "pass-kratos-random: random unavailable")
	}
}
