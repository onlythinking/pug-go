package server

import (
	"context"
	"fmt"
	"net/http"
)

//健康检查 为什么是z? 观看 https://vimeo.com/173610242
func HandleHealthz(hctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"status": "ok"}`)
	})
}
