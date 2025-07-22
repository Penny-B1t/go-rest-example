package handlers

import (
	errors2 "errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"go-rest-example/internal/db"
	"go-rest-example/internal/logger"
	"go-rest-example/internal/model/data"
	"go-rest-example/internal/model/external"
	"go-rest-example/internal/util"
)

type ReportsHandler struct {
	rsRepo db.ReportsDataService
	dsRepo db.DevicesDataService
	logger *logger.AppLogger
}

func NewReportsHandler(lgr *logger.AppLogger, rsRepo db.ReportsDataService)(*ReportsHandler, error){
	if lgr == nil || rsRepo == nil {
		return nil, errors2.New("missing required parameters to create orders handler")
	}

	return &ReportsHandler{rsRepo: rsRepo, logger: lgr}, nil
}

// Create handles POST /report.
// TODO : device 레페지토리 findDevice 선언 필요 
func(d *ReportsHandler) Report(c *gin.Context){
	lgr, requestID := d.logger.WithReqID(c)
	var reportReq external.ReportReq

	// 0. BODY -> JSON 직렬화 
	if err := c.ShouldBindBodyWithJSON(&reportReq); err != nil {
		d.abortWithAPIError(c, lgr, http.StatusBadRequest, "Invalid report request body", requestID, err)
		return
	}

	// 1. 디바이스 존재 여부 검증 : 선언 필요
	findDevice, err := d.dsRepo.GetByID(c, reportReq.ProductNumber)
	if err != nil {
		// 커스텀 에러 반환 abortWithAPIError
		return 
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
	_, err = d.rsRepo.Create(c, &report)
	if err != nil {
		d.abortWithAPIError(c, lgr, http.StatusBadRequest, "faild Create report row", requestID, err)
		return 
	}
	// 5. 제어 로직 생성
	power := 0  // 재부팅이 3회 이상 반복된 경우 
	if findDevice.ReTry >= 3 {
		power = 1
	}
	// device 필드에 count 확인 
	reboot := 0 // 에러코드가 0이 아닌 경우  / device 필드에 count 증가가 
	if power != 1 && reportReq.ErrorCode != 0 {
		reboot = 1
	}

	// 주기 보고 시간 할당 
	reportRes := data.DeviceUpdate{
		ReportCycleSec : 100,
		PowerOff       : power,
		Reboot         : reboot,
	} 

	// 5. 응답 진행
	c.JSON(http.StatusCreated, reportRes)
}

// Create handles GET /update
func(d *ReportsHandler) Update(c *gin.Context){
	lgr, requestID := d.logger.WithReqID(c)

	// 0. 쿼리 파라미터 획득 
	i := c.Query("ProductNumber")

	findDevice, err := d.dsRepo.GetByID(c,i)
	if err != nil {
		// 커스텀 에러 선언 필요 
		return
	}

	// 1. UpdateCheck가 허용이면서 FirmwareVersion 버전이 최신이 아닌경우 
	if findDevice.FirmwareVersion != "" && findDevice.UpdateCheck != 0 {
		// 커스텀 에러 선언 필요 
		return 
	}

	// 2. 펌웨어 정보 획득
	// TODO 경로 정보 획득 구성 필요 
	path := "test"

	// 3. 유틸 메서드 경로 유효성 검사
	// 내부에서 파일 존재 여부도 검사 
	err = util.PathValid(path)
	if err != nil {
		// 커스텀 에러 선언 필요 
		return 
	}

	// 5. 체크썸 계산 (옵션)
	// 내부에서 파일을 읽어 체크썸 생성 - 추가 보안 필요시 사용할 것 
	// https://stackoverflow.com/questions/15879136/how-to-calculate-sha256-file-checksum-in-go
	// util.CheckSum(path)

	// 6. 체크썸 및 파일 정보 전송 
	// golang gin file fileattach 차이 인지 필요 스트리밍으로 구현할 것인가? 
	c.File(path)
}

// 에러 발생 시 응답 생성 역할 수행 
// 
func(d *ReportsHandler) abortWithAPIError(
	c *gin.Context,
	lgr zerolog.Logger,
	status int,
	message, debugID string,
	err error,
){
	apiErr := &external.APIError{
		HTTPStatusCode : status,
		Message : message,
		DebugID: debugID,
	}

	event := lgr.Error().Int("HttpStatusCode",status)
	if err != nil {
		event.Err(err)
	}

	event.Msg(message)
	c.AbortWithStatusJSON(status, apiErr)
}