package db

import (
	"context"
	"time"

	"go-rest-example/internal/logger"
	"go-rest-example/internal/model"
	"go-rest-example/internal/model/domain"
	"go-rest-example/internal/model/entity"
)

// 필수 상수 선언
const (
	DefSchema = "reports"
	DefLimit  = 50
)

// ReportsRepo를 통해 사용할 메서드를 제약하고 규정하기 위한 인터페이스 
type ReportsDataService interface {
	Create(ctx context.Context, di *domain.DeviceInfo) (int64, error)
	GetAll(ctx context.Context) (*[]domain.DeviceInfo, error)
	GetByProductNumber(ctx context.Context, productNumber string) (*[]domain.DeviceInfo, error)
	Delete(ctx context.Context, productNumber string)  error
}

// reports 테이블을 접근하기 위한 커넥션 관리
type ReportsRepo struct {
	connection DBTX
	logger     *logger.AppLogger
}

func NewReportsRepo(lgr *logger.AppLogger, db DBTX) *ReportsRepo {
	return &ReportsRepo{
		connection: db,
		logger:     lgr,
	}
}

// 주기보고 정보 row 생성
func (r *ReportsRepo) Create(ctx context.Context, di *domain.DeviceInfo) (int64, error) {
	// ReportAt 설정 (테이블에 DEFAULT가 없으면)
	di.ReportAt = time.Now()

	query := `
		INSERT INTO reports (ProductNumber, BatteryPercent, Lat, Lon, TemperatureCelsius, IP, ErrorCode, ReportAt, ReportedStatus)
		VALUES (:ProductNumber, :BatteryPercent, :Lat, :Lon, :TemperatureCelsius, :IP, :ErrorCode, :ReportAt, :ReportedStatus)`

	result, err := r.connection.NamedExecContext(ctx, query, di)

	if err != nil {
		r.logger.Error().Err(err).Msg("failed to create device_info")
		return 0, err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastID, nil
}

// 장비 식별자 추가 필요 
func (r *ReportsRepo) GetAll(ctx context.Context) (*[]domain.DeviceInfo, error) {
	var reports []domain.DeviceInfo
	query := `
		SELECT ProductNumber, BatteryPercent, Lat, Lon, TemperatureCelsius, IP, ErrorCode, ReportAt, ReportedStatus
		FROM reports
		ORDER BY ReportAt DESC
		LIMIT ?`

	err := r.connection.SelectContext(ctx, &reports, query, DefLimit)
	if err != nil {
		r.logger.Error().Err(err).Msg("failed to select all reports with sqlx")
		return nil, err
	}
	return &reports, nil
}

// Device ID에 해당하는 정보 획득 (단일 반환으로 변경)
func (r *ReportsRepo) GetByProductNumber(ctx context.Context, productNumber string) (*[]domain.DeviceInfo, error) {
	var reports []entity.DeviceInfo
	query := `
		SELECT ProductNumber, BatteryPercent, Lat, Lon, TemperatureCelsius, IP, ErrorCode, ReportAt, ReportedStatus
		FROM reports
		WHERE ProductNumber = ?
		ORDER BY ReportAt DESC
		LIMIT ?`

	err := r.connection.SelectContext(ctx, &reports, query, productNumber, DefLimit)
	if err != nil {
		r.logger.Error().Err(err).Str("productNumber", productNumber).Msg("failed to select reports by product number with sqlx")
		return nil, err
	}
	return toDomainDeviceInfoSlice(reports), nil
}

// Device ID에 해당하는 정보 제거
func (r *ReportsRepo) Delete(ctx context.Context, productNumber string) error {
	query := "DELETE FROM reports WHERE ProductNumber = ?"

	result, err := r.connection.ExecContext(ctx, query, productNumber)
	if err != nil {
		r.logger.Error().Err(err).Msg("failed to delete device_info")
		return err
	}

	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

// toDomainDevice는 entity를 domain 객체로 변환하는 매퍼 함수
func toDomainDeviceInfo(e entity.DeviceInfo) *domain.DeviceInfo {
    return &domain.DeviceInfo{
        ReportID:   	e.ReportID,
		ProductNumber:  e.ProductNumber,
		BatteryPercent: e.BatteryPercent,
		Lat:      	 	e.Lat,
		Lon:  			e.Lon,
		TemperatureCelsius:   e.TemperatureCelsius,
		IP:    			e.IP,
		ErrorCode:      e.ErrorCode,
		ReportAt: 		e.ReportAt,
		ReportedStatus: model.DeviceStatus(e.ReportedStatus),
    }
}

func toDomainDeviceInfoSlice(entities []entity.DeviceInfo) *[]domain.DeviceInfo {
    // 1. 결과로 반환할 domain 슬라이스를 미리 할당합니다.
    // len(entities)를 사용하여 필요한 만큼의 공간을 정확히 만들어 성능을 최적화합니다.
    domainDevices := make([]domain.DeviceInfo, 0, len(entities))

    // 2. entity 슬라이스를 순회합니다.
    for _, e := range entities {
        // 3. 각 entity를 domain 객체로 변환하여 새로운 슬라이스에 추가합니다.
        if d := toDomainDeviceInfo(e); d != nil {
            domainDevices = append(domainDevices, *d)
        }
    }

    // 4. 변환이 완료된 슬라이스를 반환합니다.
    return &domainDevices
}