// Example Consul API code to put, get and delete a key from the K/V store
package main

import (
	"flag"
	"fmt"
	"log"

	consul "github.com/hashicorp/consul/api"
)

func main() {
	var key string
	var value string

	flag.StringVar(&key, "key", "example", "name of the key")
	flag.StringVar(&value, "value", "whatever-value", "value of the key")

	flag.Parse()

	// Create API client
	client, err := consul.NewClient(consul.DefaultConfig())
	if err != nil {
		log.Fatalf("create client error: %v\n", err)
	}

	// Create handle to /v1/kv API endpoints
	kv := client.KV()

	kvPair := &consul.KVPair{Key: key, Value: []byte(value)}

	_, err = kv.Put(kvPair, nil)
	if err != nil {
		log.Fatalf("write key error: %v\n", err)
	}

	fmt.Printf("'%s' key written\n", key)

	kvPair, _, err = kv.Get(key, nil)
	if err != nil {
		log.Fatalf("lookup key error: %v\n", err)
	}

	if kvPair != nil {
		fmt.Printf("'%s' key exists and its value is '%s'\n", key, kvPair.Value)
	} else {
		log.Fatalf("'%s' key does not exist, something went wrong\n", key)
	}

	_, err = kv.Delete(key, nil)
	if err != nil {
		log.Fatalf("delete key error: %v\n", err)
	}

	fmt.Printf("'%s' key deleted\n", key)
}
