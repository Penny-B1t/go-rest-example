package model

type ServiceEnv struct {
	Name string   // 서비스 환경 이름 
	Port string   // 포트 번호
	DBname string // 데이터베이스 이름
	LogLevel string // 로깅 레벨
}