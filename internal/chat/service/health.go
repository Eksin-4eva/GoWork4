package service

import "net/http"

// HandleHealth 是给容器健康检查用的最小端点。
func (s *Service) HandleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
