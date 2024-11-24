// Copyright 2024 JongHoon Shim and The unisys Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build linux

package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// metricsHandler 헬스 체크 핸들러
//
// Parameters:
//   - c: HTTP 요청 및 응답과 관련된 정보를 포함하는 객체
func metricsHandler(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}

// healthHandler 헬스 체크 핸들러
//
// Parameters:
//   - c: HTTP 요청 및 응답과 관련된 정보를 포함하는 객체
func healthHandler(c *gin.Context) {
	c.AbortWithStatus(http.StatusOK)
}
