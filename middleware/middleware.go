package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

type Config struct {
	AuthToken      string
	AllowedOrigins []string
}

func New(cfg Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := allowedOrigin(r, cfg.AllowedOrigins)
		if origin != "" {
			writeCORS(w, origin)
		}

		if isPreflight(r) {
			if origin == "" {
				http.Error(w, "cors origin not allowed", http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if requiresAuth(r) {
			if cfg.AuthToken == "" {
				http.Error(w, "authentication is not configured", http.StatusServiceUnavailable)
				return
			}
			if !validBearer(r, cfg.AuthToken) {
				w.Header().Set("WWW-Authenticate", `Bearer realm="swiki"`)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func ParseAllowedOrigins(raw string) []string {
	var origins []string
	for _, origin := range strings.Split(raw, ",") {
		origin = strings.TrimSuffix(strings.TrimSpace(origin), "/")
		if origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}

func allowedOrigin(r *http.Request, allowed []string) string {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return ""
	}
	for _, candidate := range allowed {
		if candidate == "*" || strings.EqualFold(candidate, origin) {
			return origin
		}
	}
	return ""
}

func writeCORS(w http.ResponseWriter, origin string) {
	header := w.Header()
	header.Set("Access-Control-Allow-Origin", origin)
	header.Add("Vary", "Origin")
	header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	header.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept")
	header.Set("Access-Control-Max-Age", "600")
}

func isPreflight(r *http.Request) bool {
	return r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""
}

func requiresAuth(r *http.Request) bool {
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	}

	if r.Method == http.MethodPost && (r.URL.Path == "/swiki/search" || r.URL.Path == "/swiki/page/preview") {
		return false
	}

	return true
}

func validBearer(r *http.Request, expected string) bool {
	const prefix = "bearer "
	authorization := r.Header.Get("Authorization")
	if len(authorization) <= len(prefix) || !strings.EqualFold(authorization[:len(prefix)], prefix) {
		return false
	}
	provided := strings.TrimSpace(authorization[len(prefix):])
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}
