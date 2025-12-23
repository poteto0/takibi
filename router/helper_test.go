package router

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_hasParamPrefic(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "has param prefix",
			path:     ":foo",
			expected: true,
		},
		{
			name:     "not has param prefix",
			path:     "foo",
			expected: false,
		},
		{
			name:     "empty string",
			path:     "",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, hasPathParamPrefix(test.path))
		})
	}
}

func Test_isSupportedHttpMethod(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		expected bool
	}{
		{
			name:     "not supported method",
			method:   "FOO",
			expected: false,
		},
		{
			name:     "empty method",
			method:   "",
			expected: false,
		},
		{
			name:     "supported method Get",
			method:   http.MethodGet,
			expected: true,
		},
		{
			name:     "supported method Post",
			method:   http.MethodPost,
			expected: true,
		},
		{
			name:     "supported method Put",
			method:   http.MethodPut,
			expected: true,
		},
		{
			name:     "supported method Patch",
			method:   http.MethodPatch,
			expected: true,
		},
		{
			name:     "supported method Delete",
			method:   http.MethodDelete,
			expected: true,
		},
		{
			name:     "supported method Head",
			method:   http.MethodHead,
			expected: true,
		},
		{
			name:     "supported method Options",
			method:   http.MethodOptions,
			expected: true,
		},
		{
			name:     "supported method Trace",
			method:   http.MethodTrace,
			expected: true,
		},
		{
			name:     "supported method Connect",
			method:   http.MethodConnect,
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, isSupportedHttpMethod(test.method))
		})
	}
}
