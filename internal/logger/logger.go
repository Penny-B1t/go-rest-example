package logger

import (
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"go-rest-example/internal/util"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)


var (
	setUpOne sync.Once
	appLogger *AppLogger
)

type AppLogger struct {
	zLogger zerolog.Logger
}

func Setup(logLevel, envName string) *AppLogger {
	setUpOne.Do(func(){
		appLogger = &AppLogger{}

		lvl := parseLogLevel(logLevel)
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.TimeFieldFormat = time.RFC3339Nano
		var logDest io.Writer
		logDest = os.Stdout
		if util.IsDevMode(envName){
			logDest = zerolog.ConsoleWriter{ Out: logDest }
		}

		appLogger.zLogger = zerolog.New(logDest).With().Caller().Timestamp().Logger().Level(lvl)

	})

	return appLogger
}

// x-nf-reuqest-id 식별자를 추출하기 위한 함수
func (l *AppLogger) WithReqID(ctx *gin.Context) (zerolog.Logger, string) {
	if rID := ctx.Request.Context().Value(util.ContextKey(util.RequestIdentifier)); rID != nil{
		if reqID, ok := rID.(string); ok {
			return l.zLogger.With().Str("req_id", reqID).Logger(), reqID
		}

		return l.zLogger, ""
	}
	return l.zLogger, ""
}

func (l *AppLogger) Fatal() *zerolog.Event {
	return l.zLogger.Fatal()
}

func (l *AppLogger) Error() *zerolog.Event {
	return l.zLogger.Error()
}

func (l *AppLogger) Info() *zerolog.Event {
	return l.zLogger.Info()
}

func (l *AppLogger) Debug() *zerolog.Event {
	return l.zLogger.Debug()
}

func parseLogLevel(level string) zerolog.Level {
	switch strings.ToLower(level){
		case "debug":
			return zerolog.DebugLevel
		case "info":
			return zerolog.InfoLevel
		case "error":
			return zerolog.ErrorLevel
		case "fatal":
			return zerolog.FatalLevel
		default:
			return zerolog.InfoLevel
	}
}
