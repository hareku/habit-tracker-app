package api

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	firebase "firebase.google.com/go/auth"
	"github.com/hareku/habit-tracker-app/internal/auth"
	"github.com/hareku/habit-tracker-app/internal/repository"
	"github.com/hareku/habit-tracker-app/internal/repository/repositorytest"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

func TestHTTPHandler_showHabitPage(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	uid := auth.UserID("123")
	ctx := auth.SetUserID(context.Background(), uid)

	authn := NewMockAuthenticator(ctrl)
	authn.EXPECT().GetUser(gomock.Any(), gomock.Any()).Times(1).
		Return(&firebase.UserRecord{
			UserInfo: &firebase.UserInfo{
				UID:         uid.String(),
				DisplayName: "test",
			},
		}, nil)

	repo := NewMockDynamoRepository(ctrl)

	seeder := repositorytest.NewSeeder()
	habit := seeder.SeedHabit(uid, nil)

	repo.EXPECT().FindHabit(gomock.Any(), uid, habit.ID).Times(1).Return(habit, nil)
	repo.EXPECT().ListLatestChecksWithLimit(gomock.Any(), uid, habit.ID, int32(7)).Times(1).Return([]*repository.DynamoCheck{
		seeder.SeedCheck(uid, habit.ID, "2021-01-01", nil),
		seeder.SeedCheck(uid, habit.ID, "2021-01-02", nil),
	}, nil)

	h := NewHTTPHandler(&NewHTTPHandlerInput{
		AuthMiddleware: noopMiddleware,
		CSRFMiddleware: noopMiddleware,
		Authenticator:  authn,
		Repository:     repo,
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", fmt.Sprintf("/habits/%s", habit.ID), nil)
	r = r.WithContext(ctx)
	h.ServeHTTP(w, r)

	require.Equal(t, 200, w.Result().StatusCode)
	snapshotHTML(t, w.Result().Body)
}
