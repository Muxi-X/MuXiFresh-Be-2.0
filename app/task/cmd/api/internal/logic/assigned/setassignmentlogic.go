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

type SetAssignmentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetAssignmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetAssignmentLogic {
	return &SetAssignmentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetAssignmentLogic) SetAssignment(req *types.SetAssignmentReq) (resp *types.SetAssignmentResp, err error) {

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
	//布置
	setAssignmentResp, err := l.svcCtx.AssignmentClient.SetAssignment(l.ctx, &assignmentclient.SetAssignmentReq{
		AssignmentID: req.AssignmentID,
		Group:        req.Group,
		TitleText:    req.TitleText,
		Content:      req.Content,
		Urls:         req.Urls,
		Deadline:     req.Deadline,
		Semester:     req.Semester,
		Year:         int64(req.Year),
	})
	if err != nil {
		return nil, err
	}
	return &types.SetAssignmentResp{Flag: setAssignmentResp.Flag}, nil
}
