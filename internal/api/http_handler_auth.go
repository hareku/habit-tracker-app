package api

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gorilla/csrf"
	"github.com/hareku/habit-tracker-app/internal/auth"
)

func (h *HTTPHandler) showLoginPage(w http.ResponseWriter, r *http.Request) {
	h.writePage(w, r, http.StatusOK, TemplatePageLogin, map[string]interface{}{
		"CSRFHiddenInput": csrf.TemplateField(r),
	})
}

func (h *HTTPHandler) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		MaxAge: 0,
	})

	h.redirect(w, "/login")
}

func (h *HTTPHandler) deleteAccount(w http.ResponseWriter, r *http.Request) {
	if err := h.Authenticator.DeleteUser(r.Context(), auth.MustGetUserID(r.Context())); err != nil {
		h.handleError(w, r, fmt.Errorf("delete account: %w", err))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		MaxAge: 0,
	})

	h.redirect(w, "/login")
}

func (h *HTTPHandler) storeSessionCookie(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	tk := r.PostFormValue("idToken")
	if tk == "" {
		http.Error(w, "missing idToken in the request body.", http.StatusBadRequest)
		return
	}

	cookie, err := h.Authenticator.SessionCookie(r.Context(), tk)
	if err != nil {
		http.Error(w, "invalid idToken.", http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    cookie,
		MaxAge:   int((time.Hour * 24 * 14).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.Secure,
	})

	h.redirect(w, "/")
}

// https://firebase.google.com/docs/auth/web/redirect-best-practices?hl=ja&authuser=1#proxy-requests
func (h *HTTPHandler) handleFirebaseAuth(w http.ResponseWriter, r *http.Request) {
	firebaseAuthDomain := "https://habittrackerapp-1da2d.firebaseapp.com"
	target, err := url.Parse(firebaseAuthDomain)
	if err != nil {
		h.handleError(w, r, fmt.Errorf("parse firebaseAuthDomain: %w", err))
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
	}

	proxy.ServeHTTP(w, r)
}
