package habit

import (
	"embed"
	"html/template"
	"net/http"
	"path"

	"github.com/go-chi/chi/v5"
)

//go:embed templates/*
var templates embed.FS

type NewHTTPHandlerInput struct {
	AuthMiddleware Middleware
	CSRFMiddleware Middleware
	Authenticator  *FirebaseAuthenticator
	Repository     *DynamoRepository
	Secure         bool
}

type HTTPHandler struct {
	Chi           *chi.Mux
	Authenticator *FirebaseAuthenticator
	Repository    *DynamoRepository
	Secure        bool
	tmpls         map[string]*template.Template
}

func NewHTTPHandler(in *NewHTTPHandlerInput) *HTTPHandler {
	h := &HTTPHandler{
		Authenticator: in.Authenticator,
		Repository:    in.Repository,
		Secure:        in.Secure,
	}

	common := template.Must(template.ParseFS(templates, "templates/_*.html"))
	tmpls := map[string]*template.Template{}
	pages := []string{"top.html", "login.html", "habit.html"}
	for _, page := range pages {
		tmpls[page] = template.Must(
			template.Must(common.Clone()).
				ParseFS(templates, path.Join("templates", page)),
		)
	}
	h.tmpls = tmpls

	r := chi.NewMux()
	r.Use(in.CSRFMiddleware)

	r.Group(func(r chi.Router) {
		r.Use(in.AuthMiddleware)
		r.Get("/", h.showTopPage)
		r.Get("/habits/{habitUUID}", h.showHabitPage)
		r.Post("/habits", h.createHabit)
		r.Post("/checks", h.createCheck)
		r.Post("/archive-habit", h.archiveHabit)
		r.Post("/unarchive-habit", h.unarchiveHabit)
		r.Post("/delete-habit", h.deleteHabit)
		r.Post("/delete-check", h.deleteCheck)
		r.Post("/logout", h.logout)
		r.Post("/delete-account", h.deleteAccount)
	})
	r.Get("/login", h.showLoginPage)
	r.Post("/session-cookie", h.storeSessionCookie)
	h.Chi = r

	return h
}

func (h *HTTPHandler) redirect(w http.ResponseWriter, loc string) {
	w.Header().Set("Location", loc)
	w.WriteHeader(http.StatusFound)
}
