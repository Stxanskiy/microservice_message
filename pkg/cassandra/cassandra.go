package cassandra

import "github.com/gocql/gocql"

func NewSession(cassandraURL string) (*gocql.Session, error) {
	cluster := gocql.NewCluster(cassandraURL)
	cluster.Keyspace = "messaging_keyspace"
	cluster.Consistency = gocql.Quorum
	return cluster.CreateSession()
}
