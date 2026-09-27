package model

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/mon"
)

var (
	ErrNotFound        = mon.ErrNotFound
	ErrInvalidObjectId = errors.New("invalid objectId")
	// ErrVersionConflict 表示报名配置的乐观锁版本不匹配（含非首写却不存在）。
	ErrVersionConflict = errors.New("recruit setting version conflict")
)
