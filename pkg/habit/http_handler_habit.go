package habit

import (
	"fmt"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/gorilla/csrf"
)

func (h *HTTPHandler) showHabitPage(w http.ResponseWriter, r *http.Request) {
	hid := h.mustHabitUUID(r)

	ctx := r.Context()
	uid := MustGetUserID(ctx)
	userRec, err := h.Authenticator.GetUser(ctx, uid)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	habit, err := h.Repository.FindHabit(ctx, uid, hid)
	if err != nil {
		h.handleError(w, r, fmt.Errorf("find a habit: %w", err))
		return
	}
	checks, err := h.Repository.ListLatestChecksWithLimit(ctx, uid, hid, 7)
	if err != nil {
		h.handleError(w, r, fmt.Errorf("list latest checks: %w", err))
		return
	}

	h.writePage(w, r, http.StatusOK, TemplatePageHabit, map[string]interface{}{
		"CSRFHiddenInput": csrf.TemplateField(r),
		"User":            userRec.UserInfo,
		"Habit":           habit,
		"Checks":          checks,
		"NextCheckDate": func() string {
			if len(checks) == 0 {
				return ""
			}

			latest, _ := time.Parse("2006-01-02", checks[0].Date)
			return latest.Add(24 * time.Hour).Format("2006-01-02")
		}(),
	})
}

func (h *HTTPHandler) createHabit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := MustGetUserID(ctx)

	title := r.PostFormValue("title")
	cnt := utf8.RuneCountInString(title)
	if cnt == 0 || cnt > 50 {
		http.Error(w, "Habit title length must be less than 50", http.StatusUnprocessableEntity)
		return
	}

	habit, err := h.Repository.CreateHabit(ctx, uid, title)
	if err != nil {
		h.handleError(w, r, fmt.Errorf("create a habit: %w", err))
		return
	}

	h.redirect(w, fmt.Sprintf("/habits/%s", habit.UUID))
}

func (h *HTTPHandler) updateHabit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	in := DynamoRepositoryUpdateHabitInput{
		UserID:    MustGetUserID(ctx),
		HabitUUID: uuid.MustParse(r.PostFormValue("habit_uuid")),
	}

	title := r.PostFormValue("title")
	cnt := utf8.RuneCountInString(title)
	if cnt == 0 || cnt > 50 {
		http.Error(w, "Habit title length must be less than 50", http.StatusUnprocessableEntity)
		return
	}
	in.Title = title

	if err := h.Repository.UpdateHabit(ctx, &in); err != nil {
		h.handleError(w, r, fmt.Errorf("update a habit: %w", err))
		return
	}

	h.redirect(w, fmt.Sprintf("/habits/%s", in.HabitUUID))
}

func (h *HTTPHandler) deleteHabit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := MustGetUserID(ctx)
	hid := uuid.MustParse(r.PostFormValue("habit_uuid"))

	err := h.Repository.DeleteHabit(ctx, uid, hid)
	if err != nil {
		h.handleError(w, r, fmt.Errorf("delete a habit: %w", err))
		return
	}

	h.redirect(w, "/")
}
