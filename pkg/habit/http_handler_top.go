package habit

import (
	"html/template"
	"log"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gorilla/csrf"
)

func (h *HTTPHandler) showTopPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := MustGetUserID(ctx)
	userRec, err := h.Authenticator.GetUser(ctx, uid)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	habits, err := h.Repository.AllHabits(ctx, uid)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	checks, err := h.Repository.ListLastWeekChecks(ctx, uid)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	for _, habit := range habits {
		for _, check := range checks {
			if check.HabitUUID != habit.UUID {
				continue
			}

			if habit.LatestCheck == nil || habit.LatestCheck.Date < check.Date {
				habit.LatestCheck = check
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	if err := h.tmpls["top.html"].Execute(w, struct {
		CSRFHiddenInput template.HTML
		User            *auth.UserInfo
		Habits          []*DynamoHabit
	}{
		CSRFHiddenInput: csrf.TemplateField(r),
		User:            userRec.UserInfo,
		Habits:          habits,
	}); err != nil {
		log.Printf("Failed to write index page: %s", err)
	}
}
