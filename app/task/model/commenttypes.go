package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Comment struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserId       primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	SubmissionID primitive.ObjectID `bson:"submission_id,omitempty" json:"submission_id,omitempty"`
	Content      string             `bson:"content,omitempty" json:"content,omitempty"`
	UpdateAt     time.Time          `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt     time.Time          `bson:"createAt,omitempty" json:"createAt,omitempty"`
	FatherId     primitive.ObjectID `bson:"father_id,omitempty" json:"father_id,omitempty"`
}
