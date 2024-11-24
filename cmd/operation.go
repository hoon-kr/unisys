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

package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/meloncoffee/unisys/config"
	"github.com/meloncoffee/unisys/internal/logger"
	"github.com/meloncoffee/unisys/internal/server"
	"github.com/meloncoffee/unisys/pkg/util/file"
	"github.com/meloncoffee/unisys/pkg/util/goroutine"
	"github.com/meloncoffee/unisys/pkg/util/process"
	"github.com/spf13/cobra"
)

// unisysOperation 메인 동작 제어 구조체
type unisysOperation struct{}

// start unisys 모듈 가동
//
// Parameters:
//   - cmd: cobra 명령어 정보 구조체
//
// Returns:
//   - error: 정상 종료(nil), 비정상 종료(error)
func (u *unisysOperation) start(cmd *cobra.Command) error {
	// 작업 경로를 현재 프로세스가 위치한 경로로 변경
	err := file.ChangeWorkPathToModulePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
		return err
	}

	// 이미 동작 중인 프로세스가 존재하는지 확인
	var pid int
	if IsRunning(&pid, config.PidFilePath) {
		fmt.Fprintf(os.Stdout, "[INFO] there is already a process in operation (pid:%d)\n", pid)
		return nil
	}

	// 데몬 프로세스 생성
	err = process.DaemonizeProcess()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
		return err
	}

	// 현재 프로세스 PID 저장
	config.RunConf.Pid = os.Getpid()

	// 현재 프로세스 PID를 파일에 기록
	err = file.WriteDataToTextFile(config.PidFilePath, config.RunConf.Pid, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
		return err
	}

	// 디버그 모드 체크 (디버그 모드일 경우 stdout, stderr 출력)
	if cmd.Use == "debug" {
		config.RunConf.DebugMode = true
	} else {
		os.Stdout = nil
		os.Stderr = nil
	}

	// 시그널 설정
	sigChan := SetupSignal()
	defer signal.Stop(sigChan)

	// 고루틴 관리 구조체 생성
	gm := goroutine.NewGoroutineManager()
	// 패닉 핸들러 설정
	gm.PanicHandler = u.panicHandler

	// 초기화
	u.initialization(gm)
	// 종료 시 자원 정리
	defer u.finalization(gm)

	logger.Log.LogInfo("start %s (pid:%d, mode:%s)", config.ModuleName, config.RunConf.Pid,
		func() string {
			if config.RunConf.DebugMode {
				return "debug"
			}
			return "normal"
		}())

	// 작업에 등록된 모든 고루틴 가동
	gm.StartAll()

	// 종료 시그널 대기 (SIGINT, SIGTERM, SIGUSR1)
	sig := <-sigChan
	logger.Log.LogInfo("received %s (signum:%d)", sig.String(), sig)

	return nil
}

// stop unisys 모듈 정지
//
// Parameters:
//   - cmd: cobra 명령어 정보 구조체
//
// Returns:
//   - error: 정상 종료(nil), 비정상 종료(error)
func (u *unisysOperation) stop(cmd *cobra.Command) error {
	// 작업 경로를 현재 프로세스가 위치한 경로로 변경
	err := file.ChangeWorkPathToModulePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
		return err
	}

	// 이미 동작 중인 프로세스가 존재하는지 확인
	var pid int
	if !IsRunning(&pid, config.PidFilePath) {
		return nil
	}

	// 서버에 정지 시그널 전송 (SIGTERM)
	if err := process.SendSignal(pid, syscall.SIGTERM); err != nil {
		fmt.Fprintf(os.Stderr, "[WARNING] %v\n", err)
		return err
	}

	return nil
}

// initialization 초기화
//
// Parameters:
//   - gm: 고루틴 동작 관리 구조체
func (u *unisysOperation) initialization(gm *goroutine.GoroutineManager) {
	// 설정 파일 로드
	config.Conf.LoadConfig(config.ConfFilePath)
	// 로거 초기화
	logger.Log.InitializeLogger()

	// 메인 서버를 고루틴 작업에 등록
	var server server.Server
	gm.AddTask("server", server.Run)
}

// finalization 종료 시 자원 정리
//
// Parameters:
//   - gm: 고루틴 동작 관리 구조체
func (u *unisysOperation) finalization(gm *goroutine.GoroutineManager) {
	// 작업에 등록된 모든 고루틴 종료
	gm.StopAll(10 * time.Second)

	// 로그 자원 정리
	logger.Log.FinalizeLogger()
}

var unisysOper unisysOperation

// startCmd unisys 시작 명령
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Run unisys (normal mode)",
	RunE:  WrapCommandFuncForCobra(unisysOper.start),
}

// debugCmd unisys 시작 명령 (디버그)
var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Run unisys (debug mode)",
	RunE:  WrapCommandFuncForCobra(unisysOper.start),
}

// stopCmd unisys 정지 명령
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop unisys",
	RunE:  WrapCommandFuncForCobra(unisysOper.stop),
}

// panicHandler 패닉 핸들러
//
// Parameters:
//   - panicErr: 패닉 에러
func (u *unisysOperation) panicHandler(panicErr interface{}) {
	logger.Log.LogError("panic occurred: %v", panicErr)
	process.SendSignal(config.RunConf.Pid, syscall.SIGUSR1)
}
