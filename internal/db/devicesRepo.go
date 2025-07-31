package db

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go-rest-example/internal/logger"
	"go-rest-example/internal/model/data"
	"go-rest-example/internal/model/external"
)

// DeviceRepo를 통해 사용할 메서드를 제약하고 규정하기 위한 인터페이스
type DevicesDataService interface {
	Create(ctx context.Context, di *data.Device) (int64, error) 
	GetAll(ctx context.Context) (*[]data.Device, error)
	GetByProductNumber(ctx context.Context, ID string) (*data.Device, error)
	Update(ctx context.Context, ID string, parmas *external.UpdateDeviceParams) error
	Delete(ctx context.Context, ID string) error
}

// Device 테이블을 접근하기 위한 커넥션 관리
type DevicesRepo struct {
	connection DBTX
	logger     *logger.AppLogger
}

func NewDevicesRepo(lgr *logger.AppLogger, db DBTX) *DevicesRepo {

	return &DevicesRepo{
		connection: db,
		logger:     lgr,
	}
}

func (d *DevicesRepo) Create(ctx context.Context, device *data.Device)(int64, error){

	d.logger.Info().Msg("call create")
	
	query := `
        INSERT INTO devices (ProductNumber, MacAddress, FirmwareVersion, LastSeenAt, CreatedAt, ReTry, UpdateCheck, Status)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

    // ReTry와 UpdateCheck에 0을 명시적으로 전달합니다.
    result, err := d.connection.ExecContext(ctx, query,
        device.ProductNumber,
        device.MacAddress,
        device.FirmwareVersion,
        device.LastSeenAt,
        device.CreatedAt,
        0, 
        0,
        device.Status,
    )

	if err != nil {
		d.logger.Error().Err(err).Msg("[deviceRepo] failed to create device with sqlx")
		return 0, err
	}

	lastID, err := result.LastInsertId()

	if err != nil {
		d.logger.Error().Err(err).Msg("[deviceRepo] non create device with sqlx")
		return 0, err
	}

	return lastID, nil
}

func (d *DevicesRepo) GetAll(ctx context.Context) (*[]data.Device, error){

	query := `
	SELECT InternalID, ProductNumber, MacAddress, FirmwareVersion, LastSeenAt, CreatedAt, ReTry, UpdateCheck, Status
	FROM devices`

	var devices []data.Device

	err := d.connection.SelectContext(ctx, &devices, query)

	if err != nil {
		d.logger.Error().Err(err).Msg("[deviceRepo] non select device with sqlx")
		return nil, err
	}

	return &devices, nil
}

func (d *DevicesRepo) GetByProductNumber(ctx context.Context, productNumber string) (*data.Device, error){
	var device data.Device

	query := `
		SELECT InternalID, ProductNumber, MacAddress, FirmwareVersion, LastSeenAt, CreatedAt, ReTry, UpdateCheck, Status
		FROM devices WHERE ProductNumber = ?`

	err := d.connection.GetContext(ctx, &device, query, productNumber)

	if err != nil {
		d.logger.Error().Err(err).Msg("[deviceRepo] failed to select device with sqlx")
		return nil, err
	}

	 return &device, nil
}

func (d *DevicesRepo) Update(ctx context.Context, productNumber string, parmas *external.UpdateDeviceParams) error{

	query, args := d.GenerateUpdateQuery(parmas)
	if query == "" || args == nil {
		return errors.New("[deviceRepo] non Query")
	}

	// 식별자 추가 
	args = append(args, productNumber)

	// 5. 쿼리 실행
    result, err := d.connection.ExecContext(ctx,query, args...)
    if err != nil {
        d.logger.Error().Err(err).Msg("[deviceRepo] failed to update device")
        return err
    }

	// 6. 실제로 변경이 일어났는지 확인 
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		d.logger.Error().Err(err).Msg("[deviceRepo] non update device")
		return err
	}

	return nil
}

func (d *DevicesRepo) Delete(ctx context.Context, productNumber string)  error {
	query := "DELETE FROM devices WHERE ProductNumber = ?"

	result, err := d.connection.ExecContext(ctx, query, productNumber)
	if err != nil {
		d.logger.Error().Err(err).Msg("[deviceRepo] failed to delete devices")
		return err
	}

	_, err = result.RowsAffected()
	if err != nil {
		d.logger.Error().Err(err).Msg("[deviceRepo] non Delete device)")
		return err
	}

	return nil
}

// 조건문 생성 기능만을 담당하는 함수 : 역할 분리 
func (d *DevicesRepo) GenerateUpdateQuery(params *external.UpdateDeviceParams) (string, []interface{}) {
	setClauses := []string{}
	args := []interface{}{}

	// 2. 파라미터로 받은 값들을 확인하며 쿼리 조립 (MySQL용 ? 플레이스홀더 사용)
	if params.FirmwareVersion != nil {
		setClauses = append(setClauses, "FirmwareVersion = ?")
		args = append(args, *params.FirmwareVersion)
	}

	if params.LastSeenAt != nil {
		setClauses = append(setClauses, "LastSeenAt = ?")
		args = append(args, *params.LastSeenAt)
	}

	if params.ReTry != nil {
		setClauses = append(setClauses, "ReTry = ?")
		args = append(args, *params.ReTry)
	}

	if params.UpdateCheck != nil {
		setClauses = append(setClauses, "UpdateCheck = ?")
		args = append(args, *params.UpdateCheck)
	}

	if params.Status != nil {
		setClauses = append(setClauses, "Status = ?")
		args = append(args, *params.Status)
	}

	// 3. 변경할 내용이 없으면 아무것도 하지 않고 종료
	if len(setClauses) == 0 {
		return "", nil
	}

	// 4. 최종 쿼리문 생성 (WHERE 조건의 ID는 마지막에 추가)
	query := fmt.Sprintf(
		"UPDATE devices SET %s WHERE ProductNumber = ?",
		strings.Join(setClauses, ", "),
	)

	return query, args
}