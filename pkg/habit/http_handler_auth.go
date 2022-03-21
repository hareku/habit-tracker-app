package habit

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/csrf"
)

func (h *HTTPHandler) showLoginPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if err := h.tmpls["login.html"].Execute(w, struct {
		CSRFHiddenInput template.HTML
	}{
		CSRFHiddenInput: csrf.TemplateField(r),
	}); err != nil {
		log.Printf("Failed to write login page: %s", err)
	}
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
	if err := h.Authenticator.DeleteUser(r.Context(), MustGetUserID(r.Context())); err != nil {
		http.Error(w, "Failed to delete your account.", http.StatusInternalServerError)
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
