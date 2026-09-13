package logic

import (
	formmodel "MuXiFresh-Be-2.0/app/form/model"
	"context"
	"errors"

	"MuXiFresh-Be-2.0/app/schedule/rpc/internal/svc"
	"MuXiFresh-Be-2.0/app/schedule/rpc/pb"
	"MuXiFresh-Be-2.0/common/ctxData"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/metadata"
)

const (
	ctxCallerIDKey = ctxData.CallerIDKey
)

// callerIDFromCtx 从 grpc metadata 提取调用者身份；缺失/为空/非合法 ObjectID 一律拒绝（fail closed）。
func callerIDFromCtx(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("缺少用户身份")
	}
	vals := md.Get(ctxCallerIDKey)
	if len(vals) == 0 || vals[0] == "" {
		return "", errors.New("缺少用户身份")
	}
	if _, err := primitive.ObjectIDFromHex(vals[0]); err != nil {
		return "", errors.New("非法的用户身份")
	}
	return vals[0], nil
}

type CheckLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckLogic {
	return &CheckLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CheckLogic) Check(in *pb.CheckReq) (*pb.CheckResp, error) {
	// 归属校验：调用者身份由 API 层经 grpc metadata 注入，RPC 直连不可信（fail closed）
	callerID, err := callerIDFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	if callerID != in.UserId {
		return nil, errors.New("无权查看该进度")
	}

	f := &formmodel.EntryForm{}
	f, err = l.svcCtx.EntryFormClient.FindOneByUserId(l.ctx, in.UserId)
	if err != nil && err != formmodel.ErrNotFound {
		return nil, err
	}
	if f == nil {
		// 未交表用户：FindOneByUserId 返回 nil, ErrNotFound，
		// 这里兜底用空表单，避免后续 f.Major/f.Group nil 解引用导致进程崩溃。
		f = &formmodel.EntryForm{}
	}

	s, err := l.svcCtx.ScheduleClient.FindOne(l.ctx, in.ScheduleID)
	if err != nil {
		return nil, err
	}
	// schedule 归属校验：进度必须属于调用者本人，防止直连用他人 schedule_id 读取录取状态
	if s.UserID.Hex() != callerID {
		return nil, errors.New("无权查看该进度")
	}

	userInfo, err := l.svcCtx.UserInfoClient.FindOne(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.CheckResp{
		Name:            userInfo.Name,
		School:          userInfo.School,
		Major:           f.Major,
		Group:           f.Group,
		EntryFormStatus: s.EntryFormStatus,
		AdmissionStatus: s.AdmissionStatus,
	}, nil
}
