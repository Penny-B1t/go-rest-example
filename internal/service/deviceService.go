package service

import (
	"go-rest-example/internal/db"
	"go-rest-example/internal/model/data"
	"go-rest-example/internal/model/external"
	"time"

	"github.com/gin-gonic/gin"
)

// DeviceService 인터페이스 (의존성 역전을 위해)
type IDeviceService interface {
	Create(c *gin.Context, deviceReq external.DeviceReq) error 
	GetAll(c *gin.Context)(*[]data.Device, error)
	GetByID(c *gin.Context, ID string)(*data.Device, error)
}

// 실제 구현체
type DeviceService struct {
	dsRepo db.DevicesDataService
}

// 팩토리 함수
func NewDeviceService(dsRepo db.DevicesDataService) IDeviceService {
	return &DeviceService{
		dsRepo: dsRepo,
	}
}

// Create handles POST /device.
func(d *DeviceService) Create(c *gin.Context, deviceReq external.DeviceReq) error {

	// 2. DB 중복 객체 존재 여부 확인
	findDevice, err := d.dsRepo.GetByID(c, deviceReq.ProductNumber)
	if err != nil || findDevice != nil {
		// 상위 레이어에게 전달
		return err
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
		// 상위 레이어에게 전달
		return err
	}

	// TODO 4. 실제로 생성되었는지 검사

	return nil
}

// Select handles GET /device.
func(d *DeviceService) GetAll(c *gin.Context)(*[]data.Device, error){
	// 0. 데이터 레이어를 통한 정보 획득 
	devices, err := d.dsRepo.GetAll(c)
	if err != nil {
		return nil, err
	}

	// 1. 정보 반환 
	return devices, nil
}

// Select handles GET /device/ID=.
func(d *DeviceService) GetByID(c *gin.Context, ID string)(*data.Device, error){

	// 1. 데이터 레이어를 통한 정보 획득 
	findDevice, err := d.dsRepo.GetByID(c, ID)
	if err != nil {
		// 커스텀 에러 선언 필요 
		return nil, err
	}

	// 2. 정보 반환
	return findDevice, nil
}


