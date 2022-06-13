package base

import (
	"context"
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

func (s instrumentingService) GetUser(ctx context.Context, username string) (res usermodel.User, err error) {
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
