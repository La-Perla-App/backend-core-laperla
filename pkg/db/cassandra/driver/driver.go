package driver

import (
	"github.com/La-Perla-App/backend-core-laperla/pkg/db/cassandra/cqlhooks/hooks"
	"github.com/gocql/gocql"
)

func NewSession(config gocql.ClusterConfig) (*gocql.Session, error) {
	cluster := gocql.NewCluster(config.Hosts...)
	cluster.Keyspace = config.Keyspace
	cluster.Consistency = config.Consistency
	cluster.Authenticator = config.Authenticator
	cluster.QueryObserver = hooks.NewLogHook()
	cluster.BatchObserver = hooks.NewLogHook()

	return cluster.CreateSession()
}
