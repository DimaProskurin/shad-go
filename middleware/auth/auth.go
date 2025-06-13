//go:build !solution

package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

type User struct {
	Name  string
	Email string
}

type userKey struct{}

func ContextUser(ctx context.Context) (*User, bool) {
	u, ok := ctx.Value(userKey{}).(*User)
	return u, ok
}

var ErrInvalidToken = errors.New("invalid token")

type TokenChecker interface {
	CheckToken(ctx context.Context, token string) (*User, error)
}

func CheckAuth(checker TokenChecker) func(next http.Handler) http.Handler {
	mdw := func(next http.Handler) http.Handler {
		f := func(w http.ResponseWriter, r *http.Request) {
			authHdr := r.Header.Get("Authorization")
			if len(authHdr) == 0 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			token, ok := strings.CutPrefix(authHdr, "Bearer ")
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			u, err := checker.CheckToken(r.Context(), token)
			if err != nil {
				if errors.Is(err, ErrInvalidToken) {
					w.WriteHeader(http.StatusUnauthorized)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
				return
			}

			if u == nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userKey{}, u)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(f)
	}
	return mdw
}
