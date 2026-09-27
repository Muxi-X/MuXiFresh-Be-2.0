package logic

import (
	"context"
	"errors"

	"MuXiFresh-Be-2.0/app/form/api/internal/svc"
	"MuXiFresh-Be-2.0/app/form/api/internal/types"
	"MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SetRecruitDeadlineLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetRecruitDeadlineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetRecruitDeadlineLogic {
	return &SetRecruitDeadlineLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// SetRecruitDeadline 以乐观锁写入指定届次的报名截止时间，仅管理员可写。
// rev 为乐观锁版本号：首次写入传 0，之后回传上次返回的 rev；冲突时提示刷新重试。
func (l *SetRecruitDeadlineLogic) SetRecruitDeadline(req *types.SetRecruitDeadlineReq) (resp *types.SetRecruitDeadlineResp, err error) {
	userId := ctxData.GetUserIdFromCtx(l.ctx)
	if userId == "" {
		return nil, errors.New("身份缺失")
	}
	userInfo, err := l.svcCtx.UserInfoModelClient.FindOne(l.ctx, userId)
	if err != nil {
		return nil, err
	}
	if userInfo.UserType != globalKey.Admin && userInfo.UserType != globalKey.SuperAdmin {
		return nil, errors.New("permission denied")
	}

	if _, _, ok := model.ParseCycle(req.Cycle); !ok {
		return nil, errors.New("invalid cycle")
	}
	if req.Rev < 0 {
		return nil, errors.New("invalid rev")
	}
	deadline, err := model.ParseDeadline(req.Deadline)
	if err != nil {
		return nil, errors.New("invalid deadline")
	}
	operatorID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, errors.New("非法的用户身份")
	}

	newRev, err := l.svcCtx.RecruitSettingModel.SetDeadline(l.ctx, req.Cycle, deadline, req.Rev, operatorID, userInfo.UserType)
	if err != nil {
		if errors.Is(err, model.ErrVersionConflict) {
			return nil, errors.New("deadline has been modified, please refresh")
		}
		return nil, err
	}

	return &types.SetRecruitDeadlineResp{
		Cycle:    req.Cycle,
		Deadline: model.FormatDeadline(deadline),
		Rev:      newRev,
	}, nil
}
