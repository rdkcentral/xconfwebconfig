/**
 * Copyright 2025 Comcast Cable Communications Management, LLC
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
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	conversion "github.com/rdkcentral/xconfwebconfig/protobuf"

	"github.com/gorilla/mux"
	"google.golang.org/protobuf/proto"
	"gotest.tools/assert"
)

var (
	testConfig = "../config/sample_xconfwebconfig.conf"
	sc         *common.ServerConfig
)

func GetTestConfig() string {
	return testConfig
}

func ExecuteRequest(r *http.Request, handler http.Handler) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)
	return recorder
}

func GetTestXconfServer(testConfigFile string) (*xwhttp.XconfServer, *mux.Router) {
	if _, err := os.Stat(testConfigFile); os.IsNotExist(err) {
		panic(err)
	}

	os.Setenv("XPC_KEY", "/d7BEl8H78rVtiMwGvughqQMHhSCIGsTuTncw8+q1bo=")
	os.Setenv("SAT_CLIENT_ID", "foo")
	os.Setenv("SAT_CLIENT_SECRET", "bar")
	os.Setenv("AWS_S3_SSEC_KEY", "testAwsS3SsecKey")
	os.Setenv("MD5_AWS_S3_SSEC_KEY", "testMd5AwsS3SsecKey")
	os.Setenv("AWS_ACCESS_KEY", "testAwsAccessKey")
	os.Setenv("AWS_SECRET_KEY", "testAwsSecretKey")
	os.Setenv("MTLS_AWS_ACCESS_KEY", "testMtlsAwsAccessKey")
	os.Setenv("MTLS_AWS_SECRET_KEY", "testMtlsAwsSecretKey")
	os.Setenv("SKY_AWS_ACCESS_KEY", "testSkyAwsAccessKey")
	os.Setenv("SKY_AWS_SECRET_KEY", "testSkyAwsSecretKey")
	os.Setenv("X1_SSR_KEYS", "test-key-1;test-key-2;test-key-3")
	os.Setenv("DATABASE_USER", "cassandra")
	os.Setenv("DATABASE_PASSWORD", "cassandra")
	// these keys below are FAKE keys that were randomly created to be the correct number of digits for testing purposes only, these keys are not used in CI or PROD
	os.Setenv("SKY_PARTNER_KEYS", "test:2687d34a2f2a4e9db5657ad2782c867c2406c4b21a33c7ec6be479137a4cfa77df3ecacdaf7fe1db5385e2aca2e69397b8ad2c59285e50d0bddb9d4de07fbbe1;partner1:40ce6eb1f2b2d00d8c8d2ccf55dfd2913c7cb92cfebaaefa86e1c209e37c6bf145aa30cbacf78f00366201d1421f59cd69f8ef3d2bcfa049241f7d6ca1ac28a6")
	os.Setenv("SECURITY_TOKEN_KEY", "dGVzdC1jMDctZDBkMS00MTBiLTg5Y2EtNmM1NWY1ZTU=")

	var err error
	sc, err = common.NewServerConfig(testConfigFile)
	if err != nil {
		panic(err)
	}
	server := xwhttp.NewXconfServer(sc, true, nil)
	xwhttp.InitSatTokenManager(server, true)
	router := server.GetRouter(true)

	XconfSetup(server, router)
	return server, router
}

func GetTestXconfServerDefault() (*xwhttp.XconfServer, *mux.Router) {
	testConfigFile := "/app/xconfwebconfig/xconfwebconfig.conf"
	return GetTestXconfServer(testConfigFile)
}

func AddCpeToDeviceTable(server *xwhttp.XconfServer, serialNum string, ecmMac string) {
	c := server.DatabaseClient.(*db.CassandraClient)
	stmt := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s"."pod_cpe_account" (pod_id text, cpe_mac text, PRIMARY KEY (pod_id))`, c.GetDeviceKeyspace())
	err := c.Query(stmt).Exec()
	stmt = fmt.Sprintf(`INSERT INTO %s.pod_cpe_account (pod_id,cpe_mac) VALUES ('%s','%s');`, c.GetDeviceKeyspace(), serialNum, ecmMac)
	err = c.Query(stmt).Exec()
	fmt.Printf("error: %+v", err)
}

func TruncateDeviceTable(server *xwhttp.XconfServer) {
	c := server.DatabaseClient.(*db.CassandraClient)
	stmt := fmt.Sprintf(`TRUNCATE "%s"."pod_cpe_account"`, c.GetDeviceKeyspace())
	c.Query(stmt).Exec()
}

func SetupGroupServiceMockServerOkResponse(t *testing.T, server xwhttp.XconfServer, path string, cpeGroup *conversion.CpeGroup) *httptest.Server {
	mockedGroupServiceResponse, _ := proto.Marshal(cpeGroup)
	groupServiceMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusOK)
				w.Write(mockedGroupServiceResponse)
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))
	server.SetGroupServiceHost(groupServiceMockServer.URL)
	targetGroupServiceHost := server.GroupServiceHost()
	assert.Equal(t, groupServiceMockServer.URL, targetGroupServiceHost)
	return groupServiceMockServer
}

func SetupGroupServiceMockServer500Response(t *testing.T, server xwhttp.XconfServer) *httptest.Server {
	groupServiceMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error Msg"))
		}))
	server.SetGroupServiceHost(groupServiceMockServer.URL)
	targetGroupServiceHost := server.GroupServiceHost()
	assert.Equal(t, groupServiceMockServer.URL, targetGroupServiceHost)
	return groupServiceMockServer
}

func SetupSatServiceMockServerOkResponse(t *testing.T, server xwhttp.XconfServer) *httptest.Server {
	mockedSatServiceResponse := []byte(`{"access_token":"one_mock_token","expires_in":86400,"scope":"scope1 scope2 scope3","token_type":"Bearer"}`)
	SatServiceMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(mockedSatServiceResponse)
		}))
	server.SetSatServiceHost(SatServiceMockServer.URL)
	targetSatServiceHost := server.SatServiceHost()
	assert.Equal(t, SatServiceMockServer.URL, targetSatServiceHost)
	return SatServiceMockServer
}

func SetupSatServiceMockServerErrorResponse(t *testing.T, server xwhttp.XconfServer) *httptest.Server {
	satServiceMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
	server.SetSatServiceHost(satServiceMockServer.URL)
	targetSatServiceHost := server.SatServiceHost()
	assert.Equal(t, satServiceMockServer.URL, targetSatServiceHost)
	return satServiceMockServer
}

func SetupDeviceServiceMockServerOkResponse(t *testing.T, server xwhttp.XconfServer, path string) *httptest.Server {
	mockedDeviceServiceResponse := []byte(`{"status":200,"data":{"account_id":"testAccountId", "cpe_mac":"testCpeMac", "timezone": "America/New_York", "partner_id": "partner"}}`)
	deviceServiceMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusOK)
				w.Write(mockedDeviceServiceResponse)
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))
	server.SetDeviceServiceHost(deviceServiceMockServer.URL)
	targetDeviceServiceHost := server.DeviceServiceHost()
	assert.Equal(t, deviceServiceMockServer.URL, targetDeviceServiceHost)
	return deviceServiceMockServer
}

func SetupDeviceServiceMockServerOkResponseDynamic(t *testing.T, server xwhttp.XconfServer, response []byte, path string) *httptest.Server {
	deviceServiceMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusOK)
				w.Write(response)
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))
	server.SetDeviceServiceHost(deviceServiceMockServer.URL)
	targetDeviceServiceHost := server.DeviceServiceHost()
	assert.Equal(t, deviceServiceMockServer.URL, targetDeviceServiceHost)
	return deviceServiceMockServer
}

func SetupDeviceServiceMockServerErrorResponse(t *testing.T, server xwhttp.XconfServer, path string) *httptest.Server {
	deviceServiceMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))
	server.SetDeviceServiceHost(deviceServiceMockServer.URL)
	targetDeviceServiceHost := server.DeviceServiceHost()
	assert.Equal(t, deviceServiceMockServer.URL, targetDeviceServiceHost)
	return deviceServiceMockServer
}

func SetupAccountServiceMockServerOkResponse(t *testing.T, server xwhttp.XconfServer, path string) *httptest.Server {
	mockedAccountServiceResponse := []byte(`[{"data":{"serviceAccountId":"testServiceAccountUri","partner":"testPartnerId"},"id":"testId"}]`)
	accountServiceMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusOK)
				w.Write(mockedAccountServiceResponse)
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))
	server.SetAccountServiceHost(accountServiceMockServer.URL)
	targetAccountServiceHost := server.AccountServiceHost()
	assert.Equal(t, accountServiceMockServer.URL, targetAccountServiceHost)
	return accountServiceMockServer
}

func SetupAccountServiceMockServerOkResponseDynamic(t *testing.T, server xwhttp.XconfServer, response []byte, path string) *httptest.Server {
	accountServiceMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusOK)
				w.Write(response)
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))
	server.SetAccountServiceHost(accountServiceMockServer.URL)
	targetAccountServiceHost := server.AccountServiceHost()
	assert.Equal(t, accountServiceMockServer.URL, targetAccountServiceHost)
	return accountServiceMockServer
}

func SetupAccountServiceMockServerOkResponseDynamicTwoCalls(t *testing.T, server xwhttp.XconfServer, response []byte, response2 []byte, path string, path2 string) *httptest.Server {
	accountServiceMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusOK)
				w.Write(response)
			} else if strings.Contains(r.RequestURI, path2) {
				w.WriteHeader(http.StatusOK)
				w.Write(response2)
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))
	server.SetAccountServiceHost(accountServiceMockServer.URL)
	targetAccountServiceHost := server.AccountServiceHost()
	assert.Equal(t, accountServiceMockServer.URL, targetAccountServiceHost)
	return accountServiceMockServer
}

func SetupAccountServiceMockServerEmptyResponse(t *testing.T, server xwhttp.XconfServer, path string) *httptest.Server {
	mockedAccountServiceResponse := []byte(`[]`)
	accountServiceMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusOK)
				w.Write(mockedAccountServiceResponse)
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))
	server.SetAccountServiceHost(accountServiceMockServer.URL)
	targetAccountServiceHost := server.AccountServiceHost()
	assert.Equal(t, accountServiceMockServer.URL, targetAccountServiceHost)
	return accountServiceMockServer
}

func SetupAccountServiceMockServerErrorResponse(t *testing.T, server xwhttp.XconfServer, path string) *httptest.Server {
	accountServiceMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))
	server.SetAccountServiceHost(accountServiceMockServer.URL)
	targetAccountServiceHost := server.AccountServiceHost()
	assert.Equal(t, accountServiceMockServer.URL, targetAccountServiceHost)
	return accountServiceMockServer
}

func SetupTaggingMockServerOkResponse(t *testing.T, server xwhttp.XconfServer, path string) *httptest.Server {
	mockedTaggingResponse := []byte(`["value1", "value2", "value3"]`)
	taggingMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusOK)
				w.Write(mockedTaggingResponse)
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))
	server.SetTaggingHost(taggingMockServer.URL)
	targetTaggingHost := server.TaggingHost()
	assert.Equal(t, taggingMockServer.URL, targetTaggingHost)
	return taggingMockServer
}

func SetupTaggingMockServerOkResponseDynamic(t *testing.T, server xwhttp.XconfServer, response string, path string) *httptest.Server {
	mockedTaggingResponse := []byte(response)
	taggingMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusOK)
				w.Write(mockedTaggingResponse)
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))

	server.SetTaggingHost(taggingMockServer.URL)
	targetTaggingHost := server.TaggingHost()
	assert.Equal(t, taggingMockServer.URL, targetTaggingHost)
	return taggingMockServer
}

func SetupTaggingMockServerEmptyResponse(t *testing.T, server xwhttp.XconfServer, path string) *httptest.Server {
	mockedTaggingResponse := []byte(`[]`)
	taggingMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusOK)
				w.Write(mockedTaggingResponse)
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))
	server.SetTaggingHost(taggingMockServer.URL)
	targetTaggingHost := server.TaggingHost()
	assert.Equal(t, taggingMockServer.URL, targetTaggingHost)
	return taggingMockServer
}

func SetupTaggingMockServer404Response(t *testing.T, server xwhttp.XconfServer, path string) *httptest.Server {
	taggingMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Error Msg"))
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))
	server.SetTaggingHost(taggingMockServer.URL)
	targetTaggingHost := server.TaggingHost()
	assert.Equal(t, taggingMockServer.URL, targetTaggingHost)
	return taggingMockServer
}

func SetupTaggingMockServer500Response(t *testing.T, server xwhttp.XconfServer, path string) *httptest.Server {
	taggingMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Error Msg"))
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))
	server.SetTaggingHost(taggingMockServer.URL)
	targetTaggingHost := server.TaggingHost()
	assert.Equal(t, taggingMockServer.URL, targetTaggingHost)
	return taggingMockServer
}

func SetupGroupServiceFTMockServerOkResponse(t *testing.T, server xwhttp.XconfServer, path string, xdasHashes *conversion.XdasHashes) *httptest.Server {
	mockedGroupServiceResponse, _ := proto.Marshal(xdasHashes)
	groupServiceMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusOK)
				w.Write(mockedGroupServiceResponse)
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))
	server.SetGroupServiceHost(groupServiceMockServer.URL)
	targetGroupServiceHost := server.GroupServiceHost()
	assert.Equal(t, groupServiceMockServer.URL, targetGroupServiceHost)
	return groupServiceMockServer
}

func SetupGroupServiceFTMockServerOkResponseMultipleCalls(t *testing.T, server xwhttp.XconfServer, path1 string, path2 string, path3 string, xdasHashes1 *conversion.XdasHashes, xdasHashes2 *conversion.XdasHashes, xdasHashes3 *conversion.XdasHashes) *httptest.Server {
	mockedGroupServiceResponse1, _ := proto.Marshal(xdasHashes1)
	mockedGroupServiceResponse2, _ := proto.Marshal(xdasHashes2)
	mockedGroupServiceResponse3, _ := proto.Marshal(xdasHashes3)
	groupServiceMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path1) {
				w.WriteHeader(http.StatusOK)
				w.Write(mockedGroupServiceResponse1)
			} else if strings.Contains(r.RequestURI, path2) {
				w.WriteHeader(http.StatusOK)
				w.Write(mockedGroupServiceResponse2)
			} else if strings.Contains(r.RequestURI, path3) {
				w.WriteHeader(http.StatusOK)
				w.Write(mockedGroupServiceResponse3)
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))
	server.SetGroupServiceHost(groupServiceMockServer.URL)
	targetGroupServiceHost := server.GroupServiceHost()
	assert.Equal(t, groupServiceMockServer.URL, targetGroupServiceHost)
	return groupServiceMockServer
}

func SetupGroupServiceHashesMockServerOkResponse(t *testing.T, server xwhttp.XconfServer, xdasHashes *conversion.XdasHashes) *httptest.Server {
	mockedGroupServiceResponse, _ := proto.Marshal(xdasHashes)
	groupServiceMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(mockedGroupServiceResponse)
		}))
	server.SetGroupServiceHost(groupServiceMockServer.URL)
	targetGroupServiceHost := server.GroupServiceHost()
	assert.Equal(t, groupServiceMockServer.URL, targetGroupServiceHost)
	return groupServiceMockServer
}
