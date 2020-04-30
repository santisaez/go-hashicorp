// Example Consul API code to monitor a key using Blocking Queries: https://www.consul.io/api/features/blocking.html
package main

import (
	"flag"
	"log"
	"time"

	consul "github.com/hashicorp/consul/api"
)

// Watch watches for changes in a given key from Consul's K/V
func Watch(kv *consul.KV, key string) {
	var lastIndex uint64

	log.Printf("watching '%s' key\n", key)

	for {
		// Blocking Query: wait until the next index or the timeout is reached
		blockingQuery := &consul.QueryOptions{WaitIndex: lastIndex}

		kvPair, meta, err := kv.Get(key, blockingQuery)
		if err != nil {
			log.Printf("lookup key error: %v\n", err)

			time.Sleep(1 * time.Second)
			continue
		}

		lastIndex = meta.LastIndex

		if kvPair != nil {
			log.Printf("LastIndex: %d, '%s' key value is '%s'\n", lastIndex, key, kvPair.Value)
		} else {
			log.Printf("LastIndex: %d, key '%s' not found\n", lastIndex, key)
		}

		// TODO: rate limit outgoing queries to avoid overloading the server if the index starts to churn quickly
		// https://www.consul.io/api/features/blocking.html#implementation-details
	}
}

func main() {
	var key string
	flag.StringVar(&key, "key", "example", "name of the key to watch")
	flag.Parse()

	// Create API client
	client, err := consul.NewClient(consul.DefaultConfig())
	if err != nil {
		log.Fatal("create client error: ", err)
	}

	// Create handle to /v1/kv API endpoints
	kv := client.KV()

	go Watch(kv, key)

	select {}
}
