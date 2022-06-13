package main

import (
	"context"
	"flag"
	"fmt"
	"log/syslog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"accessdoor/base"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/gorilla/handlers"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	UsersServiceName  = "users"
	EventsServiceName = "events"
	httpurl           = "http://"
)

var (
	proxyservices = map[string]string{
		UsersServiceName:  "",
		EventsServiceName: "",
	}
)

func main() {

	var (
		serviceName          = flag.String("service.name", "dooraccess", "Name of microservice")
		basePath             = flag.String("service.base.path", "dooraccess", "Name of microservice")
		version              = flag.String("service.version", "v1", "Version of microservice")
		httpAddr             = flag.String("http.addr", "localhost", "This is the addr at which http requests are accepted (Default localhost)")
		httpPort             = flag.Int("http.port", 8090, "This is the port at which http requests are accepted (Default :8080)")
		metricsPort          = flag.Int("metrics.port", 8092, "HTTP metrics listen address (Default 8082)")
		dataType             = flag.String("service.datatype", "test", "default Test/qa")
		consulAddr           = flag.String("consul.addr", "localhost:8500", "consul address (Default localhost:8500)")
		serverTimeout        = flag.Int64("service.timeout", 200000, "service timeout in milliseconds")
		sysLogAddress        = flag.String("syslog.address", "localhost:514", "default location for the syslogger")
		usersgetuserURL      = flag.String("proxy.getuserurl", "/users/v1/getuser", "user proxy url")
		usersauthenticateURL = flag.String("proxy.authenticateurl", "/users/v1/authenticate", "user proxy url")
		usersupdateaccessURL = flag.String("proxy.updatedaccess", "/users/v1/updateuseraccess", "user proxy url")
		eventsupdatURL       = flag.String("proxy.eventupdate", "/events/v1/updateevent", "events proxy url")
		geteventsURL         = flag.String("proxy.getevent", "/events/v1/getevents", "events proxy url")
		maxAttempts          = flag.Int("outbound.service.attempts", 1, "max attempts for API")
		apiMaxTime           = flag.Int("outbound.service.maxtime", 500000, "maxTime for API in milliseconds")
	)
	flag.Parse()
	errs := make(chan error)

	sysLogger, err := syslog.Dial("udp", *sysLogAddress, syslog.LOG_EMERG|syslog.LOG_LOCAL6, *serviceName)
	if err != nil {
		fmt.Printf("exit: %v\n", err)
		return
	}
	defer sysLogger.Close()

	var logger log.Logger
	{
		logger = log.NewJSONLogger(os.Stdout)
		logger = log.With(logger, "serviceName", *serviceName)
		logger = log.With(logger, "ip", *httpAddr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	consulClient, registrar, err := base.Register(*serviceName, *consulAddr, *httpAddr, *httpPort, []string{}, logger)
	if err != nil || registrar == nil {
		logger.Log("exit", err)
		return
	}
	registrar.Register()
	//get service ip from consul. since the instance ips are dynamic using consul agent to fetch them.
	for service, _ := range proxyservices {
		serviceinfo, _, _ := consulClient.Health().Service(service, "", true, nil)
		proxyservices[service] = httpurl + serviceinfo[0].Service.Address + ":" + strconv.Itoa(serviceinfo[0].Service.Port)
	}
	labelNames := []string{"method"}
	constLabels := map[string]string{"serviceName": *serviceName, "version": *version, "dataType": *dataType}
	getuserinfoURL, err := url.Parse(proxyservices[UsersServiceName] + *usersgetuserURL)
	if err != nil {
		logger.Log("error while parsing getuserinfoURL" + err.Error())
	}

	authenticateuserURL, err := url.Parse(proxyservices[UsersServiceName] + *usersauthenticateURL)
	if err != nil {
		logger.Log("error while parsing authenticateuserURL" + err.Error())
	}

	updateuseraccessURL, err := url.Parse(proxyservices[UsersServiceName] + *usersupdateaccessURL)
	if err != nil {
		logger.Log("error while parsing updateuseraccessURL" + err.Error())
	}

	updateeventURL, err := url.Parse(proxyservices[EventsServiceName] + *eventsupdatURL)
	if err != nil {
		logger.Log("error while parsing updateevents" + err.Error())
	}

	eventsgetURL, err := url.Parse(proxyservices[EventsServiceName] + *geteventsURL)
	if err != nil {
		logger.Log("error while parsing eventsget" + err.Error())
	}

	var eventsService base.EventsService
	eventsService = base.NewEventsProxy(context.Background(),
		base.ProxyConfig{
			URL:         eventsgetURL,
			Method:      http.MethodGet,
			MaxAttempts: *maxAttempts,
			MaxTime:     time.Duration(*apiMaxTime) * time.Millisecond,
		},
		base.ProxyConfig{
			URL:         updateeventURL,
			Method:      http.MethodPost,
			MaxAttempts: *maxAttempts,
			MaxTime:     time.Duration(*apiMaxTime) * time.Millisecond,
		},
		logger)(eventsService)
	eventsService = base.NewEventsProxyLoggingMiddleware(logger)(eventsService)
	eventsService = base.NewEventsProxyInstrumentingService(labelNames,
		prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Name:        "proxy_request_count",
			Help:        "Number of requests received.",
			ConstLabels: constLabels,
		}, labelNames),
		prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Name:        "proxy_outbound_err_count",
			Help:        "Number of errors.",
			ConstLabels: constLabels,
		}, labelNames),
		prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Name:        "proxy_outbound_request_latency_seconds",
			Help:        "Total duration of requests in request_latency_seconds.",
			ConstLabels: constLabels,
		}, labelNames))(eventsService)

	var usersService base.UsersService
	usersService = base.NewUsersProxy(context.Background(),
		base.ProxyConfig{
			URL:         getuserinfoURL,
			Method:      http.MethodGet,
			MaxAttempts: *maxAttempts,
			MaxTime:     time.Duration(*apiMaxTime) * time.Millisecond,
		},
		base.ProxyConfig{
			URL:         authenticateuserURL,
			Method:      http.MethodPost,
			MaxAttempts: *maxAttempts,
			MaxTime:     time.Duration(*apiMaxTime) * time.Millisecond,
		},
		base.ProxyConfig{
			URL:         updateuseraccessURL,
			Method:      http.MethodPost,
			MaxAttempts: *maxAttempts,
			MaxTime:     time.Duration(*apiMaxTime) * time.Millisecond,
		},
		logger)(usersService)
	usersService = base.NewUsersProxyLoggingMiddleware(logger)(usersService)
	usersService = base.NewUsersProxyInstrumentingService(labelNames,
		prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Name:        "outbound_request_count",
			Help:        "Number of requests received.",
			ConstLabels: constLabels,
		}, labelNames),
		prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Name:        "outbound_err_count",
			Help:        "Number of errors.",
			ConstLabels: constLabels,
		}, labelNames),
		prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Name:        "outbound_request_latency_seconds",
			Help:        "Total duration of requests in request_latency_seconds.",
			ConstLabels: constLabels,
		}, labelNames))(usersService)

	var s base.Service
	{

		s = base.NewService(logger, usersService, eventsService)
		s = base.NewLoggingMiddleware(logger)(s)
		s = base.NewInstrumentingService(labelNames, prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Name:        "request_count",
			Help:        "Number of requests received.",
			ConstLabels: constLabels,
		}, labelNames),
			prometheus.NewCounterFrom(stdprometheus.CounterOpts{
				Name:        "err_count",
				Help:        "Number of errors.",
				ConstLabels: constLabels,
			}, labelNames),
			prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
				Name:        "request_latency_seconds",
				Help:        "Total duration of requests in request_latency_seconds.",
				ConstLabels: constLabels,
			}, labelNames),
			s)
	}

	h := base.MakeHTTPHandler(s, logger, *version, *basePath)
	h = http.TimeoutHandler(h, time.Duration(*serverTimeout)*time.Millisecond, "")

	httpServer := http.Server{
		Addr:    ":" + strconv.Itoa(*httpPort),
		Handler: handlers.RecoveryHandler(handlers.RecoveryLogger(base.NewPanicLogger(logger)))(h),
	}

	go func() {
		errs <- httpServer.ListenAndServe()
	}()

	metricsServer := http.Server{
		Addr:    ":" + strconv.Itoa(*metricsPort),
		Handler: promhttp.Handler(),
	}

	go func() {
		logger.Log("transport", "HTTP", "addr", *metricsPort)
		errs <- metricsServer.ListenAndServe()
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	errMain := <-errs
	//exit gracefully
	errMetricsServer := metricsServer.Shutdown(context.Background())
	errHTTPServer := httpServer.Shutdown(context.Background())
	logger.Log("exit", errMain, "httpErr", errHTTPServer, "metricsErr", errMetricsServer)

}
