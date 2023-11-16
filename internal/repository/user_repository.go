package repository

import (
	"context"
	"fmt"

	"go-form-hub/internal/database"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type User struct {
	ID        int64   `db:"id"`
	Username  string  `db:"username"`
	FirstName string  `db:"first_name"`
	LastName  string  `db:"last_name"`
	Password  string  `db:"password"`
	Email     string  `db:"email"`
	Avatar    *string `db:"avatar"`
}

type userDatabaseRepository struct {
	db      database.ConnPool
	builder squirrel.StatementBuilderType
}

func NewUserDatabaseRepository(db database.ConnPool, builder squirrel.StatementBuilderType) UserRepository {
	return &userDatabaseRepository{
		db:      db,
		builder: builder,
	}
}

func (r *userDatabaseRepository) getTableName() string {
	return fmt.Sprintf("%s.user", r.db.GetSchema())
}

func (r *userDatabaseRepository) FindAll(ctx context.Context) (users []*User, err error) {
	query, _, err := r.builder.Select("id", "username", "first_name", "last_name", "password", "email", "avatar").
		From(r.getTableName()).ToSql()
	if err != nil {
		return nil, fmt.Errorf("user_repository find_by_username failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("user_repository find_by_username failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	rows, err := tx.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("user_repository find_by_username failed to execute query: %e", err)
	}

	users, err = r.fromRows(rows)

	return users, err
}

func (r *userDatabaseRepository) FindByUsername(ctx context.Context, username string) (user *User, err error) {
	query, args, err := r.builder.Select("id", "username", "first_name", "last_name", "password", "email", "avatar").
		From(r.getTableName()).
		Where(squirrel.Eq{"username": username}).Limit(1).ToSql()
	if err != nil {
		return nil, fmt.Errorf("user_repository find_by_username failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("user_repository find_by_username failed to begin transaction: %e", err)
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	row := tx.QueryRow(ctx, query, args...)

	user, err = r.fromRow(row)
	return user, err
}

func (r *userDatabaseRepository) FindByEmail(ctx context.Context, email string) (user *User, err error) {
	query, args, err := r.builder.Select("id", "username", "first_name", "last_name", "password", "email", "avatar").
		From(r.getTableName()).
		Where(squirrel.Eq{"email": email}).Limit(1).ToSql()
	if err != nil {
		return nil, fmt.Errorf("user_repository find_by_email failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("user_repository find_by_email failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	row := tx.QueryRow(ctx, query, args...)
	user, err = r.fromRow(row)
	return user, err
}

func (r *userDatabaseRepository) FindByID(ctx context.Context, id int64) (user *User, err error) {
	query, args, err := r.builder.Select("id", "username", "first_name", "last_name", "password", "email", "avatar").
		From(r.getTableName()).
		Where(squirrel.Eq{"id": id}).Limit(1).ToSql()
	if err != nil {
		return nil, fmt.Errorf("user_repository find_by_id failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("user_repository find_by_id failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	row := tx.QueryRow(ctx, query, args...)
	user, err = r.fromRow(row)
	return user, err
}

func (r *userDatabaseRepository) Insert(ctx context.Context, user *User) (int64, error) {
	query, args, err := r.builder.Insert(r.getTableName()).
		Columns("username", "first_name", "last_name", "password", "email", "avatar").
		Values(user.Username, user.FirstName, user.LastName, user.Password, user.Email, user.Avatar).
		Suffix("RETURNING id").ToSql()
	if err != nil {
		return 0, fmt.Errorf("user_repository insert failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("user_repository insert failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	row := tx.QueryRow(ctx, query, args...)
	if row == nil {
		err = fmt.Errorf("user_repository insert failed to execute query: %e", err)
		return 0, err
	}

	var id int64
	err = row.Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("user_repository insert failed to return id: %e", err)
	}

	return id, nil
}

func (r *userDatabaseRepository) Update(ctx context.Context, id int64, user *User) error {
	query, args, err := r.builder.Update(r.getTableName()).
		Set("username", user.Username).
		Set("first_name", user.FirstName).
		Set("last_name", user.LastName).
		Set("password", user.Password).
		Set("email", user.Email).
		Set("avatar", user.Avatar).
		Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return fmt.Errorf("user_repository update failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("user_repository update failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("user_repository update failed to execute query: %e", err)
	}

	return nil
}

func (r *userDatabaseRepository) Delete(ctx context.Context, id int64) error {
	query, args, err := r.builder.Delete(r.getTableName()).
		Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return fmt.Errorf("user_repository delete failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("user_repository delete failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("user_repository delete failed to execute query: %e", err)
	}

	return nil
}

func (r *userDatabaseRepository) fromRows(rows pgx.Rows) ([]*User, error) {
	defer func() {
		rows.Close()
	}()

	users := []*User{}

	for rows.Next() {
		user, err := r.fromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("user_repository failed to scan row: %e", err)
		}

		users = append(users, user)
	}

	return users, nil
}

func (r *userDatabaseRepository) fromRow(row pgx.Row) (*User, error) {
	user := &User{}
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.Email,
		&user.Avatar,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("user_repository failed to scan row: %e", err)
	}

	return user, nil
}
