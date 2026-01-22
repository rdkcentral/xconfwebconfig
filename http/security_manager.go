package http

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/go-akka/configuration"
	log "github.com/sirupsen/logrus"
)

const (
	HTTP_CLIENT_PROTOCOL           = "http"
	HTTPS_CLIENT_PROTOCOL          = "https"
	MTLS_CLIENT_PROTOCOL           = "mtls"
	MTLS_OPTIONAL_CLIENT_PROTOCOL  = "mtls-optional"
	MTLS_RECOVERY_CLIENT_PROTOCOL  = "mtls-recovery"
	SECURITY_TOKEN_KEY             = "xds"
	SECURITY_TOKEN_CLIENT_PROTOCOL = "clientProtocol"
	SECURITY_TOKEN_ESTB_MAC        = "estbMac"
	SECURITY_TOKEN_ESTB_IP         = "estbIP"
	SECURITY_TOKEN_MODEL           = "model"
	SECURITY_TOKEN_PARTNER         = "partnerId"
	SECURITY_TOKEN_FW_FILENAME     = "firmwareFilename"
	SECURITY_TOKEN_FW_VERSION      = "fwVersion"
	SECURITY_TOKEN_FW_DOWNLOAD_TS  = "fwDownloadTs"
	SECURITY_TOKEN_LOG_UPLOAD_TS   = "logUploadTs"
	URL_PROTOCOL_PREFIX            = "http://"
	FQDN_CHECK                     = "://"
)

// Define a custom Base64 encoding with a custom alphabet
var SecurityTokenCustomBase64Encoding = base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_").WithPadding(base64.NoPadding)

type SecurityTokenPathConfig struct {
	UrlPathMap             map[string]bool
	UrlHostKeywordMap      map[string]bool
	TimestampKey           string
	FilenameInTokenEnabled bool
}

type SecurityTokenConfig struct {
	SecurityTokenOnlyForNewOfferedFwEnabled bool
	SkipSecurityTokenClientProtocolSet      util.Set
	SecurityTokenHostKeyword                string
	SecurityTokenGroupServiceEnabled        bool
}

func NewSecurityTokenConfig(conf *configuration.Config) *SecurityTokenConfig {
	securityTokenOnlyForNewOfferedFwEnabled := conf.GetBoolean("xconfwebconfig.xconf.security_token_only_for_new_offered_fw_enabled")
	clientProtocolString := conf.GetString("xconfwebconfig.xconf.skip_security_token_client_protocol_set")
	clientProtocolSet := util.NewSet()
	if !util.IsBlank(clientProtocolString) {
		for _, clientProtocol := range strings.Split(clientProtocolString, ";") {
			clientProtocolSet.Add(strings.ToLower(clientProtocol))
		}
	}
	hostKeyword := conf.GetString("xconfwebconfig.xconf.security_token_host_keyword")
	securityTokenGroupServiceEnabled := conf.GetBoolean("xconfwebconfig.xconf.security_token_group_service_enabled")

	return &SecurityTokenConfig{
		SecurityTokenOnlyForNewOfferedFwEnabled: securityTokenOnlyForNewOfferedFwEnabled,
		SkipSecurityTokenClientProtocolSet:      clientProtocolSet,
		SecurityTokenHostKeyword:                hostKeyword,
		SecurityTokenGroupServiceEnabled:        securityTokenGroupServiceEnabled,
	}
}

func NewFirmwareNonMtlSsrTokenPathConfig(conf *configuration.Config) *SecurityTokenPathConfig {
	pathEnabledMap := util.CreateConfigMapStringBool(conf, "xconfwebconfig.xconf.firmware_ssr_token_paths")
	hostKeywordEnabledMap := util.CreateConfigMapStringBool(conf, "xconfwebconfig.xconf.firmware_ssr_token_host_keywords")
	// firmware download API requires filename to be validated in token
	return &SecurityTokenPathConfig{
		UrlPathMap:             pathEnabledMap,
		UrlHostKeywordMap:      hostKeywordEnabledMap,
		TimestampKey:           SECURITY_TOKEN_FW_DOWNLOAD_TS,
		FilenameInTokenEnabled: true,
	}
}

func NewLogUploaderNonMtlSsrTokenPathConfig(conf *configuration.Config) *SecurityTokenPathConfig {
	pathEnabledMap := util.CreateConfigMapStringBool(conf, "xconfwebconfig.xconf.loguploader_ssr_token_paths")
	hostKeywordEnabledMap := util.CreateConfigMapStringBool(conf, "xconfwebconfig.xconf.loguploader_ssr_token_host_keywords")
	// loguploader API does not require filename to be validated in token
	return &SecurityTokenPathConfig{
		UrlPathMap:             pathEnabledMap,
		UrlHostKeywordMap:      hostKeywordEnabledMap,
		TimestampKey:           SECURITY_TOKEN_LOG_UPLOAD_TS,
		FilenameInTokenEnabled: false,
	}
}

func (s *SecurityTokenPathConfig) getSecurityToken(deviceInfo map[string]string, fields log.Fields) string {
	if CanSkipSecurityTokenForClientProtocol(deviceInfo) {
		log.WithFields(fields).Debugf("Client protocol type is in token generation skip list, no security token needed")
		return ""
	}
	token := ""
	// we will use mac address without colons as the token
	// 1. if group service is enabled, we will store token info in group service to look up later
	// 2. if group service is not enabled, we will get the token info from Cassandra (all fields needed are already in penetration table, so just return token)
	// token will be set to the mac address without colons
	token = util.AlphaNumericMacAddress(deviceInfo[SECURITY_TOKEN_ESTB_MAC])
	if util.IsBlank(token) {
		log.WithFields(fields).Errorf("Mac address is missing, not generating security token")
		return ""
	}
	if Ws.SecurityTokenConfig.SecurityTokenGroupServiceEnabled {
		// add token info to Group Service to look up later, if disabled, we will get the info from Cassandra Penetration Metrics table (will be written later in existing flow)
		deviceInfo[s.TimestampKey] = fmt.Sprintf("%d", time.Now().Unix())
		// removing macAddress from deviceInfo since it's already present as the key
		delete(deviceInfo, SECURITY_TOKEN_ESTB_MAC)
		log.WithFields(fields).Debugf("Adding security token to group service")
		err := Ws.GroupServiceSyncConnector.AddSecurityTokenInfo(token, deviceInfo, fields)
		if err != nil {
			log.WithFields(fields).Errorf("Error adding security token to group service, err=%+v", err)
		}
	}
	return token
}

func (s *SecurityTokenPathConfig) addTokenToUrl(deviceInfo map[string]string, urlString string, isFqdn bool, fields log.Fields) string {
	securityToken := s.getSecurityToken(deviceInfo, fields)
	if util.IsBlank(securityToken) {
		return urlString
	}
	fields[fmt.Sprintf("%s.token", SECURITY_TOKEN_KEY)] = securityToken
	urlStringWithProtocol := urlString
	// add protocol so we can parse the url properly
	if isFqdn {
		urlStringWithProtocol = fmt.Sprintf("%s%s", URL_PROTOCOL_PREFIX, urlString)
	}

	u, err := url.Parse(urlStringWithProtocol)
	if err != nil {
		log.WithFields(fields).Errorf("Error parsing url to add security token, err=%+v", err)
		return urlString
	}
	path := u.Path
	// path already has a leading slash
	path = fmt.Sprintf("/%s/%s%s", SECURITY_TOKEN_KEY, securityToken, path)
	u.Path = path
	uString := u.String()
	// remove protocol
	if isFqdn {
		uString = strings.Replace(uString, URL_PROTOCOL_PREFIX, "", 1)
	}
	return uString
}

func (s *SecurityTokenPathConfig) AddSecurityTokenToUrl(deviceInfo map[string]string, urlString string, fields log.Fields) string {
	fields = common.FilterLogFields(fields)
	isFqdn := isUrlFqdn(urlString)
	if !s.doesUrlNeedToken(urlString, isFqdn, fields) {
		return urlString
	}
	return s.addTokenToUrl(deviceInfo, urlString, isFqdn, fields)
}

func isUrlFqdn(urlString string) bool {
	return !strings.Contains(urlString, FQDN_CHECK)
}

func (s *SecurityTokenPathConfig) doesUrlNeedToken(urlString string, isFqdn bool, fields log.Fields) bool {
	if isFqdn {
		for ssrHost, isEnabled := range s.UrlHostKeywordMap {
			if strings.Contains(strings.ToLower(urlString), ssrHost) {
				if !isEnabled {
					log.WithFields(fields).Debugf("Security token feature is disabled for FQDN host keyword, no security token needed, keyword=%s, url=%s", ssrHost, urlString)
					return false
				}
				log.WithFields(fields).Debugf("Matched FQDN host with keyword to add security token, keyword=%s, url=%s", ssrHost, urlString)
				return true
			}
		}
		log.WithFields(fields).Debugf("No FQDN host keyword found in url to add security token to, url=%s", urlString)
		return false
	}

	url, err := url.Parse(urlString)
	if err != nil {
		log.WithFields(fields).Errorf("Error parsing url to add security token, err=%+v", err)
		return false
	}
	for ssrPath, isEnabled := range s.UrlPathMap {
		if strings.Contains(strings.ToLower(url.Path), ssrPath) {
			if !isEnabled {
				log.WithFields(fields).Debugf("Security token feature is disabled for path, no security token needed, path=%s, url=%s", ssrPath, urlString)
				return false
			}
			log.WithFields(fields).Debugf("Matched ssr path to add security token, path=%s, url=%s", ssrPath, urlString)
			return true
		}
	}
	log.WithFields(fields).Debugf("No ssr path found in url to add security token to, url=%s", urlString)
	return false
}

func CanSkipSecurityTokenForClientProtocol(deviceInfo map[string]string) bool {
	clientProtocol := strings.ToLower(deviceInfo[SECURITY_TOKEN_CLIENT_PROTOCOL])
	return Ws.SecurityTokenConfig.SkipSecurityTokenClientProtocolSet.Contains(clientProtocol)
}
