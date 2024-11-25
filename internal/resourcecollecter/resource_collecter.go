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
Package resourcecollecter 리소스 수집 패키지
*/
package resourcecollecter

import (
	"context"
	"sync"
	"time"

	"github.com/meloncoffee/unisys/config"
	"github.com/meloncoffee/unisys/internal/logger"
	"github.com/meloncoffee/unisys/pkg/util/goroutine"
	"github.com/meloncoffee/unisys/pkg/util/resource"
)

// Resource 리소스 정보 구조체
type Resource struct {
	CPUUsageRate   float64
	MemUsageRate   float64
	DiskUsageRate  float64
	NetworkTraffic []resource.NetworkTraffic
}

var (
	// 전역 리소스 변수 선언
	GlobalResource Resource
	GlobalResMutex sync.RWMutex
)

// SetGlobalResource 전역 리소스 구조체 정보 업데이트
//
// Parameters:
//   - r: 리소스 정보 구조체
func SetGlobalResource(r *Resource) {
	GlobalResMutex.Lock()
	defer GlobalResMutex.Unlock()
	GlobalResource = Resource{
		CPUUsageRate:   r.CPUUsageRate,
		MemUsageRate:   r.MemUsageRate,
		DiskUsageRate:  r.DiskUsageRate,
		NetworkTraffic: append([]resource.NetworkTraffic{}, r.NetworkTraffic...),
	}
}

// GetGlobalResource 전역 리소스 구조체의 복사본 반환
//
// Returns:
//   - Resource: 복사된 리소스 정보 구조체
func GetGlobalResource() Resource {
	GlobalResMutex.RLock()
	defer GlobalResMutex.RUnlock()
	return Resource{
		CPUUsageRate:   GlobalResource.CPUUsageRate,
		MemUsageRate:   GlobalResource.MemUsageRate,
		DiskUsageRate:  GlobalResource.DiskUsageRate,
		NetworkTraffic: append([]resource.NetworkTraffic{}, GlobalResource.NetworkTraffic...),
	}
}

// ResourceCollecter 리소스 수집 구조체
type ResourceCollecter struct{}

// CollectResource 리소스 수집
//
// Parameters:
//   - ctx: 종료 컨텍스트
func (rc *ResourceCollecter) CollectResource(ctx context.Context) {
	var wg sync.WaitGroup
	var timeout time.Duration = 0

	for goroutine.WaitTimeout == goroutine.WaitCancelWithTimeout(ctx, timeout) {
		var res Resource

		// 3초 주기로 리소스 수집
		timeout = 3 * time.Second

		// CPU 사용률 획득
		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error
			res.CPUUsageRate, err = rc.getCPUUsageRate()
			if err != nil {
				logger.Log.LogWarn("failed to get CPU usage rate: %v", err)
			}
		}()

		// 메모리 사용률 획득
		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error
			res.MemUsageRate, err = rc.getMemUsageRate()
			if err != nil {
				logger.Log.LogWarn("failed to get memory usage rate: %v", err)
			}
		}()

		// 디스크 사용률 획득
		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error
			res.DiskUsageRate, err = rc.getDiskUsageRate()
			if err != nil {
				logger.Log.LogWarn("failed to get disk usage rate: %v", err)
			}
		}()

		// 네트워크 트래픽량 획득
		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error
			res.NetworkTraffic, err = rc.getNetworkTraffic()
			if err != nil {
				logger.Log.LogWarn("failed to get network traffic: %v", err)
			}
		}()

		// 고루틴 종료 대기
		wg.Wait()

		// 글로벌 리소스 구조체에 정보 업데이트
		SetGlobalResource(&res)

		if config.RunConf.DebugMode {
			logger.Log.LogDebug("CPU Usage Rate: %.2f%%", res.CPUUsageRate)
			logger.Log.LogDebug("Memory Usage Rate: %.2f%%", res.MemUsageRate)
			logger.Log.LogDebug("Disk Usage Rate: %.2f%%", res.DiskUsageRate)
			for _, traffic := range res.NetworkTraffic {
				logger.Log.LogDebug("Network Traffic - Interface: %s, InboundBps: %.2fbps, OutboundBps: %.2fbps",
					traffic.Interface, traffic.InboundBps, traffic.OutboundBps)
			}
		}
	}
}

// getCPUUsageRate CPU 사용률 획득
//
// Returns:
//   - float64: CPU 사용률
//   - error: 성공(nil), 실패(error)
func (rc *ResourceCollecter) getCPUUsageRate() (float64, error) {
	// 이전 CPU 상태 정보 획득
	prevStat, err := resource.GetCPUStat()
	if err != nil {
		return 0.0, err
	}

	// 1초 대기
	time.Sleep(1 * time.Second)

	// 현재 CPU 상태 정보 획득
	currStat, err := resource.GetCPUStat()
	if err != nil {
		return 0.0, err
	}

	// CPU 사용률 반환
	return resource.CalculateCPURate(prevStat, currStat), nil
}

// getMemUsageRate 메모리 사용률 획득
//
// Returns:
//   - float64: 메모리 사용률
//   - error: 성공(nil), 실패(error)
func (rc *ResourceCollecter) getMemUsageRate() (float64, error) {
	// 메모리 상태 정보 획득
	memStat, err := resource.GetMemStat()
	if err != nil {
		return 0.0, err
	}

	// 메모리 사용률 반환
	return resource.CalculateMemRate(memStat), nil
}

// getDiskUsageRate 디스크 사용률 획득
//
// Returns:
//   - float64: 디스크 사용률
//   - error: 성공(nil), 실패(error)
func (rc *ResourceCollecter) getDiskUsageRate() (float64, error) {
	// 디스크 상태 정보 획득
	diskStat, err := resource.GetDiskStat("/")
	if err != nil {
		return 0.0, err
	}

	// 디스크 사용률 반환
	return resource.CalculateDiskRate(diskStat), nil
}

// getNetworkTraffic 네트워크 트래픽량 획득
//
// Returns:
//   - float64: 디스크 사용률
//   - error: 성공(nil), 실패(error)
func (rc *ResourceCollecter) getNetworkTraffic() ([]resource.NetworkTraffic, error) {
	// 이전 네트워크 트래픽 획득
	prev, err := resource.GetAllNetworkTraffic()
	if err != nil {
		return nil, err
	}

	// 1초 대기
	intervalSec := 1.0
	time.Sleep(time.Duration(intervalSec) * time.Second)

	// 현재 네트워크 트래픽 획득
	current, err := resource.GetAllNetworkTraffic()
	if err != nil {
		return nil, err
	}

	// 네트워크 트래픽량 계산
	trafficList, err := resource.CalculateNetworkTraffic(prev, current, intervalSec)
	if err != nil {
		return nil, err
	}

	return trafficList, nil
}
