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

/*
Package server 메인 서버 패키지
*/
package server

import (
	"context"
	"crypto/tls"
	"net/http"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/meloncoffee/unisys/config"
	"github.com/meloncoffee/unisys/internal/logger"
	"github.com/meloncoffee/unisys/pkg/util/process"
	"golang.org/x/crypto/acme/autocert"
)

// Server 메인 서버 정보 구조체
type Server struct{}

// Run 메인 서버 가동
//
// Parameters:
//   - ctx: 종료 컨텍스트
func (s *Server) Run(ctx context.Context) {
	var err error
	isTLS := true
	port := config.Conf.Server.Port

	// 서버 공통 설정
	server := &http.Server{
		// gin 엔진 설정
		Handler: s.newGinRouterEngine(),
		// 요청 타임아웃 10초 설정
		ReadTimeout: 10 * time.Second,
		// 응답 타임아웃 10초 설정
		WriteTimeout: 10 * time.Second,
		// 요청 헤더 최대 크기를 1MB로 설정
		MaxHeaderBytes: 1 << 20,
	}

	// Auto TLS 사용 옵션 체크
	if config.Conf.Server.AutoTLS.Enabled {
		// Auto TLS 옵션 유효성 검사
		if config.Conf.Server.AutoTLS.Host == "" ||
			config.Conf.Server.AutoTLS.CertPath == "" {
			logger.Log.LogError("invalid auto TLS options (host:%s, cert path: %s)",
				config.Conf.Server.AutoTLS.Host, config.Conf.Server.AutoTLS.CertPath)
			process.SendSignal(config.RunConf.Pid, syscall.SIGUSR1)
			return
		}

		// Let's Encrypt 기반의 auto TLS(HTTPS) 구성
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(config.Conf.Server.AutoTLS.Host),
			Cache:      autocert.DirCache(config.Conf.Server.AutoTLS.CertPath),
		}

		// 서버 주소 및 TLS 설정
		server.Addr = ":https"
		server.TLSConfig = &tls.Config{GetCertificate: m.GetCertificate}
		port = 443
	} else if config.Conf.Server.TLSEnabled {
		// TLS 인증서 파일 옵션 유효성 검사
		if config.Conf.Server.TLSCertificateFile == "" ||
			config.Conf.Server.TLSPrivateKeyFile == "" {
			logger.Log.LogError("invalid TLS certificate files (cert: %s, private key: %s)",
				config.Conf.Server.TLSCertificateFile, config.Conf.Server.TLSPrivateKeyFile)
			process.SendSignal(config.RunConf.Pid, syscall.SIGUSR1)
			return
		}

		// TLS 설정
		tlsConf := &tls.Config{
			// 최소 TLS 지원 버전 설정
			MinVersion: tls.VersionTLS12,
		}
		if tlsConf.NextProtos == nil {
			// 애플리케이션 계층 프로토콜(HTTP/1.1, HTTP/2) 설정
			tlsConf.NextProtos = []string{"h2", "http/1.1"}
		}

		// TLS 인증서 로드
		tlsConf.Certificates = make([]tls.Certificate, 1)
		tlsConf.Certificates[0], err = tls.LoadX509KeyPair(config.Conf.Server.TLSCertificateFile,
			config.Conf.Server.TLSPrivateKeyFile)
		if err != nil {
			logger.Log.LogError("failed to load https cert file: %v", err)
			process.SendSignal(config.RunConf.Pid, syscall.SIGUSR1)
			return
		}

		// 서버 주소 및 TLS 설정
		server.Addr = ":" + strconv.Itoa(port)
		server.TLSConfig = tlsConf
	} else {
		// 서버 주소 설정 (TLS 사용 안함)
		server.Addr = ":" + strconv.Itoa(port)
		isTLS = false
	}

	// HTTP 서버 가동
	if isTLS {
		go func() {
			err := server.ListenAndServeTLS("", "")
			if err != nil && err != http.ErrServerClosed {
				logger.Log.LogError("server error occurred: %v", err)
				process.SendSignal(config.RunConf.Pid, syscall.SIGUSR1)
			}
		}()
	} else {
		go func() {
			err := server.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				logger.Log.LogError("server error occurred: %v", err)
				process.SendSignal(config.RunConf.Pid, syscall.SIGUSR1)
			}
		}()
	}

	logger.Log.LogInfo("server listening on port %d", port)

	// 서버 종료 신호 대기
	<-ctx.Done()

	// 종료 신호를 받았으면 graceful shutdown을 위해 5초 타임아웃 설정
	timeout := time.Duration(config.Conf.Server.ShutdownTimeout) * time.Second
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 서버 종료
	err = server.Shutdown(shutdownCtx)
	if err != nil {
		logger.Log.LogWarn("server shutdown: %v", err)
		return
	}

	logger.Log.LogInfo("server shutdown on port %d", port)
}

// newRouterEngine gin 엔진 생성
//
// Returns:
//   - *gin.Engine: gin 엔진
func (s *Server) newGinRouterEngine() *gin.Engine {
	// gin 모드 설정
	gin.SetMode(func() string {
		if config.RunConf.DebugMode {
			return gin.DebugMode
		}
		return gin.ReleaseMode
	}())

	// gin 라우터 생성
	r := gin.New()
	// 요청/응답 정보 로깅 미들웨어 등록
	r.Use(s.ginLoggerMiddleware())
	// 복구 미들웨어 등록
	r.Use(gin.Recovery())
	// 버전 정보 미들웨어 등록
	r.Use(s.versionMiddleware())

	// 요청 핸들러 등록
	r.GET(config.Conf.API.MetricURI, metricsHandler)
	r.GET(config.Conf.API.HealthURI, healthHandler)
	r.GET("/version", versionHandler)
	r.GET("/", rootHandler)

	return r
}

// ginLoggerMiddleware gin 요청/응답 정보 로깅 미들웨어
//
// Returns:
//   - gin.HandlerFunc: gin 로거 핸들러
func (s *Server) ginLoggerMiddleware() gin.HandlerFunc {
	// 로깅에서 제외할 경로 설정
	excludePath := map[string]struct{}{
		config.Conf.API.MetricURI: {},
		config.Conf.API.HealthURI: {},
	}

	return func(c *gin.Context) {
		// 요청 시작 시간 획득
		start := time.Now()
		// 요청 경로 획득
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		// 요청 처리
		c.Next()

		// 제외할 경로는 로깅하지 않음
		if _, ok := excludePath[path]; ok {
			return
		}

		// 요청 종료 시간 및 latency 계산
		end := time.Now()
		latency := end.Sub(start)

		// 로그 메시지 설정
		var logMsg string
		if len(c.Errors) > 0 {
			logMsg = c.Errors.String()
		} else {
			logMsg = "Request"
		}
		// 상태 코드 획득
		statusCode := c.Writer.Status()
		// 요청 메서드 획득
		method := c.Request.Method
		// 요청 클라이언트 IP 획득
		clientIP := c.ClientIP()
		// 사용자 에이전트 획득
		userAgent := c.Request.UserAgent()
		// 응답 바디 사이즈 획득
		resBodySize := c.Writer.Size()

		// 로그 출력 (상태 코드에 따른 로그 레벨 설정)
		if statusCode >= 500 {
			logger.Log.LogError("[%d] %s %s (IP: %s, Latency: %v, UA: %s, ResSize: %d) %s",
				statusCode, method, path, clientIP, latency, userAgent, resBodySize, logMsg)
		} else if statusCode >= 400 {
			logger.Log.LogWarn("[%d] %s %s (IP: %s, Latency: %v, UA: %s, ResSize: %d) %s",
				statusCode, method, path, clientIP, latency, userAgent, resBodySize, logMsg)
		} else {
			logger.Log.LogInfo("[%d] %s %s (IP: %s, Latency: %v, UA: %s, ResSize: %d) %s",
				statusCode, method, path, clientIP, latency, userAgent, resBodySize, logMsg)
		}
	}
}

// versionMiddleware 버전 정보 미들웨어
//
// Returns:
//   - gin.HandlerFunc: 버전 정보 핸들러
func (s *Server) versionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-UNISYS-VERSION", config.Version)
		c.Next()
	}
}
