package api

import (
	"fmt"
	"net/http"

	"github.com/hareku/habit-tracker-app/internal/auth"
)

func (h *HTTPHandler) archiveHabit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := auth.MustGetUserID(ctx)
	hid := r.PostFormValue("habit_id")

	if err := h.Repository.ArchiveHabit(ctx, uid, hid); err != nil {
		h.handleError(w, r, fmt.Errorf("archive a habit: %w", err))
		return
	}

	h.redirect(w, "/")
}

func (h *HTTPHandler) unarchiveHabit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := auth.MustGetUserID(ctx)
	hid := r.PostFormValue("habit_id")

	if err := h.Repository.UnarchiveHabit(ctx, uid, hid); err != nil {
		h.handleError(w, r, fmt.Errorf("unarchive a habit: %w", err))
		return
	}

	h.redirect(w, "/")
}
