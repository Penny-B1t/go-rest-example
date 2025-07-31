package handlers

import (
	// errors2 "errors"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"go-rest-example/internal/logger"
	"go-rest-example/internal/model/external"
	"go-rest-example/internal/service"

	error2 "go-rest-example/internal/error"
)

type DevicesHandler struct {
	dService service.IDeviceService
	logger *logger.AppLogger
}

func NewDevicesHandler(lgr *logger.AppLogger,dService service.IDeviceService )(*DevicesHandler, error){

	if lgr == nil || dService == nil {
		return nil, errors.New("Handler require null")
	}

	return &DevicesHandler{dService: dService, logger: lgr}, nil
}


// Create handles POST /device.
func(d *DevicesHandler) Create(c *gin.Context){
	// lgr, requestID := d.logger.WithReqID(c)
	var deviceReq external.DeviceReq

	// 0. BODY -> JSON 직렬화
	err := c.ShouldBindBodyWithJSON(&deviceReq) 
	if err != nil {
		c.Error(error2.NewBadValidateError(err))
		return 
	}

	// 1. 객체 유효성 검사 
	err = deviceReq.Validate()
	if err != nil{
		c.Error(error2.NewBadValidateError(err))
		return 
	}

	err = d.dService.Create(c, deviceReq)
	if err != nil{
		c.Error(err)
		return 
	}

	c.String(http.StatusCreated, "update is ok" )
}

// Select handles GET /device.
func(d *DevicesHandler) GetAll(c *gin.Context){
	// 0. 데이터 레이어를 통한 정보 획득 
	devices, err := d.dService.GetAll(c)
	if err != nil{
		_ = c.Error(err)
		return 
	}

	// 1. 정보 반환 
	c.JSON(http.StatusCreated, devices)
}

// Select handles GET /device/ID=.
func(d *DevicesHandler) GetByID(c *gin.Context){
	// 0. 쿼리 파라미터 획득 
	i := c.Param("ID") 

	// 1. 데이터 레이어를 통한 정보 획득 
	findDevice, err := d.dService.GetByProductNumber(c,i)
	if err != nil{
		_ = c.Error(err)
		return 
	}

	// 2. 정보 반환
	c.JSON(http.StatusCreated, findDevice)
}