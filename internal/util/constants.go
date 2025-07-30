package util

import "time"

type ContextKey string

const RequestIdentifier = "X-Request-ID"

// 서비스 레이어에서 사용할 기본 타임아웃을 상수로 정의합니다.
const DefaultServiceTimeout = 5 * time.Second