package service

import (
	"context"
	"database/sql"
	"errors"
	"go-rest-example/internal/db"
	"go-rest-example/internal/model/data"
	"go-rest-example/internal/model/external"
	"go-rest-example/internal/util"
	"time"

	error2 "go-rest-example/internal/error"
)

// DeviceService 인터페이스 (의존성 역전을 위해)
type IDeviceService interface {
	Create(parentCtx context.Context, deviceReq external.DeviceReq) error 
	GetAll(parentCtx context.Context)(*[]data.Device, error)
	GetByProductNumber(parentCtx context.Context, ProductNumber string)(*data.Device, error)
}

// 실제 구현체
type DeviceService struct {
	// UoW 객체 할당 
	uow db.IUnitOfWork
}

// 팩토리 함수
func NewDeviceService(uow db.IUnitOfWork) IDeviceService {
	return &DeviceService{
		uow: uow,
	}
}

// Create handles POST /device.
func (d *DeviceService) Create(parentCtx context.Context, deviceReq external.DeviceReq) error {

	ctx, cancel := context.WithTimeout(parentCtx, util.DefaultServiceTimeout)
	defer cancel()

	return d.uow.Execute(ctx, func(work db.IUnitOfWork) error {

		// 1. 객체 생성을 위한 도메인 엔티티 생성
		newDevice := data.Device{
			InternalID 	  : 1, 
			ProductNumber : deviceReq.ProductNumber,
			MacAddress    : deviceReq.MacAddress,
			FirmwareVersion : deviceReq.FirmwareVersion,  
			LastSeenAt    : time.Now(),
			CreatedAt     : time.Now(),
			ReTry         : 0,
			UpdateCheck   : 0,
			Status        : data.ReportPowerOn,
		}

		_, err := work.Device().Create(ctx, &newDevice)
		if err != nil {
			return error2.NewInternalServerError(err)
		}
		
		// 2. 최초 보고 로직 생성 
		newReport := data.DeviceInfo{
				ReportID          : 1,
				ProductNumber     : deviceReq.ProductNumber,
				BatteryPercent    : 0,
				Lat               : 0,
				Lon               : 0,
				TemperatureCelsius: 0,
				IP                : "255.255.255.255",
				ErrorCode         : 0,
				ReportAt          : time.Now(),
				ReportedStatus    : data.ReportPowerOn,
		}

		_, err = work.Report().Create(ctx, &newReport)
		if err != nil {
			return error2.NewInternalServerError(err)
		}

		return nil
	})
}

// Select handles GET /device.
func(d *DeviceService) GetAll(parentCtx context.Context)(*[]data.Device, error){


	ctx, cancel := context.WithTimeout(parentCtx, util.DefaultServiceTimeout)
	defer cancel()

	// 0. 데이터 레이어를 통한 정보 획득 
	devices, err := d.uow.Device().GetAll(ctx)

	// 1. 에러 처리 구간
	if err != nil {
		if errors.Is(err, sql.ErrNoRows){
			return nil, error2.NewNotFoundError("Device Get All request to faile", err)
		}
		return nil, error2.NewInternalServerError(err)
	}

	// 2. 정보 반환 
	return devices, nil
}

// Select handles GET /device/ID=.
func(d *DeviceService) GetByProductNumber(parentCtx context.Context, ProductNumber string)(*data.Device, error){

	ctx, cancel := context.WithTimeout(parentCtx, util.DefaultServiceTimeout)
	defer cancel()

	// 1. 데이터 레이어를 통한 정보 획득 
	findDevice, err := d.uow.Device().GetByProductNumber(ctx, ProductNumber)

	// 2. 에러 처리 구간 
	if err != nil {
		if errors.Is(err, sql.ErrNoRows){
			return nil, error2.NewNotFoundError("Device Get request to faile", err)
		}
		return nil, error2.NewInternalServerError(err)
	}

	// 2. 정보 반환
	return findDevice, nil
}


