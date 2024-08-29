package db

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sigv4-auth-cassandra-gocql-driver-plugin/sigv4"
	"github.com/go-akka/configuration"
	"github.com/gocql/gocql"
	log "github.com/sirupsen/logrus"
)

func awsKeyspaceClient(conf *configuration.Config, testOnly bool) (*CassandraClient, error) {
	// init
	hosts := conf.GetStringList("xconfwebconfig.database.hosts")
	cluster := gocql.NewCluster(hosts...)

	cluster.Consistency = gocql.LocalQuorum
	cluster.ProtoVersion = int(conf.GetInt32("xconfwebconfig.database.protocolversion", ProtocolVersion))
	cluster.DisableInitialHostLookup = DisableInitialHostLookup
	cluster.Timeout = time.Duration(conf.GetInt32("xconfwebconfig.database.timeout_in_sec", 1)) * time.Second
	cluster.ConnectTimeout = time.Duration(conf.GetInt32("xconfwebconfig.database.connect_timeout_in_sec", 1)) * time.Second
	cluster.NumConns = int(conf.GetInt32("xconfwebconfig.database.connections", DefaultConnections))
	cluster.Port = int(conf.GetInt64("xconfwebconfig.database.port", DefaultPort))

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

	awsRegion, err := getAwsRegionForCassandra(conf)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	var auth sigv4.AwsAuthenticator = sigv4.NewAwsAuthenticator()
	auth.Region = awsRegion

	isRoleBasedAccessEnabled := conf.GetBoolean("xconfwebconfig.database.role_based_access_enabled")
	if isRoleBasedAccessEnabled {
		sess, err := session.NewSession(&aws.Config{
			Region: aws.String(awsRegion)},
		)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}

		value, err := sess.Config.Credentials.Get()
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}

		auth.AccessKeyId = value.AccessKeyID
		auth.SecretAccessKey = value.SecretAccessKey
		auth.SessionToken = value.SessionToken
	} else {
		auth.AccessKeyId = conf.GetString("xconfwebconfig.database.access_key_id")
		auth.SecretAccessKey = conf.GetString("xconfwebconfig.database.secret_access_key")
	}
	cluster.Authenticator = auth

	awsKeySpaceCaPath := conf.GetString("xconfwebconfig.database.aws_keyspace_ca_path")
	cluster.SslOpts = &gocql.SslOptions{
		CaPath:                 awsKeySpaceCaPath,
		EnableHostVerification: false,
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

func getAwsRegionForCassandra(conf *configuration.Config) (string, error) {
	awsRegion := conf.GetString("xconfwebconfig.database.aws_region")
	if len(awsRegion) == 0 {
		awsRegion = os.Getenv("AWS_REGION")
	}

	if len(awsRegion) == 0 {
		return "", fmt.Errorf("%s", "Aws region is not provided")
	}

	return awsRegion, nil
}
