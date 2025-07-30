package service

import (
	"context"
	"go-rest-example/internal/db"
	"go-rest-example/internal/model/data"
	"go-rest-example/internal/model/external"
	"go-rest-example/internal/util"
	"time"
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

		// 2. DB 중복 객체 존재 여부 확인
		findDevice, err := work.Device().GetByProductNumber(ctx, deviceReq.ProductNumber)
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

		_, err = work.Device().Create(ctx, &newDevice)
		if err != nil {
			// 상위 레이어에게 전달
			return err
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
	if err != nil {
		return nil, err
	}

	// 1. 정보 반환 
	return devices, nil
}

// Select handles GET /device/ID=.
func(d *DeviceService) GetByProductNumber(parentCtx context.Context, ProductNumber string)(*data.Device, error){

	ctx, cancel := context.WithTimeout(parentCtx, util.DefaultServiceTimeout)
	defer cancel()

	// 1. 데이터 레이어를 통한 정보 획득 
	findDevice, err := d.uow.Device().GetByProductNumber(ctx, ProductNumber)
	if err != nil {
		// 커스텀 에러 선언 필요 
		return nil, err
	}

	// 2. 정보 반환
	return findDevice, nil
}


