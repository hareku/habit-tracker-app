package habit

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

func (h *HTTPHandler) archiveHabit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := MustGetUserID(ctx)
	hid := uuid.MustParse(r.PostFormValue("habit_uuid"))

	_, err := h.Repository.ArchiveHabit(ctx, uid, hid)
	if err != nil {
		h.handleError(w, r, fmt.Errorf("archive a habit: %w", err))
		return
	}

	h.redirect(w, "/")
}

func (h *HTTPHandler) unarchiveHabit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := MustGetUserID(ctx)
	hid := uuid.MustParse(r.PostFormValue("habit_uuid"))

	_, err := h.Repository.UnarchiveHabit(ctx, uid, hid)
	if err != nil {
		h.handleError(w, r, fmt.Errorf("unarchive a habit: %w", err))
		return
	}

	h.redirect(w, "/")
}
