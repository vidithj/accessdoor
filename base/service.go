package base

import (
	"accessdoor/api"
	"accessdoor/model"
	"context"
	"errors"
	eventmodel "events/model"
	"strings"
	"time"
	usermodel "users/model"

	"github.com/go-kit/kit/log"
)

//Service ...
type Service interface {
	Check(ctx context.Context) (bool, error)
	GetUser(ctx context.Context, username string) (model.UserResponse, error)
	UpdateUserAccess(ctx context.Context, req usermodel.UpdateAccessRequest) error
	DoorAuthenticate(ctx context.Context, req usermodel.DoorAuthenticate) (bool, error)
}

type baseService struct {
	logger        log.Logger
	usersService  UsersService
	eventsService EventsService
}

//NewService ...
func NewService(l log.Logger, usersService UsersService, eventsService EventsService) Service {
	return baseService{
		logger:        l,
		usersService:  usersService,
		eventsService: eventsService,
	}
}

//Check ...
func (s baseService) Check(ctx context.Context) (bool, error) {
	return true, nil
}

func (s baseService) GetUser(ctx context.Context, username string) (model.UserResponse, error) {
	userinformation, err := s.usersService.GetUser(ctx, username)
	if err != nil {
		return model.UserResponse{}, err
	}
	userevents, err := s.eventsService.GetEvents(ctx, username)
	if err != nil {
		return model.UserResponse{}, err
	}
	return api.FormatEvents(userinformation, userevents), nil
}
func (s baseService) UpdateUserAccess(ctx context.Context, req usermodel.UpdateAccessRequest) error {
	userinfo, err := s.usersService.GetUser(ctx, req.Username)
	if err != nil {
		return err
	}
	if userinfo.IsAdmin {
		_, err := s.usersService.UpdateUserAccess(ctx, req)
		if err != nil {
			return err
		}
	} else {
		return errors.New("only admin users can update access")
	}
	return nil
}
func (s baseService) DoorAuthenticate(ctx context.Context, req usermodel.DoorAuthenticate) (bool, error) {
	hasaccess, err := s.usersService.DoorAuthenticate(ctx, req)
	if err != nil {
		return false, err
	}
	if !strings.Contains(hasaccess, "not") {
		s.eventsService.UpdateEvents(ctx, eventmodel.UpdateEventRequest{
			Username: req.Username,
			Event: map[string]int64{
				req.AccessDoor: time.Now().Unix(),
			},
		})
		return true, nil
	} else {
		return false, errors.New("User does not have access to " + req.AccessDoor)
	}
}
