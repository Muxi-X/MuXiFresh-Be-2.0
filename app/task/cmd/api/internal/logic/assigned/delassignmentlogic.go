package assigned

import (
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/assignment/assignmentclient"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/userclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"
	"context"
	"errors"

	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DelAssignmentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除指定作业(题目)
func NewDelAssignmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DelAssignmentLogic {
	return &DelAssignmentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DelAssignmentLogic) DelAssignment(req *types.DelAssignmentReq) (resp *types.DelAssignmentResp, err error) {
	//管理员身份认证
	getUserTypeResp, err := l.svcCtx.UserClient.GetUserType(l.ctx, &userclient.GetUserTypeReq{
		UserId: ctxData.GetUserIdFromCtx(l.ctx),
	})
	if err != nil {
		return nil, err
	}
	if getUserTypeResp.UserType != globalKey.Admin && getUserTypeResp.UserType != globalKey.SuperAdmin {
		return nil, errors.New("permission denied")
	}
	delAssignment, err := l.svcCtx.AssignmentClient.DelAssignment(l.ctx, &assignmentclient.DelAssignmentReq{
		AssignmentID: req.AssignmentID,
	})
	return &types.DelAssignmentResp{
		Flag: delAssignment.Flag,
	}, nil
}
