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
package http

import (
	"fmt"
	"time"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

// SatTokenMgr - token manager
type SatTokenMgr struct {
	*SatToken
	testOnly bool
}

// SatToken - response object of sat token from SatService
type SatToken struct {
	Token    string `json:"token"`
	Source   string `json:"source"`
	KeyName  string `json:"key_name"`
	Expiry   string `json:"expiry"`
	TokenTTL int    `json:"token_ttl"`
}

// stm variable
var (
	stm *SatTokenMgr
	Ws  *XconfServer
)

// InitSatTokenManager - init sat token manager
func InitSatTokenManager(ws *XconfServer, args ...bool) {
	log.Debug("init of sat token manager")
	Ws = ws
	stm = NewSatTokenMgr(args...)
}

// NewSatTokenMgr - new SAT token manager
func NewSatTokenMgr(args ...bool) *SatTokenMgr {
	if len(args) > 0 {
		arg := args[0]
		return &SatTokenMgr{
			SatToken: &SatToken{},
			testOnly: arg,
		}
	}
	return &SatTokenMgr{
		SatToken: &SatToken{},
		testOnly: false,
	}
}

// GetSatTokenManager - return Sattoken manager object
func GetSatTokenManager() *SatTokenMgr {
	return stm
}

// GetSatToken logic as below...
//  1. try to check local cache, has token, token is still valid
//  2. try to get from REDIS, has token, token is still valid
//  3. try to get the token from SatService, update local cache and redis Cache
func (s *SatTokenMgr) GetSatToken(fields log.Fields) (string, error) {
	tkn, err := GetLocalSatToken(fields)
	if err != nil {
		return "", err
	}
	return tkn.Token, nil
}

func (s *SatTokenMgr) TestOnly() bool {
	return s.testOnly
}

func (s *SatTokenMgr) SetTestOnly(testOnly bool) {
	s.testOnly = testOnly
}

// GetLocalSatToken - get local sattoken
func GetLocalSatToken(fields log.Fields) (*SatToken, error) {
	if stm.TestOnly() {
		return stm.SatToken, nil
	}
	fields = common.FilterLogFields(fields)

	// check for if we have token or not as well as token is expired or not
	if stm.Token == "" || stm.IsTokenExpired(fields) {
		log.WithFields(fields).Debug("no local token found or expired, getting token from SatService")
		err := SetLocalSatToken(fields)
		if err != nil {
			return nil, err
		}
		return stm.SatToken, nil
	}
	log.WithFields(fields).Debug("used local token")
	return stm.SatToken, nil
}

// SetLocalSatToken - setting up local sat token from SatService
func SetLocalSatToken(fields log.Fields) error {
	// going to SatService to get sattoken
	cb2Token, err := Ws.GetSatTokenFromSatService(fields)
	if err != nil {
		// SatService had error
		return err
	}
	name := Ws.SatServiceConnector.SatServiceName()
	keyname := fmt.Sprintf("sat_token_%s", name)
	stm.SatToken = &SatToken{
		Token:    cb2Token.AccessToken,
		Source:   name,
		KeyName:  keyname,
		Expiry:   GetTokenExpiryTime(),
		TokenTTL: cb2Token.ExpiresIn,
	}
	return nil
}

// GetSatTokenFromSatService - getting sat token from SatService
func GetSatTokenFromSatService(fields log.Fields) (*SatToken, error) {
	log.WithFields(fields).Debug("getting sat token from SatService")
	satToken, err := Ws.GetSatTokenFromSatService(fields)
	if err != nil {
		log.WithFields(fields).Errorf("unable to get token from SatService")
		return nil, err
	}

	name := Ws.SatServiceConnector.SatServiceName()
	keyname := fmt.Sprintf("sat_token_%s", name)

	return &SatToken{
		Token:    satToken.AccessToken,
		Source:   name,
		KeyName:  keyname,
		Expiry:   GetTokenExpiryTime(),
		TokenTTL: satToken.ExpiresIn,
	}, nil
}

// IsTokenExpired - making sure token is still valid
func (xst *SatToken) IsTokenExpired(fields log.Fields) bool {
	expireTs, err := time.Parse("2006-01-02 15:04:05", xst.Expiry) //"2020-03-21 23:45:01"
	if err != nil {
		log.WithFields(fields).Errorf("unable to parse expiry string to timestamp")
		return true
	}
	return util.UtcCurrentTimestamp().After(expireTs)
}

// GetTokenExpiryTime - expiration time of sat token
func GetTokenExpiryTime() string {
	addTime := Ws.Config.GetInt32("xconfwebconfig.sat.SAT_REFRESH_FREQUENCY_IN_HOUR")*60 - Ws.Config.GetInt32("xconfwebconfig.sat.SAT_REFRESH_BUFFER_IN_MINS")
	return util.UtcOffsetTimestamp(int(addTime) * 60).Format("2006-01-02 15:04:05")
}
