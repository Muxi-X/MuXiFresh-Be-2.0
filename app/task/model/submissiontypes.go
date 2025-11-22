package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Submission struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserId       primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	AssignmentID primitive.ObjectID `bson:"assignment_id,omitempty" json:"assignment_id,omitempty"`
	Urls         []string           `bson:"urls,omitempty" json:"urls,omitempty"`
	Status       string             `bson:"status,omitempty" json:"status,omitempty"`
	Version      int64              `bson:"version,omitempty" json:"version,omitempty"`
	UpdateAt     time.Time          `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt     time.Time          `bson:"createAt,omitempty" json:"createAt,omitempty"`
}

// 按userID分组接收响应的结果
type SubmissionStats struct {
	UserId     primitive.ObjectID `bson:"_id,omitempty" json:"user_id,omitempty"`
	Status     string             `bson:"status,omitempty" json:"status,omitempty"`
	VersionNum int64              `bson:"version_num,omitempty" json:"version_num,omitempty"`
}
