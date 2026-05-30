package takibi_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/poteto0/takibi"
	"github.com/poteto0/takibi/interfaces"
	"github.com/stretchr/testify/assert"
)

type TestBindings struct{}

func TestCamp_Get(t *testing.T) {
	app := takibi.New(&TestBindings{})
	app.Get("/hello", func(c interfaces.IContext[TestBindings]) error {
		return c.Text("Hello, World!")
	})

	resp := app.Camp("GET", "/hello")
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	body, _ := io.ReadAll(resp.Raw().Body)
	assert.Equal(t, "Hello, World!", string(body))
}

func TestCamp_PostWithBody(t *testing.T) {
	app := takibi.New(&TestBindings{})
	app.Post("/echo", func(c interfaces.IContext[TestBindings]) error {
		var reqBody map[string]string
		json.NewDecoder(c.Req().Raw().Body).Decode(&reqBody)
		return c.Json(reqBody)
	})

	resp := app.Camp("POST", "/echo",
		interfaces.Header("Content-Type", "application/json"),
		interfaces.Body(map[string]string{"msg": "echo"}),
	)

	assert.Equal(t, http.StatusOK, resp.StatusCode())

	var respBody map[string]string
	resp.Unmarshall(&respBody)
	assert.Equal(t, "echo", respBody["msg"])
}

func TestNew_WithMaxBodyBytes(t *testing.T) {
	type Bindings struct{}
	type Payload struct {
		Message string `json:"message"`
	}

	t.Run("body exceeding custom limit returns error in handler", func(t *testing.T) {
		app := takibi.NewWithOption(&Bindings{}, interfaces.TakibiOption{MaxBodyBytes: 20})

		var handlerErr error
		app.Post("/upload", func(c interfaces.IContext[Bindings]) error {
			handlerErr = c.Req().Unmarshall(&Payload{})
			return handlerErr
		})

		bigBody := strings.Repeat("x", 100)
		jsonBody := `{"message":"` + bigBody + `"}`
		app.Camp("POST", "/upload",
			interfaces.Header("Content-Type", "application/json"),
			interfaces.Body(bytes.NewBufferString(jsonBody)),
		)

		var maxBytesErr *http.MaxBytesError
		assert.ErrorAs(t, handlerErr, &maxBytesErr)
	})

	t.Run("body within custom limit succeeds", func(t *testing.T) {
		app := takibi.NewWithOption(&Bindings{}, interfaces.TakibiOption{MaxBodyBytes: 1024})

		var result Payload
		app.Post("/upload", func(c interfaces.IContext[Bindings]) error {
			if err := c.Req().Unmarshall(&result); err != nil {
				return err
			}
			return c.Text("ok")
		})

		resp := app.Camp("POST", "/upload",
			interfaces.Header("Content-Type", "application/json"),
			interfaces.Body(map[string]string{"message": "hello"}),
		)

		assert.Equal(t, http.StatusOK, resp.StatusCode())
		assert.Equal(t, "hello", result.Message)
	})
}

func TestCamp_Json(t *testing.T) {
	app := takibi.New(&TestBindings{})
	app.Get("/hello", func(c interfaces.IContext[TestBindings]) error {
		return c.Json(map[string]string{"message": "Hello, JSON!"})
	})

	resp := app.Camp("GET", "/hello",
		interfaces.Header("Content-Type", "application/json"),
	)
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	jsonData, err := resp.Json()
	assert.Nil(t, err)
	assert.Equal(t, "Hello, JSON!", jsonData["message"])
}
