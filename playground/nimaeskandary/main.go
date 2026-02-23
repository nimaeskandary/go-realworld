package main

import (
	"context"
	"log"

	"github.com/nimaeskandary/go-realworld/pkg/user/types"
	"github.com/nimaeskandary/go-realworld/playground"

	"github.com/samber/mo"
)

func main() {
	ctx := context.Background()
	f, cm := playground.SetupStandardSystem(ctx)
	defer func() {
		e := recover()
		if e != nil {
			log.Printf("Panic occurred: %v", e)
		}
		cm.Cleanup()
	}()

	u, err := f.UserService.CreateUser(ctx, user_types.UpsertUserParams{
		Username: "jdoe",
		Email:    "jdoe@example.com",
		Bio:      mo.Some("I am John Doe"),
		Image:    mo.Some("https://example.com/jdoe.jpg"),
		Token:    mo.Some("token"),
	})

	if err != nil {
		panic(err)
	}

	println("Created user:", u.Id.String())
}
