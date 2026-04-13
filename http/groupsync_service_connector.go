package http

import (
	"crypto/tls"
	"fmt"

	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/go-akka/configuration"
)

const (
	Accept                    = "Accept"
	ContentType               = "Content-Type"
	ApplicationProtobufHeader = "application/x-protobuf"
)

type GroupServiceSyncConnector interface {
	GroupServiceSyncHost() string
	SetGroupServiceSyncHost(host string)
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
		return &DefaultGroupServiceSync{
			HttpClient: NewHttpClient(conf, groupServiceSyncServiceName, tlsConfig),
			host:       fmt.Sprintf("%s%s", host, path),
		}
	}
}

func (c *DefaultGroupServiceSync) GroupServiceSyncHost() string {
	return c.host
}

func (c *DefaultGroupServiceSync) SetGroupServiceSyncHost(host string) {
	c.host = host
}

func protobufHeaders() map[string]string {
	headers := make(map[string]string)
	headers[Accept] = ApplicationProtobufHeader
	headers[ContentType] = ApplicationProtobufHeader
	return headers
}
