package internal

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/nimaeskandary/go-realworld/pkg/database/types"
	"github.com/nimaeskandary/go-realworld/pkg/observability/types"
	"github.com/nimaeskandary/go-realworld/pkg/user/types"

	"github.com/google/uuid"
	"github.com/samber/mo"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/scan"
)

const (
	usersTableName     = "users"
	followersTableName = "user_followers"
)

type postgresUserRepo struct {
	db     db_types.PostgresRealWorldAppDb
	logger obs_types.Logger
}

func NewPostgresUserRepository(db db_types.PostgresRealWorldAppDb, logger obs_types.Logger) user_types.UserRepository {
	return &postgresUserRepo{db: db, logger: logger}
}

func (r *postgresUserRepo) UpsertUser(ctx context.Context, user user_types.User) (user_types.User, error) {
	createCols := []string{"id", "username", "email", "bio", "image", "created_at", "updated_at"}
	updateCols := []string{"username", "email", "bio", "image", "updated_at"}
	q := psql.Insert(
		im.Into(usersTableName, createCols...),
		im.Values(
			psql.Arg(
				user.Id,
				user.Username,
				user.Email,
				user.Bio,
				user.Image,
				time.UnixMilli(user.CreatedAtMillis),
				time.UnixMilli(user.UpdatedAtMillis),
			),
		),
		im.OnConflict("id").DoUpdate(
			im.SetExcluded(updateCols...),
		),
		im.Returning("*"),
	)

	result, err := bob.One(ctx, bob.NewDB(r.db.GetDB()), q, scan.StructMapper[postgresUser]())
	if err != nil {
		return user_types.User{}, fmt.Errorf("error with user upsert query, user=%v: %w", user, err)
	}

	return fromPostgresUser(result), nil
}

func (r *postgresUserRepo) DeleteUser(ctx context.Context, id uuid.UUID) error {
	q := psql.Delete(
		dm.From(usersTableName),
		dm.Where(psql.Quote("id").EQ(psql.Arg(id.String()))),
	)

	_, err := bob.Exec(ctx, bob.NewDB(r.db.GetDB()), q)
	if err != nil {
		return fmt.Errorf("error with user delete query, id=%v: %w", id.String(), err)
	}
	return nil
}

func (r *postgresUserRepo) GetUserByUsername(ctx context.Context, username string) (mo.Option[user_types.User], error) {
	q := psql.Select(
		sm.Columns("*"),
		sm.From(usersTableName),
		sm.Where(psql.Quote("username").EQ(psql.Arg(username))),
	)

	result, err := bob.One(ctx, bob.NewDB(r.db.GetDB()), q, scan.StructMapper[postgresUser]())
	if err != nil {
		if err == sql.ErrNoRows {
			return mo.None[user_types.User](), nil
		}
		return mo.None[user_types.User](), fmt.Errorf("error with get user by username query, username=%v: %w", username, err)
	}

	return mo.Some(fromPostgresUser(result)), nil
}

func (r *postgresUserRepo) GetUserById(ctx context.Context, id uuid.UUID) (mo.Option[user_types.User], error) {
	q := psql.Select(
		sm.Columns("*"),
		sm.From(usersTableName),
		sm.Where(psql.Quote("id").EQ(psql.Arg(id.String()))),
	)

	result, err := bob.One(ctx, bob.NewDB(r.db.GetDB()), q, scan.StructMapper[postgresUser]())
	if err != nil {
		if err == sql.ErrNoRows {
			return mo.None[user_types.User](), nil
		}
		return mo.None[user_types.User](), fmt.Errorf("error with get user by id query, id=%v: %w", id.String(), err)
	}

	return mo.Some(fromPostgresUser(result)), nil
}

func (r *postgresUserRepo) GetUserByEmail(ctx context.Context, email string) (mo.Option[user_types.User], error) {
	q := psql.Select(
		sm.Columns("*"),
		sm.From(usersTableName),
		sm.Where(psql.Quote("email").EQ(psql.Arg(email))),
	)

	result, err := bob.One(ctx, bob.NewDB(r.db.GetDB()), q, scan.StructMapper[postgresUser]())
	if err != nil {
		if err == sql.ErrNoRows {
			return mo.None[user_types.User](), nil
		}
		return mo.None[user_types.User](), fmt.Errorf("error with get user by email query, email=%v: %w", email, err)
	}

	return mo.Some(fromPostgresUser(result)), nil
}

func (r *postgresUserRepo) IsFollowing(ctx context.Context, followedByUserId uuid.UUID, followingUserId uuid.UUID) (bool, error) {
	q, args, err := psql.Select(
		sm.Columns(psql.F("EXISTS",
			psql.Select(
				sm.From(followersTableName),
				sm.Columns("followed_by_user_id"),
				sm.Where(
					psql.And(
						psql.Quote("followed_by_user_id").EQ(psql.Arg(followedByUserId)),
						psql.Quote("following_user_id").EQ(psql.Arg(followingUserId)),
					),
				),
			),
		)),
	).Build(ctx)

	if err != nil {
		return false, fmt.Errorf("error building is following query: %w", err)
	}

	var exists bool
	err = r.db.GetDB().QueryRowContext(ctx, q, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error executing is following query: %w", err)
	}

	return exists, nil
}

func (r *postgresUserRepo) Follow(ctx context.Context, followedByUserId uuid.UUID, followingUserId uuid.UUID) error {
	q := psql.Insert(
		im.Into(followersTableName, "followed_by_user_id", "following_user_id"),
		im.Values(psql.Arg(followedByUserId), psql.Arg(followingUserId)),
		im.OnConflict("followed_by_user_id", "following_user_id").DoNothing(),
	)

	_, err := bob.Exec(ctx, bob.NewDB(r.db.GetDB()), q)

	if err != nil {
		return fmt.Errorf("error executing follow query: %w", err)
	}

	return nil
}

func (r *postgresUserRepo) Unfollow(ctx context.Context, followedByUserId uuid.UUID, followingUserId uuid.UUID) error {
	q := psql.Delete(
		dm.From(followersTableName),
		dm.Where(
			psql.And(
				psql.Quote("followed_by_user_id").EQ(psql.Arg(followedByUserId)),
				psql.Quote("following_user_id").EQ(psql.Arg(followingUserId)),
			),
		),
	)

	_, err := bob.Exec(ctx, bob.NewDB(r.db.GetDB()), q)
	if err != nil {
		return fmt.Errorf("error executing unfollow query: %w", err)
	}

	return nil
}

type postgresUser struct {
	Id        uuid.UUID         `db:"id"`
	Username  string            `db:"username"`
	Email     string            `db:"email"`
	Bio       mo.Option[string] `db:"bio"`
	Image     mo.Option[string] `db:"image"`
	CreatedAt time.Time         `db:"created_at"`
	UpdatedAt time.Time         `db:"updated_at"`
}

func fromPostgresUser(from postgresUser) user_types.User {
	return user_types.User{
		Id:              from.Id,
		Username:        from.Username,
		Email:           from.Email,
		Bio:             from.Bio,
		Image:           from.Image,
		CreatedAtMillis: from.CreatedAt.UnixMilli(),
		UpdatedAtMillis: from.UpdatedAt.UnixMilli(),
	}
}
