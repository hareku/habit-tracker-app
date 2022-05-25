package habit

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (h *HTTPHandler) createCheck(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	ctx := r.Context()
	uid := MustGetUserID(ctx)
	hid := uuid.MustParse(r.PostFormValue("habit_uuid"))

	date := r.PostFormValue("date")
	layout := "2006-01-02"
	if _, err := time.Parse(layout, date); err != nil {
		http.Error(w, fmt.Sprintf("Check date format must be %q", layout), http.StatusUnprocessableEntity)
		return
	}

	_, err := h.Repository.CreateCheck(ctx, uid, hid, date)
	if err != nil {
		log.Printf("Failed to create a check: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	h.redirect(w, "/")
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
