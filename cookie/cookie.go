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

	if options.Prefix == constants.CookiePrefixSecure {
		cookie.Name = constants.CookieSecurePrefix + name
		cookie.Secure = true
	}
	if options.Prefix == constants.CookiePrefixHost {
		cookie.Name = constants.CookieHostPrefix + name
		cookie.Secure = true
		cookie.Path = "/"
		cookie.Domain = ""
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

func SetSignedCookie[T any](ctx interfaces.IContext[T], name, value, secret string, opts *CookieOptions) bool {
	s := securecookie.New([]byte(secret), nil)
	encoded, err := s.Encode(name, value)
	if err != nil {
		return false
	}
	return SetCookie(ctx, name, encoded, opts)
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

func DeleteCookie[T any](ctx interfaces.IContext[T], name string, opts *CookieOptions) bool {
	cookie := &http.Cookie{
		Name:   name,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	
	options := opts
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
		
		if options.Prefix == constants.CookiePrefixSecure {
			cookie.Name = constants.CookieSecurePrefix + name
		}
		if options.Prefix == constants.CookiePrefixHost {
			cookie.Name = constants.CookieHostPrefix + name
		}
	}

	http.SetCookie(ctx.Response(), cookie)
	return true
}
