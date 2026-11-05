package http

import (
	"os"
	"strings"
	"testing"

	"github.com/go-akka/configuration"
	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"
)

type pprofTestCassandraConnector struct{}

func (pprofTestCassandraConnector) NewCassandraClient(*configuration.Config, bool) (*db.CassandraClient, error) {
	return &db.CassandraClient{}, nil
}

func TestNewXconfServer_PprofEnabled(t *testing.T) {
	tests := []struct {
		name           string
		config         *common.ServerConfig
		expectedEnable bool
	}{
		{
			name:           "disabled by default",
			config:         sc,
			expectedEnable: false,
		},
	}

	configBytes, err := os.ReadFile(testConfigFile)
	if err != nil {
		t.Fatalf("failed to read test config: %v", err)
	}
	configText := strings.Replace(string(configBytes),
		"metrics_enabled = true",
		"metrics_enabled = true\n        pprof_enabled = true",
		1,
	)
	if configText == string(configBytes) {
		t.Fatal("failed to add pprof_enabled to test config")
	}
	enabledConfig, err := common.NewServerConfigFromText(configText)
	if err != nil {
		t.Fatalf("failed to parse enabled test config: %v", err)
	}
	tests = append(tests, struct {
		name           string
		config         *common.ServerConfig
		expectedEnable bool
	}{
		name:           "enabled by configuration",
		config:         enabledConfig,
		expectedEnable: true,
	})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := NewXconfServer(test.config, true, &ExternalConnectors{
				CassandraConnector: pprofTestCassandraConnector{},
			})
			if got := server.PprofEnabled(); got != test.expectedEnable {
				t.Errorf("PprofEnabled() = %v, want %v", got, test.expectedEnable)
			}
		})
	}
}
