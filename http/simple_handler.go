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
	"net/http"

	"github.com/rdkcentral/xconfwebconfig/common"
)

func (s *XconfServer) VersionHandler(w http.ResponseWriter, r *http.Request) {
	version := common.Version{
		CodeGitCommit:   s.GetString("xconfwebconfig.code_git_commit"),
		BuildTime:       s.GetString("xconfwebconfig.build_time"),
		BinaryVersion:   common.BinaryVersion,
		BinaryBranch:    common.BinaryBranch,
		BinaryBuildTime: common.BinaryBuildTime,
	}
	WriteOkResponse(w, r, version)
}

func (s *XconfServer) InfoVersionHandler(w http.ResponseWriter, r *http.Request) {
	version := common.InfoVersion{
		ProjectName:    s.GetString("xconfwebconfig.ProjectName"),
		ProjectVersion: s.GetString("xconfwebconfig.ProjectVersion"),
		ServiceName:    s.GetString("xconfwebconfig.ServiceName"),
		ServiceVersion: s.GetString("xconfwebconfig.ServiceVersion"),
		Source:         s.GetString("xconfwebconfig.Source"),
		Rev:            s.GetString("xconfwebconfig.Rev"),
		GitBranch:      s.GetString("xconfwebconfig.GitBranch"),
		GitBuildTime:   s.GetString("xconfwebconfig.GitBuildTime"),
		GitCommitId:    s.GetString("xconfwebconfig.GitCommitId"),
		GitCommitTime:  s.GetString("xconfwebconfig.GitCommitTime"),
	}
	versionTemplate := "<html><body>&lt;ServiceInfo>" +
		"<br> &nbsp;&nbsp; &lt;projectName&gt;" + version.ProjectName + "&lt;/projectName&gt;" +
		"<br> &nbsp;&nbsp; &lt;projectVersion&gt;" + version.ProjectVersion + "&lt;/projectVersion&gt;" +
		"<br> &nbsp;&nbsp; &lt;serviceName&gt;" + version.ServiceName + "&lt;/serviceName&gt;" +
		"<br> &nbsp;&nbsp; &lt;serviceVersion&gt;" + version.ServiceVersion + "&lt;/serviceVersion&gt;" +
		"<br> &nbsp;&nbsp; &lt;source&gt;" + version.Source + "&lt;/source&gt;" +
		"<br> &nbsp;&nbsp; &lt;rev&gt;" + version.Rev + "&lt;/rev&gt;" +
		"<br> &nbsp;&nbsp; &lt;gitBranch&gt;" + version.GitBranch + "&lt;/gitBranch&gt;" +
		"<br> &nbsp;&nbsp; &lt;gitBuildTime&gt;" + version.GitBuildTime + "&lt;/gitBuildTime&gt;" +
		"<br> &nbsp;&nbsp; &lt;gitCommitId&gt;" + version.GitCommitId + "&lt;/gitCommitId&gt;" +
		"<br> &nbsp;&nbsp; &lt;gitCommitTime&gt;" + version.GitCommitTime + "&lt;/gitCommitTime&gt;" +
		"<br>&lt;/ServiceInfo></body></html>"
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(versionTemplate))
}

func (s *XconfServer) MonitorHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-length", "0")
}

func (s *XconfServer) HealthZHandler(w http.ResponseWriter, r *http.Request) {
	WriteOkResponse(w, r, nil)
}

func (s *XconfServer) NotificationHandler(w http.ResponseWriter, r *http.Request) {
	WriteOkResponse(w, r, nil)
}

func (s *XconfServer) ServerConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write(s.ConfigBytes())
}

const (
	Default404ResponseTemplate = `<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=UTF-8"/>
<title>Error 404 Not Found</title>
</head>
<body><h2>HTTP ERROR 404</h2>
<p>Problem accessing %v. Reason:
<pre>    Not Found</pre></p><hr><i><small>Powered by Jetty://</small></i><hr/>

</body>
</html>`
)

func (s *XconfServer) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	body := fmt.Sprintf(Default404ResponseTemplate, r.URL)
	w.Write([]byte(body))
}
