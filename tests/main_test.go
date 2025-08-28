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
package tests

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/dataapi"
	"github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"

	"github.com/gorilla/mux"
)

var (
	testConfigFile string
	sc             *common.ServerConfig
	server         *xwhttp.XconfServer
	noServerErr    = fmt.Errorf("server is not init")
)

/*
Code is:
Copyright (c) 2023 The Gorilla Authors. All rights reserved.
Licensed under the BSD-3 License
*/
func Walk(r *mux.Router) {
	err := r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			fmt.Println("ROUTE:", pathTemplate)
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			fmt.Println("Path regexp:", pathRegexp)
		}
		queriesTemplates, err := route.GetQueriesTemplates()
		if err == nil {
			fmt.Println("Queries templates:", strings.Join(queriesTemplates, ","))
		}
		queriesRegexps, err := route.GetQueriesRegexp()
		if err == nil {
			fmt.Println("Queries regexps:", strings.Join(queriesRegexps, ","))
		}
		methods, err := route.GetMethods()
		if err == nil {
			fmt.Println("Methods:", strings.Join(methods, ","))
		}
		fmt.Println()
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	testConfigFile = "/app/xconfwebconfig/xconfwebconfig.conf"
	if _, err := os.Stat(testConfigFile); os.IsNotExist(err) {
		testConfigFile = "../config/sample_xconfwebconfig.conf"
		if _, err := os.Stat(testConfigFile); os.IsNotExist(err) {
			panic(fmt.Errorf("config file problem %v", err))
		}
	}
	fmt.Printf("testConfigFile=%v\n", testConfigFile)

	os.Setenv("XPC_KEY", "testXpcKey")
	os.Setenv("SAT_CLIENT_ID", "foo")
	os.Setenv("SAT_CLIENT_SECRET", "bar")
	os.Setenv("AWS_ACCESS_KEY", "testAwsAccessKey")
	os.Setenv("AWS_SECRET_KEY", "testAwsSecretKey")
	os.Setenv("MTLS_AWS_ACCESS_KEY", "testMtlsAwsAccessKey")
	os.Setenv("MTLS_AWS_SECRET_KEY", "testMtlsAwsSecretKey")
	os.Setenv("SKY_AWS_ACCESS_KEY", "testSkyAwsAccessKey")
	os.Setenv("SKY_AWS_SECRET_KEY", "testSkyAwsSecretKey")
	os.Setenv("X1_SSR_KEYS", "test-key-1;test-key-2;test-key-3")
	os.Setenv("SECURITY_TOKEN_KEY", "testSecurityTokenKey")
	os.Setenv("AWS_S3_SSEC_KEY", "testAwsS3SsecKey")
	os.Setenv("MD5_AWS_S3_SSEC_KEY", "testMd5AwsS3SsecKey")

	var err error
	sc, err = common.NewServerConfig(testConfigFile)
	if err != nil {
		panic(err)
	}
	server = xwhttp.NewXconfServer(sc, true, nil)
	defer server.Server.Close()
	xwhttp.InitSatTokenManager(server, true)
	server.SetupMocks()

	// start clean
	db.SetDatabaseClient(server.DatabaseClient)
	defer server.DatabaseClient.Close()
	server.DatabaseClient.SetUp()
	// server.DatabaseClient.TearDown()

	// setup router
	router := server.GetRouter(true)

	// setup Xconf APIs and tables
	dataapi.XconfSetup(server, router)

	log.SetOutput(io.Discard)

	// tear down to start clean
	server.TearDown()

	returnCode := m.Run()

	// tear down to clean up
	server.TearDown()

	os.Exit(returnCode)
}
