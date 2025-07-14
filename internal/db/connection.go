package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-rest-example/internal/logger"
	"time"

	_ "github.com/go-sql-driver/mysql" // 데이터베이스 드라이버 구현체 추가
)

// --- 새로 추가된 인터페이스들 ---

// DBTX는 데이터베이스 쿼리 실행기(sql.DB 또는 sql.Tx)에 대한 인터페이스입니다.
// Repository 레이어가 이 인터페이스에 의존하게 하여 테스트 용이성을 높입니다.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// DBManager는 데이터베이스 연결의 생명주기를 관리하는 인터페이스입니다.
type DBManager interface {
	DB() DBTX // DB 또는 Tx를 나타내는 DBTX 인터페이스 반환
	Ping() error
	Disconnect() error
}

// --- 기존 코드 (일부 수정) ---

type MariaDBCredentials struct {
	User     string
	Password string
	Host     string
	Port     int
	Database string
}

type MariaDBManager struct {
	db     *sql.DB
	logger *logger.AppLogger
}

// 컴파일 타임에 MariaDBManager가 DBManager 인터페이스를 구현하는지 확인합니다.
var _ DBManager = (*MariaDBManager)(nil)

var (
	ErrInvalidConnURL    = errors.New("failed to connect to DB, as the connection string is invalid")
	ErrConnectionEstablish = errors.New("failed to establish connection to DB")
	ErrClientInit        = errors.New("failed to initialize DB client")
	ErrConnectionLeak    = errors.New("unable to disconnect from DB, potential connection leak")
	ErrPingDB            = errors.New("failed to ping DB")
)

func NewMariaDBManager(creds *MariaDBCredentials, lgr *logger.AppLogger) (DBManager, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", // parseTime=true 추가 권장
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

	mgr := &MariaDBManager{
		db:     db,
		logger: lgr,
	}

	if err := mgr.Ping(); err != nil {
		// Ping 실패 시 생성된 db 객체를 닫아주는 것이 좋습니다.
		db.Close()
		return nil, err
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return mgr, nil
}

// DB - DBTX 인터페이스를 반환합니다. *sql.DB는 DBTX를 구현하므로 그대로 반환할 수 있습니다.
func (m *MariaDBManager) DB() DBTX {
	return m.db
}

func (m *MariaDBManager) Disconnect() error {
	m.logger.Info().Msg("disconnecting from MariaDB")
	return m.db.Close()
}

func (m *MariaDBManager) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.db.PingContext(ctx); err != nil {
		m.logger.Error().Err(err).Msg("failed to ping DB")
		return ErrPingDB
	}
	return nil
}

func MaskConnectionDSN(creds *MariaDBCredentials) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		"******",
		"******",
		creds.Host,
		creds.Port,
		creds.Database,
	)
}