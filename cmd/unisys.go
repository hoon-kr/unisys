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
Package cmd 명령 처리 패키지
*/
package cmd

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/hoon-kr/unisys/config"
	"github.com/hoon-kr/unisys/pkg/util/process"
	"github.com/spf13/cobra"
	"go.uber.org/automaxprocs/maxprocs"
)

// unisysCmd 최상위 루트 명령어
var unisysCmd = &cobra.Command{
	Use:     "unisys",
	Short:   "",
	Long:    "",
	Version: config.Version,
}

// init cmd 패키지 임포트 시 자동 초기화
func init() {
	unisysCmd.AddCommand(startCmd)
	unisysCmd.AddCommand(debugCmd)
	unisysCmd.AddCommand(stopCmd)
}

// Execute 명령어 실행
func Execute() {
	// GOMAXPROCS 값 최적화
	undo, err := maxprocs.Set()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[WARNING] failed to set GOMAXPROCS: %v\n", err)
	}
	defer undo()

	// 명령어 및 플래그 처리
	err = unisysCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// WrapCommandFuncForCobra cobra.Command의 RunE 필드 랩핑 함수
//
// Parameters:
//   - function: 명령어 함수
//
// Returns:
//   - error: 정상 종료(nil), 비정상 종료(error)
func WrapCommandFuncForCobra(function func(cmd *cobra.Command) error) func(cmd *cobra.Command, _ []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		cmd.SilenceErrors = true
		return function(cmd)
	}
}

// IsRunning 모듈이 동작 중인지 확인
//
// Parameters:
//   - pid: PID 저장 변수
//   - pidFilePath: PID 파일 경로
//
// Returns:
//   - bool: 동작(true), 미동작(false)
func IsRunning(pid *int, pidFilePath string) bool {
	if pid == nil {
		return false
	}

	file, err := os.Open(pidFilePath)
	if err != nil {
		return false
	}
	defer file.Close()

	// PID 값 읽기
	pidStr, err := io.ReadAll(file)
	if err != nil {
		return false
	}

	// PID 값을 정수로 변환
	*pid, err = strconv.Atoi(string(pidStr))
	if err != nil {
		return false
	}

	// 프로세스 동작 확인
	return process.IsProcessRun(*pid)
}

// SetupSignal 시그널 설정
//
// Returns:
//   - chan os.Signal: signal channel
func SetupSignal() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	// 수신할 시그널 설정 (SIGINT, SIGTERM, SIGUSR1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)
	// 무시할 시그널 설정
	signal.Ignore(syscall.SIGABRT, syscall.SIGALRM, syscall.SIGFPE, syscall.SIGHUP,
		syscall.SIGILL, syscall.SIGPROF, syscall.SIGQUIT, syscall.SIGTSTP,
		syscall.SIGVTALRM)

	return sigChan
}
