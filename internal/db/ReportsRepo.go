package db

import (
	"context"
	"errors"
	"strconv"
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
	Create(ctx context.Context, di *data.DeviceInfo) (string, error)
	GetAll(ctx context.Context) (*[]data.DeviceInfo, error)
	GetByID(ctx context.Context, ID string) (*[]data.DeviceInfo, error)
	Delete(ctx context.Context, ID string)  error
}

// reports 테이블을 접근하기 위한 커넥션 관리
type ReportsRepo struct {
	connection DBTX
	logger     *logger.AppLogger
}

func NewReportsRepo(lgr *logger.AppLogger, db DBTX) (*ReportsRepo, error) {
	if lgr == nil || db == nil {
		return nil, ErrInvalidReportRequired
	}
	return &ReportsRepo{
		connection: db,
		logger:     lgr,
	}, nil
}

// 주기보고 정보 row 생성
func (d *ReportsRepo) Create(ctx context.Context, di *data.DeviceInfo) (string, error) {
	// ReportAt 설정 (테이블에 DEFAULT가 없으면)
	di.ReportAt = time.Now()

	query := "INSERT INTO reports (ProductNumber, BatteryPercent, Lat, Lon, TemperatureCelsius, IP, ErrorCode, ReportAt, ReportedStatus) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"

	result, err := d.connection.ExecContext(ctx, query,
		di.ProductNumber,
		di.BatteryPercent, 
		di.Lat, 
		di.Lon, 
		di.TemperatureCelsius, 
		di.IP, 
		di.ErrorCode, 
		di.ReportAt, 
		di.ReportedStatus)

	if err != nil {
		d.logger.Error().Err(err).Msg("failed to create device_info")
		return "", ErrFailedToCreateReportInfo
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(lastID, 10), nil
}

// 장비 식별자 추가 필요 
func (d *ReportsRepo) GetAll(ctx context.Context) (*[]data.DeviceInfo, error) {
	query := "SELECT ProductNumber, BatteryPercent, Lat, Lon, TemperatureCelsius, IP, ErrorCode, ReportAt, ReportedStatus FROM reports LIMIT ?"

	rows, err := d.connection.QueryContext(ctx, query, DefLimit)
	if err != nil {
		d.logger.Error().Err(err).Msg("failed to select device_info")
		return nil, ErrFailedToSelectReportInfo
	}
	defer rows.Close()

	var responseData []data.DeviceInfo

	for rows.Next() {
		var device data.DeviceInfo
		err := rows.Scan(
			&device.ProductNumber, &device.BatteryPercent, &device.Lat, &device.Lon, &device.TemperatureCelsius,
			&device.IP, &device.ErrorCode, &device.ReportAt,  &device.ReportedStatus,
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

// Device ID에 해당하는 정보 획득 (단일 반환으로 변경)
func (d *ReportsRepo) GetByID(ctx context.Context, ID string) (*[]data.DeviceInfo, error) {
	query := "SELECT ProductNumber, BatteryPercent, Lat, Lon, TemperatureCelsius, IP, ErrorCode, ReportAt, ReportedStatus FROM reports WHERE ProductNumber = ? LIMIT ?"

	rows, err := d.connection.QueryContext(ctx, query, ID, DefLimit)
	if err != nil {
		d.logger.Error().Err(err).Msg("failed to select device_info by ID")
		return nil, ErrFailedToSelectReportInfo
	}
	
	defer rows.Close()

	var responseData []data.DeviceInfo

	for rows.Next() {
		var device data.DeviceInfo
		err := rows.Scan(
			&device.ProductNumber,&device.BatteryPercent, &device.Lat, &device.Lon, &device.TemperatureCelsius,
			&device.IP, &device.ErrorCode, &device.ReportAt, &device.ReportedStatus,
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

// Device ID에 해당하는 정보 제거
func (d *ReportsRepo) Delete(ctx context.Context, ID string) error {
	query := "DELETE FROM reports WHERE ProductNumber = ?"

	result, err := d.connection.ExecContext(ctx, query, ID)
	if err != nil {
		d.logger.Error().Err(err).Msg("failed to delete device_info")
		return ErrFailedToDeleteReportInfo
	}

	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}