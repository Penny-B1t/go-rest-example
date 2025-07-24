package gorestexample

import (
	"context"
	"errors"
	"fmt"
	"go-rest-example/internal/db"
	"go-rest-example/internal/logger"
	"go-rest-example/internal/model"
	"go-rest-example/internal/server"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	// "syrconv"
	// "time"
	// internal
)

// 상수 선언언
const (
	serviceName = ""
	defaultPort = "8080"
	defaultLogLevel = "info"
)

var version string

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Service %s exited with error: %v (exit code: %d) \n",
		 serviceName, err, exitCode(err))
		os.Exit(exitCode(err))
	}
}

func run() error {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 에러 채널을 생성하여, 치명적인 오류 수신 시 종료
	errChan := make(chan error, 1)

	svcenv, envErr := getEnvConfig()
	if envErr != nil {
		return envErr
	}

	// 로그 모듈 setup
	lgr := logger.Setup(svcenv.LogLevel, svcenv.Name)
	
	//setup : database 연결
	dbConnMgr, dbErr := setupDB(lgr, svcenv)
	if dbErr != nil {
		return dbErr
	}

	go func(){
		errChan <- server.Start(svcenv, lgr, dbConnMgr)
	}()

	lgr.Info().
		Str("name", serviceName).
		Str("environment", svcenv.Name).
		Str("started at", time.Now().UTC().Format(time.RFC3339)).
		Str("version", version).
		Msg("starting the service")

	// Wait until termination or a critical error
	select {
	case <-ctx.Done():
		lgr.Info().Msg("graceful shutdown signal received")
		err := <-errChan // wait for go routines to exit
		cleanup(lgr, dbConnMgr)
		return err
	case err := <-errChan:
		lgr.Error().Err(err).Msg("something went wrong")
		cleanup(lgr, dbConnMgr)
		return err
	}
	
}

// .env 파일을 읽어 시스템 동작에 사용한다.
// builder 패턴을 사용하여 환경 변수 객체의 주소값을 반환한다.
// TODO ServiceEnv 구조체 형태 변경에 따른 조건식 변경 
func getEnvConfig() (*model.ServiceEnv, error) {

	// 작업 환경 구분
	// 기본값 local
	envName := os.Getenv("enviroment")
	if envName == "" {
		envName = "local"
	}

	// 포트 번호를 확인하는 작업
	// 기본값 8080
	port := os.Getenv("port")
	if port == "" {
		port = defaultPort
	}

	// DB 명칭을 확인
	// 필수
	dbname := os.Getenv("dbname")
	if dbname == "" {
		return nil, errors.New("dbname is required")
	}
	

	// 로그 레벨 지정
	logLevel := os.Getenv("logLevel")
	if logLevel == "" {
		logLevel = defaultLogLevel
	}

	//
	envConfigurations := &model.ServiceEnv{
		Name : envName,
		Port : port,
		DBname : dbname,
		LogLevel : logLevel,
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
	portInt, err := strconv.Atoi(svcEnv.Port)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %v", err)
	}
	connOpts := &db.MariaDBCredentials{
		User:     "root",
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