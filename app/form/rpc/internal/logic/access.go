package logic

import (
	"context"
	"errors"

	"MuXiFresh-Be-2.0/app/form/rpc/internal/svc"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/metadata"
)

// callerIDFromCtx 从 grpc metadata 提取调用者身份；缺失/为空/非合法 ObjectID 一律拒绝（fail closed）。
// 调用者身份由 API 层从 JWT 上下文注入，RPC 直连不可信；根治依赖 RPC 层统一鉴权中间件。
func callerIDFromCtx(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("缺少用户身份")
	}
	vals := md.Get(ctxData.CallerIDKey)
	if len(vals) == 0 || vals[0] == "" {
		return "", errors.New("缺少用户身份")
	}
	if _, err := primitive.ObjectIDFromHex(vals[0]); err != nil {
		return "", errors.New("非法的用户身份")
	}
	return vals[0], nil
}

func isAdminType(ctx context.Context, svcCtx *svc.ServiceContext, callerID string) (bool, error) {
	caller, err := svcCtx.UserInfoModel.FindOne(ctx, callerID)
	if err != nil {
		if errors.Is(err, mon.ErrNotFound) {
			return false, errors.New("无权查看该报名表")
		}
		return false, err
	}
	if caller == nil {
		return false, errors.New("无权查看该报名表")
	}
	return caller.UserType == globalKey.Admin || caller.UserType == globalKey.SuperAdmin, nil
}

// checkEntryFormReadAccess 校验调用者能否读取指定报名表：本人或 Admin/SuperAdmin。
// 报名表不存在与无权访问归一为同一错误，不泄露存在性。
func checkEntryFormReadAccess(ctx context.Context, svcCtx *svc.ServiceContext, callerID, entryFormID string) error {
	form, err := svcCtx.FormClient.FindOne(ctx, entryFormID)
	if err != nil {
		if errors.Is(err, mon.ErrNotFound) {
			return errors.New("无权查看该报名表")
		}
		return err
	}
	if form.UserId.Hex() == callerID {
		return nil
	}
	admin, err := isAdminType(ctx, svcCtx, callerID)
	if err != nil {
		return err
	}
	if !admin {
		return errors.New("无权查看该报名表")
	}
	return nil
}

// checkEntryFormWriteAccess 校验调用者能否修改指定报名表：仅本人。
// （管理员审阅不需要改学生报名表，避免越权写入。）
func checkEntryFormWriteAccess(ctx context.Context, svcCtx *svc.ServiceContext, callerID, entryFormID string) error {
	form, err := svcCtx.FormClient.FindOne(ctx, entryFormID)
	if err != nil {
		if errors.Is(err, mon.ErrNotFound) {
			return errors.New("无权修改该报名表")
		}
		return err
	}
	if form.UserId.Hex() != callerID {
		return errors.New("无权修改该报名表")
	}
	return nil
}
