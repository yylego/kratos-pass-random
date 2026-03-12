// Package passkratosrandom: Probabilistic request blocking middleware with chaos testing support
// Provides random pass-through rate with configurable odds
// Good fit in testing, chaos engineering, and staged rollout scenarios
// Enables blocking specific request ratio within designated route scope
//
// passkratosrandom: 概率性请求阻断中间件，支持混沌测试
// 提供可配置概率的随机放行率控制
// 适用于压力测试、混沌工程和灰度发布场景
// 可在指定路由范围内阻断一定百分比的请求
package passkratosrandom

import (
	"context"
	"math/rand"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/yylego/kratos-auth/authkratos"
	"github.com/yylego/neatjson/neatjsons"
)

type Config struct {
	routeScope *authkratos.RouteScope
	rate       float64
	spanHooks  []authkratos.NewSpanHookFunc
	debugMode  bool
}

func NewConfig(routeScope *authkratos.RouteScope, passRate float64) *Config {
	return &Config{
		routeScope: routeScope,
		rate:       passRate,
		spanHooks:  nil,
		debugMode:  false,
	}
}

func (c *Config) WithDebugMode(debugMode bool) *Config {
	c.debugMode = debugMode
	return c
}

// WithNewSpanHook appends a span hook creation callback
//
// WithNewSpanHook 追加一个 span 钩子创建回调函数
func (c *Config) WithNewSpanHook(fn authkratos.NewSpanHookFunc) *Config {
	c.spanHooks = append(c.spanHooks, fn)
	return c
}

// NewMiddleware creates middleware that fails requests with configured rate
//
// NewMiddleware 让接口有一定概率失败
func NewMiddleware(cfg *Config, logger log.Logger) middleware.Middleware {
	slog := log.NewHelper(logger)
	slog.Infof(
		"pass-kratos-random: new middleware side=%v operations=%d rate=%v debug-mode=%v",
		cfg.routeScope.Side,
		len(cfg.routeScope.OperationSet),
		cfg.rate,
		authkratos.BooleanToNum(cfg.debugMode),
	)
	if cfg.debugMode {
		slog.Debugf("pass-kratos-random: new middleware route-scope: %s", neatjsons.S(cfg.routeScope))
	}
	return selector.Server(middlewareFunc(cfg, logger)).Match(matchFunc(cfg, logger)).Build()
}

func matchFunc(cfg *Config, logger log.Logger) selector.MatchFunc {
	slog := log.NewHelper(logger)

	return func(ctx context.Context, operation string) bool {
		defer authkratos.RunSpanHooks(ctx, cfg.spanHooks, "pass-kratos-random-match")()

		if match := cfg.routeScope.Match(operation); !match {
			if cfg.debugMode {
				slog.Debugf("pass-kratos-random: operation=%s side=%v match=%d next -> skip random", operation, cfg.routeScope.Side, authkratos.BooleanToNum(match))
			}
			return false
		}
		// 设置rate=0.6就是有60%的概率通过
		// 设置rate=1或者>1就是肯定通过，设置为0或负数就必然不通过
		ratePass := rand.Float64() < cfg.rate

		// 是否进入拦截器，拦截器会拦截请求
		// 因此这里求逆值，通过的不拦截，不通过的拦截
		match := !ratePass
		if cfg.debugMode {
			if match {
				slog.Debugf("pass-kratos-random: operation=%s side=%v match=%d next -> goto unavailable", operation, cfg.routeScope.Side, authkratos.BooleanToNum(match))
			} else {
				slog.Debugf("pass-kratos-random: operation=%s side=%v match=%d skip -- skip unavailable", operation, cfg.routeScope.Side, authkratos.BooleanToNum(match))
			}
		}
		return match
	}
}

func middlewareFunc(cfg *Config, logger log.Logger) middleware.Middleware {
	slog := log.NewHelper(logger)

	return func(handleFunc middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			defer authkratos.RunSpanHooks(ctx, cfg.spanHooks, "pass-kratos-random")()

			if cfg.debugMode {
				slog.Debugf("pass-kratos-random: random match unavailable")
			}
			//当已经命中概率的时候，就直接返回错误
			return nil, errors.ServiceUnavailable("RANDOM_RATE_UNAVAILABLE", "pass-kratos-random: random unavailable")
		}
	}
}
