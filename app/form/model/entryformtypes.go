package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EntryForm struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserId primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	// Cycle 标识报名所属届次（如 2026autumn）。同一用户每个届次至多一份表，
	// 往届表在重报时保留，见 CycleOf / FindOneByUserId。
	Cycle         string `bson:"cycle,omitempty" json:"cycle,omitempty"`
	Avatar        string `bson:"avatar,omitempty" json:"avatar,omitempty"`
	Major         string `bson:"major,omitempty" json:"major,omitempty"`
	Grade         string `bson:"grade,omitempty" json:"grade,omitempty"`
	Gender        string `bson:"gender,omitempty" json:"gender,omitempty"`
	Phone         string `bson:"phone,omitempty" json:"phone,omitempty"`
	Group         string `bson:"group,omitempty" json:"group,omitempty"`
	Reason        string `bson:"reason,omitempty" json:"reason,omitempty"`
	Knowledge     string `bson:"knowledge,omitempty" json:"knowledge,omitempty"`
	SelfIntro     string `bson:"selfIntro,omitempty" json:"selfIntro,omitempty"`
	ExtraQuestion string `bson:"extraQuestion,omitempty" json:"extraQuestion,omitempty"`
	// InterviewComment 是管理员填写的面评 markdown 正文。仅审阅侧读写，
	// 不进入候选人可见的报名表响应（见 form/rpc CheckForm）。
	InterviewComment string `bson:"interviewComment,omitempty" json:"interviewComment,omitempty"`
	// InterviewCommentRev 是面评的版本号，用作并发写入的乐观锁：每次写入 +1，
	// 缺省/0 表示从未填写。
	InterviewCommentRev int64     `bson:"interviewCommentRev,omitempty" json:"interviewCommentRev,omitempty"`
	UpdateAt            time.Time `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt            time.Time `bson:"createAt,omitempty" json:"createAt,omitempty"`
}
