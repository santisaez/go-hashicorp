// Example Consul API code to query a service and list the instances
package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"

	consul "github.com/hashicorp/consul/api"
)

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

	// Query the catalog of Consul
	instanceList, _, err := catalog.Service(serviceName, "", nil)
	if err != nil {
		log.Fatal("query service error: ", err)
	}

	// Check if the service exists in Consul
	if len(instanceList) < 1 || instanceList[0].ServiceName != serviceName {
		fmt.Printf("service '%s' not found\n", serviceName)
	} else {
		var nodeList []string

		for _, instance := range instanceList {
			host := instance.ServiceAddress
			port := instance.ServicePort

			nodeList = append(nodeList, fmt.Sprintf("%s:%d", host, port))
		}

		sort.Strings(nodeList)

		fmt.Printf("%s: %s\n", serviceName, strings.Join(nodeList, ","))
	}
}
