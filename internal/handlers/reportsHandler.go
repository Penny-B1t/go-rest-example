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
)

type ReportsHandler struct {
	dvRepo db.ReportsDataService
	logger *logger.AppLogger
}

func NewDeviceHandler(lgr *logger.AppLogger, dvRepo db.ReportsDataService)(*ReportsHandler, error){
	if lgr == nil || dvRepo == nil {
		return nil, errors2.New("missing required parameters to create orders handler")
	}

	return &ReportsHandler{dvRepo: dvRepo, logger: lgr}, nil
}

// 디바이스 신규 등록을 담당하는 API
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
	// findDevice, err := .findByID()
	// if err != nil {
	// 	return 
	// }

	// 2. 디바이스 식별자 할당 
	PN := findDevice.ProductNumber

	// 3. 정보 업데이트 객체 준비 
	DI := data.DeviceInfo{	
		ReportID           : 1,
		ProductNumber      : PN,
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
	_, err := d.dvRepo.Create(c, &DI)
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

// 디바이스 상세 정보를 획득한다.
// Create handles GET /update
func(d *ReportsHandler) Update(c *gin.Context){
	// 상위 미들웨어에서 인증 처리 

	// 0. 
}

// 디바이스 상세 정보를 획득한다.
// Create handles GET /setting
func(d *ReportsHandler) Setting(c *gin.Context){
	
}

// 에러 발생 시 응답 생성 역할 수행 
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