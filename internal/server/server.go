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
)

// Server 메인 서버 정보 구조체
type Server struct {
	ctx context.Context
	// wg  sync.WaitGroup
}

// Run 메인 서버 가동
//
// Parameters:
//   - ctx: 종료 컨텍스트
func (s *Server) Run(ctx context.Context) {
	// 종료 컨텍스트 저장
	s.ctx = ctx

	// // TLS 인증서 로드
	// cert, err := tls.LoadX509KeyPair(config.Conf.Server.TLSCertificateFile,
	// 	config.Conf.Server.TLSPrivateKeyFile)
	// if err != nil {
	// 	logger.Log.LogError("%v", err)
	// 	process.SendSignal(config.RunConf.Pid, syscall.SIGUSR1)
	// 	return
	// }

	// // TLS 설정
	// tlsConfig := &tls.Config{
	// 	// 인증서 증명
	// 	Certificates: []tls.Certificate{cert},
	// 	// TLS 버전 설정
	// 	MinVersion: tls.VersionTLS12,
	// 	MaxVersion: tls.VersionTLS13,
	// }

}
