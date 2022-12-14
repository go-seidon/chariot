package restmiddleware

import (
	"net/http"

	"github.com/go-seidon/chariot/api/restapp"
	"github.com/go-seidon/chariot/internal/session"
	"github.com/go-seidon/provider/serialization"
	"github.com/go-seidon/provider/status"
)

type sessionAuth struct {
	sessionClient session.Session
	serializer    serialization.Serializer
	feature       string
}

func (m *sessionAuth) Handle(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.FormValue("token")
		if token == "" {
			token = r.URL.Query().Get("token")
		}

		if token == "" {
			response := &restapp.ResponseBodyInfo{
				Code:    status.ACTION_FORBIDDEN,
				Message: "token is not specified",
			}
			info, _ := m.serializer.Marshal(response)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write(info)
			return
		}

		_, err := m.sessionClient.VerifySession(r.Context(), session.VerifySessionParam{
			Token:   token,
			Feature: m.feature,
		})
		if err != nil {
			response := &restapp.ResponseBodyInfo{
				Code:    status.ACTION_FORBIDDEN,
				Message: err.Message,
			}
			info, _ := m.serializer.Marshal(response)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write(info)
			return
		}

		h.ServeHTTP(w, r)
	})
}

type SessionAuthParam struct {
	SessionClient session.Session
	Serializer    serialization.Serializer
	Feature       string
}

func NewSessionAuth(p SessionAuthParam) *sessionAuth {
	return &sessionAuth{
		sessionClient: p.SessionClient,
		serializer:    p.Serializer,
		feature:       p.Feature,
	}
}
