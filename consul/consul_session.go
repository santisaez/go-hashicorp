// Example Consul API code that manages the lifecycle of a session, features:
//
// - It creates, renews in a goroutine and destroys the session when `SIGINT` is received
// - Using the session it acquires the lock of a key in the K/V store
// - When the TTL of the session expires or the session is destroyed the key is deleted
//
// See: https://www.consul.io/docs/internals/sessions.html

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"

	consul "github.com/hashicorp/consul/api"
)

// RenewSession renews a session in Consul
func RenewSession(session *consul.Session, initialTTL string, sessionId string, doneCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("periodically starting to renew '%s'\n", sessionId)

	err := session.RenewPeriodic(initialTTL, sessionId, nil, doneCh)
	if err != nil {
		log.Fatalf("session periodic renew error: %v\n", err)
	}

	log.Printf("stopping to renew '%s'\n", sessionId)
}

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

	// Create handles to API endpoints
	kv := client.KV()
	session := client.Session()

	// Set up the session to delete the associated lock helds when it's destroyed or the TTL expires
	sessionEntry := &consul.SessionEntry{
		Behavior:  consul.SessionBehaviorDelete,
		TTL:       "10s",
		LockDelay: 0,
	}

	sessionId, _, err := session.Create(sessionEntry, nil)
	if err != nil {
		log.Fatalf("create session error: %v\n", err)
	}

	log.Printf("new session created: %s\n", sessionId)

	// When `doneCh` channel is closed it will stop renewing the session
	doneCh := make(chan struct{})

	var wg sync.WaitGroup

	wg.Add(1)
	go RenewSession(session, "1s", sessionId, doneCh, &wg)

	kvPair := &consul.KVPair{
		Key:     key,
		Value:   []byte(value),
		Session: sessionId,
	}

	acquired, _, err := kv.Acquire(kvPair, nil)
	if err != nil {
		log.Fatalf("key acquire error: %v\n", err)
	}

	if !acquired {
		log.Fatalf("can't acquire '%s' key\n", key)
	} else {
		log.Printf("'%s' key lock acquired\n", key)
	}

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)

	select {
	case sig := <-sigCh:
		log.Printf("'%s' signal received\n", sig)

		// Closing the channel will stop invoking Session.Renew
		close(doneCh)

		// Wait until RenewSession ends, it will destroy the session
		wg.Wait()

		return
	}
}
