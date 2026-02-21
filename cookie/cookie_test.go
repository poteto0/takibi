package cookie_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/poteto0/takibi/cookie"
	"github.com/poteto0/takibi/interfaces"
	"github.com/stretchr/testify/assert"
)

type mockContext[Bindings any] struct {
	env      *Bindings
	request  *http.Request
	response http.ResponseWriter
}

func (m *mockContext[Bindings]) Env() *Bindings { return m.env }
func (m *mockContext[Bindings]) Request() *http.Request {
	if m.request == nil {
		return httptest.NewRequest("GET", "/", nil)
	}
	return m.request
}
func (m *mockContext[Bindings]) Response() http.ResponseWriter {
	if m.response == nil {
		return httptest.NewRecorder()
	}
	return m.response
}
func (m *mockContext[Bindings]) Context() context.Context { return context.Background() }
func (m *mockContext[Bindings]) Reset(w http.ResponseWriter, r *http.Request) {
	m.response = w
	m.request = r
}
func (m *mockContext[Bindings]) Status(code int) interfaces.IContext[Bindings] { return m }
func (m *mockContext[Bindings]) Text(text string) error                        { return nil }
func (m *mockContext[Bindings]) Json(data any) error                           { return nil }
func (m *mockContext[Bindings]) Redirect(url string) error                     { return nil }

func TestCookie(t *testing.T) {
	t.Run("SetCookie", func(t *testing.T) {
		t.Run("basic cookie", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			ctx := &mockContext[any]{response: w, request: r}

			ok := cookie.SetCookie[any](ctx, "name", "takibi", nil)
			assert.True(t, ok)

			cookies := w.Result().Cookies()
			assert.Len(t, cookies, 1)
			assert.Equal(t, "name", cookies[0].Name)
			assert.Equal(t, "takibi", cookies[0].Value)
			assert.Equal(t, "/", cookies[0].Path) // Default
			assert.True(t, cookies[0].HttpOnly)   // Default
			assert.True(t, cookies[0].Secure)     // Default
			assert.Equal(t, http.SameSiteStrictMode, cookies[0].SameSite) // Default
		})

		t.Run("with options", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			ctx := &mockContext[any]{response: w, request: r}

			expires := time.Now().Add(24 * time.Hour)
			ok := cookie.SetCookie[any](ctx, "name", "value", &cookie.CookieOptions{
				Expires:  expires,
				Path:     "/api",
				Domain:   "example.com",
				Secure:   false,
				HttpOnly: false,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   3600,
			})
			assert.True(t, ok)

			cookies := w.Result().Cookies()
			assert.Len(t, cookies, 1)
			assert.Equal(t, "name", cookies[0].Name)
			assert.Equal(t, "value", cookies[0].Value)
			assert.Equal(t, "/api", cookies[0].Path)
			assert.Equal(t, "example.com", cookies[0].Domain)
			assert.False(t, cookies[0].Secure)
			assert.False(t, cookies[0].HttpOnly)
			assert.Equal(t, http.SameSiteLaxMode, cookies[0].SameSite)
			assert.Equal(t, 3600, cookies[0].MaxAge)
			assert.WithinDuration(t, expires, cookies[0].Expires, time.Second)
		})
		
		t.Run("prefix secure", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			ctx := &mockContext[any]{response: w, request: r}
			
			ok := cookie.SetCookie[any](ctx, "name", "value", &cookie.CookieOptions{
				Prefix: "secure",
			})
			assert.True(t, ok)
			
			cookies := w.Result().Cookies()
			assert.Len(t, cookies, 1)
			assert.Equal(t, "__Secure-name", cookies[0].Name)
			assert.True(t, cookies[0].Secure) // Force secure
		})

		t.Run("prefix host", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			ctx := &mockContext[any]{response: w, request: r}
			
			ok := cookie.SetCookie[any](ctx, "name", "value", &cookie.CookieOptions{
				Prefix: "host",
			})
			assert.True(t, ok)
			
			cookies := w.Result().Cookies()
			assert.Len(t, cookies, 1)
			assert.Equal(t, "__Host-name", cookies[0].Name)
			assert.True(t, cookies[0].Secure) // Force secure
			assert.Equal(t, "/", cookies[0].Path) // Force path /
			assert.Empty(t, cookies[0].Domain) // Force no domain
		})
	})

	t.Run("GetCookie", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			r.AddCookie(&http.Cookie{Name: "name", Value: "takibi"})
			ctx := &mockContext[any]{request: r}

			c, ok := cookie.GetCookie[any](ctx, "name")
			assert.True(t, ok)
			assert.NotNil(t, c)
			assert.Equal(t, "name", c.Name)
			assert.Equal(t, "takibi", c.Value)
		})

		t.Run("fail", func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			ctx := &mockContext[any]{request: r}

			c, ok := cookie.GetCookie[any](ctx, "name")
			assert.False(t, ok)
			assert.Nil(t, c)
		})
	})
	
	t.Run("GetCookies", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.AddCookie(&http.Cookie{Name: "name1", Value: "value1"})
		r.AddCookie(&http.Cookie{Name: "name2", Value: "value2"})
		ctx := &mockContext[any]{request: r}
		
		cookies := cookie.GetCookies[any](ctx)
		assert.Len(t, cookies, 2)
		assert.Equal(t, "name1", cookies[0].Name)
		assert.Equal(t, "name2", cookies[1].Name)
	})
	
	t.Run("DeleteCookie", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		ctx := &mockContext[any]{response: w, request: r}
		
		ok := cookie.DeleteCookie[any](ctx, "name", nil)
		assert.True(t, ok)
		
		cookies := w.Result().Cookies()
		assert.Len(t, cookies, 1)
		assert.Equal(t, "name", cookies[0].Name)
		assert.Equal(t, "", cookies[0].Value)
		assert.Equal(t, -1, cookies[0].MaxAge)
		assert.True(t, cookies[0].Expires.Before(time.Now()))
	})

	t.Run("SignedCookie", func(t *testing.T) {
		t.Run("Set and Get success", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			ctx := &mockContext[any]{response: w, request: r}
			// gorilla/securecookie default hash key size is 32 or 64 bytes
			// but it accepts any length key for HMAC.
			// However, AES block size is 16, 24, or 32 bytes for encryption.
			// securecookie.New(hashKey, blockKey)
			// If blockKey is nil, it only signs.
			secret := "secret-key-must-be-32-bytes-long!!" 

			// Fix signature to match implementation: SetSignedCookie(ctx, name, value, secret, options)
			ok := cookie.SetSignedCookie[any](ctx, "name", "takibi", secret, nil)
			assert.True(t, ok)

			cookies := w.Result().Cookies()
			assert.Len(t, cookies, 1)
			assert.Equal(t, "name", cookies[0].Name)
			assert.NotEqual(t, "takibi", cookies[0].Value) // Should be encoded

			// Reuse cookie for GetSignedCookie
			// In a real request, the client sends back the cookie.
			// httptest.NewRequest doesn't automatically pick up cookies from recorder.
			reqWithCookie := httptest.NewRequest("GET", "/", nil)
			reqWithCookie.AddCookie(cookies[0])
			
			ctxWithCookie := &mockContext[any]{response: w, request: reqWithCookie}

			// Fix signature to match implementation: GetSignedCookie(ctx, name, secret)
			c, ok := cookie.GetSignedCookie[any](ctxWithCookie, "name", secret)
			assert.True(t, ok)
			assert.NotNil(t, c)
			assert.Equal(t, "takibi", c.Value)
		})

		t.Run("GetSignedCookie fail with wrong secret", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			ctx := &mockContext[any]{response: w, request: r}
			secret := "secret-key-must-be-32-bytes-long!!"
			wrongSecret := "wrong-key-must-be-32-bytes-long!!"

			cookie.SetSignedCookie[any](ctx, "name", "takibi", secret, nil)
			cookies := w.Result().Cookies()
			
			reqWithCookie := httptest.NewRequest("GET", "/", nil)
			reqWithCookie.AddCookie(cookies[0])
			ctxWithCookie := &mockContext[any]{response: w, request: reqWithCookie}

			c, ok := cookie.GetSignedCookie[any](ctxWithCookie, "name", wrongSecret)
			assert.False(t, ok)
			assert.Nil(t, c)
		})

		t.Run("GetSignedCookie fail with tampered value", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			ctx := &mockContext[any]{response: w, request: r}
			secret := "secret-key-must-be-32-bytes-long!!"

			cookie.SetSignedCookie[any](ctx, "name", "takibi", secret, nil)
			cookies := w.Result().Cookies()
			
			tamperedCookie := cookies[0]
			tamperedCookie.Value += "tampered"
			
			reqWithCookie := httptest.NewRequest("GET", "/", nil)
			reqWithCookie.AddCookie(tamperedCookie)
			ctxWithCookie := &mockContext[any]{response: w, request: reqWithCookie}

			c, ok := cookie.GetSignedCookie[any](ctxWithCookie, "name", secret)
			assert.False(t, ok)
			assert.Nil(t, c)
		})
	})
}
