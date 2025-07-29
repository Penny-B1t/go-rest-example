package service

import (
	"time"

	"github.com/gin-gonic/gin"

	"go-rest-example/internal/db"
	"go-rest-example/internal/model/data"
	"go-rest-example/internal/model/external"
	"go-rest-example/internal/util"
)

// ReportService 인터페이스 (의존성 역전을 위해)
type IReportService interface {
	Report(c *gin.Context, reportReq external.ReportReq) ( *external.DeviceUpdate, error)
	Update(c *gin.Context, ID string) (string, error) 
}

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

func(d *ReportService) Report(c *gin.Context, reportReq external.ReportReq) (*external.DeviceUpdate, error){

	// 2. 디바이스 존재 여부 검증 : 선언 필요
	findDevice, err := d.uow.Device().GetByID(c, reportReq.ProductNumber)
	if err != nil {
		return nil, err
	}

	// 3. 정보 업데이트 객체 준비 
	report := data.DeviceInfo{	
		ReportID           : 1,
		ProductNumber      : findDevice.ProductNumber,
		BatteryPercent     : reportReq.BatteryPercent,
		Lat                : reportReq.Lat,
		Lon                : reportReq.Lon,
		TemperatureCelsius : reportReq.TemperatureCelsius,
		IP                 : reportReq.IP,
		ErrorCode          : reportReq.ErrorCode,
		ReportAt           : time.Now(),
		ReportedStatus     : reportReq. ReportedStatus,
	}

	// 4. repo 호출을 통한 업데이트 진행 
	_, err = d.uow.Report().Create(c, &report)
	if err != nil {
		return nil, err
	}

	// 5. 제어 로직 생성 // 재부팅이 3회 이상 반복된 경우 
	power := 0  
	if findDevice.ReTry >= 3 {
		power = 1
	}

	// 에러코드가 0이 아닌 경우  / device 필드에 count 증가가 
	reboot := 0 
	if power != 1 && reportReq.ErrorCode != 0 {
		reboot = 1
	}

	// 주기 보고 시간 할당 
	reportRes := external.DeviceUpdate{
		ReportCycleSec : 100,
		PowerOff       : power,
		Reboot         : reboot,
	} 

	// 5. 응답 진행
	return &reportRes, nil
}

// Select handles GET /report/update
func(d *ReportService) Update(c *gin.Context, ID string) (string, error) {

	findDevice, err := d.uow.Device().GetByID(c,ID)
	if err != nil {
		return "", err
	}

	// 1. UpdateCheck가 허용이면서 FirmwareVersion 버전이 최신이 아닌경우 
	if findDevice.FirmwareVersion != "" && findDevice.UpdateCheck != 0 {
		return "", err
	}

	// 2. 펌웨어 정보 획득
	path := "test"

	// 3. 유틸 메서드 경로 유효성 검사
	// 내부에서 파일 존재 여부도 검사 
	err = util.PathValid(path)
	if err != nil {
		return "",err
	}

	// 5. 체크썸 계산 (옵션)
	// 내부에서 파일을 읽어 체크썸 생성 - 추가 보안 필요시 사용할 것 
	// https://stackoverflow.com/questions/15879136/how-to-calculate-sha256-file-checksum-in-go
	// util.CheckSum(path)

	// 6. 체크썸 및 파일 정보 전송 ( 고려 )
	
	return path, nil
}