package logic

import (
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/assignment/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/assignment/pb"
	"MuXiFresh-Be-2.0/app/task/model"
	"MuXiFresh-Be-2.0/common/globalKey"
	"MuXiFresh-Be-2.0/common/xerr"
	"context"
	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetAssignmentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetAssignmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetAssignmentLogic {
	return &SetAssignmentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetAssignmentLogic) SetAssignment(in *pb.SetAssignmentReq) (*pb.SetAssignmentResp, error) {

	assignment := model.Assignment{}
	copier.Copy(&assignment, in)

	var err error

	if in.AssignmentID == globalKey.NULL {
		//布置
		err = l.svcCtx.AssignmentModelClient.Insert(l.ctx, &assignment)
	} else {
		//更新
		assignment.ID, err = primitive.ObjectIDFromHex(in.AssignmentID)
		if err != nil {
			return nil, xerr.ErrExistInvalidId.Status()
		}

		_, err = l.svcCtx.AssignmentModelClient.Update(l.ctx, &assignment)
	}

	if err != nil {
		return nil, xerr.NewErrCode(xerr.DB_ERROR).Status()
	}

	return &pb.SetAssignmentResp{
		Flag: true,
	}, nil
}
