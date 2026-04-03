package server

import (
	"fmt"
	"net/http"

	"gobili/chat/internal/config"
	"gobili/chat/internal/service"
)

func New(cfg config.Config, svc *service.Service) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", svc.HandleWebSocket)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler: mux,
	}
}
