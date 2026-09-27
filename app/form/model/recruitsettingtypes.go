package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RecruitSetting 是某个届次（YYYYautumn / YYYYspring）的报名配置，目前只有截止时间。
// 每届一条，cycle 唯一。Rev 是并发写入的乐观锁版本号，每次成功写入 +1。
// History 内嵌该届次历次变更，随同一次 CAS 写入原子落库，避免"值已改、记录未写"的缺口。
type RecruitSetting struct {
	ID       primitive.ObjectID     `bson:"_id,omitempty" json:"id,omitempty"`
	Cycle    string                 `bson:"cycle,omitempty" json:"cycle,omitempty"`
	Deadline time.Time              `bson:"deadline,omitempty" json:"deadline,omitempty"`
	Rev      int64                  `bson:"rev,omitempty" json:"rev,omitempty"`
	UpdateBy primitive.ObjectID     `bson:"updateBy,omitempty" json:"updateBy,omitempty"`
	UpdateAt time.Time              `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt time.Time              `bson:"createAt,omitempty" json:"createAt,omitempty"`
	History  []RecruitSettingChange `bson:"history,omitempty" json:"history,omitempty"`
}

// RecruitSettingChange 记录一次截止时间变更。
// OldDeadline 为写入前的值；首次写入（此前无记录）时为 nil。
type RecruitSettingChange struct {
	OldDeadline  *time.Time         `bson:"oldDeadline,omitempty" json:"oldDeadline,omitempty"`
	NewDeadline  time.Time          `bson:"newDeadline,omitempty" json:"newDeadline,omitempty"`
	Rev          int64              `bson:"rev,omitempty" json:"rev,omitempty"`
	OperatorID   primitive.ObjectID `bson:"operatorId,omitempty" json:"operatorId,omitempty"`
	OperatorType string             `bson:"operatorType,omitempty" json:"operatorType,omitempty"`
	OperatedAt   time.Time          `bson:"operatedAt,omitempty" json:"operatedAt,omitempty"`
}
