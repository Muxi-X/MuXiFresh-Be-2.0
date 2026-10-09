package model

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrNotFound        = mon.ErrNotFound
	ErrInvalidObjectId = errors.New("invalid objectId")
)

// IsDuplicateKey 报告错误是否为唯一索引冲突（E11000）。
func IsDuplicateKey(err error) bool {
	return mongo.IsDuplicateKeyError(err)
}
