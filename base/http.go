package base

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	usermodel "users/model"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

var (
	errBadRequest = errors.New("invalid request")
)

// MakeHTTPHandler mounts all of the service endpoints into an http.Handler.
func MakeHTTPHandler(s Service, logger log.Logger, version string, basePath string) http.Handler {
	r := mux.NewRouter()
	e := MakeServerEndpoints(s)

	baseRoute := "/" + basePath + "/" + version

	r.Methods(http.MethodGet).Path("/healthcheck").Handler(httptransport.NewServer(
		e.Check,
		httptransport.NopRequestDecoder,
		encodeHealthResponse,
	))
	r.Methods(http.MethodPost).Path(baseRoute + "/authenticate").Handler(httptransport.NewServer(
		e.DoorAuthenticate,
		decodedoorauthenticateRequest,
		encodeResponse,
		httptransport.ServerBefore(httptransport.PopulateRequestContext),
	))
	r.Methods(http.MethodPost).Path(baseRoute + "/updateuseraccess").Handler(httptransport.NewServer(
		e.UpdateUserAccess,
		decodeUpdateUserRequest,
		encodeResponse,
		httptransport.ServerBefore(httptransport.PopulateRequestContext),
	))
	r.Methods(http.MethodGet).Path(baseRoute + "/getuser").Handler(httptransport.NewServer(
		e.GetUser,
		decodeGetUserRequest,
		encodeResponse,
		httptransport.ServerBefore(httptransport.PopulateRequestContext),
	))
	return r
}

// errorer is implemented by all concrete response types that may contain
// errors. It allows us to change the HTTP response code without needing to
// trigger an endpoint (transport-level) error. For more information, read the
// big comment in endpoints.go.
type errorer interface {
	error() error
}

// encodeResponse is the common method to encode all response types to the
// client. I chose to do it this way because, since we're using JSON, there's no
// reason to provide anything more specific. It's certainly possible to
// specialize on a per-response (per-method) basis.
func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func encodeHealthResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	val, ok := response.(bool)
	if ok && !val {
		w.WriteHeader(http.StatusTooManyRequests)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
func decodeGetUserRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	username := r.URL.Query().Get("username")
	if username == "" {
		return "", errBadRequest
	}
	return username, nil
}

func decodeUpdateUserRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req usermodel.UpdateAccessRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	return req, nil
}

func decodedoorauthenticateRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req usermodel.DoorAuthenticate
	err = json.NewDecoder(r.Body).Decode(&req)
	return req, nil
}
