package base

import (
	"accessdoor/model"
	"context"
	eventmodel "events/model"
	"time"
	usermodel "users/model"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport/http"
	"github.com/gorilla/handlers"
)

//Middleware ...
type Middleware func(Service) Service

//NewLoggingMiddleware ...
func NewLoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return &loggingMiddleware{
			next:   next,
			logger: logger,
		}
	}
}

type loggingMiddleware struct {
	next   Service
	logger log.Logger
}

//NewPanicLogger implements the RecoveryHandler logger interface
func NewPanicLogger(logger log.Logger) handlers.RecoveryHandlerLogger {
	return panicLogger{
		logger,
	}
}

type panicLogger struct {
	log.Logger
}

//Println ....
func (pl panicLogger) Println(msgs ...interface{}) {
	for _, msg := range msgs {
		pl.Log("panic", msg)
	}
}

func (mw loggingMiddleware) Check(ctx context.Context) (res bool, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "Check", "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.Check(ctx)
}

func cid(ctx context.Context) string {
	cid, _ := ctx.Value(http.ContextKeyRequestXRequestID).(string)
	return cid
}
func xff(ctx context.Context) string {
	xff, _ := ctx.Value(http.ContextKeyRequestXForwardedFor).(string)
	return xff
}
func (mw loggingMiddleware) GetUser(ctx context.Context, username string) (res model.UserResponse, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "GetUser", "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.GetUser(ctx, username)
}

func (mw loggingMiddleware) UpdateUserAccess(ctx context.Context, req usermodel.UpdateAccessRequest) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "UpdateUserAccess", "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.UpdateUserAccess(ctx, req)
}

func (mw loggingMiddleware) DoorAuthenticate(ctx context.Context, req usermodel.DoorAuthenticate) (hasaccess bool, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "DoorAuthenticate", "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.DoorAuthenticate(ctx, req)
}

func NewUsersProxyLoggingMiddleware(logger log.Logger) UsersProxy {
	return func(next UsersService) UsersService {
		return &userLoggingMiddleware{
			next:   next,
			logger: logger,
		}
	}
}

type userLoggingMiddleware struct {
	next   UsersService
	logger log.Logger
}

func (mw userLoggingMiddleware) GetUser(ctx context.Context, username string) (resp usermodel.User, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "GetUserProxy", "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.GetUser(ctx, username)
}

func (mw userLoggingMiddleware) UpdateUserAccess(ctx context.Context, req usermodel.UpdateAccessRequest) (resp string, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "UpdateUserAccessProxy", "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.UpdateUserAccess(ctx, req)
}
func (mw userLoggingMiddleware) DoorAuthenticate(ctx context.Context, req usermodel.DoorAuthenticate) (resp string, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "DoorAuthenticateProxy", "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.DoorAuthenticate(ctx, req)
}

func NewEventsProxyLoggingMiddleware(logger log.Logger) EventsProxy {
	return func(next EventsService) EventsService {
		return &eventsLoggingMiddleware{
			next:   next,
			logger: logger,
		}
	}
}

type eventsLoggingMiddleware struct {
	next   EventsService
	logger log.Logger
}

func (mw eventsLoggingMiddleware) GetEvents(ctx context.Context, username string) (resp eventmodel.Events, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "GetEvents", "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.GetEvents(ctx, username)
}

func (mw eventsLoggingMiddleware) UpdateEvents(ctx context.Context, request eventmodel.UpdateEventRequest) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "UpdateEvents", "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.UpdateEvents(ctx, request)
}
