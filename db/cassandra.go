package db

import (
	"fmt"
	"time"
	"xconfwebconfig/security"

	"github.com/go-akka/configuration"
	"github.com/gocql/gocql"
	log "github.com/sirupsen/logrus"
)

func cassandraClient(conf *configuration.Config, testOnly bool) (*CassandraClient, error) {
	// init
	hosts := conf.GetStringList("xconfwebconfig.database.hosts")
	cluster := gocql.NewCluster(hosts...)

	cluster.Consistency = gocql.LocalQuorum
	cluster.ProtoVersion = int(conf.GetInt32("xconfwebconfig.database.protocolversion", ProtocolVersion))
	cluster.DisableInitialHostLookup = DisableInitialHostLookup
	cluster.Timeout = time.Duration(conf.GetInt32("xconfwebconfig.database.timeout_in_sec", 1)) * time.Second
	cluster.ConnectTimeout = time.Duration(conf.GetInt32("xconfwebconfig.database.connect_timeout_in_sec", 1)) * time.Second
	cluster.NumConns = int(conf.GetInt32("xconfwebconfig.database.connections", DefaultConnections))

	cluster.RetryPolicy = &gocql.DowngradingConsistencyRetryPolicy{
		[]gocql.Consistency{
			gocql.LocalQuorum,
			gocql.LocalOne,
			gocql.One,
		},
	}

	localDc := conf.GetString("xconfwebconfig.database.local_dc")
	if len(localDc) > 0 {
		cluster.PoolConfig.HostSelectionPolicy = gocql.DCAwareRoundRobinPolicy(localDc)
	}

	isSslEnabled := conf.GetBoolean("xconfwebconfig.database.is_ssl_enabled")

	var password string
	var err error

	encryptedPassword := conf.GetString("xconfwebconfig.database.encrypted_password")
	if encryptedPassword != "" {
		codec := security.NewAesCodec()
		password, err = codec.Decrypt(encryptedPassword)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
	} else {
		password = conf.GetString("xconfwebconfig.database.password")
	}

	user := conf.GetString("xconfwebconfig.database.user")
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: user,
		Password: password,
	}

	if isSslEnabled {
		cluster.SslOpts = &gocql.SslOptions{
			EnableHostVerification: false,
		}
	}

	// Use the appropriate keyspace
	var deviceKeyspace string
	if testOnly {
		cluster.Keyspace = conf.GetString("xconfwebconfig.database.test_keyspace", DefaultTestKeyspace)
		deviceKeyspace = conf.GetString("webconfig.database.device_test_keyspace", DefaultDeviceTestKeyspace)
	} else {
		cluster.Keyspace = conf.GetString("xconfwebconfig.database.keyspace", DefaultKeyspace)
		deviceKeyspace = conf.GetString("webconfig.database.device_keyspace", DefaultDeviceKeyspace)
	}
	log.Debug(fmt.Sprintf("Init CassandraClient with keyspace: %v", cluster.Keyspace))

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	devicePodTableName := conf.GetString("webconfig.database.device_pod_table_name", DefaultDevicePodTableName)

	return &CassandraClient{
		Session:            session,
		ClusterConfig:      cluster,
		sleepTime:          conf.GetInt32("xconfwebconfig.perftest.sleep_in_msecs", DefaultSleepTimeInMillisecond),
		concurrentQueries:  make(chan bool, conf.GetInt32("xconfwebconfig.database.concurrent_queries", 500)),
		localDc:            localDc,
		deviceKeyspace:     deviceKeyspace,
		devicePodTableName: devicePodTableName,
		testOnly:           testOnly,
	}, nil
}
