package passkratosrandom

import (
	"context"
	"io"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/transport/http"
	"github.com/stretchr/testify/require"
	"github.com/yylego/kratos-auth/authkratos"
)

// newTestServer starts a Kratos HTTP server with the random pass middleware and two routes:
// /api.Demo/Search is inside the scope, /api.Demo/Health is outside it
//
// newTestServer 启动带随机放行中间件的 Kratos HTTP 服务，挂两个路由：
// /api.Demo/Search 在范围内，/api.Demo/Health 不在范围内
func newTestServer(t *testing.T, cfg *Config) *httptest.Server {
	srv := http.NewServer(http.Middleware(NewMiddleware(cfg, slog.Default())))
	route := srv.Route("/")
	for path, operation := range map[string]string{"/search": "/api.Demo/Search", "/health": "/api.Demo/Health"} {
		route.GET(path, func(ctx http.Context) error {
			http.SetOperation(ctx, operation)
			handle := ctx.Middleware(func(ctx context.Context, req any) (any, error) {
				return map[string]string{"status": "ok"}, nil
			})
			res, err := handle(ctx, nil)
			if err != nil {
				return err
			}
			return ctx.Result(200, res)
		})
	}
	server := httptest.NewServer(srv)
	t.Cleanup(server.Close)
	return server
}

// sendRequest sends a GET request and returns the status code and the body
// sendRequest 发一个 GET 请求，返回状态码和响应体
func sendRequest(t *testing.T, address string) (int, string) {
	response, err := nethttp.Get(address)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, response.Body.Close())
	}()
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	return response.StatusCode, string(body)
}

// countPasses sends n requests and returns the count that get 200
// countPasses 发 n 个请求，返回拿到 200 的个数
func countPasses(t *testing.T, address string, n int) int {
	count := 0
	for range n {
		if code, _ := sendRequest(t, address); code == 200 {
			count++
		}
	}
	return count
}

const blockedBody = `{"code":503,"reason":"RANDOM_BLOCKED","message":"pass-kratos-random: request blocked at random"}`

func TestNewConfig(t *testing.T) {
	cfg := NewConfig(authkratos.NewInclude("/api.Demo/Search"), 0.8)
	require.Equal(t, 0.8, cfg.passRate)
}

func TestConfig_WithBlockErk(t *testing.T) {
	cfg := NewConfig(authkratos.NewInclude("/api.Demo/Search"), 0).
		WithBlockErk(func(ctx context.Context) *errors.Error {
			return errors.ServiceUnavailable("CHAOS_BLOCKED", "blocked on purpose")
		})
	server := newTestServer(t, cfg)

	code, body := sendRequest(t, server.URL+"/search")
	require.Equal(t, 503, code)
	require.Equal(t, `{"code":503,"reason":"CHAOS_BLOCKED","message":"blocked on purpose"}`, body)
}

func TestNewMiddleware(t *testing.T) {
	scope := authkratos.NewInclude("/api.Demo/Search")

	t.Run("rate-1-lets-each-request-through", func(t *testing.T) {
		server := newTestServer(t, NewConfig(scope, 1))
		code, body := sendRequest(t, server.URL+"/search")
		require.Equal(t, 200, code)
		require.Equal(t, `{"status":"ok"}`, body)
		require.Equal(t, 100, countPasses(t, server.URL+"/search", 100))
	})

	t.Run("rate-0-blocks-each-request", func(t *testing.T) {
		server := newTestServer(t, NewConfig(scope, 0))
		code, body := sendRequest(t, server.URL+"/search")
		require.Equal(t, 503, code)
		require.Equal(t, blockedBody, body)
		require.Equal(t, 0, countPasses(t, server.URL+"/search", 100))
	})

	t.Run("route-outside-the-scope-is-not-blocked", func(t *testing.T) {
		server := newTestServer(t, NewConfig(scope, 0))
		require.Equal(t, 100, countPasses(t, server.URL+"/health", 100))
	})

	t.Run("rate-0.5-lets-about-half-through", func(t *testing.T) {
		// With 2000 requests, the odds of landing outside 40%~60% are below one in a billion
		// 发 2000 个请求，落在 40%~60% 之外的概率小于十亿分之一
		server := newTestServer(t, NewConfig(scope, 0.5))
		count := countPasses(t, server.URL+"/search", 2000)
		require.Greater(t, count, 800)
		require.Less(t, count, 1200)
	})
}
