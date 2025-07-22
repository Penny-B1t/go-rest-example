package db

import (
	"context"
	"errors"
	"strconv"

	"go-rest-example/internal/logger"
	"go-rest-example/internal/model/data"
)

var (
	ErrInvalidDeviceRequired          = errors.New("missing required inputs to create DeviceRepo")
	ErrNothingAffrectedDevice         = errors.New("nothing affected to DeviceRepo")
	ErrFailedToCreateDevice 		  = errors.New("failed to create device")
	ErrFailedToSelectDevice 		  = errors.New("failed to select device")
	ErrFailedToUpdateDevice 		  = errors.New("failed to update device")
	ErrFailedToDeleteDevice 	      = errors.New("failed to delete device")
)

// DeviceRepo를 통해 사용할 메서드를 제약하고 규정하기 위한 인터페이스 
// 입력 타입 및 반환 타입 수정 필요 
type DevicesDataService interface {
	Create(ctx context.Context, di *data.Device) (string, error) // id 식별자를 반환한다
	GetAll(ctx context.Context) (*[]data.Device, error)
	GetByID(ctx context.Context, ID string) (*data.Device, error)
	Update(ctx context.Context, ID string) (*data.Device, error)
	Delete(ctx context.Context, ID string) (string, error)
}

// Device 테이블을 접근하기 위한 커넥션 관리
type DevicesRepo struct {
	connection DBTX
	logger     *logger.AppLogger
}

func NewDevicesRepo(lgr *logger.AppLogger, db DBTX) (*DevicesRepo, error) {
	if lgr == nil || db == nil {
		return nil, ErrInvalidReportRequired
	}
	return &DevicesRepo{
		connection: db,
		logger:     lgr,
	}, nil
}

func (d *DevicesRepo) Create(ctx context.Context, di *data.Device)(string, error){
	// 쿼리문 생성
	query := "INSERT INTO devices " +
	"( ProductNumber, MacAddress, FirmwareVersion, LastSeenAt, CreatedAt, ReTry, UpdateCheck, Status)" +
	"VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?)"

	// 쿼리문 실행
	result, err := d.connection.ExecContext(
		ctx, 
		query, 
		di.ProductNumber,
		di.MacAddress,
		di.FirmwareVersion,
		di.LastSeenAt,
		di.CreatedAt,
		0,
		0,
		data.ReportPowerOn,
	)

	if err != nil {
		d.logger.Error().Err(err).Msg("failed to create devices")
		return "", ErrFailedToCreateDevice
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return "", ErrFailedToCreateDevice
	}

	return strconv.FormatInt(lastID, 10), nil
}

func (d *DevicesRepo) GetByID(ctx context.Context, productNumber string) (*data.Device, error){
	//
	query := "SELECT InternalID, ProductNumber, MacAddress, FirmwareVersion, LastSeenAt, CreatedAt, ReTry, UpdateCheck, Status from devices WHERE ProductNumber = ?"

	row := d.connection.QueryRowContext(ctx, query, productNumber)

	var device data.Device
	err := row.Scan(
		&device.InternalID,
		&device.ProductNumber,
		&device.MacAddress,
		&device.FirmwareVersion,
		&device.LastSeenAt,
		&device.CreatedAt,
		&device.ReTry,
		&device.UpdateCheck,
		&device.Status,
	)

	 if err != nil {
		return nil, ErrFailedToSelectDevice
	 }

	 return &device, nil
}

// 업데이트 로직 파라마티에 따른 변화 고민 필요 
// TODO ! 
func (d *DevicesRepo) Update(ctx context.Context, productNumber string) (*data.Device, error){
	return nil, nil
}

func (d *DevicesRepo) Delete(ctx context.Context, productNumber string)  error {
	query := "DELETE FROM devices WHERE ProductNumber = ? "

	result, err := d.connection.ExecContext(ctx, query, productNumber)
	if err != nil {
		d.logger.Error().Err(err).Msg("failed to delete devices")
		return ErrFailedToDeleteDevice
	}

	_, err = result.RowsAffected()
	if err != nil {
		d.logger.Error().Err(err).Msg("nothing affected")
		return ErrNothingAffrectedDevice
	}

	return nil
}