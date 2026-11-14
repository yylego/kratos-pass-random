<!-- TEMPLATE (EN) BEGIN: BADGES -->

[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/yylego/kratos-pass-random/release.yml?branch=main&label=BUILD)](https://github.com/yylego/kratos-pass-random/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/yylego/kratos-pass-random)](https://pkg.go.dev/github.com/yylego/kratos-pass-random)
[![Coverage Status](https://img.shields.io/coveralls/github/yylego/kratos-pass-random/main.svg)](https://coveralls.io/github/yylego/kratos-pass-random?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.26%2B-lightgrey.svg)](https://go.dev/)
[![GitHub Release](https://img.shields.io/github/release/yylego/kratos-pass-random.svg)](https://github.com/yylego/kratos-pass-random/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yylego/kratos-pass-random)](https://goreportcard.com/report/github.com/yylego/kratos-pass-random)

<!-- TEMPLATE (EN) CLOSE: BADGES -->

# kratos-pass-random

Random request blocking middleware for chaos testing and fault injection.

---

<!-- TEMPLATE (EN) BEGIN: LANGUAGE NAVIGATION -->

## CHINESE README

[中文说明](README.zh.md)

<!-- TEMPLATE (EN) CLOSE: LANGUAGE NAVIGATION -->

## Main Features

- 🎲 **Random Blocking**: Block requests at configurable odds (0.0 to 1.0 pass rate)
- 🔥 **Chaos Testing**: Inject random 503 responses to validate resilience
- 🗺️ **Route Scope**: INCLUDE/EXCLUDE mode to choose which routes get blocked
- 🪵 **Debug Logs**: Each pass and block outcome goes to the logger at debug level, and the logger level decides if it shows

## Installation

```bash
go get github.com/yylego/kratos-pass-random
```

## Usage

```go
import (
    "log/slog"

    "github.com/go-kratos/kratos/v3/transport/http"
    "github.com/yylego/kratos-auth/authkratos"
    "github.com/yylego/kratos-pass-random/passkratosrandom"
)

// newHTTPServer lets 80% of the Search requests through and blocks 20% with 503, to check how the clients cope
func newHTTPServer(logger *slog.Logger) *http.Server {
    routeScope := authkratos.NewInclude("/api.Article/Search")
    cfg := passkratosrandom.NewConfig(routeScope, 0.8)
    return http.NewServer(http.Middleware(passkratosrandom.NewMiddleware(cfg, logger)))
}
```

Blocked requests get `503 RANDOM_RATE_UNAVAILABLE`. Routes outside the scope are not blocked.

---

<!-- TEMPLATE (EN) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-11-25 03:52:28.131064 +0000 UTC -->

## 📄 License

MIT License - see [LICENSE](LICENSE).

---

## 💬 Contact & Feedback

Contributions are welcome! Report bugs, suggest features, and contribute code:

- 🐛 **Mistake reports?** Open an issue on GitHub with reproduction steps
- 💡 **Fresh ideas?** Create an issue to discuss
- 📖 **Documentation confusing?** Report it so we can improve
- 🚀 **Need new features?** Share the use cases to help us understand requirements
- ⚡ **Performance issue?** Help us optimize through reporting slow operations
- 🔧 **Configuration problem?** Ask questions about complex setups
- 📢 **Follow project progress?** Watch the repo to get new releases and features
- 🌟 **Success stories?** Share how this package improved the workflow
- 💬 **Feedback?** We welcome suggestions and comments

---

## 🔧 Development

New code contributions, follow this process:

1. **Fork**: Fork the repo on GitHub (using the webpage UI).
2. **Clone**: Clone the forked project (`git clone https://github.com/yourname/repo-name.git`).
3. **Navigate**: Navigate to the cloned project (`cd repo-name`)
4. **Branch**: Create a feature branch (`git checkout -b feature/xxx`).
5. **Code**: Implement the changes with comprehensive tests
6. **Testing**: (Golang project) Ensure tests pass (`go test ./...`) and follow Go code style conventions
7. **Documentation**: Update documentation to support client-facing changes
8. **Stage**: Stage changes (`git add .`)
9. **Commit**: Commit changes (`git commit -m "Add feature xxx"`) ensuring backward compatible code
10. **Push**: Push to the branch (`git push origin feature/xxx`).
11. **PR**: Open a merge request on GitHub (on the GitHub webpage) with detailed description.

Please ensure tests pass and include relevant documentation updates.

---

## 🌟 Support

Welcome to contribute to this project via submitting merge requests and reporting issues.

**Project Support:**

- ⭐ **Give GitHub stars** if this project helps you
- 🤝 **Share with teammates** and (golang) programming friends
- 📝 **Write tech blogs** about development tools and workflows - we provide content writing support
- 🌟 **Join the ecosystem** - committed to supporting open source and the (golang) development scene

**Have Fun Coding with this package!** 🎉🎉🎉

<!-- TEMPLATE (EN) CLOSE: STANDARD PROJECT FOOTER -->
