package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// InterviewCommentLog 是面评的一次历史版本（写入后的全文快照），一版一条，仅落库。
// 看历史按 rev 升序取；看 diff 取相邻两版对比即可，不存 patch。
type InterviewCommentLog struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	FormID       primitive.ObjectID `bson:"formId,omitempty" json:"formId,omitempty"`
	Rev          int64              `bson:"rev,omitempty" json:"rev,omitempty"`
	Comment      string             `bson:"comment" json:"comment"`
	OperatorID   primitive.ObjectID `bson:"operatorId,omitempty" json:"operatorId,omitempty"`
	OperatorType string             `bson:"operatorType,omitempty" json:"operatorType,omitempty"`
	OperatedAt   time.Time          `bson:"operatedAt,omitempty" json:"operatedAt,omitempty"`
}
