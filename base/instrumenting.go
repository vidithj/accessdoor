package base

import (
	"accessdoor/model"
	"context"
	eventmodel "events/model"
	"time"
	usermodel "users/model"

	"github.com/go-kit/kit/metrics"
)

type instrumentingService struct {
	labelNames     []string
	requestCount   metrics.Counter
	errCount       metrics.Counter
	requestLatency metrics.Histogram
	next           Service
}

//NewInstrumentingService ...
func NewInstrumentingService(labelNames []string, counter metrics.Counter, errCounter metrics.Counter, latency metrics.Histogram,
	s Service) Service {
	return instrumentingService{
		labelNames:     labelNames,
		requestCount:   counter,
		errCount:       errCounter,
		requestLatency: latency,
		next:           s,
	}
}

func (s instrumentingService) Check(ctx context.Context) (res bool, err error) {
	defer func(begin time.Time) {
		s.instrument(begin, "Check", err)
	}(time.Now())
	return s.next.Check(ctx)
}

func (s instrumentingService) instrument(begin time.Time, methodName string, err error) {
	if len(s.labelNames) > 0 {
		s.requestCount.With(s.labelNames[0], methodName).Add(1)
		s.requestLatency.With(s.labelNames[0], methodName).Observe(time.Since(begin).Seconds())
		if err != nil {
			s.errCount.With(s.labelNames[0], methodName).Add(1)
		}
	}
}

func (s instrumentingService) GetUser(ctx context.Context, username string) (res model.UserResponse, err error) {
	defer func(begin time.Time) {
		s.instrument(begin, "GetUser", err)
	}(time.Now())
	return s.next.GetUser(ctx, username)
}

func (s instrumentingService) UpdateUserAccess(ctx context.Context, req usermodel.UpdateAccessRequest) (err error) {
	defer func(begin time.Time) {
		s.instrument(begin, "UpdateUserAccess", err)
	}(time.Now())
	return s.next.UpdateUserAccess(ctx, req)
}
func (s instrumentingService) DoorAuthenticate(ctx context.Context, req usermodel.DoorAuthenticate) (hasaccess bool, err error) {
	defer func(begin time.Time) {
		s.instrument(begin, "DoorAuthenticate", err)
	}(time.Now())
	return s.next.DoorAuthenticate(ctx, req)
}

type UserServiceInstrumentingService func(UsersService) UsersService

type userServiceInstrumentingService struct {
	is   instrumentingService
	next UsersService
}

func NewUsersProxyInstrumentingService(labelNames []string, counter metrics.Counter, errCounter metrics.Counter, latencyHistogram metrics.Histogram) UserServiceInstrumentingService {
	return func(next UsersService) UsersService {
		return userServiceInstrumentingService{
			is: instrumentingService{
				labelNames:     labelNames,
				requestCount:   counter,
				errCount:       errCounter,
				requestLatency: latencyHistogram,
			},
			next: next,
		}
	}
}

func (s userServiceInstrumentingService) GetUser(ctx context.Context, username string) (resp usermodel.User, err error) {
	defer func(begin time.Time) {
		s.is.instrument(begin, "GetUser", err)
	}(time.Now())
	return s.next.GetUser(ctx, username)
}

func (s userServiceInstrumentingService) DoorAuthenticate(ctx context.Context, req usermodel.DoorAuthenticate) (resp string, err error) {
	defer func(begin time.Time) {
		s.is.instrument(begin, "DoorAuthenticate", err)
	}(time.Now())
	return s.next.DoorAuthenticate(ctx, req)
}

func (s userServiceInstrumentingService) UpdateUserAccess(ctx context.Context, req usermodel.UpdateAccessRequest) (resp string, err error) {
	defer func(begin time.Time) {
		s.is.instrument(begin, "UpdateUserAccess", err)
	}(time.Now())
	return s.next.UpdateUserAccess(ctx, req)
}

type EventsServiceInstrumentingService func(EventsService) EventsService

type eventsServiceInstrumentingService struct {
	is   instrumentingService
	next EventsService
}

func NewEventsProxyInstrumentingService(labelNames []string, counter metrics.Counter, errCounter metrics.Counter, latencyHistogram metrics.Histogram) EventsServiceInstrumentingService {
	return func(next EventsService) EventsService {
		return eventsServiceInstrumentingService{
			is: instrumentingService{
				labelNames:     labelNames,
				requestCount:   counter,
				errCount:       errCounter,
				requestLatency: latencyHistogram,
			},
			next: next,
		}
	}
}

func (s eventsServiceInstrumentingService) GetEvents(ctx context.Context, username string) (resp eventmodel.Events, err error) {
	defer func(begin time.Time) {
		s.is.instrument(begin, "GetEvents", err)
	}(time.Now())
	return s.next.GetEvents(ctx, username)
}

func (s eventsServiceInstrumentingService) UpdateEvents(ctx context.Context, request eventmodel.UpdateEventRequest) (err error) {
	defer func(begin time.Time) {
		s.is.instrument(begin, "UpdateEvents", err)
	}(time.Now())
	return s.next.UpdateEvents(ctx, request)
}
