package cookie

import (
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/poteto0/takibi/constants"
	"github.com/poteto0/takibi/interfaces"
)

var DefaultCookieOptions = &CookieOptions{
	Path:     "/",
	HttpOnly: true,
	Secure:   true,
	SameSite: http.SameSiteStrictMode,
}

func SetCookie[T any](ctx interfaces.IContext[T], name, value string, opts *CookieOptions) bool {
	cookie := &http.Cookie{
		Name:  name,
		Value: value,
	}

	options := opts
	if options == nil {
		options = DefaultCookieOptions
	}

	cookie.Path = options.Path
	cookie.Domain = options.Domain
	cookie.Expires = options.Expires
	cookie.Secure = options.Secure
	cookie.HttpOnly = options.HttpOnly
	cookie.SameSite = options.SameSite
	cookie.MaxAge = options.MaxAge

	if options.Prefix == constants.CookieSecurePrefixMode {
		makeCookieSecure(cookie)
	}
	if options.Prefix == constants.CookieHostPrefixMode {
		makeCookieHost(cookie)
	}

	http.SetCookie(ctx.Response(), cookie)
	return true
}

func GetCookie[T any](ctx interfaces.IContext[T], name string, opts *CookieOptions) (*http.Cookie, bool) {
	if opts != nil && opts.Prefix == constants.CookieSecurePrefixMode {
		name = constants.CookieSecurePrefix + name
	}
	if opts != nil && opts.Prefix == constants.CookieHostPrefixMode {
		name = constants.CookieHostPrefix + name
	}

	c, err := ctx.Request().Cookie(name)
	if err != nil {
		return nil, false
	}
	return c, true
}

func SetSignedCookie[T any](ctx interfaces.IContext[T], name, value, secret string, opts *CookieOptions) bool {
	s := securecookie.New([]byte(secret), nil)
	encoded, err := s.Encode(name, value)
	if err != nil {
		return false
	}
	return SetCookie(ctx, name, encoded, opts)
}

func GetSignedCookie[T any](ctx interfaces.IContext[T], name, secret string, opts *CookieOptions) (*http.Cookie, bool) {
	c, ok := GetCookie(ctx, name, opts)
	if !ok {
		return nil, false
	}

	s := securecookie.New([]byte(secret), nil)
	var value string
	if err := s.Decode(name, c.Value, &value); err != nil {
		return nil, false
	}

	// Return a copy with decoded value
	decodedCookie := *c
	decodedCookie.Value = value
	return &decodedCookie, true
}

func GetCookies[T any](ctx interfaces.IContext[T]) []*http.Cookie {
	return ctx.Request().Cookies()
}

func makeCookieSecure(c *http.Cookie) {
	c.Name = constants.CookieSecurePrefix + c.Name
	c.Secure = true
}

func makeCookieHost(c *http.Cookie) {
	c.Name = constants.CookieHostPrefix + c.Name
	c.Secure = true
	c.Path = "/"
	c.Domain = ""
}
