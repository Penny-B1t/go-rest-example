package external

import (
	"go-rest-example/internal/model/data"
)

// DTO 선언 응답 혹은

// 오류에 대한 응답 DTO
// 오류에 대한 모든 내용을 출력할 될 경우 보안 취약점으로 돌아올 수 있기 때문에 행태 제한
type APIError struct {
	HTTPStatusCode int     `json:"httpStatusCode"`
	Message        string  `json:"message"`
	DebugID        string  `json:"debugId"`
	ErrorCode      string  `json:"errorCode"`
}

type ReportReq struct {
	ProductNumber string              `json:"productNumber" binding:"required"`
	BatteryPercent int                `json:"batteryPercent" binding:"required"`
	Lat float64                       `json:"lat" binding:"required"`
	Lon float64                       `json:"lon" binding:"required"`
	TemperatureCelsius float64        `json:"temperatureCelsius" binding:"required"`
	IP string                         `json:"ip" binding:"required"`
	ErrorCode int                     `json:"errorCode" binding:"required"`
	ReportedStatus data.DeviceStatus  `json:"reportedStatus" binding:"required"`
}


