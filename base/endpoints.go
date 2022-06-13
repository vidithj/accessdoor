package base

import (
	"context"
	"errors"
	usermodel "users/model"

	"github.com/go-kit/kit/endpoint"
)

//Endpoints ...
type Endpoints struct {
	Check            endpoint.Endpoint
	GetUser          endpoint.Endpoint
	UpdateUserAccess endpoint.Endpoint
	DoorAuthenticate endpoint.Endpoint
}

//MakeServerEndpoints ...
func MakeServerEndpoints(s Service) Endpoints {
	return Endpoints{
		Check:            MakeCheck(s),
		GetUser:          MakeGetUser(s),
		UpdateUserAccess: MakeUpdateUserAccess(s),
		DoorAuthenticate: MakeDoorAuthenticate(s),
	}
}

//MakeCheck ...
func MakeCheck(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		return s.Check(ctx)
	}
}
func MakeGetUser(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		username, ok := request.(string)
		if !ok {
			return nil, errors.New("Bad Request")
		}
		return s.GetUser(ctx, username)
	}
}

func MakeUpdateUserAccess(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(usermodel.UpdateAccessRequest)
		if !ok {
			return nil, errors.New("Bad Request")
		}
		return "", s.UpdateUserAccess(ctx, req)
	}
}

func MakeDoorAuthenticate(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(usermodel.DoorAuthenticate)
		if !ok {
			return nil, errors.New("Bad Request")
		}
		return s.DoorAuthenticate(ctx, req)
	}
}
