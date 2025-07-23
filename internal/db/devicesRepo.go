package db

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"go-rest-example/internal/logger"
	"go-rest-example/internal/model/data"
	"go-rest-example/internal/model/external"
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
	Create(ctx context.Context, di *data.Device) (string, error) 
	GetAll(ctx context.Context) (*[]data.Device, error)
	GetByID(ctx context.Context, ID string) (*data.Device, error)
	update(ctx context.Context, ID string, parmas *external.UpdateDeviceParams) error
	Delete(ctx context.Context, ID string) error
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

func (d *DevicesRepo) GetAll(ctx context.Context) (*[]data.Device, error){
	query := "SELECT InternalID, ProductNumber, MacAddress, FirmwareVersion, LastSeenAt, CreatedAt, ReTry, UpdateCheck, Status from devices"

	rows, err := d.connection.QueryContext(ctx, query)
	if err != nil {
		d.logger.Error().Err(err).Msg("failed to select device_info")
		return nil, ErrFailedToSelectReportInfo
	}

	defer rows.Close()

	var responseData []data.Device

	for rows.Next() {
		var device data.Device

		err := rows.Scan(
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
			d.logger.Error().Err(err).Msg("failed to scan row")
			return nil, err
		}
		
		responseData = append(responseData, device)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &responseData, nil
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

func (d *DevicesRepo) update(ctx context.Context, ID string, parmas *external.UpdateDeviceParams) error{

	query, args := d.GenerateUpdateQuery(parmas)
	if query == "" || args == nil {
		return errors.New("non Query")
	}

	// 5. 쿼리 실행
    result, err := d.connection.ExecContext(ctx, query, args...)
    if err != nil {
        d.logger.Error().Err(err).Msg("failed to update device")
        return ErrFailedToUpdateDevice
    }

	 // 6. 실제로 변경이 일어났는지 확인 (선택적)
	 rowsAffected, err := result.RowsAffected()
	 if err != nil {
		 return ErrFailedToUpdateDevice
	 }

	 if rowsAffected == 0 {
		 return ErrNothingAffrectedDevice
	 }

	return nil
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

// 조건문 생성 로직 분리
// 조건문 생성 기능만을 담당하는 함수 : 역할 분리 
func (d *DevicesRepo) GenerateUpdateQuery(parmas *external.UpdateDeviceParams) (string, []interface{}) {

	setClauses := []string{}
	args := []interface{}{}
	argId := 1

	// 2. 파라미터로 받은 값들을 확인하며 쿼리 조립


	// 4. 최종 쿼리문 생성
	query := fmt.Sprintf(
		"UPDATE devices SET %s WHERE ProductNumber = $%d",
		strings.Join(setClauses, ", "),
		argId,
	)

	if parmas.FirmwareVersion != nil {
		setClauses = append(setClauses, fmt.Sprintf("FirmwareVersion = $%d", argId))
		args = append(args, parmas.FirmwareVersion)
		argId++
	}

	if parmas.LastSeenAt != nil {
		setClauses = append(setClauses, fmt.Sprintf("LastSeenAt = $%d", argId))
        args = append(args, *parmas.LastSeenAt)
        argId++
	}

	if parmas.ReTry != nil {
		setClauses = append(setClauses, fmt.Sprintf("ReTry = $%d", argId))
        args = append(args, *parmas.ReTry)
        argId++
	}

	// 3. 변경할 내용이 없으면 아무것도 하지 않고 종료
	if len(setClauses) == 0 {
		return "", nil
	}

	return query, args
}