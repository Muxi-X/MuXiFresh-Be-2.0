package logic

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/pb"
	"MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/tool"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetEmailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetEmailLogic {
	return &SetEmailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetEmailLogic) SetEmail(in *pb.SetEmailReq) (*pb.SetEmailResp, error) {

	uid, err := primitive.ObjectIDFromHex(in.UserId)
	if err != nil {
		return nil, err
	}
	email := tool.NormalizeEmail(in.Email)

	// 占用检查：目标邮箱若已属于其他账号则拒绝（唯一索引作为并发兜底）。
	existing, err := l.svcCtx.UserInfoClient.FindByEmail(l.ctx, email)
	if err == nil && !existing.ID.IsZero() && existing.ID != uid {
		return nil, errors.New("邮箱已被占用")
	}
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}

	old, err := l.svcCtx.UserAuthClient.FindOneByUserId(l.ctx, uid)
	if err != nil {
		return nil, err
	}

	if _, err = l.svcCtx.UserAuthClient.UpdateByUserId(l.ctx, &model.UserAuth{
		Email:      email,
		UserInfoID: uid,
	}); err != nil {
		if model.IsDuplicateKey(err) {
			return nil, errors.New("邮箱已被占用")
		}
		return nil, err
	}

	if _, err = l.svcCtx.UserInfoClient.Update(l.ctx, &model.UserInfo{
		ID:    uid,
		Email: email,
	}); err != nil {
		// 补偿：userinfo 写入失败时把 userauth 的邮箱回滚为旧值，避免两集合不一致。
		if _, rbErr := l.svcCtx.UserAuthClient.UpdateByUserId(l.ctx, &model.UserAuth{
			Email:      old.Email,
			UserInfoID: uid,
		}); rbErr != nil {
			logx.WithContext(l.ctx).Errorf("setemail rollback userauth %s failed: %v", uid.Hex(), rbErr)
		}
		return nil, err
	}
	return &pb.SetEmailResp{
		Flag: true,
	}, nil
}
