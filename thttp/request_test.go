package thttp_test

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/poteto0/takibi/thttp"
	"github.com/stretchr/testify/assert"
)

func Test_NewRequest(t *testing.T) {
	// Act & Assert
	assert.NotNil(t, thttp.NewRequest(nil, nil))
}

func Test_Request_Raw(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "http://example.com", nil)
	r := thttp.NewRequest(req, nil)

	// Act & Assert
	assert.Equal(t, req, r.Raw())
}

func Test_Request_Header(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("X-Custom-Header", "value")
	r := thttp.NewRequest(req, nil)

	// Act & Assert
	assert.Equal(t, req.Header, r.Header())
}

func Test_Request_HeaderBy(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("X-Custom-Header", "value")
	r := thttp.NewRequest(req, nil)

	// Act & Assert
	assert.Equal(t, "value", r.HeaderBy("X-Custom-Header"))
	assert.Equal(t, "", r.HeaderBy("Nonexistent-Header"))
}

func Test_Request_ContentType(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("Content-Type", "application/json")
	r := thttp.NewRequest(req, nil)

	// Act & Assert
	assert.Equal(t, "application/json", r.ContentType())
}

func Test_Request_MediaType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        string
	}{
		{"plain media type", "application/json", "application/json"},
		{"media type with charset parameter", "application/json; charset=utf-8", "application/json"},
		{"uppercase is normalized to lowercase", "Application/JSON", "application/json"},
		{"empty header", "", ""},
		{"unparsable header", "application/", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			req := httptest.NewRequest("POST", "http://example.com", nil)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			r := thttp.NewRequest(req, nil)

			// Act & Assert
			assert.Equal(t, tt.want, r.MediaType())
		})
	}
}

func Test_Request_Json(t *testing.T) {
	// Arrange
	jsonBody := `{"message": "hello"}`
	req := httptest.NewRequest("POST", "http://example.com", bytes.NewBufferString(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r := thttp.NewRequest(req, nil)

	// Act
	data, err := r.Json()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{
		"message": "hello",
	}, data)
}

func Test_Request_Unmarshall(t *testing.T) {
	// Arrange
	type Payload struct {
		Message string `json:"message"`
	}
	payload := &Payload{}
	jsonBody := `{"message": "hello"}`

	t.Run("empty body is error", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest("POST", "http://example.com", nil)
		req.Header.Set("Content-Type", "application/json")
		r := thttp.NewRequest(req, nil)

		// Act
		err := r.Unmarshall(payload)

		// Assert
		assert.Error(t, err)
	})

	t.Run("unexpected content-type is error", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest("POST", "http://example.com", bytes.NewBufferString(jsonBody))
		req.Header.Set("Content-Type", "text/plain")
		r := thttp.NewRequest(req, nil)

		// Act
		err := r.Unmarshall(payload)

		// Assert
		assert.Error(t, err)
	})

	t.Run("valid json body is unmarshallable", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest("POST", "http://example.com", bytes.NewBufferString(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		r := thttp.NewRequest(req, nil)

		// Act
		err := r.Unmarshall(payload)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("content-type with charset parameter is accepted", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest("POST", "http://example.com", bytes.NewBufferString(jsonBody))
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		r := thttp.NewRequest(req, nil)

		// Act
		err := r.Unmarshall(payload)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("chunked body without content-length is unmarshallable", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest("POST", "http://example.com", bytes.NewBufferString(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		// chunked transfer encoding: ContentLength is unknown (-1)
		req.ContentLength = -1

		r := thttp.NewRequest(req, nil)

		// Act
		err := r.Unmarshall(payload)

		// Assert
		assert.NoError(t, err)
	})
}

func Test_Request_Unmarshall_BodySizeLimit(t *testing.T) {
	type Payload struct {
		Message string `json:"message"`
	}

	t.Run("body exceeding custom limit returns MaxBytesError", func(t *testing.T) {
		jsonBody := `{"message": "hello world"}`
		req := httptest.NewRequest("POST", "http://example.com", bytes.NewBufferString(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		r := thttp.NewRequest(req, &thttp.RequestOption{MaxBodyBytes: 5})

		err := r.Unmarshall(&Payload{})

		assert.Error(t, err)
		var maxBytesErr *http.MaxBytesError
		assert.True(t, errors.As(err, &maxBytesErr))
	})

	t.Run("body within custom limit succeeds", func(t *testing.T) {
		jsonBody := `{"message": "hi"}`
		req := httptest.NewRequest("POST", "http://example.com", bytes.NewBufferString(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		r := thttp.NewRequest(req, &thttp.RequestOption{MaxBodyBytes: 100})

		payload := &Payload{}
		err := r.Unmarshall(payload)

		assert.NoError(t, err)
		assert.Equal(t, "hi", payload.Message)
	})
}

func Test_Request_UnmarshallForm(t *testing.T) {
	type SignUp struct {
		Name    string  `form:"name"`
		Age     int     `form:"age"`
		Big     int64   `form:"big"`
		Score   float64 `form:"score"`
		Admin   bool    `form:"admin"`
		Untaged string  // resolved by field name
		Skip    string  `form:"-"`
		hidden  string  // unexported, must be skipped without panic
	}

	t.Run("urlencoded body binds typed fields", func(t *testing.T) {
		// Arrange
		form := url.Values{
			"name":    {"alice"},
			"age":     {"30"},
			"big":     {"9999999999"},
			"score":   {"9.5"},
			"admin":   {"true"},
			"Untaged": {"raw"},
			"-":       {"ignored"},
		}
		req := httptest.NewRequest("POST", "http://example.com", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r := thttp.NewRequest(req, nil)

		// Act
		var in SignUp
		err := r.UnmarshallForm(&in)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "alice", in.Name)
		assert.Equal(t, 30, in.Age)
		assert.Equal(t, int64(9999999999), in.Big)
		assert.Equal(t, 9.5, in.Score)
		assert.True(t, in.Admin)
		assert.Equal(t, "raw", in.Untaged)
		assert.Equal(t, "", in.Skip)
		assert.Equal(t, "", in.hidden)
	})

	t.Run("multipart/form-data value fields bind", func(t *testing.T) {
		// Arrange
		body := &bytes.Buffer{}
		w := multipart.NewWriter(body)
		_ = w.WriteField("name", "bob")
		_ = w.WriteField("age", "42")
		_ = w.Close()
		req := httptest.NewRequest("POST", "http://example.com", body)
		req.Header.Set("Content-Type", w.FormDataContentType())
		r := thttp.NewRequest(req, nil)

		// Act
		var in SignUp
		err := r.UnmarshallForm(&in)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "bob", in.Name)
		assert.Equal(t, 42, in.Age)
	})

	t.Run("missing or empty fields stay zero without error", func(t *testing.T) {
		// Arrange
		form := url.Values{"name": {"alice"}, "age": {""}}
		req := httptest.NewRequest("POST", "http://example.com", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r := thttp.NewRequest(req, nil)

		// Act
		var in SignUp
		err := r.UnmarshallForm(&in)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "alice", in.Name)
		assert.Equal(t, 0, in.Age)
	})

	t.Run("unsupported content-type is error", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest("POST", "http://example.com", strings.NewReader("name=alice"))
		req.Header.Set("Content-Type", "application/json")
		r := thttp.NewRequest(req, nil)

		// Act
		err := r.UnmarshallForm(&SignUp{})

		// Assert
		assert.Error(t, err)
	})

	t.Run("conversion failure is error", func(t *testing.T) {
		cases := map[string]url.Values{
			"int":   {"age": {"abc"}},
			"int64": {"big": {"x"}},
			"float": {"score": {"x"}},
			"bool":  {"admin": {"notabool"}},
		}
		for name, form := range cases {
			t.Run(name, func(t *testing.T) {
				req := httptest.NewRequest("POST", "http://example.com", strings.NewReader(form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				r := thttp.NewRequest(req, nil)

				err := r.UnmarshallForm(&SignUp{})

				assert.Error(t, err)
			})
		}
	})

	t.Run("malformed urlencoded body is error", func(t *testing.T) {
		// Arrange: invalid percent-encoding makes ParseForm fail
		req := httptest.NewRequest("POST", "http://example.com", strings.NewReader("name=%zz"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r := thttp.NewRequest(req, nil)

		// Act
		err := r.UnmarshallForm(&SignUp{})

		// Assert
		assert.Error(t, err)
	})

	t.Run("malformed multipart body is error", func(t *testing.T) {
		// Arrange: multipart content-type but body is not a valid multipart payload
		req := httptest.NewRequest("POST", "http://example.com", strings.NewReader("not-multipart"))
		req.Header.Set("Content-Type", "multipart/form-data; boundary=xxx")
		r := thttp.NewRequest(req, nil)

		// Act
		err := r.UnmarshallForm(&SignUp{})

		// Assert
		assert.Error(t, err)
	})

	t.Run("non-pointer dest is error", func(t *testing.T) {
		req := httptest.NewRequest("POST", "http://example.com", strings.NewReader("name=alice"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r := thttp.NewRequest(req, nil)

		assert.Error(t, r.UnmarshallForm(SignUp{}))
	})

	t.Run("nil pointer dest is error", func(t *testing.T) {
		req := httptest.NewRequest("POST", "http://example.com", strings.NewReader("name=alice"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r := thttp.NewRequest(req, nil)

		var p *SignUp
		assert.Error(t, r.UnmarshallForm(p))
	})

	t.Run("pointer to non-struct is error", func(t *testing.T) {
		req := httptest.NewRequest("POST", "http://example.com", strings.NewReader("name=alice"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r := thttp.NewRequest(req, nil)

		n := 0
		assert.Error(t, r.UnmarshallForm(&n))
	})

	t.Run("unsupported field type with value is error", func(t *testing.T) {
		type Bad struct {
			Tags []string `form:"tags"`
		}
		form := url.Values{"tags": {"a"}}
		req := httptest.NewRequest("POST", "http://example.com", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r := thttp.NewRequest(req, nil)

		assert.Error(t, r.UnmarshallForm(&Bad{}))
	})
}

func Test_Request_Queries(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "http://example.com?foo=bar&foo=baz&qux=quux", nil)
	r := thttp.NewRequest(req, nil)

	// Act
	queries := r.Queries()

	// Assert
	assert.Equal(t, map[string][]string{
		"foo": {"bar", "baz"},
		"qux": {"quux"},
	}, queries)
}

func Test_Request_QueriesBy(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "http://example.com?foo=bar&foo=baz&qux=quux", nil)
	r := thttp.NewRequest(req, nil)

	// Act
	fooValues := r.QueriesBy("foo")
	quxValues := r.QueriesBy("qux")
	nonexistentValues := r.QueriesBy("nonexistent")

	// Assert
	assert.Equal(t, []string{"bar", "baz"}, fooValues)
	assert.Equal(t, []string{"quux"}, quxValues)
	assert.Equal(t, []string(nil), nonexistentValues)
}

func Test_Request_Query(t *testing.T) {
	t.Run("get first value of query parameters", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest("GET", "http://example.com?foo=bar&foo=baz&qux=quux", nil)
		r := thttp.NewRequest(req, nil)

		// Act
		query := r.Query()

		// Assert
		assert.Equal(t, map[string]string{
			"foo": "bar",
			"qux": "quux",
		}, query)
	})
}

func Test_Request_QueryBy(t *testing.T) {
	t.Run("get first value of query parameter by key", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest("GET", "http://example.com?foo=bar&foo=baz&qux=quux", nil)
		r := thttp.NewRequest(req, nil)

		// Act
		fooValues := r.QueryBy("foo")
		quxValues := r.QueryBy("qux")
		nonexistentValues := r.QueryBy("nonexistent")

		// Assert
		assert.Equal(t, "bar", fooValues)
		assert.Equal(t, "quux", quxValues)
		assert.Equal(t, "", nonexistentValues)
	})
}
