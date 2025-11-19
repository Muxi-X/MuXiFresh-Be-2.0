package logic

import (
	"MuXiFresh-Be-2.0/app/test/rpc/testclient"
	"MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/ctxData"
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"strings"
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
		return nil, err
	}

	score, c := Exam(req.Choice)

	_, err = l.svcCtx.UserInfoClient.Update(l.ctx, &model.UserInfo{
		ID:         uid,
		TestChoice: c,
		TestResult: struct {
			LeQunXing   int64
			YouHengXing int64
			XingFenXing int64
			CongHuiXing int64
			JiaoJiXing  int64
			HuaiYiXing  int64
			WenDingXing int64
		}{
			score[0],
			score[4],
			score[3],
			score[1],
			score[5],
			score[6],
			score[2],
		},
		UpdateAt: time.Now()})
	if err != nil {
		return nil, err
	}

	testResp, err := l.svcCtx.TestClient.UserTest(l.ctx, &testclient.TestReq{
		Token: req.Authorization,
	})
	if err != nil {
		return nil, err
	}
	return &types.TestResp{
		Flag: testResp.Flag,
	}, nil
}

func Exam(choice []types.ChoiceItem) ([7]int64, []model.ChoiceItem) {
	type Question struct {
		Number     int
		TextNumber int
		Type       string
	}
	str := "1 2 3 4 5 8 9 10 13 27 28 29 30 33 34 35 36 38 51 52 53 54 55 58 59 60 61 63 64 76 77 78 79 80 82 83 84 85 86 88 89 101 102 103 104 105 107 108 109 110 111 113 114 126 127 128 129 130 132 133 134 135 136 139 151 152 153 154 156 157 158 159 160 161 164 176 177 178 179 182 183 184 185 186 187"
	arr := strings.Split(str, " ")
	var objectArr []Question
	for index, item := range arr {
		textNumber, _ := strconv.Atoi(item)
		objectArr = append(objectArr, Question{
			Number:     index + 1,
			TextNumber: textNumber,
			Type:       "",
		})
	}

	A := []int{3, 27, 51, 52, 76, 101, 126, 151, 176}
	B := []int{28, 53, 54, 77, 78, 102, 103, 127, 128, 152, 153, 177, 178}
	C := []int{4, 5, 29, 30, 55, 79, 80, 104, 105, 129, 130, 154, 179}
	F := []int{8, 33, 58, 82, 83, 107, 108, 132, 133, 157, 158, 182, 183}
	G := []int{9, 34, 59, 84, 109, 134, 159, 160, 184, 185}
	H := []int{10, 35, 36, 60, 61, 85, 86, 110, 111, 135, 136, 161, 186}
	L := []int{13, 38, 63, 64, 88, 89, 113, 114, 139, 164}

	for index := range objectArr {
		if contains(A, objectArr[index].TextNumber) {
			objectArr[index].Type = "A"
		}
		if contains(B, objectArr[index].TextNumber) {
			objectArr[index].Type = "B"
		}
		if contains(C, objectArr[index].TextNumber) {
			objectArr[index].Type = "C"
		}
		if contains(F, objectArr[index].TextNumber) {
			objectArr[index].Type = "F"
		}
		if contains(G, objectArr[index].TextNumber) {
			objectArr[index].Type = "G"
		}
		if contains(H, objectArr[index].TextNumber) {
			objectArr[index].Type = "H"
		}
		if contains(L, objectArr[index].TextNumber) {
			objectArr[index].Type = "L"
		}
	}

	score := [7]int64{}
	answerTrueB := []struct {
		TextNumber int
		TrueAnswer string
	}{
		{TextNumber: 28, TrueAnswer: "B"},
		{TextNumber: 53, TrueAnswer: "B"},
		{TextNumber: 54, TrueAnswer: "B"},
		{TextNumber: 77, TrueAnswer: "C"},
		{TextNumber: 78, TrueAnswer: "B"},
		{TextNumber: 102, TrueAnswer: "C"},
		{TextNumber: 103, TrueAnswer: "B"},
		{TextNumber: 127, TrueAnswer: "C"},
		{TextNumber: 128, TrueAnswer: "B"},
		{TextNumber: 152, TrueAnswer: "B"},
		{TextNumber: 153, TrueAnswer: "C"},
		{TextNumber: 177, TrueAnswer: "A"},
		{TextNumber: 178, TrueAnswer: "A"},
	}

	for index, answer := range choice {
		for _, item := range answerTrueB {
			if item.TextNumber == objectArr[index].TextNumber {
				if answer.Data == item.TrueAnswer {
					score[1]++
				}
			}
		}
	}

	str1 := "3.A 4.A 5.C 8.C 9.C 10.A 13.A 27.C 29.C 30.C 33.A 34.C 35.C 36.A 38.A 51.C 52.A 55.A 58.A 59.C 60.C 61.C 62.C 63.C 64.C 76.C 79.C 80.C 82.C 83.C 84.C 85.C 86.C 88.A 89.C 101.A 104.A 105.A 107.C 108.C 109.A 110.A 111.A 113.A 114.A 126.A 129.A 130.A 131.A 132.A 133.A 134.A 135.C 136.A 139.C 151.C 154.C 155.A 156.A 157.C 158.C 159.C 160.A 161.C 164.A 176.A 179.A 182.A 183.A 184.A 185.A 186.A"
	arr1 := strings.Split(str1, " ")
	var answerTrue []struct {
		TextNumber int
		TrueAnswer string
	}
	for _, item := range arr1 {
		temp := strings.Split(item, ".")
		textNumber, _ := strconv.Atoi(temp[0])
		answerTrue = append(answerTrue, struct {
			TextNumber int
			TrueAnswer string
		}{TextNumber: textNumber, TrueAnswer: temp[1]})
	}

	for index, answer := range choice {
		for _, item := range answerTrue {
			if item.TextNumber == objectArr[index].TextNumber {
				if answer.Data == "B" {
					switch objectArr[index].Type {
					case "A":
						score[0]++
					case "B":
						score[1]++
					case "C":
						score[2]++
					case "F":
						score[3]++
					case "G":
						score[4]++
					case "H":
						score[5]++
					case "L":
						score[6]++
					default:
					}
				} else if answer.Data == item.TrueAnswer {
					switch objectArr[index].Type {
					case "A":
						score[0] += 2
					case "B":
						score[1] += 2
					case "C":
						score[2] += 2
					case "F":
						score[3] += 2
					case "G":
						score[4] += 2
					case "H":
						score[5] += 2
					case "L":
						score[6] += 2
					default:
					}
				}
			}
		}
	}

	var c []model.ChoiceItem
	for _, item := range choice {
		newItem := model.ChoiceItem{
			Number: item.Number,
			Data:   item.Data,
		}
		c = append(c, newItem)
	}
	return score, c
}
func contains(slice []int, value int) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
