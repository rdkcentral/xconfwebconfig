package http

import (
	"crypto/tls"
	"fmt"

	conversion "xconfwebconfig/protobuf"
	"xconfwebconfig/util"

	"github.com/go-akka/configuration"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

const (
	Accept                    = "Accept"
	ContentType               = "Content-Type"
	ApplicationProtobufHeader = "application/x-protobuf"
)

type GroupServiceSyncConnector interface {
	GroupServiceSyncHost() string
	SetGroupServiceSyncHost(host string)
	AddSecurityTokenInfo(securityIdenfier string, deviceInfo map[string]string, fields log.Fields) error
}

type DefaultGroupServiceSync struct {
	*HttpClient
	host              string
	securityTokenPath string
}

var groupServiceSyncServiceName string

func NewGroupServiceSyncConnector(conf *configuration.Config, tlsConfig *tls.Config, externalGroupServiceSync GroupServiceSyncConnector) GroupServiceSyncConnector {
	if externalGroupServiceSync != nil {
		return externalGroupServiceSync
	} else {
		groupServiceSyncServiceName = conf.GetString("xconfwebconfig.xconf.group_sync_service_name")
		confKey := fmt.Sprintf("xconfwebconfig.%v.host", groupServiceSyncServiceName)
		host := conf.GetString(confKey)
		if util.IsBlank(host) {
			panic(fmt.Errorf("%s is required", confKey))
		}
		pathKey := fmt.Sprintf("xconfwebconfig.%v.path", groupServiceSyncServiceName)
		path := conf.GetString(pathKey)
		if util.IsBlank(path) {
			panic(fmt.Errorf("%s is required", pathKey))
		}
		tokenPathKey := fmt.Sprintf("xconfwebconfig.%v.security_token_path", groupServiceSyncServiceName)
		tokenPath := conf.GetString(tokenPathKey)
		if util.IsBlank(tokenPath) {
			panic(fmt.Errorf("%s is required", tokenPathKey))
		}
		return &DefaultGroupServiceSync{
			HttpClient:        NewHttpClient(conf, groupServiceSyncServiceName, tlsConfig),
			host:              fmt.Sprintf("%s%s", host, path),
			securityTokenPath: tokenPath,
		}
	}
}

func (c *DefaultGroupServiceSync) GroupServiceSyncHost() string {
	return c.host
}

func (c *DefaultGroupServiceSync) SetGroupServiceSyncHost(host string) {
	c.host = host
}

func (c *DefaultGroupServiceSync) AddSecurityTokenInfo(securityIdenfier string, deviceInfo map[string]string, fields log.Fields) error {
	url := fmt.Sprintf("%s%s/%s", c.host, c.securityTokenPath, securityIdenfier)
	message := conversion.XdasHashes{
		Fields: deviceInfo,
	}
	// Serialize the protobuf message to binary data.
	data, err := proto.Marshal(&message)
	if err != nil {
		return err
	}
	_, err = c.DoWithRetries("POST", url, protobufHeaders(), data, fields, groupServiceSyncServiceName)
	if err != nil {
		return err
	}

	return nil
}

func protobufHeaders() map[string]string {
	headers := make(map[string]string)
	headers[Accept] = ApplicationProtobufHeader
	headers[ContentType] = ApplicationProtobufHeader
	return headers
}
