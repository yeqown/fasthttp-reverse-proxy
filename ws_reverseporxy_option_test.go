package proxy

import (
	"net/http"
	"testing"

	"github.com/valyala/fasthttp"

	"github.com/stretchr/testify/assert"
)

func Test_defaultBuildOptionWS(t *testing.T) {
	dst := defaultBuildOptionWS()

	assert.Nil(t, dst.target)
	assert.Nil(t, dst.dialer)
	assert.Nil(t, dst.upgrader)
}

func Test_WithURL_OptionWS(t *testing.T) {
	assert.NotPanics(t, func() {
		WithURL_OptionWS("ws://localhost:8080/path")
	}, "could not panic")

	dst := defaultBuildOptionWS()
	WithURL_OptionWS("ws://localhost:8080/path").apply(dst)
	assert.NotNil(t, dst.target)
	assert.Equal(t, "ws", dst.target.Scheme)
	assert.Equal(t, "localhost:8080", dst.target.Host)
	assert.Equal(t, "8080", dst.target.Port())
	assert.Equal(t, "/path", dst.target.Path)
}

func Test_WithForwardHeadersHandlers_OptionWS(t *testing.T) {
	dst := defaultBuildOptionWS()
	WithForwardHeadersHandlers_OptionWS(func(reqHeader *fasthttp.RequestCtx) (forwardHeader http.Header) {
		return http.Header{
			"X-Test": []string{"X-Test"},
		}
	}).apply(dst)
	assert.NotNil(t, dst.fn)
}
