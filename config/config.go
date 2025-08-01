package config

// Config는 애플리케이션의 모든 설정을 담는 최상위 구조체입니다.
type Config struct {
	App    AppConfig    `mapstructure:"app"`
	Server ServerConfig `mapstructure:"server"`
	DB     DBConfig     `mapstructure:"db"`
}

// AppConfig는 애플리케이션 관련 설정을 담습니다.
type AppConfig struct {
	Environment string `mapstructure:"enviroment"`
	LogLevel    string `mapstructure:"logLevel"`
}

// ServerConfig는 서버 관련 설정을 담습니다.
type ServerConfig struct {
	Port int `mapstructure:"port"`
}

// DBConfig는 데이터베이스 연결 설정을 담습니다.
type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}
