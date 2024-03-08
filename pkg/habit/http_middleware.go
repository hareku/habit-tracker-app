package habit

import (
	"net/http"

	"github.com/gorilla/csrf"
)

type Middleware func(next http.Handler) http.Handler

func NewAuthMiddleware(authenticator *FirebaseAuthenticator) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, err := r.Cookie("session")
			if err != nil {
				redirect(w, "/login")
				return
			}

			ctx, err := authenticator.Authenticate(r.Context(), sess.Value)
			if err != nil {
				redirect(w, "/login")
				return
			}
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

func redirect(w http.ResponseWriter, loc string) {
	w.Header().Set("Location", loc)
	w.WriteHeader(http.StatusFound)
}

func NewCSRFMiddleware(key []byte, secure bool) Middleware {
	hnd := csrf.Protect(
		key,
		csrf.Secure(secure),
		csrf.Path("/"), // to prevent storing the cookie in a subpath
		csrf.TrustedOrigins([]string{"localhost:3000"}),
	)
	return func(next http.Handler) http.Handler {
		return hnd(next)
	}
}
