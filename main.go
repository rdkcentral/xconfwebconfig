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
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/dataapi"
	xhttp "github.com/rdkcentral/xconfwebconfig/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const (
	defaultConfigFile = "/app/xconfwebconfig/xconfwebconfig.conf"
)

// main function to boot up everything
func main() {
	started := time.Now().String()

	// parse flag
	configFile := flag.String("f", defaultConfigFile, "config file")
	showVersion := flag.Bool("version", false, "show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("xconfwebconfig version %s (branch %v) %v\n", common.BinaryVersion, common.BinaryBranch, common.BinaryBuildTime)
		os.Exit(0)
	}

	// read new hocon config
	sc, err := common.NewServerConfig(*configFile)
	if err != nil {
		panic(err)
	}
	server := xhttp.NewXconfServer(sc, false, nil)

	// setup logging
	logFile := server.GetString("xconfwebconfig.log.file")
	if len(logFile) > 0 {
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			fmt.Printf("ERROR opening file: %v", err)
			panic(err)
		}
		defer f.Close()
		log.SetOutput(f)
	} else {
		log.SetOutput(os.Stdout)
	}

	logFormat := server.GetString("xconfwebconfig.log.format")
	if logFormat == "text" {
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: common.LOGGING_TIME_FORMAT,
		})
	} else {
		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat: common.LOGGING_TIME_FORMAT,
			FieldMap: log.FieldMap{
				log.FieldKeyTime: "timestamp",
			},
		})
	}
	// Output to stderr instead of stdout, could also be a file.

	// default log level info
	logLevel := log.InfoLevel
	if parsed, err := log.ParseLevel(server.GetString("xconfwebconfig.log.level")); err == nil {
		logLevel = parsed
	}
	log.SetLevel(logLevel)
	// log.SetReportCaller(true)

	// SAT token INIT
	xhttp.InitSatTokenManager(server)
	xhttp.SetLocalSatToken(log.Fields{})

	// setup router
	router := server.GetRouter(false)

	// setup xconf APIs and tables
	dataapi.XconfSetup(server, router)

	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	router.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	router.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	router.Handle("/debug/pprof/block", pprof.Handler("block"))

	if server.MetricsEnabled() {
		router.Handle("/metrics", promhttp.Handler())
		appmetrics := xhttp.NewMetrics()
		metrics := server.SetWebMetrics(appmetrics)
		handler := metrics.MetricsHandler(router)
		server.Handler = handler
	} else {
		server.Handler = router
	}

	// Exit gracefully on Ctrl+C etc.
	done := make(chan bool)

	// Catch the signal and set the channel
	quit := make(chan os.Signal, 1) // Buffered channel here to fix a go vet warning
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// goroutine 1 just waits for a kill signal
		sig := <-quit
		log.Errorf("Caught the %+v signal, exiting", sig)

		// Send a message to the done channel, triggering server shutdown
		done <- true
	}()

	log.Infof("server is starting at %s, port %s", started, server.Addr)
	go func() {
		// goroutine 2 is running the server
		if err := server.ListenAndServe(); err != nil {
			log.Errorf("ListenAndServe Error %+v, exiting", err)
		}
		done <- true
	}()

	// Waiting for either a kill signal or a ListenAndServe error
	<-done

	// K8s has a default terminationGracePeriod as 30 seconds, app's wait period should be
	// slightly less. Using a wait period of 25 secs, should be enough to handle inflight reqs
	// may need to set this in config
	waitPeriod := time.Duration(25) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), waitPeriod)
	defer cancel()

	server.SetKeepAlivesEnabled(false)
	// server.Shutdown will cause "Server closed" err for ListenAndServe, terminating goroutine #2 near line #99
	if err := server.Shutdown(ctx); err != nil {
		log.Errorf("Could not shutdown the web server: %+v\n", err)
	}
	log.Info("xconfwebconfig is shutdown")
}
