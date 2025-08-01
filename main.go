package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"os"
	"os/signal"
	"syscall"

	"go-rest-example/config"
	"go-rest-example/internal/db"
	"go-rest-example/internal/logger"
	"go-rest-example/internal/server"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// 상수 선언언
const (
	serviceName = ""
	defaultPort = "8080"
	defaultLogLevel = "info"
)

var version string

func main() {

	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found or could not be loaded: %v\n", err)
	}
	
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Service %s exited with error: %v (exit code: %d) \n",
		 serviceName, err, exitCode(err))
		os.Exit(exitCode(err))
	}
}

func run() error {
	// 1. 종료 신호를 감지하는 Context 생성 
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, cfgErr := LoadConfig()
	if cfgErr != nil {
		return cfgErr
	}


	lgr := logger.Setup(cfg.App.LogLevel, cfg.App.Environment)
	dbConnMgr, dbErr := setupDB(lgr, cfg)
	if dbErr != nil {
		return dbErr
	}
	defer cleanup(lgr, dbConnMgr) 

	lgr.Info().Msgf("config: %+v", cfg)

	// 3. [REFACTOR] 서버 시작 로직 간소화
	lgr.Info().
		Str("name", serviceName).
		Str("environment", cfg.App.Environment).
		Str("version", version).
		Msg("starting the service")

	if err := server.Start(ctx, cfg, lgr, dbConnMgr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		lgr.Error().Err(err).Msg("server failed to start")
		return err
	}

	lgr.Info().Msg("service stopped")
	return nil
}

func LoadConfig() (config *config.Config, err error) {

	// 0. 기본값 지정
	viper.SetDefault("app.logLevel", "info")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("db.host", "localhost")
	viper.SetDefault("db.port", 3306)

	// 1. viper 인스턴스 생성
	viper.AddConfigPath("./config")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// 환경 변수를 자동으로 읽어오도록 설정합니다
	viper.AutomaticEnv()

	// 설정 파일을 읽습니다.
	if err := viper.ReadInConfig(); err != nil {
		// 파일이 없는 경우(ConfigFileNotFoundError)는 에러가 아니므로,
		// 로그만 남기고 무시합니다. 기본값과 환경변수로 계속 진행합니다.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// 그 외 다른 파일 관련 에러(권한 문제 등)는 실제 에러이므로 반환합니다.
			return config, fmt.Errorf("failed to read config file: %w", err)
		}
		fmt.Println("Warning: config.yaml not found. Using defaults and environment variables.")
	}


	// 설정값을 구조체 인스턴스로 변환
	err = viper.Unmarshal(&config)

	return
}



func exitCode(err error) int {
	if err == nil || errors.Is(err,context.Canceled) {
		return 0
	}
	return 1
}


func setupDB(lgr *logger.AppLogger, cfg *config.Config) (db.DBManager, error) {
	connOpts := &db.MariaDBCredentials{
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		Database: cfg.DB.Database,
	}

	dbConnMgr, dberr := db.NewMariaDBManager(connOpts, lgr, cfg.App.Environment)
	if dberr != nil {
		return nil, dberr
	}

	return dbConnMgr, nil

}
	
func cleanup(lgr *logger.AppLogger, dbConnMgr db.DBManager) {
	if err := dbConnMgr.Disconnect(); err != nil {
		lgr.Error().Err(err).Msg("failed to close DB connection, potential connection leak")
		return
	}
}