package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"os"
	"os/signal"
	"strconv"
	"syscall"

	"go-rest-example/internal/db"
	"go-rest-example/internal/logger"
	"go-rest-example/internal/model"
	"go-rest-example/internal/server"

	"github.com/joho/godotenv"
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

	// 2. 환경변수 및 의존성 설정 
	svcenv, envErr := getEnvConfig()
	if envErr != nil {
		return envErr
	}
	lgr := logger.Setup(svcenv.LogLevel, svcenv.Name)
	dbConnMgr, dbErr := setupDB(lgr, svcenv)
	if dbErr != nil {
		return dbErr
	}
	defer cleanup(lgr, dbConnMgr) 

	// 3. [REFACTOR] 서버 시작 로직 간소화
	lgr.Info().
		Str("name", serviceName).
		Str("environment", svcenv.Name).
		Str("version", version).
		Msg("starting the service")

	if err := server.Start(ctx, svcenv, lgr, dbConnMgr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		lgr.Error().Err(err).Msg("server failed to start")
		return err
	}

	lgr.Info().Msg("service stopped")
	return nil
}
	

// .env 파일을 읽어 시스템 동작에 사용한다.
// builder 패턴을 사용하여 환경 변수 객체의 주소값을 반환한다.
func getEnvConfig() (*model.ServiceEnv, error) {
	// 작업 환경 구분
	// 기본값 local
	envName := os.Getenv("environment") // 오타: environment -> enviroment (.env 파일과 일치)
	if envName == "" {
		envName = "local"
	}

	// 애플리케이션 포트 번호
	// 기본값 8080
	port := os.Getenv("port")
	if port == "" {
		port = defaultPort
	}

	// 데이터베이스 호스트 정보
	// 기본값 localhost
	host := os.Getenv("host")
	if host == "" {
		host = "localhost"
	}

	// 데이터베이스 사용자 정보
	// 기본값 root
	user := os.Getenv("user")
	if user == "" {
		user = "root"
	}

	// 데이터베이스 패스워드
	// 필수 (보안상 기본값 없음)
	password := os.Getenv("password")
	if password == "" {
		return nil, errors.New("database password is required")
	}

	// 데이터베이스 포트 번호
	// 기본값 3306 (MySQL/MariaDB 기본 포트)
	dbPort := os.Getenv("dbport")
	if dbPort == "" {
		dbPort = "3306"
	}

	// DB 명칭 확인
	// 필수
	dbname := os.Getenv("dbname")
	if dbname == "" {
		return nil, errors.New("database name is required")
	}

	// 로그 레벨 지정
	logLevel := os.Getenv("logLevel")
	if logLevel == "" {
		logLevel = defaultLogLevel
	}

	// ServiceEnv 구조체 생성 및 반환
	envConfigurations := &model.ServiceEnv{
		Name:     envName,
		Host:     host,
		User:     user,
		Password: password,
		Port:     port,
		DBPort:   dbPort,
		DBname:   dbname,
		LogLevel: logLevel,
	}

	return envConfigurations, nil
}



func exitCode(err error) int {
	if err == nil || errors.Is(err,context.Canceled) {
		return 0
	}
	return 1
}


func setupDB(lgr *logger.AppLogger, svcEnv *model.ServiceEnv) (db.DBManager, error) {
	portInt, err := strconv.Atoi(svcEnv.DBPort)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %v", err)
	}
	connOpts := &db.MariaDBCredentials{
		User:     svcEnv.User,
		Password: svcEnv.Password,
		Host:     svcEnv.Host,
		Port:     portInt,
		Database: svcEnv.DBname,
	}

	dbConnMgr, dberr := db.NewMariaDBManager(connOpts, lgr)
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