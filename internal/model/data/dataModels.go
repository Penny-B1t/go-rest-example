package data

import (
	"time"
)

type DeviceStatus string

// 디바이스 상태 상수
const (
	// 서버가 판별하는 상태
	StatusReady    DeviceStatus = "Ready"       // 정상 (주기적으로 보고가 오고 있음)

	// 디바이스가 보고하는 상태
	ReportPowerOn  DeviceStatus = "PowerOn"  // 전원 켜짐을 보고
	ReportPowerOff DeviceStatus = "PowerOff" // 전원 꺼짐을 보고
	ReportError  DeviceStatus = "ERROR"      // 상태 보고 주기 시 장비 오류 
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
	Status          DeviceStatus `db:"Status"`
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
	ReportedStatus     DeviceStatus `db:"ReportedStatus"`
}
