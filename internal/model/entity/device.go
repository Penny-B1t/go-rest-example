package entity

import (
	"go-rest-example/internal/model"
	"time"
)

// 디바이스의 고유 정보 (DB에 저장되는 모델)
type Device struct {
	InternalID      int64        `db:"InternalID"`      // DB 컬럼 이름이 'InternalID'라고 가정
	ProductNumber   string       `db:"ProductNumber"`
	MacAddress      string       `db:"MacAddress"`
	FirmwareVersion string       `db:"FirmwareVersion"`
	LastSeenAt      time.Time    `db:"LastSeenAt"`
	CreatedAt       time.Time    `db:"CreatedAt"`
	ReTry           int          `db:"ReTry"`
	UpdateCheck     int          `db:"UpdateCheck"`
	Status          model.DeviceStatus `db:"Status"`
}

// DeviceInfo는 디바이스가 서버로 주기적으로 보고하는 정보 (DTO)
type DeviceInfo struct {
	ReportID           int64        `db:"ReportID"`
	ProductNumber      string       `db:"ProductNumber"`
	BatteryPercent     int          `db:"BatteryPercent"`
	Lat                float64      `db:"Lat"`
	Lon                float64      `db:"Lon"`
	TemperatureCelsius float64      `db:"TemperatureCelsius"`
	IP                 string       `db:"IP"`
	ErrorCode          int          `db:"ErrorCode"`
	ReportAt           time.Time    `db:"ReportAt"`
	ReportedStatus     model.DeviceStatus `db:"ReportedStatus"`
}
