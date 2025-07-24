package handlers

import (
	errors2 "errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"go-rest-example/internal/db"
	"go-rest-example/internal/logger"
	"go-rest-example/internal/model/data"
	"go-rest-example/internal/model/external"
)

type DevicesHandler struct {
	dsRepo db.DevicesDataService
	logger *logger.AppLogger
}

func NewDevicesHandler(lgr *logger.AppLogger,dsRepo db.DevicesDataService )(*DevicesHandler, error){
	if lgr == nil || dsRepo == nil {
		return nil, errors2.New("missing required parameters to create orders handler")
	}

	return &DevicesHandler{dsRepo: dsRepo, logger: lgr}, nil
}


// Create handles POST /device.
func(d *ReportsHandler) Create(c *gin.Context){
	// lgr, requestID := d.logger.WithReqID(c)
	var deviceReq external.DeviceReq

	// 0. BODY -> JSON 직렬화
	err := c.ShouldBindBodyWithJSON(&deviceReq) 
	if err != nil {
		// 커스텀 에러 선언 필요 
		return 
	}

	// 1. 객체 유효성 검사 
	err = deviceReq.Validate()
	if err != nil {
		// 커스텀 에러 선언 필요 
		return 
	}

	// 2. DB 중복 객체 존재 여부 확인
	findDevice, err := d.dsRepo.GetByID(c, deviceReq.ProductNumber)
	if err != nil || findDevice != nil {
		// 커스텀 에러 선언 필요 
		return 
	}

	// 3. 객체 생성을 위한 도메인 엔티티 생성
	newDevice := data.Device{
		InternalID 	  : 1, 
		ProductNumber : deviceReq.ProductNumber,
		MacAddress    : deviceReq.MacAddress,
		FirmwareVersion : deviceReq.FirmwareVersion,  
		LastSeenAt    : time.Now(),
		CreatedAt     : time.Now(),
		ReTry         : 0,
		UpdateCheck   : 0,
		Status        : data.StatusReady,
	}

	_, err = d.dsRepo.Create(c, &newDevice)
	if err != nil {
		// 커스텀 에러 선언 필요 
		return 
	}

	c.String(http.StatusCreated, "update is ok" )
}

// Select handles GET /device.
func(d *ReportsHandler) GetAll(c *gin.Context){
	// 0. 데이터 레이어를 통한 정보 획득 

	// 1. 정보 반환 
	c.JSON(http.StatusCreated, nil)
}

// Select handles GET /device/ID=.
func(d *ReportsHandler) GetByID(c *gin.Context){
	// 0. 쿼리 파라미터 획득 
	i := c.Query("ID") 

	// 1. 데이터 레이어를 통한 정보 획득 
	findDevice, err := d.dsRepo.GetByID(c, i)
	if err != nil {
		// 커스텀 에러 선언 필요 
		return
	}

	// 2. 정보 반환
	c.JSON(http.StatusCreated, findDevice)
}