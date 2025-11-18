package logic

import (
	"MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/xerr"
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"

	"MuXiFresh-Be-2.0/app/test/api/internal/svc"
	"MuXiFresh-Be-2.0/app/test/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserTestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserTestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserTestLogic {
	return &UserTestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserTestLogic) UserTest(req *types.TestReq) (resp *types.TestResp, err error) {

	userId := ctxData.GetUserIdFromCtx(l.ctx)

	uid, err := primitive.ObjectIDFromHex(userId)

	if err != nil {
		return nil, xerr.ErrExistInvalidId
	}

	score, c := Exam(req.Choice)

	_, err = l.svcCtx.UserInfoClient.Update(l.ctx, &model.UserInfo{
		ID:         uid,
		TestChoice: c,
		TestResult: &model.ExamResult{
			LeQunXing:   score[0],
			YouHengXing: score[4],
			XingFenXing: score[3],
			CongHuiXing: score[1],
			JiaoJiXing:  score[5],
			HuaiYiXing:  score[6],
			WenDingXing: score[2],
		},
		UpdateAt: time.Now()})
	if err != nil {
		return nil, xerr.NewErrCode(xerr.DB_ERROR)
	}

	return &types.TestResp{
		Flag: true,
	}, nil
}

func Exam(choice []types.ChoiceItem) ([7]int64, []model.ChoiceItem) {
	// 初始化各个因素的得分
	factorA := 0
	factorB := 0
	factorC := 0
	factorF := 0
	factorG := 0
	factorH := 0
	factorL := 0

	// 计算各个因素的得分
	for _, c := range choice {
		switch c.Number {
		case 177, 178:
			if c.Data == "A" {
				factorB++
			}
		case 28, 53, 54, 78, 103, 128, 152:
			if c.Data == "B" {
				factorB++
			}
		case 77, 102, 127, 153:
			if c.Data == "C" {
				factorB++
			}
		case 3, 52, 101, 126, 176:
			if c.Data == "A" {
				factorA += 2
			} else if c.Data == "B" {
				factorA++
			}
		case 4, 55, 104, 105, 129, 130, 179:
			if c.Data == "A" {
				factorC += 2
			} else if c.Data == "B" {
				factorC++
			}
		case 33, 58, 132, 133, 182, 183:
			if c.Data == "A" {
				factorF += 2
			} else if c.Data == "B" {
				factorF++
			}
		case 109, 134, 184, 185, 160:
			if c.Data == "A" {
				factorG += 2
			} else if c.Data == "B" {
				factorG++
			}
		case 10, 36, 110, 111, 136, 186:
			if c.Data == "A" {
				factorH += 2
			} else if c.Data == "B" {
				factorH++
			}
		case 13, 38, 88, 113, 114, 164:
			if c.Data == "A" {
				factorL += 2
			} else if c.Data == "B" {
				factorL++
			}
		case 27, 51, 76, 151:
			if c.Data == "C" {
				factorA += 2
			} else if c.Data == "B" {
				factorA++
			}
		case 5, 29, 30, 79, 80, 154:
			if c.Data == "C" {
				factorC += 2
			} else if c.Data == "B" {
				factorC++
			}
		case 8, 82, 83, 107, 108, 157, 158:
			if c.Data == "C" {
				factorF += 2
			} else if c.Data == "B" {
				factorF++
			}
		case 9, 34, 59, 84, 159:
			if c.Data == "C" {
				factorG += 2
			} else if c.Data == "B" {
				factorG++
			}
		case 35, 60, 61, 85, 86, 135, 161:
			if c.Data == "C" {
				factorH += 2
			} else if c.Data == "B" {
				factorH++
			}
		case 63, 64, 89, 139:
			if c.Data == "C" {
				factorL += 2
			} else if c.Data == "B" {
				factorL++
			}
		}
	}

	var sc [7]int64
	sc[0] = int64(factorA)
	sc[1] = int64(factorB)
	sc[2] = int64(factorC)
	sc[3] = int64(factorF)
	sc[4] = int64(factorG)
	sc[5] = int64(factorH)
	sc[6] = int64(factorL)

	// 创建一个空的 []model.ChoiceItem
	var modelItems []model.ChoiceItem

	// 遍历 typesItems，并将每个元素转换为 model.ChoiceItem，并添加到 modelItems 中
	for _, typesItem := range choice {
		modelItem := model.ChoiceItem(typesItem) // 执行类型转换
		modelItems = append(modelItems, modelItem)
	}

	return sc, modelItems
}
