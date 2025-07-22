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
)


// 디바이스의 고유 정보 (DB에 저장되는 모델)
type Device struct {
	InternalID 	  int64 // DB에서 사용할 내부 ID auto increments 
	ProductNumber string     // 사용자가 식별하는 제품 번호 (Unique Key)
	MacAddress    string        // 디바이스의 MAC 주소 (Unique Key)
	FirmwareVersion string  
	LastSeenAt    time.Time     // 마지막으로 보고를 받은 시간
	CreatedAt     time.Time 
	ReTry         int 
	UpdateCheck   int         
	Status        DeviceStatus      // 서버가 판단하는 디바이스의 최종 상태
}

// DeviceInfo는 디바이스가 서버로 주기적으로 보고하는 정보 (DTO)
type DeviceInfo struct {
	ReportID           int64 // DB에서 사용할 내부 ID auto increments 
	ProductNumber      string
	BatteryPercent     int      // 배터리 퍼센트 (0-100)
	Lat                float64  // 위도
	Lon                float64  // 경도
	TemperatureCelsius float64  // 온도 (섭씨), 정밀도를 위해 float64 고려
	IP                 string   // IP 정보
	ErrorCode          int      // 에러 코드 (0: 정상)
	ReportAt           time.Time // 보고 시간 정보 
	ReportedStatus     DeviceStatus    // 디바이스가 보고하는 현재 상태 (예: PowerOn)
}


// DeviceUpdate는 서버가 디바이스에 응답으로 보내는 제어 정보 (DTO)
type DeviceUpdate struct {
	ReportCycleSec int  // 보고 주기 (초 단위)
	PowerOff       int // 원격 종료 명령
	Reboot         int // 원격 재부팅 명령
}