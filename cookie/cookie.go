package cookie

import (
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/poteto0/takibi/interfaces"
)

func SetCookie[T any](ctx interfaces.IContext[T], name, value string, options *CookieOptions) bool {
	cookie := &http.Cookie{
		Name:  name,
		Value: value,
	}

	if options == nil {
		cookie.Path = "/"
		cookie.HttpOnly = true
		cookie.Secure = true
		cookie.SameSite = http.SameSiteStrictMode
	} else {
		cookie.Path = options.Path
		cookie.Domain = options.Domain
		cookie.Expires = options.Expires
		cookie.Secure = options.Secure
		cookie.HttpOnly = options.HttpOnly
		cookie.SameSite = options.SameSite
		cookie.MaxAge = options.MaxAge

		if options.Prefix == "secure" {
			cookie.Name = "__Secure-" + name
			cookie.Secure = true
		}
		if options.Prefix == "host" {
			cookie.Name = "__Host-" + name
			cookie.Secure = true
			cookie.Path = "/"
			cookie.Domain = ""
		}
	}

	http.SetCookie(ctx.Response(), cookie)
	return true
}

func GetCookie[T any](ctx interfaces.IContext[T], name string) (*http.Cookie, bool) {
	c, err := ctx.Request().Cookie(name)
	if err != nil {
		return nil, false
	}
	return c, true
}

func SetSignedCookie[T any](ctx interfaces.IContext[T], name, value, secret string, options *CookieOptions) bool {
	s := securecookie.New([]byte(secret), nil)
	encoded, err := s.Encode(name, value)
	if err != nil {
		return false
	}
	return SetCookie(ctx, name, encoded, options)
}

func GetSignedCookie[T any](ctx interfaces.IContext[T], name, secret string) (*http.Cookie, bool) {
	c, err := ctx.Request().Cookie(name)
	if err != nil {
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

func DeleteCookie[T any](ctx interfaces.IContext[T], name string, options *CookieOptions) bool {
	cookie := &http.Cookie{
		Name:   name,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	
	if options != nil {
		if options.Path != "" {
			cookie.Path = options.Path
		}
		if options.Domain != "" {
			cookie.Domain = options.Domain
		}
		if options.Secure {
			cookie.Secure = true
		}
	}

	http.SetCookie(ctx.Response(), cookie)
	return true
}
