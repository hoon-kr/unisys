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
Package file 파일 유틸 패키지
*/
package file

import (
	"fmt"
	"os"
	"path/filepath"
)

// ChangeWorkPathToModulePath 실행 파일 경로로 작업 경로 변경
//
// Returns:
//   - error: 성공(nil), 실패(error)
func ChangeWorkPathToModulePath() error {
	// 현재 프로세스의 절대 경로 획득
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to absolute path: %v", err)
	}

	// 절대 경로에서 디렉터리만 추출
	dirPath := filepath.Dir(exePath)

	// 작업 경로 변경
	err = os.Chdir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to change dir: %v", err)
	}

	return nil
}

// WriteDataToTextFile 제네릭한 파일 쓰기 함수
//
// Parameters:
//   - filePath: 파일 경로
//   - data: 제네릭 타입 데이터
//   - isMakeDir: 디렉터리가 존재하지 않을 경우 생성 옵션
//
// Returns:
//   - error: 성공(nil), 실패(error)
func WriteDataToTextFile[T any](filePath string, data T, isMakeDir bool) error {
	if isMakeDir {
		// 디렉터리가 존재하지 않을 경우 생성
		dir := filepath.Dir(filePath)
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to make directory: %v", err)
		}
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "%v", data)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}
