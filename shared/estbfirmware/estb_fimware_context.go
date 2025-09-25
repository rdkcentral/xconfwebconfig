/**
 * Copyright 2022 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package estbfirmware

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

const (
	DATE_TIME_FORMATTER     = "1/2/2006 15:04"
	DATE_TIME_SEC_FORMATTER = "01/02/2006 15:04:05"
)

// ConvertedContext ... convert the map string to type based context
type ConvertedContext struct {
	Context             map[string]string   `json:"-"` // orignal Context
	EstbMac             string              `json:"estbMac,omitempty"`
	Env                 string              `json:"env,omitempty"`
	Model               string              `json:"model,omitempty"`
	FirmwareVersion     string              `json:"firmwareVersion,omitempty"`
	EcmMac              string              `json:"ecmMac,omitempty"`
	ReceiverId          string              `json:"receiverId,omitempty"`
	ControllerId        int64               `json:"controllerId,omitempty"`
	ChannelMapId        int64               `json:"channelMapId,omitempty"`
	VodId               int64               `json:"vodId,omitempty"`
	BypassFilters       map[string]struct{} `json:"-"`
	RawBypassFilters    []string            `json:"bypassFilters,omitempty"` // custom unmarshal for BypassFilters
	ForceFilters        map[string]struct{} `json:"-"`
	RawForceFilters     []string            `json:"forceFilters,omitempty"` // custom unmarshal for ForceFilters
	XconfHttpHeader     string              `json:"xconfHttpHeader,omitempty"`
	AccountId           string              `json:"accountId,omitempty"`
	Capabilities        []Capabilities      `json:"capabilities"`
	TimeZone            *time.Location      `json:"-"`
	RawTimeZone         string              `json:"timeZone"` // custom unmarshal for TimeZone
	Time                *time.Time          `json:"-"`
	RawTime             string              `json:"time"` // custom unmarshal for Time
	IpAddress           string              `json:"ipAddress,omitempty"`
	Rcdl                bool                `json:"rcdl"`
	SupportsFullHttpUrl bool                `json:"supportsFullHttpUrl"`
	RebootDecoupled     bool                `json:"rebootDecoupled"`
}

// NewConvertedContext ...
func NewConvertedContext(ctx map[string]string) *ConvertedContext {
	return GetContextConverted(ctx)
}

func (cc *ConvertedContext) UnmarshalJSON(bytes []byte) error {
	type convertedContext ConvertedContext

	err := json.Unmarshal(bytes, (*convertedContext)(cc))
	if err != nil {
		return err
	}

	if cc.RawTimeZone != "" {
		location, err := time.LoadLocation(cc.RawTimeZone)
		if err != nil {
			return err
		}
		cc.TimeZone = location

		if cc.RawTime != "" {
			t, err := time.ParseInLocation(DATE_TIME_SEC_FORMATTER, cc.RawTime, location)

			if err != nil {
				return err
			}
			cc.Time = &t
		}
	}

	cc.BypassFilters = util.NewStringSet(cc.RawBypassFilters)
	cc.ForceFilters = util.NewStringSet(cc.RawForceFilters)
	cc.Rcdl = cc.IsRcdl()
	cc.RebootDecoupled = cc.IsRebootDecoupled()
	cc.SupportsFullHttpUrl = cc.IsSupportsFullHttpUrl()
	return nil
}

func (c *ConvertedContext) GetProperties() map[string]string {
	// copy orignal context
	ctxmap := map[string]string{}
	for k, v := range c.Context {
		ctxmap[k] = v
	}

	// capanbilities
	capabilities := c.GetCapabilities()
	for _, capability := range capabilities {
		ctxmap[capability] = ""
	}

	// see Go time Location struct
	timeZone := offsetToTimeZone(c.GetTimeZoneOffset())
	ctxmap["timeZone"] = timeZone.String()
	tm := c.GetTime()
	if tm != nil {
		ctxmap["time"] = tm.String()
	} else {
		tm := time.Now().In(timeZone)
		ctxmap["time"] = tm.String()
	}

	return ctxmap
}
func GetContextConverted(ctx map[string]string) *ConvertedContext {
	c := &ConvertedContext{
		Context: ctx,
	}
	/**
	 * Requests with invalid mac addresses are junk, we don't care about
	 * them, we just return 500 error and don't even log it.
	 */

	stbmac := c.GetEStbMac()
	if _, err := shared.NewMacAddress(stbmac); err == nil {
		c.SetEstbMacConverted(stbmac)
	}

	c.SetEnvConverted(c.GetEnv())

	c.SetModelConverted(c.GetModel())

	c.SetAccountIdConverted(c.GetAccountId())

	c.FirmwareVersion = c.GetFirmwareVersion()

	c.BypassFilters = map[string]struct{}{}

	c.ForceFilters = map[string]struct{}{}

	c.Capabilities = []Capabilities{}

	ecmMac := c.GetECMMac()
	if _, err := shared.NewMacAddress(ecmMac); err == nil {
		c.SetEcmMacConverted(ecmMac)
	}

	c.ReceiverId = c.GetReceiverId()

	controlerId := c.GetControllerId()
	if i, err := strconv.ParseInt(controlerId, 10, 64); err == nil {
		c.ControllerId = i
	}

	mapId := c.GetChannelMapId()
	if i, err := strconv.ParseInt(mapId, 10, 64); err == nil {
		c.ChannelMapId = i
	}

	vodId := c.GetVodId()
	if i, err := strconv.ParseInt(vodId, 10, 64); err == nil {
		c.VodId = i
	}

	c.TimeZone = offsetToTimeZone(c.GetTimeZoneOffset())
	c.RawTimeZone = c.TimeZone.String()
	c.SetTimeZone(c.RawTimeZone)

	c.SetTimeConverted(c.GetTime())
	// c.RawTime = c.getContextValue(shared.TIME)
	// tm := c.GetTime()
	// c.RawTime = fmt.Sprintf("%d/%d/%d %d:%02d", tm.Month(), tm.Day(), tm.Year(), tm.Hour(), tm.Minute())
	c.SetTime(*c.Time)
	c.RawTime = c.getContextValue(common.TIME)

	c.IpAddress = c.GetIpAddress()

	if len(c.GetCapabilities()) != 0 {
		c.Capabilities = c.CreateCapabilitiesList()
	}

	bypassfilter := c.GetBypassFilters()
	if bypassfilter != "" {
		addFiltersIntoConverted(bypassfilter, c.BypassFilters)
	}

	forcefilter := c.GetForceFilters()
	if forcefilter != "" {
		addFiltersIntoConverted(forcefilter, c.ForceFilters)
	}

	c.SetXconfHttpHeaderConverted(c.GetXconfHttpHeader())

	return c
}

func addFiltersIntoConverted(filterStr string, filters map[string]struct{}) {
	//split := strings.Split(strings.TrimSpace(filterStr), "[,]")
	split := strings.Split(strings.TrimSpace(filterStr), ",")
	for _, f := range split {
		filters[f] = struct{}{}
	}
}

func offsetToTimeZone(offset string) *time.Location {
	if len(offset) == 0 {
		return time.UTC
	}

	aa := strings.Split(offset, ":")
	if len(aa) != 2 {
		return time.UTC
	}

	isNumber := true
	for _, c := range aa[0] {
		if !unicode.IsNumber(c) {
			isNumber = false
			break
		}
	}

	isDigit := true
	for _, c := range aa[1] {
		if !unicode.IsDigit(c) {
			isDigit = false
			break
		}
	}

	if isNumber && isDigit {
		hrInt, err1 := strconv.Atoi(aa[0])
		minInt, err2 := strconv.Atoi(aa[1])
		if err1 == nil && err2 == nil {
			loc := time.FixedZone("UTC-8", -(hrInt*3600 + minInt*60))
			return loc
		}
	}
	return time.UTC
}

func (c *ConvertedContext) CreateCapabilitiesList() []Capabilities {
	capList := []Capabilities{}
	for _, strCap := range c.GetCapabilities() {
		switch strings.ToUpper(strings.TrimSpace(strCap)) {
		case "RCDL":
			capList = append(capList, RCDL)
		case "REBOOTDECOUPLED":
			capList = append(capList, RebootCoupled)
		case "REBOOTCOUPLED":
			capList = append(capList, RebootDecoupled)
		case "SUPPORTSFULLHTTPURL":
			capList = append(capList, SupportsFullHttpUrl)
		default:
			log.Debug(fmt.Sprintf("Unknown capability will be ignored: %s", strCap))
		}
	}
	return capList
}

func (c *ConvertedContext) isThisCap(thisCap Capabilities) bool {
	if len(c.Capabilities) == 0 {
		return false
	}
	for _, cap := range c.Capabilities {
		if cap == thisCap {
			return true
		}
	}
	return false
}

func (c *ConvertedContext) IsRcdl() bool {
	return c.isThisCap(RCDL)
}

func (c *ConvertedContext) IsRebootDecoupled() bool {
	return c.isThisCap(RebootDecoupled)
}

func (c *ConvertedContext) IsSupportsFullHttpUrl() bool {
	return c.isThisCap(SupportsFullHttpUrl)
}

func (c *ConvertedContext) GetEnvConverted() string {
	return c.Env
}

func (c *ConvertedContext) SetEnvConverted(env string) {
	if env != "" {
		c.Env = shared.NewEnvironment(env, "").ID
	}
}

func (c *ConvertedContext) GetModelConverted() string {
	return c.Model
}

func (c *ConvertedContext) SetModelConverted(model string) {
	if model != "" {
		c.Model = shared.NewModel(model, "").ID
	}
}

func (c *ConvertedContext) GetFirmwareVersionConverted() string {
	return c.FirmwareVersion
}

func (c *ConvertedContext) SetFirmwareVersionConverted(firmwareVersion string) {
	c.FirmwareVersion = firmwareVersion
}

func (c *ConvertedContext) GetEcmMacConverted() string {
	return c.EcmMac
}

func (c *ConvertedContext) SetEcmMacConverted(ecmMac string) {
	c.EcmMac = ecmMac
}

func (c *ConvertedContext) GetEstbMacConverted() string {
	return c.EstbMac
}

func (c *ConvertedContext) SetEstbMacConverted(estbMac string) {
	c.EstbMac = estbMac
}

func (c *ConvertedContext) GetReceiverIdConverted() string {
	return c.ReceiverId
}

func (c *ConvertedContext) SetReceiverIdConverted(receiverId string) {
	c.ReceiverId = receiverId
}

func (c *ConvertedContext) GetControllerIdConverted() int64 {
	return c.ControllerId
}

func (c *ConvertedContext) SetControllerIdConverted(controllerId int64) {
	c.ControllerId = controllerId
}

func (c *ConvertedContext) GetChannelMapIdConverted() int64 {
	return c.ChannelMapId
}

func (c *ConvertedContext) SetChannelMapIdConverted(channelMapId int64) {
	c.ChannelMapId = channelMapId
}

func (c *ConvertedContext) GetVodIdConverted() int64 {
	return c.VodId
}

func (c *ConvertedContext) SetVodIdConverted(vodId int64) {
	c.VodId = vodId
}

func (c *ConvertedContext) GetXconfHttpHeaderConverted() string {
	return c.XconfHttpHeader
}

func (c *ConvertedContext) IsXconfHttpHeaderSecureConnection() bool {
	if c.XconfHttpHeader == common.XCONF_HTTPS_VALUE || c.XconfHttpHeader == common.XCONF_MTLS_VALUE || c.XconfHttpHeader == common.XCONF_MTLS_RECOVERY_VALUE || c.XconfHttpHeader == common.XCONF_MTLS_OPTIONAL_VALUE {
		return true
	}
	return false
}

func (c *ConvertedContext) SetXconfHttpHeaderConverted(xconfHttpHeader string) {
	c.XconfHttpHeader = xconfHttpHeader
}

func (c *ConvertedContext) GetAccountIdConverted() string {
	return c.AccountId
}

func (c *ConvertedContext) SetAccountIdConverted(accountId string) {
	c.AccountId = accountId
}

/**
 * This value will always be non-null, it is derived as follows.
 * <p>
 * If "time" parameter was sent in query string, this value will be that
 * value. No time zone offset will be applied.
 * <p>
 * If "time" parameter was not sent in query string, this value will be
 * current UTC time plus time zone offset if specified.
 * <p>
 */

func (c *ConvertedContext) GetTimeConverted() *time.Time {
	return c.Time
}

/**
 * WARNING: time zone must be set before time.
 */

func (c *ConvertedContext) SetTimeConverted(t *time.Time) {
	if t == nil {
		tmp := time.Now()
		t = &tmp
	}
	c.Time = t
}

func (c *ConvertedContext) GetIpAddressConverted() string {
	return c.IpAddress
}

func (c *ConvertedContext) SetIpAddressConverted(ipAddress string) {
	c.IpAddress = ipAddress
}

/**
 * This value will never null. It is derived as follows.
 * <p>
 * If a timeZoneOffset was sent, we use that offset to construct
 * timeZone.
 * <p>
 * If no timeZoneOffset was sent (or if it was invalid), we set timeZone
 * to utc.
 * <p>
 * If timeZone is UTC, we use the OLD and soon to be deprecated IP
 * Address + UTC time blocking filter. If timeZone is anything other
 * than UTC, we use the new local time based blocking filter. Once boot
 * blocking and download scheduling are both fixed, both time based
 * blocking filters will be deprecated.
 */

func (c *ConvertedContext) GetTimeZoneConverted() *time.Location {
	return c.TimeZone
}

func (c *ConvertedContext) GetRawTimeZoneConverted() string {
	return c.RawTimeZone
}

func (c *ConvertedContext) SetTimeZoneConverted(timeZone *time.Location) {
	c.TimeZone = timeZone
}

func (c *ConvertedContext) IsUTCConverted() bool {
	return c.TimeZone == time.UTC
}

func (c *ConvertedContext) GetCapabilitiesConverted() []Capabilities {
	return c.Capabilities
}

func (c *ConvertedContext) SetCapabilitiesConverted(capabilities []Capabilities) {
	c.Capabilities = capabilities
}

func (c *ConvertedContext) GetBypassFiltersConverted() map[string]struct{} {
	return c.BypassFilters
}

func (c *ConvertedContext) SetBypassFiltersConverted(bypassFilters map[string]struct{}) {
	c.BypassFilters = bypassFilters
}

func (c *ConvertedContext) AddBypassFiltersConverted(bypassFilter string) {
	if c.BypassFilters == nil {
		c.BypassFilters = make(map[string]struct{})
	}
	c.BypassFilters[bypassFilter] = struct{}{}
}

func (c *ConvertedContext) GetForceFiltersConverted() map[string]struct{} {
	return c.ForceFilters
}

func (c *ConvertedContext) SetForceFiltersConverted(forceFilters map[string]struct{}) {
	c.ForceFilters = forceFilters
}

func (c *ConvertedContext) AddForceFiltersConverted(forceFilter string) {
	if c.ForceFilters == nil {
		c.ForceFilters = make(map[string]struct{})
	}
	c.ForceFilters[forceFilter] = struct{}{}
}

//  These are get DATA from the oringal context (NOT from converted context)

func (c *ConvertedContext) getContextValue(key string) string {
	val, ok := c.Context[key]
	if ok {
		return val
	}
	return ""
}

func (c *ConvertedContext) setContextValue(key string, val string) {
	c.Context[key] = val
}

func (c *ConvertedContext) GetEStbMac() string {
	return c.getContextValue(common.ESTB_MAC)
}
func (c *ConvertedContext) SetEStbMac(eStbMac string) {
	c.setContextValue(common.ESTB_MAC, eStbMac)
}

func (c *ConvertedContext) GetEnv() string {
	return c.getContextValue(common.ENV)
}

func (c *ConvertedContext) SetEnv(env string) {
	c.setContextValue(common.ENV, env)
}

func (c *ConvertedContext) GetModel() string {
	return c.getContextValue(common.MODEL)
}

func (c *ConvertedContext) SetModel(model string) {
	c.setContextValue(common.MODEL, model)
}

func (c *ConvertedContext) GetFirmwareVersion() string {
	return c.getContextValue(common.FIRMWARE_VERSION)
}

func (c *ConvertedContext) SetFirmwareVersion(firmwareVersion string) {
	c.setContextValue(common.FIRMWARE_VERSION, firmwareVersion)
}

func (c *ConvertedContext) GetECMMac() string {
	return c.getContextValue(common.ECM_MAC)
}

func (c *ConvertedContext) SetECMMac(eCMMac string) {
	c.setContextValue(common.ECM_MAC, eCMMac)
}

func (c *ConvertedContext) GetReceiverId() string {
	return c.getContextValue(common.RECEIVER_ID)
}

func (c *ConvertedContext) SetReceiverId(receiverId string) {
	c.setContextValue(common.RECEIVER_ID, receiverId)
}

func (c *ConvertedContext) GetControllerId() string {
	return c.getContextValue(common.CONTROLLER_ID)
}

func (c *ConvertedContext) SetControllerId(controllerId string) {
	c.setContextValue(common.CONTROLLER_ID, controllerId)
}

func (c *ConvertedContext) GetChannelMapId() string {
	return c.getContextValue(common.CHANNEL_MAP_ID)
}

func (c *ConvertedContext) SetChannelMapId(channelMapId string) {
	c.setContextValue(common.CHANNEL_MAP_ID, channelMapId)
}

func (c *ConvertedContext) GetVodId() string {
	return c.getContextValue(common.VOD_ID)
}

func (c *ConvertedContext) SetVodId(vodId string) {
	c.setContextValue(common.VOD_ID, vodId)
}

func (c *ConvertedContext) GetAccountHash() string {
	return c.getContextValue(common.ACCOUNT_HASH)
}

func (c *ConvertedContext) SetAccountHash(accountHash string) {
	c.setContextValue(common.ACCOUNT_HASH, accountHash)
}

func (c *ConvertedContext) GetXconfHttpHeader() string {
	return c.getContextValue(common.XCONF_HTTP_HEADER)
}

func (c *ConvertedContext) SetXconfHttpHeader(xconfHttpHeader string) {
	c.setContextValue(common.XCONF_HTTP_HEADER, xconfHttpHeader)
}

/**
* This is an optional parameter used mostly for testing to override actual
* local time. This is always LOCAL time. We do NOT apply time zone offset
* to this value. If time zone offset is sent, it is assumed to have already
* been applied to this time.
 */

func (c *ConvertedContext) GetTime() *time.Time {
	stime := c.getContextValue(common.TIME)
	if stime != "" {
		// DateTimeFormat.forPattern("M/d/yyyy H:mm");
		// GO "01/02/2006 15:04" as time pkg
		t, err := time.Parse(DATE_TIME_SEC_FORMATTER, stime)
		if err == nil {
			return &t
		}
		// "9/1/2021 17:32" parse
		t, err = time.Parse(DATE_TIME_FORMATTER, stime)
		if err == nil {
			return &t
		}
		log.Error(fmt.Sprintf("Parse  DateTimeFormat failed for %s", stime))
	}
	t := time.Now()
	return &t
}
func (c *ConvertedContext) SetTime(t time.Time) {
	c.setContextValue(common.TIME, t.Format(DATE_TIME_SEC_FORMATTER))
}

func (c *ConvertedContext) GetIpAddress() string {
	return c.getContextValue(common.IP_ADDRESS)
}
func (c *ConvertedContext) SetIpAddress(ipAddress string) {
	c.setContextValue(common.IP_ADDRESS, ipAddress)
}
func (c *ConvertedContext) GetBypassFilters() string {
	return c.getContextValue(common.BYPASS_FILTERS)
}
func (c *ConvertedContext) SetBypassFilters(bypassFilters string) {
	c.setContextValue(common.BYPASS_FILTERS, bypassFilters)
}
func (c *ConvertedContext) GetForceFilters() string {
	return c.getContextValue(common.FORCE_FILTERS)
}
func (c *ConvertedContext) SetForceFilters(forceFilters string) {
	c.setContextValue(common.FORCE_FILTERS, forceFilters)
}

/**
* Tells us the STB offset from UTC
* http://joda-time.sourceforge.net/timezones.html Will be a string like
* "-04:00". From this we can derive the SBT local time.
* <p>
* The normal case will be that "time" parameter is NOT sent and
* "timeZoneOffset" parameter IS specified. In this case we will derive STB
* local time from current UTC plus this offset.
* <p>
* For testing "time" parameter may be set. If it is set, it is assumed to
* be local time, we do not apply time zone offset to it.
 */
func (c *ConvertedContext) GetTimeZone() string {
	return c.getContextValue(common.TIME_ZONE)
}

func (c *ConvertedContext) SetTimeZone(tz string) {
	c.setContextValue(common.TIME_ZONE, tz)
}

func (c *ConvertedContext) GetTimeZoneOffset() string {
	return c.getContextValue(common.TIME_ZONE_OFFSET)
}

func (c *ConvertedContext) SetTimeZoneOffset(timeZoneOffset string) {
	c.setContextValue(common.TIME_ZONE_OFFSET, timeZoneOffset)
}

func (c *ConvertedContext) GetCapabilities() []string {
	strCapabilities := c.getContextValue(common.CAPABILITIES)
	return strings.Split(strCapabilities, ",")
}

func (c *ConvertedContext) SetCapabilities(capabilities []string) {
	strCapabilities := strings.Join(capabilities, ",")
	c.setContextValue(common.CAPABILITIES, strCapabilities)
}

func (c *ConvertedContext) GetPartnerId() string {
	return c.getContextValue(common.PARTNER_ID)
}

func (c *ConvertedContext) SetPartnerId(partnerId string) {
	c.setContextValue(common.PARTNER_ID, partnerId)
}

func (c *ConvertedContext) GetAccountId() string {
	return c.getContextValue(common.ACCOUNT_ID)
}

func (c *ConvertedContext) SetAccountId(accountId string) {
	c.setContextValue(common.ACCOUNT_ID, accountId)
}

func (c *ConvertedContext) ToString() string {
	return fmt.Sprintf("estbMac=%s model=%s reportedFirmwareVersion=%s env=%s ecmMac=%s receiverId=%s  controllerId=%s channelMapId=%s vodId=%s partnerId=%s accountId=%s capabilities=%s timeZone=%v time=\"%v\"  ipAddress=%s bypassFilters=%v forceFilters=%v",
		c.GetEStbMac(),
		c.GetModel(),
		c.GetFirmwareVersion(),
		c.GetEnv(),
		c.GetECMMac(),
		c.GetReceiverId(),
		c.GetControllerId(),
		c.GetChannelMapId(),
		c.GetVodId(),
		c.GetPartnerId(),
		c.GetAccountId(),
		c.GetCapabilities(),
		c.GetTimeZoneOffset(),
		c.GetTime(),
		c.GetIpAddress(),
		c.GetBypassFilters(),
		c.GetForceFilters(),
	)
}
