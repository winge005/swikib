package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuth(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	tests := []struct {
		name       string
		method     string
		path       string
		authToken  string
		authorized string
		want       int
	}{
		{name: "read without token", method: http.MethodGet, path: "/swiki/pages/1", authToken: "secret", want: http.StatusNoContent},
		{name: "search without token", method: http.MethodPost, path: "/swiki/search", authToken: "secret", want: http.StatusNoContent},
		{name: "preview without token", method: http.MethodPost, path: "/swiki/page/preview", authToken: "secret", want: http.StatusNoContent},
		{name: "mutation without token", method: http.MethodPost, path: "/swiki/pages", authToken: "secret", want: http.StatusUnauthorized},
		{name: "mutation with token", method: http.MethodPost, path: "/swiki/pages", authToken: "secret", authorized: "Bearer secret", want: http.StatusNoContent},
		{name: "delete with token", method: http.MethodDelete, path: "/swiki/pages/1", authToken: "secret", authorized: "Bearer secret", want: http.StatusNoContent},
		{name: "wrong token", method: http.MethodDelete, path: "/swiki/pages/1", authToken: "secret", authorized: "Bearer wrong", want: http.StatusUnauthorized},
		{name: "auth not configured", method: http.MethodPost, path: "/swiki/pages", want: http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.authorized != "" {
				req.Header.Set("Authorization", tt.authorized)
			}
			rec := httptest.NewRecorder()

			New(Config{AuthToken: tt.authToken}, next).ServeHTTP(rec, req)

			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestCORS(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := New(Config{
		AuthToken:      "secret",
		AllowedOrigins: []string{"https://wiki.example"},
	}, next)

	t.Run("allowed origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/swiki/pages", nil)
		req.Header.Set("Origin", "https://wiki.example")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://wiki.example" {
			t.Fatalf("Access-Control-Allow-Origin = %q", got)
		}
		if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "" {
			t.Fatalf("credentials header should be unset, got %q", got)
		}
	})

	t.Run("disallowed origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/swiki/pages", nil)
		req.Header.Set("Origin", "https://evil.example")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("Access-Control-Allow-Origin = %q, want empty", got)
		}
	})

	t.Run("preflight allowed origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/swiki/pages", nil)
		req.Header.Set("Origin", "https://wiki.example")
		req.Header.Set("Access-Control-Request-Method", http.MethodPost)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
	})

	t.Run("preflight disallowed origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/swiki/pages", nil)
		req.Header.Set("Origin", "https://evil.example")
		req.Header.Set("Access-Control-Request-Method", http.MethodPost)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
		}
	})
}
