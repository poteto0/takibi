package middlewares

import "github.com/poteto0/takibi/interfaces"

type SecureHeadersConfig struct {
	XContentTypeOptions string
	XFrameOptions       string
	ReferrerPolicy      string
}

func DefaultSecureHeadersConfig() SecureHeadersConfig {
	return SecureHeadersConfig{
		XContentTypeOptions: "nosniff",
		XFrameOptions:       "DENY",
		ReferrerPolicy:      "strict-origin-when-cross-origin",
	}
}

func SecureHeaders[Bindings any](config ...SecureHeadersConfig) interfaces.MiddlewareFunc[Bindings] {
	cfg := DefaultSecureHeadersConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c interfaces.IContext[Bindings], next interfaces.HandlerFunc[Bindings]) error {
		h := c.Response().Header()

		if cfg.XContentTypeOptions != "" {
			h.Set("X-Content-Type-Options", cfg.XContentTypeOptions)
		}
		if cfg.XFrameOptions != "" {
			h.Set("X-Frame-Options", cfg.XFrameOptions)
		}
		if cfg.ReferrerPolicy != "" {
			h.Set("Referrer-Policy", cfg.ReferrerPolicy)
		}

		return next(c)
	}
}
