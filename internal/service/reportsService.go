package service

import (
	"context"
	"errors"

	"go-rest-example/internal/db"
	"go-rest-example/internal/model/data"
	"go-rest-example/internal/model/external"
	"go-rest-example/internal/util"
)

// ReportService 인터페이스 (의존성 역전을 위해)
type IReportService interface {
	DeviceReport(parentCtx context.Context, reportReq external.ReportReq) ( *external.DeviceUpdate, error)
	CheckForUpdate(parentCtx context.Context, ID string) (string, error) 
}

//
var (
	ErrDeviceNotFound    = errors.New("device not found")
	ErrFirmwareNotAvailable = errors.New("firmware not available or update not allowed")
)

// 실제 구현체
type ReportService struct {
	uow db.IUnitOfWork
}

// 팩토리 함수
func NewUserService(uow db.IUnitOfWork) IReportService {
	return &ReportService{
		uow: uow,
	}
}

// DeviceReport: 디바이스의 보고를 처리하고, 제어 명령을 반환합니다.
func(r *ReportService) DeviceReport(parentCtx context.Context, reportReq external.ReportReq) (*external.DeviceUpdate, error){
	// 1. 서비스 타임아웃을 포함한 새로운 Context 생성
	ctx, cancel := context.WithTimeout(parentCtx, util.DefaultServiceTimeout)
	defer cancel()

	// 1-1. 포인트로 생성하여 클로저 환경 주소를 공유하여 작업 
	var deviceUpdateResponse *external.DeviceUpdate

	// 2. 여러 DB 작업을 하나의 트랜잭션으로 묶기 위해 Unit of Work 사용
	err := r.uow.Execute(ctx, func(work db.IUnitOfWork) error {
		// 2-1. 리포지토리 획득 (이 리포지토리들은 모두 동일한 트랜잭션을 공유합니다)
		deviceRepo := work.Device()
		reportRepo := work.Report()

		// 2-2. 디바이스 존재 여부 검증
		device, err := deviceRepo.GetByProductNumber(ctx, reportReq.ProductNumber)
		if err != nil {
			if errors.Is(err, db.ErrFailedToSelectDevice) { // 리포지토리의 에러를 더 구체적인 서비스 에러로 변환
				return ErrDeviceNotFound
			}
			return err
		}

		// 2-3. 보고 정보(report) 레코드 생성
		report := data.DeviceInfo{
			ProductNumber:      device.ProductNumber,
			BatteryPercent:     reportReq.BatteryPercent,
			Lat:                reportReq.Lat,
			Lon:                reportReq.Lon,
			TemperatureCelsius: reportReq.TemperatureCelsius,
			IP:                 reportReq.IP,
			ErrorCode:          reportReq.ErrorCode,
			ReportedStatus:     reportReq.ReportedStatus,
		}
		if _, err := reportRepo.Create(ctx, &report); err != nil {
			return err
		}
		
		// 2-4. 디바이스 상태 업데이트 (필요 시)
		// 예: 마지막 접속 시간(LastSeenAt) 업데이트
		updateParams := &external.UpdateDeviceParams{
			LastSeenAt: &report.ReportAt,
		}
		// 에러 코드가 0이 아닌 경우, 재시도 횟수(ReTry) 증가
		if reportReq.ErrorCode != 0 {
			newRetryCount := device.ReTry + 1
			updateParams.ReTry = &newRetryCount
			device.ReTry = newRetryCount // 아래 제어 로직에서 사용하기 위해 로컬 변수도 업데이트
		}
		if err := deviceRepo.Update(ctx, device.ProductNumber, updateParams); err != nil {
			return err
		}

		// 3. 비즈니스 로직: 제어 명령 생성
		powerOff := 0
		if device.ReTry >= 3 { // 재부팅이 3회 이상 반복된 경우
			powerOff = 1
		}

		reboot := 0
		if powerOff != 1 && reportReq.ErrorCode != 0 { // 전원 차단 명령이 없고, 에러가 보고된 경우
			reboot = 1
		}

		// 4. 최종 응답 객체 생성
		deviceUpdateResponse = &external.DeviceUpdate{
			ReportCycleSec: 100, // 이 값은 설정(config)에서 가져오는 것이 좋습니다.
			PowerOff:       powerOff,
			Reboot:         reboot,
		}

		return nil // 에러가 없으면 uow.Execute가 트랜잭션을 커밋합니다.
	})

	return deviceUpdateResponse, err
}

// CheckForUpdate: 디바이스의 펌웨어 업데이트를 확인합니다.
func(d *ReportService) CheckForUpdate(parentCtx context.Context, ID string) (string, error) {
	// 1. 서비스 타임아웃 Context 생성
	ctx, cancel := context.WithTimeout(parentCtx, util.DefaultServiceTimeout)
	defer cancel()

	findDevice, err := d.uow.Device().GetByProductNumber(ctx,ID)
	if err != nil {
		return "", err
	}

	// 2. 업데이트 조건 확인
	// TODO: 최신 펌웨어 버전을 외부 소스(DB, 설정 파일 등)에서 가져오는 로직 필요
	const latestFirmwareVersion = "v1.2.0"
	if findDevice.UpdateCheck == 0 || findDevice.FirmwareVersion == latestFirmwareVersion {
		// 업데이트가 허용되지 않았거나, 이미 최신 버전인 경우
		return "", ErrFirmwareNotAvailable
	}

	// 3. 펌웨어 파일 경로 확인
	// TODO: 실제 펌웨어 파일 경로를 반환하는 로직 필요
	firmwarePath := "firmware/latest.bin"
	if err := util.PathValid(firmwarePath); err != nil {
		return "", err
	}

	
	// TODO: 필요 시 체크섬 계산 및 반환 로직 추가
	// https://stackoverflow.com/questions/15879136/how-to-calculate-sha256-file-checksum-in-go
	// util.CheckSum(path)

	return firmwarePath, nil
}