package habit

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"

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
		h.handleError(w, r, fmt.Errorf("all habits: %w", err))
		return
	}
	sort.Slice(habits, func(i, j int) bool {
		return habits[i].CreatedAt.UnixNano() > habits[j].CreatedAt.UnixNano()
	})

	archivedHabits, err := h.Repository.AllArchivedHabits(ctx, uid)
	if err != nil {
		h.handleError(w, r, fmt.Errorf("all archived habits: %w", err))
		return
	}

	type habit2 struct {
		*DynamoHabit
		LatestCheck *DynamoCheck
	}
	var habits2 []*habit2

	checks, err := h.Repository.ListLastWeekChecksInAllHabits(ctx, uid)
	if err != nil {
		h.handleError(w, r, fmt.Errorf("list last week checks in all habits: %w", err))
		return
	}
	for _, habit := range habits {
		h2 := &habit2{DynamoHabit: habit}
		for _, check := range checks {
			if check.HabitUUID != habit.UUID {
				continue
			}

			if h2.LatestCheck == nil || h2.LatestCheck.Date < check.Date {
				h2.LatestCheck = check
			}
		}
		habits2 = append(habits2, h2)
	}

	w.WriteHeader(http.StatusOK)
	if err := h.tmpls["top.html"].Execute(w, struct {
		CSRFHiddenInput template.HTML
		User            *auth.UserInfo
		Habits          []*habit2
		ArchivedHabits  []*DynamoHabit
	}{
		CSRFHiddenInput: csrf.TemplateField(r),
		User:            userRec.UserInfo,
		Habits:          habits2,
		ArchivedHabits:  archivedHabits,
	}); err != nil {
		log.Printf("Failed to write index page: %s", err)
	}
}
