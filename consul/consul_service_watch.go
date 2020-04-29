// Example Consul API code to monitor a service using Blocking Queries: https://www.consul.io/api/features/blocking.html
package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	consul "github.com/hashicorp/consul/api"
)

// Watch monitors a service in Consul
func Watch(catalog *consul.Catalog, serviceName string) {
	var lastIndex uint64

	log.Printf("watching '%s' service\n", serviceName)

	for {
		// Blocking Query: wait until the next index or the timeout is reached
		blockingQuery := &consul.QueryOptions{WaitIndex: lastIndex}

		instanceList, meta, err := catalog.Service(serviceName, "", blockingQuery)
		if err != nil {
			log.Printf("query service error: %s\n", err)

			time.Sleep(1 * time.Second)
			continue
		}

		lastIndex = meta.LastIndex

		if len(instanceList) < 1 || instanceList[0].ServiceName != serviceName {
			log.Printf("LastIndex: %d, service '%s' not found\n", lastIndex, serviceName)
		} else {
			var nodeList []string

			for _, instance := range instanceList {
				host := instance.ServiceAddress
				port := instance.ServicePort

				nodeList = append(nodeList, fmt.Sprintf("%s:%d", host, port))
			}

			sort.Strings(nodeList)

			log.Printf("LastIndex: %d, instances: %s\n", lastIndex, strings.Join(nodeList, ","))
		}

		// TODO: rate limit outgoing queries to avoid overloading the server if the index starts to churn quickly
		// https://www.consul.io/api/features/blocking.html#implementation-details
	}
}

func main() {
	var serviceName string
	flag.StringVar(&serviceName, "service", "example", "name of the service")
	flag.Parse()

	// Create API client
	client, err := consul.NewClient(consul.DefaultConfig())
	if err != nil {
		log.Fatal("create client error: ", err)
	}

	// Create handle to /v1/catalog API endpoints
	catalog := client.Catalog()

	// Monitor the service in Consul
	go Watch(catalog, serviceName)

	select {}
}
