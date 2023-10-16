package integration

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func ReqBuilder(t *testing.T, method, url string, body []byte, headers ...[]string) *http.Request {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	require.NoError(t, err)

	for _, header := range headers {
		req.Header.Set(header[0], header[1])
	}

	if len(headers) < 1 {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}
