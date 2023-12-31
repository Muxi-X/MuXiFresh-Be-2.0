package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Intro struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	URL      string             `bson:"url,omitempty" json:"url"`
	UpdateAt time.Time          `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt time.Time          `bson:"createAt,omitempty" json:"createAt,omitempty"`
}
