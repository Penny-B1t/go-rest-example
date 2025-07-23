package external

import (
	"errors"
	"go-rest-example/internal/model/data"
	"time"
)

// 오류 타입 선언
var (
	errordRequired = errors.New("error code is required when status is ERROR")
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

// DeviceUpdate는 서버가 디바이스에 응답으로 보내는 제어 정보 (DTO)
type DeviceUpdate struct {
	ReportCycleSec int  // 보고 주기 (초 단위)
	PowerOff       int // 원격 종료 명령
	Reboot         int // 원격 재부팅 명령
}

// 요청 DTO 
type DeviceReq struct {
	ProductNumber string     // 사용자가 식별하는 제품 번호 (Unique Key)
	MacAddress    string     // 디바이스의 MAC 주소 (Unique Key)
	FirmwareVersion string  
}

func (d *DeviceReq)Validate() error{
	return nil
}

// 기본적인 검증 수행
type ReportReq struct {
	ProductNumber      string            `json:"productNumber" binding:"required"`
	BatteryPercent     int               `json:"batteryPercent" binding:"required,min=0,max=100"`
	Lat                float64           `json:"lat" binding:"required"`
	Lon                float64           `json:"lon" binding:"required"`
	TemperatureCelsius float64           `json:"temperatureCelsius" binding:"required"`
	IP                 string            `json:"ip" binding:"required,ip"`
	ErrorCode          int               `json:"errorCode"` 
	ReportedStatus     data.DeviceStatus `json:"reportedStatus" binding:"required"`
}

// 복합 조건 검증 수행 
func(r *ReportReq)Validate() error {
	// 보고 타입이 에러인 경우 에러 타입 입력 필수 
	if r.ReportedStatus == "ERROR" && r.ErrorCode == 0 {
        return errordRequired
    }

	// 국내 위도 범위 검증
	if r.Lat < 33 || r.Lat > 34 {
        return errors.New("latitude must be between 33 and 34")
    }

	// 국내 위도 범위 검증 
	if r.Lat < 124 || r.Lat > 132 {
        return errors.New("latitude must be between 124 and 132")
    }

	// 제품 명칭 길이 : 정규 표현식 사용 고려 
	// 간단한 예시 
	if len(r.ProductNumber) > 9 {
		return errors.New("invalide to Name lange")
	}

	return nil

}

// Device를 업데이트할 때 사용할 파라미터
type UpdateDeviceParams struct {
    FirmwareVersion *string
    LastSeenAt      *time.Time
    ReTry           *int
    UpdateCheck     *int
    Status          *data.DeviceStatus
}