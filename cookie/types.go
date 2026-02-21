package cookie

import (
	"net/http"
	"time"
)

type CookieOptions struct {
	Expires  time.Time
	HttpOnly bool
	Secure   bool
	Path     string
	Domain   string
	SameSite http.SameSite
	MaxAge   int
	Prefix   string // "secure" | "host" | ""
}
