package base

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
)

// Register func
func Register(serviceName string, consulAddress string, httpAddr string, httpPort int,
	dependencies []string, logger log.Logger) (*api.Client, sd.Registrar, error) {

	rand.Seed(time.Now().UTC().UnixNano())
	// Service discovery domain. In this example we use Consul.
	var client consulsd.Client
	var consulClient *api.Client
	var err error
	{
		consulConfig := api.DefaultConfig()
		consulConfig.Address = consulAddress
		consulClient, err = api.NewClient(consulConfig)
		if err != nil {
			logger.Log("err", err)
			return nil, nil, err
		}
		client = consulsd.NewClient(consulClient)
	}

	checks := api.AgentServiceChecks{}
	checks = append(checks, &api.AgentServiceCheck{
		HTTP:                           "http://" + httpAddr + ":" + strconv.Itoa(httpPort) + "/healthcheck",
		Interval:                       "1s",
		Timeout:                        "1s",
		DeregisterCriticalServiceAfter: "72h",
		Status:                         "warning",
		Notes:                          "Service health check",
	})

	//using service checks, alias check #NRFPT
	if len(dependencies) > 0 {
		for _, dependency := range dependencies {
			checks = append(checks, &api.AgentServiceCheck{
				HTTP:                           "http://" + httpAddr + ":" + strconv.Itoa(httpPort) + "/healthcheck/" + dependency,
				Interval:                       "1s",
				Timeout:                        "1s",
				DeregisterCriticalServiceAfter: "72h",
				Status:                         "warning",
				Notes:                          "Service Check to monitor health of : " + dependency,
			})
		}
	}

	asr := api.AgentServiceRegistration{
		ID:      serviceName,
		Name:    serviceName,
		Address: httpAddr,
		Port:    httpPort,
		Checks:  checks,
	}

	return consulClient, consulsd.NewRegistrar(client, &asr, logger), err
}
