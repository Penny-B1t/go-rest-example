package model

type ServiceEnv struct {
	Name string   // 서비스 환경 이름
	Host string   // 호스트 정보  
	User string    // 유저 정보
	Password string // 로그인 정보 
	Port string   // 포트 번호
	DBPort string 
	DBname string // 데이터베이스 이름
	LogLevel string // 로깅 레벨
}