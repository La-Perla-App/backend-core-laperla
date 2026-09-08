package cassandra

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/La-Perla-App/backend-core-laperla/pkg/config"
	"github.com/La-Perla-App/backend-core-laperla/pkg/db/cassandra/driver"
	"github.com/gocql/gocql"
)

var (
	databaseInstances     = map[string]*gocql.Session{}
	databaseIntancesMutex = sync.Mutex{}
)

type ConnConfig struct {
	Hosts       []string
	Keyspace    string
	Consistency string
	Username    string
	Password    string
}

var DefaultConnectionConfig = ConnConfig{
	Hosts:       config.GetArrayStrings("cassandra.hosts"),
	Keyspace:    config.GetString("cassandra.keyspace"),
	Consistency: config.GetString("cassandra.consistency"),
	Username:    config.GetString("cassandra.username"),
	Password:    config.GetString("cassandra.password"),
}

func (c ConnConfig) ConnectionString() string {
	return fmt.Sprintf("%s-%s", strings.Join(c.Hosts, ","), c.Keyspace)
}

func Connect(config ConnConfig) (*gocql.Session, error) {
	log.Println("[CASSANDRA DEBUG] Attempting to connect...")
	connectionString := config.ConnectionString()

	databaseIntancesMutex.Lock()
	if session, ok := databaseInstances[connectionString]; ok && session != nil && !session.Closed() {
		databaseIntancesMutex.Unlock()
		log.Println("[CASSANDRA DEBUG] Returning cached session.")
		return session, nil
	}
	databaseIntancesMutex.Unlock()

	log.Printf("[CASSANDRA DEBUG] No cached session found. Creating new connection for keyspace '%s' with hosts %v", config.Keyspace, config.Hosts)
	clusterConfig := gocql.ClusterConfig{
		Hosts:       config.Hosts,
		Keyspace:    config.Keyspace,
		Consistency: gocql.ParseConsistency(config.Consistency),
		Timeout:     20 * time.Second,
	}

	if config.Username != "" && config.Password != "" {
		clusterConfig.Authenticator = gocql.PasswordAuthenticator{
			Username: config.Username,
			Password: config.Password,
		}
	}

	session, err := driver.NewSession(clusterConfig)
	log.Printf("[CASSANDRA DEBUG] driver.NewSession returned: session is nil? %t, err is: %v", session == nil, err)

	if err != nil {
		log.Printf("[CASSANDRA DEBUG] Connection failed. Returning (nil, err). Error: %v", err)
		return nil, err
	}

	databaseIntancesMutex.Lock()
	databaseInstances[connectionString] = session
	databaseIntancesMutex.Unlock()

	log.Printf("[CASSANDRA DEBUG] Connection successful. Returning session. Session is nil? %t", session == nil)
	return session, nil
}
