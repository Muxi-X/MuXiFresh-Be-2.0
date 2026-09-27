package logic

import (
	"context"
	"errors"

	"MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/userclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	defaultHistoryLimit = 50
	maxHistoryLimit     = 200
)

type GetInterviewCommentHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetInterviewCommentHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInterviewCommentHistoryLogic {
	return &GetInterviewCommentHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetInterviewCommentHistory 按 rev 升序返回某报名表的面评历史版本（仅管理员）。
// 后端不算 diff，返回每版全文，前端自行对比。
func (l *GetInterviewCommentHistoryLogic) GetInterviewCommentHistory(req *types.GetInterviewCommentHistoryReq) (resp *types.GetInterviewCommentHistoryResp, err error) {
	userId := ctxData.GetUserIdFromCtx(l.ctx)
	getUserTypeResp, err := l.svcCtx.UserClient.GetUserType(l.ctx, &userclient.GetUserTypeReq{UserId: userId})
	if err != nil {
		return nil, err
	}
	if getUserTypeResp.UserType != globalKey.Admin && getUserTypeResp.UserType != globalKey.SuperAdmin {
		return nil, errors.New("permission denied")
	}

	if _, err := primitive.ObjectIDFromHex(req.FormID); err != nil {
		return nil, errors.New("invalid form id")
	}

	logs, err := l.svcCtx.InterviewCommentLogModel.ListByFormID(l.ctx, req.FormID, clampHistoryLimit(req.Limit))
	if err != nil {
		return nil, err
	}

	names, err := resolveUserNames(l.ctx, l.svcCtx, operatorIDs(logs))
	if err != nil {
		return nil, err
	}

	rows := make([]types.InterviewCommentVersion, 0, len(logs))
	for _, entry := range logs {
		rows = append(rows, types.InterviewCommentVersion{
			Rev:          entry.Rev,
			Comment:      entry.Comment,
			OperatorName: names[entry.OperatorID.Hex()],
			OperatorType: entry.OperatorType,
			OperatedAt:   formatAuditTime(entry.OperatedAt),
		})
	}
	return &types.GetInterviewCommentHistoryResp{Rows: rows}, nil
}

// clampHistoryLimit 把请求上限归一到 [1, maxHistoryLimit]，非正数取默认值。
func clampHistoryLimit(limit int64) int64 {
	if limit <= 0 {
		return defaultHistoryLimit
	}
	if limit > maxHistoryLimit {
		return maxHistoryLimit
	}
	return limit
}

// operatorIDs 收集历史里去重后的修改人 id。
func operatorIDs(logs []*model.InterviewCommentLog) []string {
	ids := make([]string, 0, len(logs))
	for _, entry := range logs {
		if !entry.OperatorID.IsZero() {
			ids = append(ids, entry.OperatorID.Hex())
		}
	}
	return dedupeStrings(ids)
}

// resolveUserNames 按 id 批量解析用户名；查不到的 id 不进入结果。
func resolveUserNames(ctx context.Context, svcCtx *svc.ServiceContext, ids []string) (map[string]string, error) {
	names := make(map[string]string, len(ids))
	if len(ids) == 0 {
		return names, nil
	}

	userInfos, err := svcCtx.UserInfoModel.FindByUserIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, info := range userInfos {
		names[info.ID.Hex()] = info.Name
	}
	return names, nil
}
