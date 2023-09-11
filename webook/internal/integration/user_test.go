package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"geektime-basic-go/webook/internal/web"
	"geektime-basic-go/webook/ioc"
)

func reqBuilder(t *testing.T, method, url string, body []byte, headers ...[]string) *http.Request {
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

func TestUserHandler_SendSMSLoginCode(t *testing.T) {
	const sendSMSCodeUrl = "/users/login_sms/code/send"
	server := InitWebServer()
	rdb := ioc.InitRedis()
	testCases := []struct {
		name string

		before func(t *testing.T)
		after  func(t *testing.T)
		body   []byte

		wantCode   int
		wantResult web.Result
	}{
		{
			name:   "发送成功",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx := context.Background()
				key := "phone_code:login:13888888888"

				ttl, err := rdb.TTL(ctx, key).Result()
				require.NoError(t, err)
				assert.True(t, ttl > 9*time.Minute)

				val, err := rdb.GetDel(ctx, key).Result()
				require.NoError(t, err)
				assert.True(t, len(val) == 6)
			},
			body:       []byte(`{"phone": "13888888888"}`),
			wantCode:   http.StatusOK,
			wantResult: web.Result{Msg: "发送成功"},
		},
		{
			name:       "空手机号",
			before:     func(t *testing.T) {},
			after:      func(t *testing.T) {},
			body:       []byte(`{"phone": ""}`),
			wantCode:   http.StatusOK,
			wantResult: web.Result{Code: 4, Msg: "手机号码错误"},
		},
		{
			name: "发送太频繁",
			before: func(t *testing.T) {
				ctx := context.Background()
				key := "phone_code:login:13888888888"

				err := rdb.Set(ctx, key, "123456", 9*time.Minute+40*time.Second).Err()
				require.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx := context.Background()
				key := "phone_code:login:13888888888"

				val, err := rdb.GetDel(ctx, key).Result()
				require.NoError(t, err)
				assert.Equal(t, "123456", val)
			},
			body:       []byte(`{"phone": "13888888888"}`),
			wantCode:   http.StatusOK,
			wantResult: web.Result{Code: 4, Msg: "短信发送太频繁，请稍后再试"},
		},
		{
			name: "未知错误",
			before: func(t *testing.T) {
				ctx := context.Background()
				key := "phone_code:login:13888888888"

				err := rdb.Set(ctx, key, "123456", 0).Err()
				require.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx := context.Background()
				key := "phone_code:login:13888888888"

				val, err := rdb.GetDel(ctx, key).Result()
				require.NoError(t, err)
				assert.Equal(t, "123456", val)
			},
			body:       []byte(`{"phone": "13888888888"}`),
			wantCode:   http.StatusOK,
			wantResult: web.Result{Code: 5, Msg: "系统错误"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)

			req := reqBuilder(t, http.MethodPost, sendSMSCodeUrl, tc.body)
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, req)

			var webRes web.Result
			err := json.NewDecoder(recorder.Body).Decode(&webRes)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantResult, webRes)
			tc.after(t)
		})
	}

}
