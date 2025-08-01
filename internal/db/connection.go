package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-rest-example/internal/logger"
	"go-rest-example/internal/util"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL 드라이버는 그대로 사용합니다.
	"github.com/jmoiron/sqlx"          // sqlx 임포트 추가
)

/*
 @brief  데이터베이스 쿼리 실행기(sqlx.DB 또는 sqlx.Tx)에 대한 인터페이스
         sqlx의 *sqlx.DB와 *sqlx.Tx는 이 인터페이스를 모두 구현합니다.
*/
type DBTX interface {
	// --- 표준 `database/sql` 호환 메서드 ---

	// Context 사용 버전
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row

	// --- `sqlx`의 핵심 편의 기능 (구조체 스캔) ---

	// Context 사용 버전
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error

	// --- `sqlx`의 명명된 쿼리(Named Query) 지원 ---

	// Context 사용 버전
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
}

/*
 @brief  데이터베이스 연결의 생명주기를 관리하는 인터페이스
*/
type DBManager interface {
	// DB() 메서드가 이제 *sqlx.DB를 반환합니다.
	DB() *sqlx.DB
	Ping() error
	Disconnect() error
}

/*
 @brief  연결 정보 구조체 (변경 없음)
*/
type MariaDBCredentials struct {
	User     string
	Password string
	Host     string
	Port     int
	Database string
}

type MariaDBManager struct {
	// db 필드를 *sql.DB에서 *sqlx.DB로 변경합니다.
	db     *sqlx.DB
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

func NewMariaDBManager(creds *MariaDBCredentials, lgr *logger.AppLogger, envMode string) (DBManager, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		creds.User,
		creds.Password,
		creds.Host,
		creds.Port,
		creds.Database,
	)

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		lgr.Error().Err(err).Msg("failed to connect or ping DB with sqlx")
		return nil, ErrConnectionEstablish
	}

	if !util.IsDevMode(envMode) {
		dsn = MaskConnectionDSN(creds)
	}

	lgr.Info().Str("connURL", dsn).Msg("connecting to MariaDB with sqlx")

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	mgr := &MariaDBManager{
		db:     db,
		logger: lgr,
	}

	return mgr, nil
}

// DB - 이제 *sqlx.DB를 반환합니다.
func (m *MariaDBManager) DB() *sqlx.DB {
	return m.db
}

func (m *MariaDBManager) Disconnect() error {
	m.logger.Info().Msg("disconnecting from MariaDB")
	return m.db.Close()
}

func (m *MariaDBManager) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// sqlx.DB도 PingContext를 지원합니다.
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