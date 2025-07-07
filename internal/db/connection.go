package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go-rest-example/internal/logger"

	_ "github.com/go-sql-driver/mysql" // 데이테베이스 드라이브 구현체 추가
)

type MariaDBCredentials struct {
    User     string
    Password string
    Host     string
    Port     int
    Database string
}

// MariaDBManager는 *sql.DB 객체를 관리합니다.
// *sql.DB는 내부에 커넥션 풀을 가지고 있어 스레드-세이프합니다.
// wapper 방식을 통한 이벤트 모니터링 기능 추가
type MariaDBManager struct {
    db     *sql.DB
    logger *logger.AppLogger
}

var (
	ErrInvalidConnURL      = errors.New("failed to connect to DB, as the connection string is invalid")
	ErrConnectionEstablish = errors.New("failed to establish connection to DB")
	ErrClientInit          = errors.New("failed to initialize DB client")
	ErrConnectionLeak      = errors.New("unable to disconnect from DB, potential connection leak")
	ErrPingDB              = errors.New("failed to ping DB")
)

// NewMariaDBManager - MariaDB 연결을 초기화하고 Manager를 반환합니다.
func NewMariaDBManager(creds *MariaDBCredentials, lgr *logger.AppLogger) (*MariaDBManager, error) {
    
    // 1. DSN (Data Source Name) 문자열 생성
    // 형식: "user:password@tcp(host:port)/dbname?parseTime=true"
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
        creds.User,
        creds.Password,
        creds.Host,
        creds.Port,
        creds.Database,
    )

    lgr.Info().Str("connURL", MaskConnectionDSN(creds)).Msg("connecting to MariaDB")

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        lgr.Error().Err(err).Msg("failed to prepare DB connection")
        return nil, ErrClientInit
    }

    // 객체 생성 및 초기화
    // mgr은 주소 포인터 타입이므로 주소 연산자를 사용하여 주소를 저장합니다.
    mgr := &MariaDBManager{
        db:     db,
        logger: lgr,
    }

    // 2. Ping()으로 실제 연결 테스트
    if err := mgr.Ping(); err != nil {
        return nil, err
    }

    // 3. 커넥션 풀 설정 
    db.SetConnMaxLifetime(time.Minute * 3)
    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(10)

    return mgr, nil
}

// DB - sql.DB 객체를 반환하는 메서드
// 쿼리 실행이 필요할 때 이 메서드를 통해 DB 핸들을 얻습니다.
func (m *MariaDBManager) DB() *sql.DB {
    return m.db
}

// Disconnect - DB 커넥션 풀을 닫습니다.
func (m *MariaDBManager) Disconnect() error {
    m.logger.Info().Msg("disconnecting from MariaDB")
    return m.db.Close()
}

// Ping - DB 연결 상태를 확인합니다.
// defer cancel() 을 통해 컨텍스트 취소 함수를 호출하여 컨텍스트를 종료합니다.
func (m *MariaDBManager) Ping() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := m.db.PingContext(ctx); err != nil {
        m.logger.Error().Err(err).Msg("failed to ping DB")
        return ErrPingDB 
    }

    return nil
}

// newClient - creates a new Client to connect DB.

// 캡슐화를 통한 은닉화 처리리

func MaskConnectionDSN(creds *MariaDBCredentials) string {

    return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
        "######",
        "######",
        creds.Host,
        creds.Port,
        creds.Database,
    )

}