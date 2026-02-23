package fixtures

import (
	"context"
	"testing"

	http_handler_types "github.com/nimaeskandary/go-realworld/cmd/http_server/app/http_handler/types"
	article_types "github.com/nimaeskandary/go-realworld/pkg/article/types"
	auth_types "github.com/nimaeskandary/go-realworld/pkg/auth/types"
	user_types "github.com/nimaeskandary/go-realworld/pkg/user/types"
	"github.com/nimaeskandary/go-realworld/pkg/util"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

type StandardFixture struct {
	ArticleRepo    article_types.ArticleRepository
	ArticleService article_types.ArticleService
	AuthService    auth_types.AuthService
	HttpHandler    http_handler_types.HttpHandler
	UserRepo       user_types.UserRepository
	UserService    user_types.UserService
}

// SetupStandardFixture sets up a standard fixture with all dependencies injected.
// This will create new test databases which is expensive, so err on the side of using once per test file, not per sub test.
//
// To use overrides, pass in functions that return the type being overriden in the dependency tree,
// this will replace the default constructor for that type, e.g.
// override1 := func(dep SomeDepFromTree, etc) TypeBeingOverriden { return myNewmockImplementation(dep1) }
// SetupStandardFixture(t, override1)
func SetupStandardFixture(t *testing.T, overrides ...any) StandardFixture {
	testModules, err := AllTestModules(t)
	require.NoError(t, err)

	for _, constructorFn := range overrides {
		testModules = append(testModules, fx.Decorate(constructorFn))
	}

	f := StandardFixture{}

	fxApp := util.CreateFxAppAndExtract(
		testModules,
		&f.ArticleRepo,
		&f.ArticleService,
		&f.AuthService,
		&f.HttpHandler,
		&f.UserRepo,
		&f.UserService,
	)

	require.NoError(t, fxApp.Start(t.Context()))
	t.Cleanup(func() { _ = fxApp.Stop(context.Background()) })

	return f
}
