package domain

import (
	"go-rest-example/internal/model"
	"time"
)

// Device 도메인 객체
type Device struct {
    InternalID      int64
    ProductNumber   string
    MacAddress      string
    FirmwareVersion string
    LastSeenAt      time.Time
    ReTry           int
    UpdateCheck     int
    Status          model.DeviceStatus
}

// Device 도메인 규칙 정의 
func (d *Device) NeedsPowerOff() bool {
    // 예시: 재시도 횟수가 3회 이상이면 재부팅이 필요하다.
    return d.ReTry >= 3
}

func (d *Device) NeedsRebot(power int, error int) bool {
	return power != 1 && error != 0
}

func (d *Device) NewRetry(error int) int {
	if error != 0 {
		newRetryCount := d.ReTry + 1
		d.ReTry = newRetryCount 
	}
	
	return d.ReTry
}


// DeviceInfo 도메인 객체
type DeviceInfo struct {
	ReportID           int64        
	ProductNumber      string       
	BatteryPercent     int          
	Lat                float64      
	Lon                float64      
	TemperatureCelsius float64      
	IP                 string       
	ErrorCode          int          
	ReportAt           time.Time    
	ReportedStatus     model.DeviceStatus 
}
