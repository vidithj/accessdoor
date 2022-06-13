package base

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
	usermodel "users/model"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/lb"
	kithttp "github.com/go-kit/kit/transport/http"
)

//ProxyConfig ...
type ProxyConfig struct {
	URL         *url.URL
	Method      string
	MaxAttempts int
	MaxTime     time.Duration
}

func MakeProxyEndpoints(method string, config ProxyConfig, encoder kithttp.EncodeRequestFunc, decoder kithttp.DecodeResponseFunc, logger log.Logger) endpoint.Endpoint {
	var endpointer sd.FixedEndpointer
	var e endpoint.Endpoint

	e = kithttp.NewClient(
		method, config.URL,
		encoder,
		decoder,
	).Endpoint()

	endpointer = append(endpointer, e)
	balancer := lb.NewRoundRobin(endpointer)
	return lb.Retry(config.MaxAttempts, config.MaxTime, balancer)
}

type UsersProxy func(UsersService) UsersService

func NewUsersProxy(ctx context.Context, getuserconfig, authenticateuserconfig, updateaccessconfig ProxyConfig, logger log.Logger) UsersProxy {
	if getuserconfig.URL == nil || authenticateuserconfig.URL == nil || updateaccessconfig.URL == nil {
		return func(next UsersService) UsersService { return next }
	}

	getUserProxy := MakeProxyEndpoints(getuserconfig.Method, getuserconfig, encodegetUsersInfoRequest, decodeGetUsersResponse, logger)
	authenticateUserProxy := MakeProxyEndpoints(authenticateuserconfig.Method, authenticateuserconfig, encodePOSTRequest, decodeAuthenticateUsersResponse, logger)
	updateaccessUserProxy := MakeProxyEndpoints(updateaccessconfig.Method, updateaccessconfig, encodePOSTRequest, decodeUpdateAccessUsersResponse, logger)

	return func(next UsersService) UsersService {
		return &usersService{
			Context:                  ctx,
			GetUserEndpoint:          getUserProxy,
			UpdateUserAccessEndpoint: updateaccessUserProxy,
			DoorAuthenticateEndpoint: authenticateUserProxy,
			UsersService:             next,
		}
	}
}

type UsersService interface {
	GetUser(ctx context.Context, username string) (usermodel.User, error)
	UpdateUserAccess(ctx context.Context, req usermodel.UpdateAccessRequest) error
	DoorAuthenticate(ctx context.Context, req usermodel.DoorAuthenticate) (bool, error)
}

type usersService struct {
	context.Context
	GetUserEndpoint          endpoint.Endpoint
	UpdateUserAccessEndpoint endpoint.Endpoint
	DoorAuthenticateEndpoint endpoint.Endpoint
	UsersService
}

func (s usersService) GetUser(ctx context.Context, username string) (usermodel.User, error) {
	response, err := s.GetUserEndpoint(ctx, username)
	if err != nil {
		return usermodel.User{}, err
	}
	return response.(usermodel.User), nil
}
func (s usersService) UpdateUserAccess(ctx context.Context, req usermodel.UpdateAccessRequest) error {
	_, err := s.UpdateUserAccessEndpoint(ctx, req)
	if err != nil {
		return err
	}
	return nil
}
func (s usersService) DoorAuthenticate(ctx context.Context, req usermodel.DoorAuthenticate) (bool, error) {
	response, err := s.DoorAuthenticateEndpoint(ctx, req)
	if err != nil {
		return false, err
	}
	return response.(bool), nil
}
func encodegetUsersInfoRequest(ctx context.Context, r *http.Request, req interface{}) error {
	setRequestHeaders(ctx, r, req)
	username, ok := req.(string)
	if !ok {
		return errors.New("invalid username during proxy call")
	}
	if username != "" {
		q := r.URL.Query()
		q.Add("username", fmt.Sprintf("%s", username))
		r.URL.RawQuery = q.Encode()
	}
	return nil
}
func decodeGetUsersResponse(_ context.Context, r *http.Response) (interface{}, error) {
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status code incorrect for GetUserinfo. Expected: %v received %v", http.StatusOK, r.StatusCode)
	}
	var response usermodel.User
	err := json.NewDecoder(r.Body).Decode(&response)
	return response, err
}
func decodeAuthenticateUsersResponse(_ context.Context, r *http.Response) (interface{}, error) {
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status code incorrect for AuthenticateUser. Expected: %v received %v", http.StatusOK, r.StatusCode)
	}
	var response bool
	err := json.NewDecoder(r.Body).Decode(&response)
	return response, err
}
func decodeUpdateAccessUsersResponse(_ context.Context, r *http.Response) (interface{}, error) {
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status code incorrect for UpdateAccess. Expected: %v received %v", http.StatusOK, r.StatusCode)
	}
	var response string
	err := json.NewDecoder(r.Body).Decode(&response)
	return response, err
}

func encodePOSTRequest(ctx context.Context, r *http.Request, req interface{}) error {
	setRequestHeaders(ctx, r, req)
	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(payload))
	return nil
}

func setRequestHeaders(ctx context.Context, r *http.Request, req interface{}) error {
	r.Header.Set("Content-Type", "application/json;charset=utf-8")
	r.Header.Set("Accept", "*/*")
	r.Header.Set("X-Forwarded-For", xff(ctx))
	r.Header.Set("X-Request-Id", cid(ctx))
	return nil
}
