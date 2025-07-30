package db

import (
	"context"
	"errors"
	"time"

	"go-rest-example/internal/logger"
	"go-rest-example/internal/model/data"
)

// 필수 상수 선언
const (
	DefSchema = "reports"
	DefLimit  = 50
)
// orm을 사용하지 않은 이유?

// 오류 상수 선언
var (
	ErrInvalidReportRequired          = errors.New("missing required inputs to create ReportsRepo")
	ErrFailedToCreateReportInfo = errors.New("failed to create device_info")
	ErrFailedToSelectReportInfo = errors.New("failed to select device_info")
	ErrFailedToDeleteReportInfo = errors.New("failed to delete device_info")
	ErrInvalidIDSelect          = errors.New(" invalid ProductNumber")
)

// ReportsRepo를 통해 사용할 메서드를 제약하고 규정하기 위한 인터페이스 
type ReportsDataService interface {
	Create(ctx context.Context, di *data.DeviceInfo) (int64, error)
	GetAll(ctx context.Context) (*[]data.DeviceInfo, error)
	GetByProductNumber(ctx context.Context, productNumber string) (*[]data.DeviceInfo, error)
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
func (r *ReportsRepo) Create(ctx context.Context, di *data.DeviceInfo) (int64, error) {
	// ReportAt 설정 (테이블에 DEFAULT가 없으면)
	di.ReportAt = time.Now()

	query := `
		INSERT INTO reports (ProductNumber, BatteryPercent, Lat, Lon, TemperatureCelsius, IP, ErrorCode, ReportAt, ReportedStatus)
		VALUES (:ProductNumber, :BatteryPercent, :Lat, :Lon, :TemperatureCelsius, :IP, :ErrorCode, :ReportAt, :ReportedStatus)`

	result, err := r.connection.NamedExecContext(ctx, query, di)

	if err != nil {
		r.logger.Error().Err(err).Msg("failed to create device_info")
		return 0, ErrFailedToCreateReportInfo
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastID, nil
}

// 장비 식별자 추가 필요 
func (r *ReportsRepo) GetAll(ctx context.Context) (*[]data.DeviceInfo, error) {
	var reports []data.DeviceInfo
	query := `
		SELECT ProductNumber, BatteryPercent, Lat, Lon, TemperatureCelsius, IP, ErrorCode, ReportAt, ReportedStatus
		FROM reports
		ORDER BY ReportAt DESC
		LIMIT ?`

	err := r.connection.SelectContext(ctx, &reports, query, DefLimit)
	if err != nil {
		r.logger.Error().Err(err).Msg("failed to select all reports with sqlx")
		return nil, ErrFailedToSelectReportInfo
	}
	return &reports, nil
}

// Device ID에 해당하는 정보 획득 (단일 반환으로 변경)
func (r *ReportsRepo) GetByProductNumber(ctx context.Context, productNumber string) (*[]data.DeviceInfo, error) {
	var reports []data.DeviceInfo
	query := `
		SELECT ProductNumber, BatteryPercent, Lat, Lon, TemperatureCelsius, IP, ErrorCode, ReportAt, ReportedStatus
		FROM reports
		WHERE ProductNumber = ?
		ORDER BY ReportAt DESC
		LIMIT ?`

	err := r.connection.SelectContext(ctx, &reports, query, productNumber, DefLimit)
	if err != nil {
		r.logger.Error().Err(err).Str("productNumber", productNumber).Msg("failed to select reports by product number with sqlx")
		return nil, ErrFailedToSelectReportInfo
	}
	return &reports, nil
}

// Device ID에 해당하는 정보 제거
func (r *ReportsRepo) Delete(ctx context.Context, productNumber string) error {
	query := "DELETE FROM reports WHERE ProductNumber = ?"

	result, err := r.connection.ExecContext(ctx, query, productNumber)
	if err != nil {
		r.logger.Error().Err(err).Msg("failed to delete device_info")
		return ErrFailedToDeleteReportInfo
	}

	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}