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

	// 서버 설정
	Server struct {
		// 서버 리슨 포트 (DEF:8443, MIN:1, MAX:65535)
		Port               int    `yaml:"port"`
		TLSCertificateFile string `yaml:"tlsCertificateFile"`
		TLSPrivateKeyFile  string `yaml:"tlsPrivateKeyFile"`
	} `yaml:"server"`
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
	Conf.Log.MaxLogFileSize = 100
	Conf.Log.MaxLogFileBackup = 10
	Conf.Log.MaxLogFileAge = 90
	Conf.Log.CompBakLogFile = true
	Conf.Server.Port = 8443
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
	if c.Log.MaxLogFileSize < 1 || c.Log.MaxLogFileSize > 1000 {
		c.Log.MaxLogFileSize = 100
	}
	if c.Log.MaxLogFileBackup < 1 || c.Log.MaxLogFileBackup > 100 {
		c.Log.MaxLogFileBackup = 10
	}
	if c.Log.MaxLogFileAge < 1 || c.Log.MaxLogFileAge > 365 {
		c.Log.MaxLogFileAge = 90
	}
	if c.Server.Port < 0 || c.Server.Port > 65535 {
		c.Server.Port = 8443
	}

	return nil
}
