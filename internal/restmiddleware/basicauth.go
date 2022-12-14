package restmiddleware

import (
	"net/http"
	"strings"

	"github.com/go-seidon/chariot/api/restapp"
	"github.com/go-seidon/chariot/internal/auth"
	"github.com/go-seidon/provider/serialization"
	"github.com/go-seidon/provider/status"
)

type basicAuth struct {
	basicClient auth.BasicAuth
	serializer  serialization.Serializer
}

func (m *basicAuth) Handle(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auths := strings.Split(r.Header.Get("Authorization"), "Basic ")
		if len(auths) != 2 {
			response := &restapp.ResponseBodyInfo{
				Code:    status.ACTION_FORBIDDEN,
				Message: "credential is not specified",
			}
			info, _ := m.serializer.Marshal(response)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(info)
			return
		}

		credential, err := m.basicClient.CheckCredential(r.Context(), auth.CheckCredentialParam{
			AuthToken: auths[1],
		})
		if err != nil {
			response := &restapp.ResponseBodyInfo{
				Code:    status.ACTION_FORBIDDEN,
				Message: "failed check credential",
			}
			info, _ := m.serializer.Marshal(response)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(info)
			return
		}

		if !credential.IsValid() {
			response := &restapp.ResponseBodyInfo{
				Code:    status.ACTION_FORBIDDEN,
				Message: "credential is invalid",
			}
			info, _ := m.serializer.Marshal(response)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(info)
			return
		}

		h.ServeHTTP(w, r)
	})
}

type BasicAuthParam struct {
	BasicClient auth.BasicAuth
	Serializer  serialization.Serializer
}

func NewBasicAuth(p BasicAuthParam) *basicAuth {
	return &basicAuth{
		basicClient: p.BasicClient,
		serializer:  p.Serializer,
	}
}
