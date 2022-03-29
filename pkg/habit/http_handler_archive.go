package habit

import (
	"log"
	"net/http"

	"github.com/google/uuid"
)

func (h *HTTPHandler) archiveHabit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := MustGetUserID(ctx)
	hid := uuid.MustParse(r.PostFormValue("habit_uuid"))

	_, err := h.Repository.ArchiveHabit(ctx, uid, hid)
	if err != nil {
		log.Printf("Failed to archive a habit: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		log.Printf("Failed to unarchive a habit: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	h.redirect(w, "/")
}
