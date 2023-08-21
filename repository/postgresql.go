package repository

import (
	"auth_service/models"
	"context"
	"log"

	"github.com/jmoiron/sqlx"
)

type PostgresRepo struct {
	l  *log.Logger
	db *sqlx.DB
}

func NewPostgresRepo(l *log.Logger, db *sqlx.DB) IRepository {
	return PostgresRepo{l: l, db: db}
}

func (r PostgresRepo) InsertUser(ctx context.Context, u models.User) error {
	_, err := r.db.Exec("INSERT INTO users (username, password, company) VALUES ($1, $2, $3);", u.Username, u.Password, u.Company)
	return err
}

func (r PostgresRepo) CheckUserNameExists(c context.Context, username string) bool {
	var exists []bool
	err := r.db.Select(&exists, "SELECT EXISTS(SELECT 1 from users where username=$1);", username)
	if err != nil {
		r.l.Println(err)
	}

	return exists[0]
}
