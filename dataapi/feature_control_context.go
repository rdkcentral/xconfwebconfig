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
package dataapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/rdkcentral/xconfwebconfig/common"
	xhttp "github.com/rdkcentral/xconfwebconfig/http"
	conversion "github.com/rdkcentral/xconfwebconfig/protobuf"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/shared/rfc"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/btcsuite/btcutil/base58"
	log "github.com/sirupsen/logrus"
)

const (
	ATHENS_EUROPE_TZ = "Europe/Athens"
	DEFAULT_OFFSET   = "UTC+2:00"
	DTGR_PARTNER_ID  = "dt-gr"
	GR_PREFIX        = "GR"
)

type PodData struct {
	AccountId string
	TimeZone  string
	PartnerId string
}

type AccountServiceData struct {
	AccountId string
	TimeZone  string
	PartnerId string
}

type PreprocessedData struct {
	AccountId                   string
	PartnerId                   string
	Model                       string
	ApplicationType             string
	Env                         string
	FwVersion                   string
	EstbIp                      string
	Experience                  string
	RfcHash                     string
	RfcRulesEngineHash          string
	RfcPostProcessingHash       string
	CtxHash                     string
	OfferedFwVersion            string
	OfferedFwRfcHash            string
	OfferedFwRfcRulesEngineHash string
}

type ContextData struct {
	Mac     string `json:"mac"`
	Model   string `json:"model"`
	Partner string `json:"partner"`
	//FirmwareVersion string   `json:"firmwareVersion"`
	SerialNum  string   `json:"serialNum"`
	Experience string   `json:"experience"`
	AccountId  string   `json:"accountId"`
	Tags       []string `json:"tags"` // tags from tagging service and xdas ft
}

func NewContextDataFromContextMap(contextMap map[string]string, tags []string) ContextData {
	return ContextData{
		Mac:     contextMap[common.ESTB_MAC_ADDRESS],
		Model:   contextMap[common.MODEL],
		Partner: contextMap[common.PARTNER_ID],
		//FirmwareVersion: contextMap[common.FIRMWARE_VERSION],
		SerialNum:  contextMap[common.SERIAL_NUM],
		Experience: contextMap[common.EXPERIENCE],
		AccountId:  contextMap[common.ACCOUNT_ID],
		Tags:       tags,
	}
}

func CalculateHashForContextData(data ContextData) (string, error) {
	if data.Tags != nil {
		sort.Strings(data.Tags)
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return util.CalculateHash(string(jsonData)), nil
}

func CompareHashWithXDAS(contextMap map[string]string, xdasHash string, tags []string) (bool, error) {
	contextData := NewContextDataFromContextMap(contextMap, tags)
	calculatedHash, err := CalculateHashForContextData(contextData)
	log.Debugf("Calculated ctx hash: %s, XDAS ctx_hash: %s", calculatedHash, xdasHash)
	if err != nil {
		return false, err
	}
	return calculatedHash == xdasHash, nil
}

// getAccountInfoFromGrpService attempts to retrieve account data
func getAccountInfoFromGrpService(ws *xhttp.XconfServer, contextMap map[string]string, fields log.Fields) (*PodData, *AccountServiceData) {
	var podData *PodData
	var td *AccountServiceData

	var xAccountId *conversion.XBOAccount
	var err error
	var macAddress string
	if util.IsValidMacAddress(contextMap[common.ESTB_MAC_ADDRESS]) {
		macAddress = contextMap[common.ESTB_MAC_ADDRESS]
		macPart := util.RemoveNonAlphabeticSymbols(contextMap[common.ESTB_MAC_ADDRESS])
		xAccountId, err = ws.GroupServiceConnector.GetAccountIdData(macPart, fields)
	}

	if xAccountId == nil && err != nil {
		if util.IsValidMacAddress(contextMap[common.ECM_MAC_ADDRESS]) {
			macAddress = contextMap[common.ECM_MAC_ADDRESS]
			macPart := util.RemoveNonAlphabeticSymbols(contextMap[common.ECM_MAC_ADDRESS])
			xAccountId, err = ws.GroupServiceConnector.GetAccountIdData(macPart, fields)
		}
	}
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Error getting accountId information from Grp Service for ecmMac=%s", macAddress)
		xhttp.IncreaseGrpServiceNotFoundResponseCounter(contextMap[common.MODEL])
		return nil, nil
	}
	if xAccountId != nil {
		accountId := xAccountId.GetAccountId()
		contextMap[common.ACCOUNT_ID] = accountId
		contextMap[common.ACCOUNT_HASH] = util.CalculateHash(accountId)

		accountProducts, err := ws.GroupServiceConnector.GetAccountProducts(accountId, fields)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Errorf("Error getting accountProducts information from Grp Service for AccountId=%s", accountId)
			return nil, nil
		}

		//Extract Partner and TimeZone from ADA response
		if timeZone, ok := accountProducts["TimeZone"]; ok {
			contextMap[common.TIME_ZONE] = timeZone
		}

		if partner, ok := accountProducts["Partner"]; ok && partner != "" {
			contextMap[common.PARTNER_ID] = strings.ToUpper(partner)
		}

		if countryCode, ok := accountProducts["CountryCode"]; ok {
			contextMap[common.COUNTRY_CODE] = countryCode
		}

		if raw, ok := accountProducts["AccountProducts"]; ok && raw != "" {
			var ap map[string]string
			err := json.Unmarshal([]byte(accountProducts["AccountProducts"]), &ap)
			if err == nil {
				for key, val := range ap {
					contextMap[key] = val
				}

				if State, ok := accountProducts["State"]; ok {
					contextMap[common.ACCOUNT_STATE] = State
				}
			} else {
				log.WithFields(fields).Error("Failed to unmarshall AccountProducts")
			}

		}

		xhttp.IncreaseGrpServiceFetchCounter(contextMap[common.MODEL], contextMap[common.PARTNER_ID])
		log.WithFields(fields).Debugf("AddContextForPods AcntId='%s' ,AccntPrd='%v' retrieved from xac/ada", contextMap[common.ACCOUNT_ID], contextMap)
		// Create PodData and AccountServiceData with retrieved information
		podData = &PodData{
			AccountId: contextMap[common.ACCOUNT_ID],
			PartnerId: contextMap[common.PARTNER_ID],
			TimeZone:  contextMap[common.TIME_ZONE],
		}

		td = &AccountServiceData{
			AccountId: contextMap[common.ACCOUNT_ID],
			PartnerId: contextMap[common.PARTNER_ID],
			TimeZone:  contextMap[common.TIME_ZONE],
		}
	}

	return podData, td
}

func AddContextForPods(ws *xhttp.XconfServer, contextMap map[string]string, satToken string, vargs ...log.Fields) (*PodData, *AccountServiceData) {
	var fields log.Fields
	var podData *PodData
	var td *AccountServiceData
	if len(vargs) > 0 {
		fields = vargs[0]
	} else {
		fields = log.Fields{}
	}

	tfields := common.FilterLogFields(fields)
	if Xc.EnableXacGroupService {
		podData, td = getAccountInfoFromGrpService(ws, contextMap, fields)
	}

	if podData == nil {
		log.WithFields(fields).Warn("Fallback Trying via Old Account Service,Failed to Get AccountId via Grp Service")
		if Xc.EnableMacAccountServiceCall && strings.HasPrefix(strings.ToUpper(contextMap[common.SERIAL_NUM]), Xc.AccountServiceMacPrefix) {
			AccountServiceDeviceObject, err := ws.AccountServiceConnector.GetDevices(common.SERIAL_NUMBER_PARAM, contextMap[common.SERIAL_NUM], satToken, fields)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Errorf("Error getting AccountService device information: serialNum=%s", contextMap[common.SERIAL_NUM])
				return podData, td
			}

			td = &AccountServiceData{
				AccountId: AccountServiceDeviceObject.DeviceData.ServiceAccountUri,
				PartnerId: AccountServiceDeviceObject.DeviceData.Partner,
			}

			if AccountServiceDeviceObject.DeviceData.ServiceAccountUri == "" {
				log.WithFields(tfields).Infof("No account found in AccountService for XLE device: serialNum=%s", contextMap[common.SERIAL_NUM])
				return podData, td
			}
			podData = &PodData{
				AccountId: AccountServiceDeviceObject.DeviceData.ServiceAccountUri,
				PartnerId: strings.ToUpper(AccountServiceDeviceObject.DeviceData.Partner),
			}
			if util.IsUnknownValue(contextMap[common.ACCOUNT_ID]) && podData.AccountId != "" {
				contextMap[common.ACCOUNT_ID] = podData.AccountId
			}
			AccountServiceAccountObject, err := ws.AccountServiceConnector.GetAccountData(podData.AccountId, satToken, fields)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Errorf("Error getting AccountService account information for XLE device: accountId=%s, serialNum=%s", podData.AccountId, contextMap[common.SERIAL_NUM])
				return podData, td
			}

			td.TimeZone = AccountServiceAccountObject.AccountData.AccountAttributes.TimeZone

			if AccountServiceAccountObject.AccountData.AccountAttributes.TimeZone == "" {
				log.WithFields(tfields).Infof("No timezone found in AccountService for XLE device: accountId=%s, serialNum=%s", AccountServiceDeviceObject.DeviceData.ServiceAccountUri, contextMap[common.SERIAL_NUM])
				return podData, td
			}
			podData.TimeZone = AccountServiceAccountObject.AccountData.AccountAttributes.TimeZone
			log.WithFields(tfields).Infof("Successfully got AccountService information  for XLE device: accountId=%s, serialNum=%s", AccountServiceDeviceObject.DeviceData.ServiceAccountUri, contextMap[common.SERIAL_NUM])
			xhttp.IncreaseAccountFetchCounter(contextMap[common.MODEL], AccountServiceDeviceObject.DeviceData.Partner)
		} else if Xc.EnableDeviceDBLookup && contextMap[common.SERIAL_NUM] != "" && !strings.HasPrefix(contextMap[common.MODEL], GR_PREFIX) {
			ecmMacAddress, err := ws.GetEcmMacFromPodTable(contextMap[common.SERIAL_NUM])
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Errorf("Error looking up pod information from odp db: serialNum=%s", contextMap[common.SERIAL_NUM])
				return podData, td
			}
			normalizedEcmMac, err := util.MacAddrComplexFormat(ecmMacAddress)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Errorf("Mac address from odp db is invalid: ecmMac=%s, serialNum=%s", ecmMacAddress, contextMap[common.SERIAL_NUM])
				return podData, td
			}
			AccountServiceDeviceObject, err := ws.AccountServiceConnector.GetDevices(common.ECM_MAC_PARAM, normalizedEcmMac, satToken, fields)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Errorf("Error getting AccountService device information: ecmMac=%s, serialNum=%s", normalizedEcmMac, contextMap[common.SERIAL_NUM])
				return podData, td
			}

			td = &AccountServiceData{
				AccountId: AccountServiceDeviceObject.DeviceData.ServiceAccountUri,
				PartnerId: AccountServiceDeviceObject.DeviceData.Partner,
			}

			if AccountServiceDeviceObject.DeviceData.ServiceAccountUri == "" {
				log.WithFields(tfields).Infof("No account found in AccountService: ecmMac=%s, serialNum=%s", normalizedEcmMac, contextMap[common.SERIAL_NUM])
				return podData, td
			}
			podData = &PodData{
				AccountId: AccountServiceDeviceObject.DeviceData.ServiceAccountUri,
				PartnerId: strings.ToUpper(AccountServiceDeviceObject.DeviceData.Partner),
			}
			if util.IsUnknownValue(contextMap[common.ACCOUNT_ID]) && podData.AccountId != "" {
				contextMap[common.ACCOUNT_ID] = podData.AccountId
			}
			accountServiceAccountObject, err := ws.AccountServiceConnector.GetAccountData(podData.AccountId, satToken, fields)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Errorf("Error getting AccountService account information: accountId=%s, ecmMac=%s, serialNum=%s", podData.AccountId, normalizedEcmMac, contextMap[common.SERIAL_NUM])
				return podData, td
			}

			td.TimeZone = accountServiceAccountObject.AccountData.AccountAttributes.TimeZone

			if accountServiceAccountObject.AccountData.AccountAttributes.TimeZone == "" {
				log.WithFields(tfields).Infof("No timezone found in AccountService: accountId=%s, ecmMac=%s, serialNum=%s", AccountServiceDeviceObject.DeviceData.ServiceAccountUri, normalizedEcmMac, contextMap[common.SERIAL_NUM])
				return podData, td
			}
			podData.TimeZone = accountServiceAccountObject.AccountData.AccountAttributes.TimeZone
			xhttp.IncreaseAccountFetchCounter(contextMap[common.MODEL], AccountServiceDeviceObject.DeviceData.Partner)
		} else if Xc.EnableDeviceService && contextMap[common.SERIAL_NUM] != "" && !strings.HasPrefix(contextMap[common.MODEL], GR_PREFIX) {
			deviceServiceObject, err := ws.DeviceServiceConnector.GetMeshPodAccountBySerialNum(contextMap[common.SERIAL_NUM], fields)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Errorf("Error getting Device Service information: serialNum=%s", contextMap[common.SERIAL_NUM])
				return podData, td
			}
			if deviceServiceObject.Status == http.StatusOK && util.IsUnknownValue(contextMap[common.ACCOUNT_ID]) && deviceServiceObject.DeviceServiceData.AccountId != "" {
				contextMap[common.ACCOUNT_ID] = deviceServiceObject.DeviceServiceData.AccountId
			}
			if deviceServiceObject.DeviceServiceData != nil {
				podData = &PodData{
					AccountId: deviceServiceObject.DeviceServiceData.AccountId,
					PartnerId: deviceServiceObject.DeviceServiceData.PartnerId,
					TimeZone:  deviceServiceObject.DeviceServiceData.TimeZone,
				}
			}
		}
	}
	return podData, td
}

func AddFeatureControlContextFromAccountService(ws *xhttp.XconfServer, contextMap map[string]string, satToken string, vargs ...log.Fields) *AccountServiceData {
	var td *AccountServiceData
	var accountId string
	var fields log.Fields
	if len(vargs) > 0 {
		fields = vargs[0]
	} else {
		fields = log.Fields{}
	}
	var err error
	if Xc.EnableXacGroupService {
		if util.IsValidMacAddress(contextMap[common.ESTB_MAC_ADDRESS]) || util.IsValidMacAddress(contextMap[common.ECM_MAC_ADDRESS]) {
			var xAccountId *conversion.XBOAccount
			var err error
			var macAddress string
			if util.IsValidMacAddress(contextMap[common.ESTB_MAC_ADDRESS]) {
				macAddress = contextMap[common.ESTB_MAC_ADDRESS]
				macPart := util.RemoveNonAlphabeticSymbols(contextMap[common.ESTB_MAC_ADDRESS])
				xAccountId, err = ws.GroupServiceConnector.GetAccountIdData(macPart, fields)
			}

			if xAccountId == nil && err != nil {
				if util.IsValidMacAddress(contextMap[common.ECM_MAC_ADDRESS]) {
					macAddress = contextMap[common.ECM_MAC_ADDRESS]
					macPart := util.RemoveNonAlphabeticSymbols(contextMap[common.ECM_MAC_ADDRESS])
					xAccountId, err = ws.GroupServiceConnector.GetAccountIdData(macPart, fields)
				}
			}

			if err != nil {
				log.WithFields(log.Fields{"error": err}).Errorf("Error getting accountId information from Grp Service for ecmMac=%s", macAddress)
				xhttp.IncreaseGrpServiceNotFoundResponseCounter(contextMap[common.MODEL])
			} else {
				if xAccountId != nil && xAccountId.GetAccountId() != "" {
					accountId = xAccountId.GetAccountId()
					contextMap[common.ACCOUNT_ID] = accountId
				}

				accountProducts, err := ws.GroupServiceConnector.GetAccountProducts(accountId, fields)
				if err != nil {
					log.WithFields(log.Fields{"error": err}).Errorf("Error getting accountProducts information from Grp Service for AccountId=%s", accountId)
				} else {
					if partner, ok := accountProducts["Partner"]; ok && partner != "" {
						contextMap[common.PARTNER_ID] = strings.ToUpper(partner)
					}
					td = &AccountServiceData{
						AccountId: accountId,
						PartnerId: contextMap[common.PARTNER_ID],
					}
					contextMap[common.ACCOUNT_HASH] = util.CalculateHash(accountId)

					if countryCode, ok := accountProducts["CountryCode"]; ok {
						contextMap[common.COUNTRY_CODE] = countryCode
					}

					if raw, ok := accountProducts["AccountProducts"]; ok && raw != "" {
						var ap map[string]string
						err := json.Unmarshal([]byte(accountProducts["AccountProducts"]), &ap)
						if err == nil {
							for key, val := range ap {
								contextMap[key] = val
							}

							if State, ok := accountProducts["State"]; ok {
								contextMap[common.ACCOUNT_STATE] = State
							}
						} else {
							log.WithFields(fields).Error("Failed to unmarshall AccountProducts")
						}
					}

					xhttp.IncreaseGrpServiceFetchCounter(contextMap[common.MODEL], contextMap[common.PARTNER_ID])
					log.WithFields(fields).Debugf("AddFeatureControlContextFromAccountService AcntId='%s' ,AccntPrd='%v'  retrieved from xac/ada", contextMap[common.ACCOUNT_ID], contextMap)
					return td
				}
			}
		}
	}

	if Xc.EnableAccountService {
		if util.IsUnknownValue(contextMap[common.ACCOUNT_ID]) || util.IsUnknownValue(contextMap[common.PARTNER_ID]) || util.IsUnknownValue(contextMap[common.ACCOUNT_HASH]) {
			log.WithFields(fields).Warn("Fallback Trying via Old Account Service,Failed to Get AccountId via Grp Service due to Flag Disabled or err")
			var accountServiceObject xhttp.AccountServiceDevices
			if util.IsValidMacAddress(contextMap[common.ESTB_MAC_ADDRESS]) {
				accountServiceObject, err = ws.AccountServiceConnector.GetDevices(common.HOST_MAC_PARAM, contextMap[common.ESTB_MAC_ADDRESS], satToken, fields)
				if err == nil {
					td = &AccountServiceData{
						AccountId: accountServiceObject.DeviceData.ServiceAccountUri,
						PartnerId: accountServiceObject.DeviceData.Partner,
					}
				}
			}
			if accountServiceObject.Id == "" && accountServiceObject.DeviceData.Partner == "" && accountServiceObject.DeviceData.ServiceAccountUri == "" && util.IsValidMacAddress(contextMap[common.ECM_MAC_ADDRESS]) {
				accountServiceObject, err = ws.AccountServiceConnector.GetDevices(common.ECM_MAC_PARAM, contextMap[common.ECM_MAC_ADDRESS], satToken, fields)
				td = &AccountServiceData{
					AccountId: accountServiceObject.DeviceData.ServiceAccountUri,
					PartnerId: accountServiceObject.DeviceData.Partner,
				}
			}
			if accountServiceObject.IsEmpty() {
				xhttp.IncreaseAccountServiceEmptyResponseCounter(contextMap[common.MODEL])
			}

			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Error getting AccountService information")
			} else {
				if accountServiceObject.DeviceData.Partner != "" {
					contextMap[common.PARTNER_ID] = strings.ToUpper(accountServiceObject.DeviceData.Partner)
				}
				if util.IsUnknownValue(contextMap[common.ACCOUNT_ID]) && accountServiceObject.DeviceData.ServiceAccountUri != "" {
					contextMap[common.ACCOUNT_ID] = accountServiceObject.DeviceData.ServiceAccountUri
				}
				if util.IsUnknownValue(contextMap[common.ACCOUNT_HASH]) && accountServiceObject.DeviceData.ServiceAccountUri != "" {
					contextMap[common.ACCOUNT_HASH] = util.CalculateHash(accountServiceObject.DeviceData.ServiceAccountUri)
				}
				xhttp.IncreaseAccountFetchCounter(contextMap[common.MODEL], contextMap[common.PARTNER_ID])
			}

			if Xc.RfcReturnCountryCode {
				// query for account data to get country code only if accountId is not empty or unknown
				if contextMap[common.ACCOUNT_ID] != "" && !util.IsUnknownValue(contextMap[common.ACCOUNT_ID]) {
					if Xc.RfcCountryCodeModelsSet.Contains(contextMap[common.MODEL]) && Xc.RfcCountryCodePartnersSet.Contains(contextMap[common.PARTNER_ID]) {
						var accountObject xhttp.Account
						accountObject, err = ws.AccountServiceConnector.GetAccountData(contextMap[common.ACCOUNT_ID], satToken, fields)
						if err != nil {
							log.WithFields(log.Fields{"error": err}).Error("Error getting AccountService account information")
						} else {
							// Add countryCode to contextMap if available
							if accountObject.AccountData.AccountAttributes.CountryCode != "" {
								contextMap[common.COUNTRY_CODE] = accountObject.AccountData.AccountAttributes.CountryCode
							} else {
								contextMap[common.COUNTRY_CODE] = ""
							}
						}
					}
				}
			}
		}
	}
	return td
}

func NormalizeFeatureControlContext(ws *xhttp.XconfServer, r *http.Request, contextMap map[string]string, fields log.Fields) {
	NormalizeCommonContext(contextMap, common.ESTB_MAC_ADDRESS, common.ECM_MAC_ADDRESS)
	estbIp := util.GetIpAddress(r, contextMap[common.ESTB_IP], fields)
	contextMap[common.ESTB_IP] = estbIp
	// check if request is for partner
	if contextMap[common.APPLICATION_TYPE] == shared.STB {
		if appType := GetApplicationTypeFromPartnerId(contextMap[common.PARTNER_ID]); appType != "" {
			contextMap[common.APPLICATION_TYPE] = appType
		}
	}
}

// AddFeatureControlContext ..
func AddFeatureControlContext(ws *xhttp.XconfServer, r *http.Request, contextMap map[string]string, configSetHash string, vargs ...log.Fields) (*PodData, []string, *AccountServiceData) {
	var fields log.Fields
	var podData *PodData
	var td *AccountServiceData
	if len(vargs) > 0 {
		fields = vargs[0]
	} else {
		fields = log.Fields{}
	}

	contextMap[common.PASSED_PARTNER_ID] = contextMap[common.PARTNER_ID]

	// getting local sat token
	localToken, err := xhttp.GetLocalSatToken(fields)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Error getting sat token from SatService")
		return podData, nil, td
	}
	satToken := localToken.Token

	if Xc.EnableXacGroupService {
		//when accountId is already present ,getting account products,countrycode,partner directly from Xdas ada keyspace
		if contextMap[common.ACCOUNT_ID] != "" || !util.IsUnknownValue(contextMap[common.ACCOUNT_ID]) {
			accountProducts, err := ws.GroupServiceConnector.GetAccountProducts(contextMap[common.ACCOUNT_ID], fields)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Errorf("Error getting accountProducts information from Grp Service for AccountId=%s", contextMap[common.ACCOUNT_ID])
			} else {
				if partner, ok := accountProducts["Partner"]; ok && partner != "" {
					contextMap[common.PARTNER_ID] = strings.ToUpper(partner)
				}
				td = &AccountServiceData{
					AccountId: contextMap[common.ACCOUNT_ID],
					PartnerId: contextMap[common.PARTNER_ID],
				}
				contextMap[common.ACCOUNT_HASH] = util.CalculateHash(contextMap[common.ACCOUNT_ID])

				if countryCode, ok := accountProducts["CountryCode"]; ok {
					contextMap[common.COUNTRY_CODE] = countryCode
				}

				if TimeZone, ok := accountProducts["TimeZone"]; ok {
					contextMap[common.TIME_ZONE] = TimeZone
				}

				if raw, ok := accountProducts["AccountProducts"]; ok && raw != "" {
					var ap map[string]string
					err := json.Unmarshal([]byte(accountProducts["AccountProducts"]), &ap)
					if err == nil {
						for key, val := range ap {
							contextMap[key] = val
						}

						if State, ok := accountProducts["State"]; ok {
							contextMap[common.ACCOUNT_STATE] = State
						}
					} else {
						log.WithFields(fields).Error("Failed to unmarshall AccountProducts")
					}
				}

				xhttp.IncreaseGrpServiceFetchCounter(contextMap[common.MODEL], contextMap[common.PARTNER_ID])
				log.WithFields(fields).Debugf("AddFeatureControlContextFromAccountService AcntId='%s' ,AccntPrd='%v'  retrieved from xac/ada", contextMap[common.ACCOUNT_ID], contextMap)
			}
		}
	}

	// if/else statement to check if we should call DeviceService or AccountService
	if strings.EqualFold("XPC", contextMap[common.ACCOUNT_MGMT]) && util.IsUnknownValue(contextMap[common.ACCOUNT_ID]) {
		podData, td = AddContextForPods(ws, contextMap, satToken, fields)
		xhttp.IncreaseUnknownIdCounter(contextMap[common.MODEL], contextMap[common.PARTNER_ID])
	} else if util.IsUnknownValue(contextMap[common.ACCOUNT_ID]) || util.IsUnknownValue(contextMap[common.PARTNER_ID]) || util.IsUnknownValue(contextMap[common.ACCOUNT_HASH]) {
		td = AddFeatureControlContextFromAccountService(ws, contextMap, satToken, fields)
		xhttp.IncreaseUnknownIdCounter(contextMap[common.MODEL], contextMap[common.PARTNER_ID])
	}
	tags := AddContextFromTaggingService(ws, contextMap, satToken, configSetHash, true, fields)
	ftTags := AddGroupServiceFTContext(Ws, common.ESTB_MAC_ADDRESS, contextMap, false, fields)
	CompareTaggingSources(contextMap, tags, ftTags, fields)
	tags = append(tags, ftTags...)
	return podData, tags, td
}

// PostProcessFeatureControl
func PostProcessFeatureControl(ws *xhttp.XconfServer, contextMap map[string]string, isSecuredConnection bool, podData *PodData) []rfc.FeatureResponse {
	featureResponses := []rfc.FeatureResponse{}
	if contextMap[common.ACCOUNT_MGMT] == "xpc" && strings.HasPrefix(contextMap[common.MODEL], GR_PREFIX) {
		if util.IsUnknownValue(contextMap[common.PASSED_PARTNER_ID]) {
			partnerFeature := CreatePartnerIdFeature(strings.ToLower(DTGR_PARTNER_ID))
			featureResponses = append(featureResponses, rfc.CreateFeatureResponseObject(partnerFeature))
		} else if contextMap[common.PASSED_PARTNER_ID] != "" {
			partnerFeature := CreatePartnerIdFeature(strings.ToLower(contextMap[common.PASSED_PARTNER_ID]))
			featureResponses = append(featureResponses, rfc.CreateFeatureResponseObject(partnerFeature))
		}
	} else if util.IsUnknownValue(contextMap[common.PASSED_PARTNER_ID]) && contextMap[common.PARTNER_ID] != "" {
		partnerFeature := CreatePartnerIdFeature(strings.ToLower(contextMap[common.PARTNER_ID]))
		featureResponses = append(featureResponses, rfc.CreateFeatureResponseObject(partnerFeature))
	}

	if contextMap[common.ACCOUNT_MGMT] == "xpc" && strings.HasPrefix(contextMap[common.MODEL], GR_PREFIX) {
		timeZoneFeatureResp := createUnknownAccountIdFeature(contextMap)
		featureResponses = append(featureResponses, timeZoneFeatureResp)
	} else if isSecuredConnection && Xc.ReturnAccountId && contextMap[common.ACCOUNT_ID] != "" {
		accountIdFeature := CreateAccountIdFeature(contextMap[common.ACCOUNT_ID])
		accountIdFeatureResponse := rfc.CreateFeatureResponseObject(accountIdFeature)
		if podData != nil {
			accountIdFeatureResponse["accountId"] = podData.AccountId
			if podData.PartnerId != "" {
				accountIdFeatureResponse["partnerId"] = strings.ToLower(podData.PartnerId)
			} else {
				accountIdFeatureResponse["partnerId"] = "unknown"
			}
			if podData.TimeZone != "" {
				accountIdFeatureResponse["timeZone"] = podData.TimeZone
				loc, err := time.LoadLocation(podData.TimeZone)
				if err == nil {
					t := time.Now().In(loc)
					accountIdFeatureResponse["tzUTCOffset"] = fmt.Sprintf("UTC%s", t.Format("-07:00"))
				} else {
					accountIdFeatureResponse["tzUTCOffset"] = "unknown"
				}

			} else {
				accountIdFeatureResponse["timeZone"] = "unknown"
				accountIdFeatureResponse["tzUTCOffset"] = "unknown"
			}
		}
		if Xc.RfcReturnCountryCode {
			if Xc.RfcCountryCodeModelsSet.Contains(contextMap[common.MODEL]) && Xc.RfcCountryCodePartnersSet.Contains(contextMap[common.PARTNER_ID]) {
				accountIdFeatureResponse[common.COUNTRY_CODE] = contextMap[common.COUNTRY_CODE]
			}
		}
		featureResponses = append(featureResponses, accountIdFeatureResponse)
	}

	if Xc.ReturnAccountHash && contextMap[common.ACCOUNT_HASH] != "" {
		accountHashFeature := rfc.Feature{
			Name:               "AccountHash",
			FeatureName:        "AccountHash",
			EffectiveImmediate: true,
			Enable:             true,
			ConfigData: map[string]string{
				common.TR181_DEVICE_TYPE_ACCOUNT_HASH: contextMap[common.ACCOUNT_HASH],
			},
		}
		featureResponses = append(featureResponses, rfc.CreateFeatureResponseObject(accountHashFeature))
	}

	return featureResponses
}

func createUnknownAccountIdFeature(contextMap map[string]string) rfc.FeatureResponse {
	accountIdFeature := CreateAccountIdFeature("unknown")
	accountIdFeatureResponse := rfc.CreateFeatureResponseObject(accountIdFeature)
	if util.IsUnknownValue(contextMap[common.PASSED_PARTNER_ID]) {
		accountIdFeatureResponse["partnerId"] = strings.ToLower(DTGR_PARTNER_ID)
	} else if contextMap[common.PASSED_PARTNER_ID] != "" {
		accountIdFeatureResponse["partnerId"] = strings.ToLower(contextMap[common.PASSED_PARTNER_ID])
	}
	accountIdFeatureResponse["timeZone"] = ATHENS_EUROPE_TZ
	accountIdFeatureResponse["tzUTCOffset"] = GetTimezoneOffset()
	return accountIdFeatureResponse
}

func GetTimezoneOffset() string {
	now := time.Now()
	location, err := time.LoadLocation(ATHENS_EUROPE_TZ)
	if err != nil {
		log.Errorf("Error loading location for timezone: %s, err=%+v", ATHENS_EUROPE_TZ, err)
		return DEFAULT_OFFSET
	}
	nowInAthens := now.In(location)
	return fmt.Sprintf("UTC%s", nowInAthens.Format("-07:00"))
}

func GetPreprocessedRfcData(ws *xhttp.XconfServer, contextMap map[string]string, fields log.Fields) *PreprocessedData {
	estbMacAddress, ok := contextMap[common.ESTB_MAC_ADDRESS]
	if !ok {
		log.WithFields(common.FilterLogFields(fields)).Debug("No estbMacAddress address provided, not looking up pre-cook data")
		return nil
	}
	xdasData, err := ws.GroupServiceConnector.GetRfcPrecookDetails(util.AlphaNumericMacAddress(estbMacAddress), fields)
	if err != nil {

		log.WithFields(common.FilterLogFields(fields)).Errorf("Error getting rfc pre-cook data from XDAS, err=%+v", err)
		return nil
	}
	precookEstbIp := util.ConvertIpBytesToString(xdasData.GetEstbIp())
	precookRfcHash := base58.Encode(xdasData.GetRfcHash())
	precookOfferedFwRfcHash := base58.Encode(xdasData.GetOfferedFwRfcHash())
	precookRfcRulesEngineHash := base58.Encode(xdasData.GetRfcPrimaryHash())
	precookOfferedFwRfcRulesEngineHash := base58.Encode(xdasData.GetOfferedFwRfcPrimaryHash())
	precookRfcPostProcessingHash := base58.Encode(xdasData.GetRfcPostProcessingHash())
	ctxHash := base58.Encode(xdasData.CtxHash)
	precookData := &PreprocessedData{
		AccountId:                   xdasData.GetAccountId(),
		PartnerId:                   xdasData.GetPartner(),
		Model:                       xdasData.GetModel(),
		ApplicationType:             xdasData.GetApplicationType(),
		Env:                         xdasData.GetEnv(),
		FwVersion:                   xdasData.GetFwVersion(),
		EstbIp:                      precookEstbIp,
		Experience:                  xdasData.GetExperience(),
		RfcHash:                     precookRfcHash,
		RfcRulesEngineHash:          precookRfcRulesEngineHash,
		RfcPostProcessingHash:       precookRfcPostProcessingHash,
		CtxHash:                     ctxHash,
		OfferedFwVersion:            xdasData.GetOfferedFwVersion(),
		OfferedFwRfcHash:            precookOfferedFwRfcHash,
		OfferedFwRfcRulesEngineHash: precookOfferedFwRfcRulesEngineHash,
	}

	return precookData
}

func GetPreprocessedRfcPostProcessResponse(postProcessConfigsetHash string, fields log.Fields) *[]rfc.FeatureResponse {
	var postProcessingPrecookResponse []rfc.FeatureResponse
	tfields := common.FilterLogFields(fields)

	log.WithFields(tfields).Debug("Looking up precook post-processing response from XPC table.")
	if responseBytes, _, err := Ws.DatabaseClient.GetPrecookDataFromXPC(postProcessConfigsetHash); err != nil {
		log.WithFields(tfields).Errorf("Error during lookup of RFC post-processing response from XPC table, configsetHash=%s, err=%v", postProcessConfigsetHash, err)
		return nil
	} else if err := json.Unmarshal(responseBytes, &postProcessingPrecookResponse); err != nil {
		log.WithFields(tfields).Errorf("Error during unmarshel of RFC post-processing response from XPC table, configsetHash=%s, err=%v", postProcessConfigsetHash, err)
		return nil
	}
	return &postProcessingPrecookResponse
}

func GetPreprocessedRfcRulesEngineResponse(rulesEngineConfigsetHash string, fields log.Fields) *[]rfc.FeatureResponse {
	var ruleEnginePrecookResponse []rfc.FeatureResponse
	tfields := common.FilterLogFields(fields)

	log.WithFields(tfields).Debug("Looking up precook rules engine response from XPC table")
	if responseBytes, _, err := Ws.DatabaseClient.GetPrecookDataFromXPC(rulesEngineConfigsetHash); err != nil {
		log.WithFields(tfields).Errorf("Error getting RFC rules engine response from XPC table, configsetHash=%s, err=%v", rulesEngineConfigsetHash, err)
		return nil
	} else if err := json.Unmarshal(responseBytes, &ruleEnginePrecookResponse); err != nil {
		log.WithFields(tfields).Errorf("Error unmarshalling RFC rules engine response from XPC table, configsetHash=%s, err=%+v", rulesEngineConfigsetHash, err)
		return nil
	}
	return &ruleEnginePrecookResponse
}

func canPrecookRfcResponses(tfields log.Fields) bool {
	var isPrecookEnabled bool
	var startTime string
	var endTime string
	isPrecookEnabled = Xc.EnableRfcPrecook
	startTime = Xc.RfcPrecookStartTime
	endTime = Xc.RfcPrecookEndTime

	if !isPrecookEnabled {
		log.WithFields(tfields).Debugf("RfcPrecook is disabled, not using pre-cooked data.")
		return false
	}
	t := time.Now().In(Xc.RfcPrecookTimeZone)
	timeString := t.Format(Xc.RfcPrecookTimeFormat)
	if len(startTime) == 0 || len(endTime) == 0 {
		return true
	}
	if startTime < endTime {
		if timeString < startTime || timeString > endTime {
			log.WithFields(tfields).Debugf("Current time is not within rfc precook window, not using pre-cooked data. Current time=%s, startTime=%s, endTime=%s", timeString, startTime, endTime)
			return false
		}
		return true
	}
	if timeString > startTime || timeString < endTime {
		return true
	}
	log.WithFields(tfields).Debugf("Current time is not within rfc precook window, not using pre-cooked data. Current time=%s, startTime=%s, endTime=%s", timeString, startTime, endTime)
	return false
}

func CreateAccountIdFeature(accountId string) rfc.Feature {
	return rfc.Feature{
		Name:               "AccountId",
		FeatureName:        "AccountId",
		EffectiveImmediate: true,
		Enable:             true,
		ConfigData: map[string]string{
			common.TR181_DEVICE_TYPE_ACCOUNT_ID: accountId,
		},
	}
}

func CreatePartnerIdFeature(partnerId string) rfc.Feature {
	return rfc.Feature{
		Name:               common.SYNDICATION_PARTNER,
		FeatureName:        common.SYNDICATION_PARTNER,
		EffectiveImmediate: true,
		Enable:             true,
		ConfigData: map[string]string{
			common.TR181_DEVICE_TYPE_PARTNER_ID: partnerId,
		},
	}
}

func generatePrecookDataChangedMetrics(contextMap map[string]string, precookData *PreprocessedData, fields log.Fields) {
	tfields := common.FilterLogFields(fields)
	if contextMap[common.MODEL] != precookData.Model {
		log.WithFields(tfields).Infof("Model changed from precook %s to %s", precookData.Model, contextMap[common.MODEL])
		xhttp.IncreaseModelChangedCounter(contextMap[common.PARTNER_ID], contextMap[common.MODEL])
	}
	if contextMap[common.PARTNER_ID] != precookData.PartnerId {
		log.WithFields(tfields).Infof("PartnerId changed from precook %s to %s", precookData.PartnerId, contextMap[common.PARTNER_ID])
		xhttp.IncreasePartnerChangedCounter(contextMap[common.PARTNER_ID], contextMap[common.MODEL])
	}
	if contextMap[common.FIRMWARE_VERSION] != precookData.FwVersion {
		log.WithFields(tfields).Infof("FirmwareVersion changed from precook %s to %s", precookData.FwVersion, contextMap[common.FIRMWARE_VERSION])
		xhttp.IncreaseFwVersionChangedCounter(contextMap[common.PARTNER_ID], contextMap[common.MODEL])
	}

	if contextMap[common.EXPERIENCE] != precookData.Experience {
		log.WithFields(tfields).Infof("Experience changed from precook %s to %s", precookData.Experience, contextMap[common.EXPERIENCE])
		xhttp.IncreaseExperienceChangedCounter(contextMap[common.PARTNER_ID], contextMap[common.MODEL])
	}

	if contextMap[common.ACCOUNT_ID] != precookData.AccountId {
		log.WithFields(tfields).Infof("AccountId changed from precook %s to %s", precookData.AccountId, contextMap[common.ACCOUNT_ID])
		xhttp.IncreaseAccountIdChangedCounter(contextMap[common.PARTNER_ID], contextMap[common.MODEL])
	}
}

func generatePrecookDataChangedIn200Metrics(contextMap map[string]string, precookData *PreprocessedData, fields log.Fields) {
	tfields := common.FilterLogFields(fields)
	if contextMap[common.MODEL] != precookData.Model {
		log.WithFields(tfields).Infof("Model changed from precook  %s to %s in 200 response", precookData.Model, contextMap[common.MODEL])
		xhttp.IncreaseModelChangedIn200Counter(contextMap[common.PARTNER_ID], contextMap[common.MODEL])
	}
	if contextMap[common.PARTNER_ID] != precookData.PartnerId {
		log.WithFields(tfields).Infof("PartnerId changed from precook %s to %s in 200 response", precookData.PartnerId, contextMap[common.PARTNER_ID])
		xhttp.IncreasePartnerChangedIn200Counter(contextMap[common.PARTNER_ID], contextMap[common.MODEL])
	}
	if contextMap[common.FIRMWARE_VERSION] != precookData.FwVersion {
		log.WithFields(tfields).Infof("FirmwareVersion changed from precook %s to %s in 200 response", precookData.FwVersion, contextMap[common.FIRMWARE_VERSION])
		xhttp.IncreaseFwVersionChangedIn200Counter(contextMap[common.PARTNER_ID], contextMap[common.MODEL])
	}

	if contextMap[common.EXPERIENCE] != precookData.Experience {
		log.WithFields(tfields).Infof("Experience changed from precook %s to %s in 200 response", precookData.Experience, contextMap[common.EXPERIENCE])
		xhttp.IncreaseExperienceChangedIn200Counter(contextMap[common.PARTNER_ID], contextMap[common.MODEL])
	}

	if contextMap[common.ACCOUNT_ID] != precookData.AccountId {
		log.WithFields(tfields).Infof("AccountId changed fromp precook %s to %s in 200 response", precookData.AccountId, contextMap[common.ACCOUNT_ID])
		xhttp.IncreaseAccountIdChangedIn200Counter(contextMap[common.PARTNER_ID], contextMap[common.MODEL])
	}
}
