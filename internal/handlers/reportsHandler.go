package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"go-rest-example/internal/logger"
	"go-rest-example/internal/model/external"
	"go-rest-example/internal/service"
)

type ReportsHandler struct {
	rpService service.IReportService
	logger *logger.AppLogger
}

// 오류 코드와 메서드 타입 사용하여 동작의 의미를 명확히 할 것 
func NewReportsHandler(lgr *logger.AppLogger, rpService service.IReportService) (*ReportsHandler, error) {

	if lgr == nil || rpService == nil {
		return nil, errors.New("Handler require null")
	}

	return &ReportsHandler{
		rpService : rpService,
		logger: lgr,
	}, nil
}

// Create handles GET /report/.
func(d *ReportsHandler) Report(c *gin.Context){
	lgr, requestID := d.logger.WithReqID(c)
	var reportReq external.ReportReq

	// 0. BODY -> JSON 직렬화 
	if err := c.ShouldBindBodyWithJSON(&reportReq); err != nil {
		d.abortWithAPIError(c, lgr, http.StatusBadRequest, "Invalid report request body", requestID, err)
		return
	}

	// 1. 객체 유효성 검사 
	err := reportReq.Validate()
	if err != nil {
		d.abortWithAPIError(c, lgr, http.StatusBadRequest, "Invalid report request body", requestID, err)
		return 
	}

	reportRes, err := d.rpService.Report(c, reportReq)
	if err != nil {
		//  커스텀 에러 선언 
		return
	}

	// 5. 응답 진행
	c.JSON(http.StatusCreated, reportRes)
}

// Select handles GET /report/update
func(d *ReportsHandler) Update(c *gin.Context){
	// lgr, requestID := d.logger.WithReqID(c)

	// 0. 쿼리 파라미터 획득 
	i := c.Query("ProductNumber")

	path, err := d.rpService.Update(c, i)
	if err != nil {
		return 
	}

	c.File(path)
}

// 에러 발생 시 응답 생성 역할 수행 
// 공용 에러 핸들러로 전환 필요 
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