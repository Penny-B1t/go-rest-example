package model

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