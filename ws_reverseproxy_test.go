package proxy

import (
	"testing"

	"github.com/valyala/fasthttp"
)

func Test_wsCopyResponseHeader(t *testing.T) {
	type args struct {
		dst fasthttp.ResponseHeader
		src fasthttp.ResponseHeader
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wsCopyResponseHeader(tt.args.dst, tt.args.src)
		})
	}
}
