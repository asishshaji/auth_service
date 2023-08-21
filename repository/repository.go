package repository

import (
	"auth_service/models"
	"context"
)

type IRepository interface {
	InsertUser(context.Context, models.User) error
	CheckUserNameExists(context.Context, string) bool
}
