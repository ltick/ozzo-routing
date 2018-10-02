package proxy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ltick/tick-routing"
	"github.com/stretchr/testify/assert"
)

func TestProxy(t *testing.T) {
	const backendResponse = "I am the backend"
	const backendStatus = 200
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(backendStatus)
		w.Write([]byte(backendResponse))
	}))
	defer backend.Close()
	backendURL, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatal(err)
	}
	h := ProxyHandler([]*Proxy{
		&Proxy{HostRule: "(www)?.example.com", MethodRule: "GET", UriRule: "/(\\w+).html", UpstreamURL: backendURL, UpstreamHeader: &http.Header{}},
	})
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://www.example.com/a.html", nil)
	c := routing.NewContext(res, req)
	err = h(c)
	assert.Nil(t, err, "return value is nil")
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "I am the backend", res.Body.String())

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "http://www.example.com/a", nil)
	c = routing.NewContext(res, req)
	err = h(c)
	assert.Nil(t, err, "return value is nil")
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "", res.Body.String())

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "http://example.com/a.html", nil)
	c = routing.NewContext(res, req)
	err = h(c)
	assert.Nil(t, err, "return value is nil")
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "", res.Body.String())
}
