package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/sonyamoonglade/poison-tg/pkg/logger"
	"go.uber.org/zap"
)

const baseURL = "http://localhost:8000"

func readBody(rc io.ReadCloser) []byte {
	b, err := io.ReadAll(rc)
	if err != nil {
		panic(err)
	}
	rc.Close()
	return b
}

func newBody(b interface{}) io.Reader {
	bodyBytes, err := json.Marshal(b)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(bodyBytes)
}

func newJsonRequest(method, url string, body any) *http.Request {
	req, _ := http.NewRequest(method, buildURL(url), newBody(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func printResponseDetails(res *http.Response) {
	body := res.Body
	bits, err := io.ReadAll(body)
	defer body.Close()
	if err != nil {
		panic(err)
	}
	res.Body = io.NopCloser(bytes.NewReader(bits))
	logger.Get().Info("completed test req",
		zap.String("X-Request-ID", res.Header.Get("X-Request-ID")),
		zap.String("URL", res.Request.URL.Path),
		zap.String("Body", string(bits)))
}

func printResponseBody(body io.ReadCloser) {
	defer body.Close()
	b, err := io.ReadAll(body)
	if err != nil {
		panic(err)
	}
	logger.Get().Sugar().Debugf("response body: %s\n", string(b))
}

func buildURL(path string) string {
	return baseURL + path
}

func StringPtr(s string) *string { return &s }

func IntPtr[N int | int8 | int16 | int32 | int64](n N) *N { return &n }
