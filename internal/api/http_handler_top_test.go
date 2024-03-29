package api

import (
	"context"
	"net/http/httptest"
	"testing"

	firebase "firebase.google.com/go/auth"
	"github.com/hareku/habit-tracker-app/internal/auth"
	"github.com/hareku/habit-tracker-app/internal/repository"
	"github.com/hareku/habit-tracker-app/internal/repository/repositorytest"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

func TestHTTPHandler_showTopPage(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	ctx := context.Background()
	uid := auth.UserID("123")
	ctx = auth.SetUserID(ctx, uid)

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
	habits := []*repository.DynamoHabit{
		seeder.SeedHabit(uid, nil),
		seeder.SeedHabit(uid, nil),
	}

	repo.EXPECT().AllHabits(gomock.Any(), gomock.Any()).Times(1).Return(habits, nil)
	repo.EXPECT().AllArchivedHabits(gomock.Any(), gomock.Any()).Times(1).Return(nil, nil)
	repo.EXPECT().ListLastWeekChecksInAllHabits(gomock.Any(), gomock.Any()).Times(1).Return([]*repository.DynamoCheck{
		seeder.SeedCheck(uid, habits[0].ID, "2021-01-01", nil),
	}, nil)

	h := NewHTTPHandler(&NewHTTPHandlerInput{
		AuthMiddleware: noopMiddleware,
		CSRFMiddleware: noopMiddleware,
		Authenticator:  authn,
		Repository:     repo,
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r = r.WithContext(ctx)
	h.ServeHTTP(w, r)

	require.Equal(t, 200, w.Result().StatusCode)
	snapshotHTML(t, w.Result().Body)
}
