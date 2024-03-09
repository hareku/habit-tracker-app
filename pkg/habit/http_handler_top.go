package habit

import (
	"fmt"
	"net/http"
	"sort"

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

	h.writePage(w, r, http.StatusOK, TemplatePageTop, map[string]interface{}{
		"CSRFHiddenInput": csrf.TemplateField(r),
		"User":            userRec.UserInfo,
		"Habits":          habits,
		"ArchivedHabits":  archivedHabits,
	})
}
