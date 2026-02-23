package helpers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nimaeskandary/go-realworld/pkg/user/types"
	"github.com/samber/mo"

	"github.com/stretchr/testify/require"
)

func GenUser() user_types.User {
	rand := time.Now().UnixNano()
	return user_types.User{
		Id:              uuid.New(),
		Username:        fmt.Sprintf("testuser-%v", rand),
		Email:           fmt.Sprintf("testuser-%v@example.com", rand),
		Bio:             mo.Some("I am a test user"),
		Image:           mo.Some("https://example.com/testuser.jpg"),
		CreatedAtMillis: now.UnixMilli(),
		UpdatedAtMillis: now.UnixMilli(),
	}
}

func GenUserWithUpsertParams(params user_types.UpsertUserParams) user_types.User {
	return user_types.User{
		Id:              uuid.New(),
		Username:        params.Username,
		Email:           params.Email,
		Bio:             params.Bio,
		Image:           params.Image,
		CreatedAtMillis: now.UnixMilli(),
		UpdatedAtMillis: now.UnixMilli(),
	}
}

// UserMatchesUpsertParams returns a function that checks if a user matches the given UpsertUserParams, ignoring fields like Id and timestamps
func UserMatchesUpsertParams(p user_types.UpsertUserParams) func(u user_types.User) bool {
	return func(u user_types.User) bool {
		return u.Username == p.Username &&
			u.Email == p.Email &&
			u.Bio == p.Bio &&
			u.Image == p.Image
	}
}

func CreateUsers(t *testing.T, userService user_types.UserService, n int) []user_types.User {
	users := make([]user_types.User, n)
	for i := range n {
		u := GenUser()
		createdUser, err := userService.CreateUser(t.Context(), user_types.UpsertUserParams{
			Username: u.Username,
			Email:    u.Email,
			Bio:      u.Bio,
			Image:    u.Image,
		})
		require.NoError(t, err)
		t.Cleanup(func() { _ = userService.DeleteUser(context.Background(), createdUser.Id) })
		users[i] = createdUser
	}
	return users
}
