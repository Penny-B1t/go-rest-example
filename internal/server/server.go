package server

import (
	"io"
	"sync"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"go-rest-example/internal/db"
	"go-rest-example/internal/logger"
	"go-rest-example/internal/middleware"
	"go-rest-example/internal/model"
	"go-rest-example/internal/util"
)

// 서버 시작 시 한번만 동작하는 것을 보장하기 위해 사용
var startOnce sync.Once


func Start(svcEnv *model.ServiceEnv, lgr *logger.AppLogger, dbMgr db.MariaDBManager) error {

	var err error
	var r *gin.Engine

	// 초기화 로직을 한번만 실행하기 위해 사용
	startOnce.Do(func() {
		r, err = WebRouter(svcEnv, lgr, dbMgr)
		lgr.Info().Msg("Registered routes")
		for _, item := range r.Routes() {
			lgr.Info().Str("method", item.Method).Str("path", item.Path).Send()
		}
		if err != nil {
			return
		}
		err = r.Run(":" + svcEnv.Port)
	})

	return err
}


// 경로 정보를 지정하고, 의존성을 주입하는 역할을 수행한다.
func WebRouter(svcEnv *model.ServiceEnv, lgr *logger.AppLogg
	er, dbMgr db.MariaDBManager) (*gin.Engine, error ){


	// 1. 환경 변수에 따라서 콘솔에 변화를 준다
	ginMode := gin.ReleaseMode
	if util.IsDevMode(svcEnv.Name){
		ginMode = gin.DebugMode
		gin.ForceConsoleColor()
	}



	gin.SetMode(ginMode)
	gin.EnableJsonDecoderDisallowUnknownFields()

	// 2. 미들웨어 등록
	// 기본적인 서비스 구성 미들웨어 지정
	gin.DefaultWriter = io.Discard
	router := gin.New();
	
	router.Use(gin.Recovery())
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(middleware.ReqIDMiddleware())
	router.Use(middleware.ResponseHeadersMiddleware())
	router.Use(middleware.RequestLogMiddleware(lgr))


	// 3. 라우터 등록

	// 4. 라우터 객체 반환
	return nil, nil
}
