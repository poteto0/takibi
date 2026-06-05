package thttp_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
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
