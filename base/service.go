package base

import (
	"context"

	usermodel "users/model"

	"github.com/go-kit/kit/log"
)

//Service ...
type Service interface {
	Check(ctx context.Context) (bool, error)
	GetUser(ctx context.Context, username string) (usermodel.User, error)
	UpdateUserAccess(ctx context.Context, req usermodel.UpdateAccessRequest) error
	DoorAuthenticate(ctx context.Context, req usermodel.DoorAuthenticate) (bool, error)
}

type baseService struct {
	logger       log.Logger
	usersService UsersService
}

//NewService ...
func NewService(l log.Logger, usersService UsersService) Service {
	return baseService{
		logger:       l,
		usersService: usersService,
	}
}

//Check ...
func (s baseService) Check(ctx context.Context) (bool, error) {
	return true, nil
}

func (s baseService) GetUser(ctx context.Context, username string) (usermodel.User, error) {
	return s.usersService.GetUser(ctx, username)
}
func (s baseService) UpdateUserAccess(ctx context.Context, req usermodel.UpdateAccessRequest) error {
	return s.usersService.UpdateUserAccess(ctx, req)
}
func (s baseService) DoorAuthenticate(ctx context.Context, req usermodel.DoorAuthenticate) (bool, error) {
	return s.usersService.DoorAuthenticate(ctx, req)
}
