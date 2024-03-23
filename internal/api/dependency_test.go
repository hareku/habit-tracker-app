package api

import "net/http"

//go:generate mockgen -package ${GOPACKAGE} -destination mock_${GOFILE} -source dependency.go

// noopMiddleware is a middleware that does nothing.
var noopMiddleware = func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
