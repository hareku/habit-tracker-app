package habit

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"path"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	formmethod "github.com/hareku/form-method-go"
	slogchi "github.com/samber/slog-chi"
)

type NewHTTPHandlerInput struct {
	AuthMiddleware Middleware
	CSRFMiddleware Middleware
	Authenticator  *FirebaseAuthenticator
	Repository     *DynamoRepository
	Secure         bool
}

type HTTPHandler struct {
	Authenticator *FirebaseAuthenticator
	Repository    *DynamoRepository
	Secure        bool

	mux   *chi.Mux
	tmpls map[TypeTemplatePage]*template.Template
}

func NewHTTPHandler(in *NewHTTPHandlerInput) *HTTPHandler {
	h := &HTTPHandler{
		Authenticator: in.Authenticator,
		Repository:    in.Repository,
		Secure:        in.Secure,
	}

	common := template.Must(template.ParseFS(templates, "templates/_*.html")).
		Funcs(template.FuncMap{
			"method_field": func(method string) template.HTML {
				return formmethod.TemplateField(method)
			},
		})
	tmpls := map[TypeTemplatePage]*template.Template{}
	for _, page := range ListPages() {
		tmpls[TypeTemplatePage(page)] = template.Must(
			template.Must(common.Clone()).
				ParseFS(templates, path.Join("templates", page)),
		)
	}
	h.tmpls = tmpls

	r := chi.NewMux()
	r.Use(formmethod.Middleware)
	r.Use(slogchi.New(slog.Default()))
	r.Use(middleware.Recoverer)
	r.Use(in.CSRFMiddleware)

	r.Group(func(r chi.Router) {
		r.Use(in.AuthMiddleware)
		r.Get("/", h.showTopPage)
		r.Get("/habits/{habitUUID}", h.showHabitPage)
		r.Post("/habits", h.createHabit)
		r.Post("/checks", h.createCheck)
		r.Post("/update-habit", h.updateHabit)
		r.Post("/archive-habit", h.archiveHabit)
		r.Post("/unarchive-habit", h.unarchiveHabit)
		r.Post("/delete-habit", h.deleteHabit)
		r.Delete("/habits/{habitUUID}/checks", h.deleteCheck)
		r.Post("/logout", h.logout)
		r.Post("/delete-account", h.deleteAccount)
	})
	r.Get("/login", h.showLoginPage)
	r.Post("/session-cookie", h.storeSessionCookie)
	h.mux = r

	return h
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *HTTPHandler) redirect(w http.ResponseWriter, loc string) {
	w.Header().Set("Location", loc)
	w.WriteHeader(http.StatusFound)
}

func (h *HTTPHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	slog.ErrorContext(r.Context(), err.Error())
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (h *HTTPHandler) writePage(w http.ResponseWriter, r *http.Request, status int, page TypeTemplatePage, data interface{}) {
	var buf bytes.Buffer // write to buffer first to prevent partial writes
	if err := h.tmpls[page].Execute(&buf, data); err != nil {
		h.handleError(w, r, fmt.Errorf("execute template: %w", err))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if _, err := buf.WriteTo(w); err != nil {
		h.handleError(w, r, fmt.Errorf("write page to response: %w", err))
	}
}
func (h *HTTPHandler) mustHabitUUID(r *http.Request) uuid.UUID {
	str := chi.URLParam(r, "habitUUID")
	if str == "" {
		panic("habitUUID is empty")
	}
	return uuid.MustParse(str)
}
