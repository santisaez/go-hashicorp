// Example Consul API code to register a service with a TTL check, features:
//
// - Starts an HTTP server on a random TCP port
// - It registers itself in Consul using a TTL check (dead man's switch mechanism)
// - It runs a goroutine to send regular heartbeats and keep the service alive in Consul
// - Graceful Shutdown: the service is deregistered when `SIGINT` is received

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/google/uuid"
	consul "github.com/hashicorp/consul/api"
)

// ConsulService holds the data of a service to register in Consul
type ConsulService struct {
	Name    string
	Address string
	Port    int
	ID      string
}

// NewConsulService creates a new ConsulService
func NewConsulService(name string, address string, port int) *ConsulService {
	service := new(ConsulService)

	service.Name = name
	service.Address = address
	service.Port = port
	service.ID = fmt.Sprintf("%s-%s", service.Name, uuid.New().String())

	return service
}

// NewListener creates a listener on a random TCP port
func NewListener() (listener net.Listener, host string, port int, err error) {
	// If the port is `:0` a random port is chosen
	listener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, "", 0, fmt.Errorf("listen error: %s", err)
	}

	// Get the chosen random TCP port
	address := listener.Addr().String()
	host, portString, err := net.SplitHostPort(address)
	if err != nil {
		return nil, "", 0, fmt.Errorf("network address parse error: ", err)
	}

	port, err = strconv.Atoi(portString)
	if err != nil {
		return nil, "", 0, fmt.Errorf("invalid port: ", err)
	}

	return listener, host, port, nil
}

// UpdateTTL sends regular heartbeats to Consul to keep the service alive
func UpdateTTL(agent *consul.Agent, service *ConsulService) {
	ticker := time.NewTicker(1 * time.Second)

	for range ticker.C {
		err := agent.UpdateTTL("service:"+service.ID, "", consul.HealthPassing)
		if err != nil {
			log.Printf("update TTL error: %s\n", err)
		}
	}
}

// Start starts an HTTP server
func Start(listener net.Listener, service *ConsulService) {
	log.Printf("listening on: %s:%d\n", service.Address, service.Port)

	err := http.Serve(listener, nil)
	if err != nil {
		log.Fatal("serve error: ", err)
	}
}

// Stop deregisters the service in Consul
func Stop(agent *consul.Agent, service *ConsulService) {
	err := agent.ServiceDeregister(service.ID)
	if err != nil {
		log.Fatal("service deregister error: ", err)
	}
}

func main() {
	var serviceName string
	flag.StringVar(&serviceName, "service", "example", "name of the service")
	flag.Parse()

	// Create listener: localhost + random TCP port
	listener, host, port, err := NewListener()
	if err != nil {
		log.Fatal("create listener error: ", err)
	}

	service := NewConsulService(serviceName, host, port)

	// Create API client
	client, err := consul.NewClient(consul.DefaultConfig())
	if err != nil {
		log.Fatal("create client error: ", err)
	}

	// Create handle to /v1/agent API endpoints
	agent := client.Agent()

	serviceReg := &consul.AgentServiceRegistration{
		Name:    service.Name,
		ID:      service.ID,
		Address: service.Address,
		Port:    service.Port,
		Check: &consul.AgentServiceCheck{
			TTL:                            "5s",
			DeregisterCriticalServiceAfter: "1m",
		},
	}

	// Register the service with the local agent of Consul
	err = agent.ServiceRegister(serviceReg)
	if err != nil {
		log.Fatal("service register error: ", err)
	}

	log.Printf("registered myself as '%s' service\n", service.Name)

	// Send heartbeat TTL signals to keep the service alive
	go UpdateTTL(agent, service)

	// Start an HTTP server
	go Start(listener, service)

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)

	// Wait and perform a graceful shutdown when SIGINT signal is received
	select {
	case sig := <-sigChan:
		log.Printf("'%s' signal received, graceful shutdown\n", sig)
		Stop(agent, service)
	}
}
