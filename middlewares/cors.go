package middlewares

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/poteto0/takibi/interfaces"
)

type CorsConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	ExposeHeaders    []string
	MaxAge           int
}

func DefaultCorsConfig() CorsConfig {
	return CorsConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodPatch},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		MaxAge:       86400,
	}
}

func Cors[Bindings any](config ...CorsConfig) interfaces.MiddlewareFunc[Bindings] {
	cfg := DefaultCorsConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c interfaces.IContext[Bindings], next interfaces.HandlerFunc[Bindings]) error {
		req := c.Request()
		res := c.Response()
		origin := req.Header.Get("Origin")

		// Check allowed origins
		allowOrigin := ""
		for _, o := range cfg.AllowOrigins {
			if o == "*" {
				allowOrigin = "*"
				break
			}
			if o == origin {
				allowOrigin = origin
				break
			}
		}

		if allowOrigin != "" {
			res.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			if cfg.AllowCredentials {
				res.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if len(cfg.ExposeHeaders) > 0 {
				res.Header().Set("Access-Control-Expose-Headers", strings.Join(cfg.ExposeHeaders, ", "))
			}
		}

		// Handle preflight
		if req.Method == http.MethodOptions {
			if len(cfg.AllowMethods) > 0 {
				res.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowMethods, ", "))
			}
			if len(cfg.AllowHeaders) > 0 {
				res.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowHeaders, ", "))
			}
			if cfg.MaxAge > 0 {
				res.Header().Set("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAge))
			}
			res.WriteHeader(http.StatusNoContent)
			return nil
		}

		return next(c)
	}
}
