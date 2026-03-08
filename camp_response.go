package takibi

import (
	"encoding/json"
	"net/http"

	"github.com/poteto0/takibi/interfaces"
)

type campResponse struct {
	raw *http.Response
}

func newCampResponse(resp *http.Response) interfaces.ICampResponse {
	return &campResponse{
		raw: resp,
	}
}

func (c *campResponse) StatusCode() int {
	return c.raw.StatusCode
}

func (c *campResponse) Raw() *http.Response {
	return c.raw
}

func (c *campResponse) Unmarshall(v any) error {
	return json.NewDecoder(c.raw.Body).Decode(v)
}

func (c *campResponse) Json() (map[string]any, error) {
	var resp map[string]any
	if err := c.Unmarshall(&resp); err != nil {
		return nil, err
	}
	return resp, nil
}
