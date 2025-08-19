package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-rest-example/internal/logger"
	"go-rest-example/internal/model"
	"go-rest-example/internal/model/domain"
	"go-rest-example/internal/model/entity"
	"go-rest-example/internal/model/external"
)

// DeviceRepo를 통해 사용할 메서드를 제약하고 규정하기 위한 인터페이스
type DevicesDataService interface {
	Create(ctx context.Context, di *domain.Device) (int64, error) 
	GetAll(ctx context.Context) (*[]domain.Device, error)
	GetByProductNumber(ctx context.Context, ID string) (*domain.Device, error)
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

func (d *DevicesRepo) Create(ctx context.Context, device *domain.Device)(int64, error){

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
        time.Now(),
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

func (d *DevicesRepo) GetAll(ctx context.Context) (*[]domain.Device, error){

	query := `
	SELECT InternalID, ProductNumber, MacAddress, FirmwareVersion, LastSeenAt, CreatedAt, ReTry, UpdateCheck, Status
	FROM devices`

	var devices []entity.Device

	err := d.connection.SelectContext(ctx, &devices, query)

	if err != nil {
		d.logger.Error().Err(err).Msg("[deviceRepo] non select device with sqlx")
		return nil, err
	}

	return toDomainDeviceSlice(devices), nil
}

func (d *DevicesRepo) GetByProductNumber(ctx context.Context, productNumber string) (*domain.Device, error){
	var device entity.Device

	query := `
		SELECT InternalID, ProductNumber, MacAddress, FirmwareVersion, LastSeenAt, CreatedAt, ReTry, UpdateCheck, Status
		FROM devices WHERE ProductNumber = ?`

	err := d.connection.GetContext(ctx, &device, query, productNumber)

	if err != nil {
		d.logger.Error().Err(err).Msg("[deviceRepo] failed to select device with sqlx")
		return nil, err
	}

	 return toDomainDevice(device), nil
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
	if params.FirmwareVersion != "" {
		setClauses = append(setClauses, "FirmwareVersion = ?")
		args = append(args, params.FirmwareVersion)
	}

	if !params.LastSeenAt.IsZero() {
		setClauses = append(setClauses, "LastSeenAt = ?")
		args = append(args, params.LastSeenAt)
	}

	if params.ReTry != 0 {
		setClauses = append(setClauses, "ReTry = ?")
		args = append(args, params.ReTry)
	}

	if params.UpdateCheck != 0 {
		setClauses = append(setClauses, "UpdateCheck = ?")
		args = append(args, params.UpdateCheck)
	}

	if params.Status != "" {
		setClauses = append(setClauses, "Status = ?")
		args = append(args, params.Status)
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

// toDomainDevice는 entity를 domain 객체로 변환하는 매퍼 함수
func toDomainDevice(e entity.Device) *domain.Device {
    return &domain.Device{
        InternalID:      e.InternalID,
        ProductNumber:   e.ProductNumber,
        MacAddress:      e.MacAddress,
		FirmwareVersion: e.FirmwareVersion,
		LastSeenAt:      e.LastSeenAt,
		ReTry:           e.ReTry,
		UpdateCheck:     e.UpdateCheck,
        Status:          model.DeviceStatus(e.Status),
    }
}

func toDomainDeviceSlice(entities []entity.Device) *[]domain.Device {
    // 1. 결과로 반환할 domain 슬라이스를 미리 할당합니다.
    // len(entities)를 사용하여 필요한 만큼의 공간을 정확히 만들어 성능을 최적화합니다.
    domainDevices := make([]domain.Device, 0, len(entities))

    // 2. entity 슬라이스를 순회합니다.
    for _, e := range entities {
        // 3. 각 entity를 domain 객체로 변환하여 새로운 슬라이스에 추가합니다.
        if d := toDomainDevice(e); d != nil {
            domainDevices = append(domainDevices, *d)
        }
    }

    // 4. 변환이 완료된 슬라이스를 반환합니다.
    return &domainDevices
}