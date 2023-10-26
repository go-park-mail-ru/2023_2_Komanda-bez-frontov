package repository

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx"
)

type User struct {
	ID        int64  `db:"id"`
	Username  string `db:"username"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Password  string `db:"password"`
	Email     string `db:"email"`
}

type userDatabaseRepository struct {
	db      *pgx.ConnPool
	builder squirrel.StatementBuilderType
}

func NewUserDatabaseRepository(db *pgx.ConnPool, builder squirrel.StatementBuilderType) UserRepository {
	return &userDatabaseRepository{
		db:      db,
		builder: builder,
	}
}

func (r *userDatabaseRepository) FindAll(ctx context.Context) ([]*User, error) {
	query, _, err := r.builder.Select("id", "username", "first_name", "last_name", "password", "email").From("user").ToSql()
	if err != nil {
		return nil, fmt.Errorf("user_repository find_by_username failed to build query: %e", err)
	}

	rows, err := r.db.QueryEx(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("user_repository find_by_username failed to execute query: %e", err)
	}

	return r.fromRows(rows)
}

func (r *userDatabaseRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	query, _, err := r.builder.Select("id", "username", "first_name", "last_name", "password", "email").From("user").Where(squirrel.Eq{"username": username}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("user_repository find_by_username failed to build query: %e", err)
	}

	row := r.db.QueryRowEx(ctx, query, nil, username)
	if err != nil {
		return nil, fmt.Errorf("user_repository find_by_username failed to execute query: %e", err)
	}

	return r.fromRow(row)
}

func (r *userDatabaseRepository) FindByID(ctx context.Context, id int64) (*User, error) {
	query, _, err := r.builder.Select("id", "username", "first_name", "last_name", "password", "email").From("user").Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("user_repository find_by_id failed to build query: %e", err)
	}

	row := r.db.QueryRowEx(ctx, query, nil, id)
	if err != nil {
		return nil, fmt.Errorf("user_repository find_by_id failed to execute query: %e", err)
	}

	return r.fromRow(row)
}

func (r *userDatabaseRepository) Insert(ctx context.Context, user *User) (int64, error) {
	query, args, err := r.builder.Insert(fmt.Sprintf("%s.user", "testrepository")).
		Columns("username", "first_name", "last_name", "password", "email").
		Values(user.Username, user.FirstName, user.LastName, user.Password, user.Email).
		Suffix("RETURNING id").ToSql()
	if err != nil {
		return 0, fmt.Errorf("user_repository insert failed to build query: %e", err)
	}

	row := r.db.QueryRowEx(
		ctx,
		query,
		nil,
		args...,
	)
	// if err != nil {
	// 	return 0, fmt.Errorf("user_repository insert failed to execute query: %e", err)
	// }

	if row == nil {
		return 0, fmt.Errorf("user_repository insert failed to execute query: %e", err)
	}

	var id int64
	err = row.Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("user_repository insert failed to return id: %e", err)
	}

	return id, nil

}

func (r *userDatabaseRepository) Update(ctx context.Context, id int64, user *User) error {
	return nil
}

func (r *userDatabaseRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (r *userDatabaseRepository) fromRows(rows *pgx.Rows) ([]*User, error) {
	defer func() {
		rows.Close()
	}()

	users := []*User{}

	for rows.Next() {
		user := &User{}

		err := rows.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Password, &user.Email)
		if err != nil {
			return nil, fmt.Errorf("user_repository failed to scan row: %e", err)
		}

		users = append(users, user)
	}

	return users, nil
}

func (r *userDatabaseRepository) fromRow(row *pgx.Row) (*User, error) {
	if row == nil {
		return nil, fmt.Errorf("user_repository row is nil")
	}

	user := &User{}
	err := row.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Password, &user.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("user_repository failed to scan row: %e", err)
	}

	return user, nil
}
