package server

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"go-rest-example/internal/db"
	"go-rest-example/internal/handlers"
	"go-rest-example/internal/logger"
	"go-rest-example/internal/middleware"
	"go-rest-example/internal/model"
	"go-rest-example/internal/service"
	"go-rest-example/internal/util"
)

// 서버 시작 시 한번만 동작하는 것을 보장하기 위해 사용
var startOnce sync.Once


func Start(ctx context.Context, svcEnv *model.ServiceEnv, lgr *logger.AppLogger, dbMgr db.DBManager) error {
	// 1. Gin 라우터 설정 (이전과 동일)
	router, err := WebRouter(svcEnv, lgr, dbMgr)
	if err != nil {
		return err
	}
	lgr.Info().Msg("Registered routes")
	for _, item := range router.Routes() {
		lgr.Info().Str("method", item.Method).Str("path", item.Path).Send()
	}

	// 2. 표준 http.Server 생성 및 설정
	srv := &http.Server{
		Addr:    ":" + svcEnv.Port,
		Handler: router,
	}

	// 3. 별도의 고루틴에서 서버 종료 신호를 감지
	go func() {
		// main의 context가 Done 되면(종료 신호 수신 시) Shutdown을 호출
		<-ctx.Done()
		lgr.Info().Msg("Shutdown signal received, starting graceful shutdown")

		// Shutdown을 위한 별도의 타임아웃 컨텍스트 생성
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// http.Server의 Shutdown 메서드를 호출하여 우아한 종료 시작
		if err := srv.Shutdown(shutdownCtx); err != nil {
			lgr.Error().Err(err).Msg("Server Shutdown Failed")
		}
	}()

	// 4. 서버 시작. 이 호출은 서버가 종료될 때까지 블로킹됩니다.
	lgr.Info().Str("address", srv.Addr).Msg("Starting server")
	err = srv.ListenAndServe()

	// 5. ListenAndServe는 정상적인 Shutdown 후에는 http.ErrServerClosed 에러를 반환합니다.
	// 이 경우는 실제 에러가 아니므로 nil을 반환하여 정상 종료를 알립니다.
	if errors.Is(err, http.ErrServerClosed) {
		lgr.Info().Msg("Server closed gracefully")
		return nil
	}

	return err
}


// 경로 정보를 지정하고, 의존성을 주입하는 역할을 수행한다.
func WebRouter(svcEnv *model.ServiceEnv, lgr *logger.AppLogger, dbMgr db.DBManager) (*gin.Engine, error ){


	// 1. 환경 변수에 따라서 콘솔에 변화를 준다
	ginMode := gin.ReleaseMode
	if util.IsDevMode(svcEnv.Name){
		ginMode = gin.DebugMode
		gin.ForceConsoleColor()
	}

	gin.SetMode(ginMode)
	gin.EnableJsonDecoderDisallowUnknownFields()

	// 2. 미들웨어 등록
	gin.DefaultWriter = io.Discard
	router := gin.New();
	
	router.Use(gin.Recovery())

	router.Use(middleware.ErrorHandler(lgr))

	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(middleware.ReqIDMiddleware())
	router.Use(middleware.ResponseHeadersMiddleware())
	router.Use(middleware.RequestLogMiddleware(lgr))


	// Health Check 도메인
	status, sHandlerErr := handlers.NewStatusHandler(lgr, dbMgr)
	if sHandlerErr != nil {
		return nil, sHandlerErr
	}
	router.GET("/healthz", status.CheckStatus)

	// 성능 모니터링
	// internalAPIGrp := router.Group("/internal")
	// internalAPIGrp.Use(middleware.InternalAuthMiddleware()) // use special auth middleware to handle internal employees
	// 프로메테우스와의 연동 계획
	// pprof.RouteRegister(internalAPIGrp, "pprof")
	// router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 0. 데이터 레이어 획득 
	d := dbMgr.DB()
	uow := db.NewUnitOfWork(d, lgr)
	if uow == nil {
		return nil, errors.New("Server uow initialize faile")
	}

	// 0. 서비스 레이어 획득
	rpService := service.NewUserService(uow)
	if rpService == nil {
		return nil, errors.New("Server rpService initialize faile")
	}

	dvService := service.NewDeviceService(uow)
	if dvService == nil {
		return nil, errors.New("Server dvService initialize faile")
	}

	deviceHandler, deviceHandlerErr := handlers.NewDevicesHandler(lgr, dvService)
	if deviceHandlerErr != nil {
		return nil, deviceHandlerErr
	}
	
	// 0. 의존성 주입 및 라우터 등록 
	deviceAPIGrp := router.Group("/device")
	deviceAPIGrp.Use(middleware.AuthMiddleware())  // 디바이스 인증 과정을 담당하는 미들웨어
	deviceAPIGrp.POST("",deviceHandler.Create)
	deviceAPIGrp.GET("",deviceHandler.GetAll)
	deviceAPIGrp.GET("/:ID",deviceHandler.GetByID)

	// repot API 등록 
	reportHandler, reportHandlerErr := handlers.NewReportsHandler(lgr, rpService)
	if reportHandlerErr != nil {
		return nil, reportHandlerErr
	}

	reportAPIGrp := router.Group("/report")
	reportAPIGrp.POST("",reportHandler.Report)
	reportAPIGrp.PATCH("",reportHandler.Update)

	// 4. 라우터 객체 반환
	return router, nil
}
