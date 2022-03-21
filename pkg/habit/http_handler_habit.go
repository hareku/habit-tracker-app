package habit

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
)

func (h *HTTPHandler) showHabitPage(w http.ResponseWriter, r *http.Request) {
	hidStr := chi.URLParam(r, "habitUUID")
	if hidStr == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	hid := uuid.MustParse(hidStr)

	ctx := r.Context()
	uid := MustGetUserID(ctx)
	userRec, err := h.Authenticator.GetUser(ctx, uid)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	habit, err := h.Repository.FindHabit(ctx, uid, hid)
	if err != nil {
		log.Printf("Failed to find a habit[%s]: %s", hid, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	checks, err := h.Repository.ListChecks(ctx, uid, hid)
	if err != nil {
		log.Printf("Failed to list checks of habit[%s]: %s", hid, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := h.tmpls["habit.html"].Execute(w, struct {
		CSRFHiddenInput template.HTML
		User            *auth.UserInfo
		Habit           *DynamoHabit
		Checks          []*DynamoCheck
	}{
		CSRFHiddenInput: csrf.TemplateField(r),
		User:            userRec.UserInfo,
		Habit:           habit,
		Checks:          checks,
	}); err != nil {
		log.Printf("Failed to write habit page: %s", err)
	}
}

func (h *HTTPHandler) createHabit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := MustGetUserID(ctx)

	title := r.PostFormValue("title")
	if title == "" {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	habit, err := h.Repository.CreateHabit(ctx, uid, title)
	if err != nil {
		log.Printf("Failed to create a habit: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	h.redirect(w, fmt.Sprintf("/habits/%s", habit.UUID))
}

func (h *HTTPHandler) deleteHabit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := MustGetUserID(ctx)
	hid := uuid.MustParse(r.PostFormValue("habit_uuid"))

	err := h.Repository.DeleteHabit(ctx, uid, hid)
	if err != nil {
		log.Printf("Failed to delete a habit: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	h.redirect(w, "/")
}
