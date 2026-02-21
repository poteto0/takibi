package takibi_test

import (
	"encoding/json"
	"io"
	"net/http"
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
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "Hello, World!", string(body))
}

func TestCamp_PostWithBody(t *testing.T) {
	app := takibi.New(&TestBindings{})
	app.Post("/echo", func(c interfaces.IContext[TestBindings]) error {
		var reqBody map[string]string
		json.NewDecoder(c.Request().Body).Decode(&reqBody)
		return c.Json(reqBody)
	})

	resp := app.Camp("POST", "/echo",
		interfaces.Header("Content-Type", "application/json"),
		interfaces.Body(map[string]string{"msg": "echo"}),
	)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string]string
	json.NewDecoder(resp.Body).Decode(&respBody)
	assert.Equal(t, "echo", respBody["msg"])
}
