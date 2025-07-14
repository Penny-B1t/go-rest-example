package handlers

import (
	"errors"
	"go-rest-example/internal/db"
	"go-rest-example/internal/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	ErrInvalidlogger  = errors.New("failed to connect to DB, as the connection string is invalid")
	ErrClientInit     = errors.New("failed to initialize DB client")
	ErrConnectionLeak = errors.New("unable to disconnect from DB, potential connection leak")
	ErrPingDB         = errors.New("failed to ping DB")
)



type StatusHandler struct {
	dbMgr db.DBManager
	lgr   *logger.AppLogger
}

func NewStatusHandler(l *logger.AppLogger, d db.DBManager) (*StatusHandler, error) {

	if l == nil || d == nil {
		return nil, errors.New("missing required inputs to create status handler")
	}
	return &StatusHandler{
		dbMgr: d,
		lgr: l,
	}, nil
}

func (s *StatusHandler) CheckStatus(c *gin.Context) {
		var code int

		if err := s.dbMgr.Ping(); err == nil {
			// 성공 시 2** 코드 
			code = http.StatusNoContent
		} else {
			// 오류 발생시 5** 코드
			s.lgr.Error().Msg("failed to ping DB")
			code = http.StatusFailedDependency
		}

		// 상태 반환 
		c.JSON(code, nil)
}