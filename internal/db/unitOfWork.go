package db

import (
	"database/sql"
	"go-rest-example/internal/logger"
)

/*
  @breif  데이터레이어 흐름과 제어를 관리
  @mathod Execute  리포지토리를 생성하고 반환하는 메서드
  @mathod Device   트랙젠션 내에 비즈니스 로직을 실행
*/
type IUnitOfWork interface {
	Device() DevicesDataService
	Report() ReportsDataService
	Execute(fn func(IUnitOfWork)error)error
	Cleanup()
}

/*
  @breif IUnitOfWork를 상속 받은 구현체
*/
type unitOfWork struct {
	db *sql.DB
	tx *sql.Tx
	logger *logger.AppLogger
}

/*
  @breif  객체 생성 위임 받은 팩토리 함수 
  @param  db          쿼리 실행기
  @return IUnitOfWork 
*/
func NewUnitOfWork(db *sql.DB) IUnitOfWork {
	return &unitOfWork{db: db}
}

/*
  @breif 트랜젝션 여부에 따라 repository 레이어 DB 객체 상태 변경
  @return DevicesDataService 
  @return error
*/
func (u *unitOfWork) Device() DevicesDataService {
	// 트랙젠션이 활성화되지 않은 경우 일반 DB 객체 반환 
	var executor DBTX = u.db
	if u.tx != nil {
		executor = u.tx
	}
	return NewDevicesRepo(u.logger, executor)
}

/*
  @breif 트랜젝션 여부에 따라 repository 레이어 DB 객체 상태 변경
  @return ReportsDataService 
  @return error
*/
func (u *unitOfWork) Report() ReportsDataService {
	// 트랙젠션이 활성화되지 않은 경우 일반 DB 객체 반환 
	var executor DBTX = u.db
	if u.tx != nil {
		executor = u.tx
	}
	return NewReportsRepo(u.logger, executor)
}

/*
  @breif 트랜젝션 작업 흐름을 관리하는 함수
         생성과 업데이트와 같은 yt
  @param fn    고차함수 
*/
func (u *unitOfWork) Execute(fn func(IUnitOfWork)error)error{

	// 1. 트랜젝션 실행
	tx, err := u.db.Begin()
	if err != nil {
		return err
	}

	// 구조체 필드 할당
	u.tx = tx

	// 2. 고차함수를 통해 내용 실행
	err = fn(u)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			u.logger.Error().Err(err).Msg("커스텀 에러")
			return rbErr
		}
		return err		 
	}

	return tx.Commit()
}

func (u *unitOfWork) Cleanup() {
	u.tx = nil
}

