package habit

import (
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
)

func (h *HTTPHandler) createCheck(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	ctx := r.Context()
	uid := MustGetUserID(ctx)
	hid := uuid.MustParse(r.PostFormValue("habit_uuid"))
	date := r.PostFormValue("date")

	_, err := h.Repository.CreateCheck(ctx, uid, hid, date)
	if err != nil {
		log.Printf("Failed to create a check: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	h.redirect(w, fmt.Sprintf("/habits/%s", hid))
}

func (h *HTTPHandler) deleteCheck(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	ctx := r.Context()
	uid := MustGetUserID(ctx)
	hid := uuid.MustParse(r.PostFormValue("habit_uuid"))
	date := r.PostFormValue("date")

	err := h.Repository.DeleteCheck(ctx, uid, hid, date)
	if err != nil {
		log.Printf("Failed to delete a check: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	h.redirect(w, fmt.Sprintf("/habits/%s", hid))
}
