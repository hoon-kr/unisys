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
Package config 설정 관리 패키지
*/
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	BuildTime  = "unknown" // 빌드 시 값 세팅됨
	Version    = "1.0.0"
	ModuleName = "unisys"
)

const (
	ConfFilePath = "conf/unisys.yaml"
	PidFilePath  = "var/unisys.pid"
	LogFilePath  = "log/unisys.log"
)

// Config 전역 설정 정보 구조체
type Config struct {
	// 서버 설정
	Server struct {
		// 서버 리슨 포트 (DEF:8443, MIN:1, MAX:65535)
		Port int `yaml:"port"`
		// 서버 셧다운 타임아웃 (DEF:5sec, MIN:1sec, MAX:20sec)
		ShutdownTimeout int `yaml:"shutdownTimeout"`
		// TLS 사용 설정 (DEF:false)
		TLSEnabled bool `yaml:"tlsEnabled"`
		// TLS 인증서 파일 경로
		TLSCertificateFile string `yaml:"tlsCertificateFile"`
		// 서버 Private Key 파일 경로
		TLSPrivateKeyFile string `yaml:"tlsPrivateKeyFile"`
		// Let's Encrypt 사용 설정
		AutoTLS AutoTLSYaml `yaml:"autoTLS"`
	} `yaml:"server"`

	// API 설정
	API struct {
		// 애플리케이션 메트릭을 제공하는 엔드포인트 (DEF: /metrics)
		MetricURI string `yaml:"metricURI"`
		// 애플리케이션 상태 점검을 위한 엔드포인트 (DEF: /health)
		HealthURI string `yaml:"healthURI"`
		// 서버 상태 정보를 제공하는 엔드포인트 (DEF: /sys/stats)
		SysStatURI string `yaml:"sysStatURI"`
	} `yaml:"api"`

	// 로그 설정
	Log struct {
		// 최대 로그 파일 사이즈 (DEF:100MB, MIN:1MB, MAX:1000MB)
		MaxLogFileSize int `yaml:"maxLogFileSize"`
		// 최대 로그 파일 백업 개수 (DEF:10, MIN:1, MAX:100)
		MaxLogFileBackup int `yaml:"maxLogFileBackup"`
		// 최대 백업 로그 파일 유지 기간(일) (DEF:90, MIN:1, MAX:365)
		MaxLogFileAge int `yaml:"maxLogFileAge"`
		// 백업 로그 파일 압축 여부 (DEF:true, ENABLE:true, DISABLE:false)
		CompBakLogFile bool `yaml:"compressBackupLogFile"`
	} `yaml:"log"`
}

// AutoTLSYaml Let's Encrypt 설정 구조체
type AutoTLSYaml struct {
	// AutoTLS(Let's Encrypt) 사용 여부 (DEF:false)
	Enabled bool `yaml:"enabled"`
	// Let's Encrypt로 부터 자동 발급된 TLS 인증서가 저장될 경로 (DEF:.cache)
	CertPath string `yaml:"certPath"`
	// TLS 인증서를 발급받을 도메인
	Host string `yaml:"host"`
}

// RunConfig 런타임 전역 설정 정보 구조체
type RunConfig struct {
	DebugMode bool
	Pid       int
}

var Conf Config
var RunConf RunConfig

// init config 패키지 임포트 시 자동 초기화
func init() {
	Conf.Server.Port = 8443
	Conf.Server.ShutdownTimeout = 5
	Conf.Server.TLSEnabled = false
	Conf.Server.TLSCertificateFile = ""
	Conf.Server.TLSPrivateKeyFile = ""
	Conf.Server.AutoTLS.Enabled = false
	Conf.Server.AutoTLS.CertPath = ".cache"
	Conf.Server.AutoTLS.Host = ""
	Conf.API.MetricURI = "/metrics"
	Conf.API.HealthURI = "/health"
	Conf.API.SysStatURI = "/sys/stats"
	Conf.Log.MaxLogFileSize = 100
	Conf.Log.MaxLogFileBackup = 10
	Conf.Log.MaxLogFileAge = 90
	Conf.Log.CompBakLogFile = true
}

// LoadConfig 설정 파일 로드
//
// Parameters:
//   - filePath: 설정 파일 경로
//
// Returns:
//   - error: 성공(nil), 실패(error)
func (c *Config) LoadConfig(filePath string) error {
	// YAML 설정 파일 열기
	file, err := os.Open(ConfFilePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// YAML 디코더 생성
	decoder := yaml.NewDecoder(file)

	// YAML 파싱 및 디코딩
	err = decoder.Decode(&Conf)
	if err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}

	// 설정 값 유효성 검사
	if c.Server.Port < 0 || c.Server.Port > 65535 {
		c.Server.Port = 8443
	}
	if c.Server.ShutdownTimeout < 0 || c.Server.ShutdownTimeout > 20 {
		c.Server.ShutdownTimeout = 5
	}
	if c.Log.MaxLogFileSize < 1 || c.Log.MaxLogFileSize > 1000 {
		c.Log.MaxLogFileSize = 100
	}
	if c.Log.MaxLogFileBackup < 1 || c.Log.MaxLogFileBackup > 100 {
		c.Log.MaxLogFileBackup = 10
	}
	if c.Log.MaxLogFileAge < 1 || c.Log.MaxLogFileAge > 365 {
		c.Log.MaxLogFileAge = 90
	}

	return nil
}
