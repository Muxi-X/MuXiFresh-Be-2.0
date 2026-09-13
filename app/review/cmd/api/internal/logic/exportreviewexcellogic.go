package logic

import (
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/userclient"
	"MuXiFresh-Be-2.0/common/convert"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"strconv"
	"time"
)

type ExportReviewExcelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 一键导出 Excel（包含每个组的名单）
func NewExportReviewExcelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExportReviewExcelLogic {
	return &ExportReviewExcelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ExportReviewExcelLogic) ExportReviewExcel(req *types.ExportReviewExcelReq) (*bytes.Buffer, string, error) {
	//管理员认证
	getUserTypeResp, err := l.svcCtx.UserClient.GetUserType(l.ctx, &userclient.GetUserTypeReq{
		UserId: ctxData.GetUserIdFromCtx(l.ctx),
	})
	if err != nil {
		return nil, "", err
	}
	if getUserTypeResp.UserType != globalKey.Admin && getUserTypeResp.UserType != globalKey.SuperAdmin {
		return nil, "", errors.New("permission denied")
	}

	//秋招
	startTime := time.Date(req.Year, time.July, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(req.Year, time.December, 31, 23, 59, 59, 999999999, time.UTC)
	//春招
	if req.Season == "spring" {
		startTime = time.Date(req.Year, time.January, 1, 0, 0, 0, 0, time.UTC)
		endTime = time.Date(req.Year, time.June, 31, 23, 59, 59, 999999999, time.UTC)
	}
	if req.Season == "" {
		startTime = time.Date(req.Year, time.January, 1, 0, 0, 0, 0, time.UTC)
		endTime = time.Date(req.Year, time.December, 31, 23, 59, 59, 999999999, time.UTC)
	}
	rows, err := buildReviewRows(l.ctx, l.svcCtx, req.Group, req.School, req.Grade, req.Status, startTime, endTime)
	if err != nil {
		return nil, "", err
	}
	// --- 生成 Excel ---
	f := excelize.NewFile()

	groupNames := []struct{ en, cn string }{
		{"Product", "产品组"},
		{"Design", "设计组"},
		{"Frontend", "前端组"},
		{"Backend", "后端组"},
		{"Android", "安卓组"},
		{"Operation", "运营组"},
	}

	// 按组拆 sheet：req.Group 为空导出所有组（空组也建表头），否则仅该组
	targets := groupNames
	if req.Group != "" {
		targets = nil
		for _, g := range groupNames {
			if g.en == req.Group {
				targets = []struct{ en, cn string }{g}
				break
			}
		}
		if targets == nil {
			targets = groupNames
		}
	}

	byGroup := make(map[string][]types.Row)
	for _, r := range rows {
		byGroup[r.Group] = append(byGroup[r.Group], r)
	}

	headers := []string{"姓名", "年级", "学校", "组别", "性别", "专业", "电话", "QQ", "报名表ID", "录取状态", "知识储备", "报名理由", "自我简介", "附加问题"}

	for idx, g := range targets {
		sheet := g.cn
		if idx == 0 {
			f.SetSheetName("Sheet1", sheet)
		} else {
			f.NewSheet(sheet)
		}
		for i, h := range headers {
			f.SetCellValue(sheet, string(rune('A'+i))+"1", h)
		}
		for rowIdx, r := range byGroup[g.en] {
			row := strconv.Itoa(rowIdx + 2)
			f.SetCellValue(sheet, "A"+row, r.Name)
			f.SetCellValue(sheet, "B"+row, r.Grade)
			f.SetCellValue(sheet, "C"+row, r.School)
			f.SetCellValue(sheet, "D"+row, convert.GroupCvtChinese(r.Group))
			f.SetCellValue(sheet, "E"+row, r.Gender)
			f.SetCellValue(sheet, "F"+row, r.Major)
			f.SetCellValue(sheet, "G"+row, r.Phone)
			f.SetCellValue(sheet, "H"+row, r.QQ)
			f.SetCellValue(sheet, "I"+row, r.FormID)
			f.SetCellValue(sheet, "J"+row, r.AdmissionStatus)
			f.SetCellValue(sheet, "K"+row, r.Understanding)
			f.SetCellValue(sheet, "L"+row, r.Reason)
			f.SetCellValue(sheet, "M"+row, r.SelfIntro)
			f.SetCellValue(sheet, "N"+row, r.ExtraQuestion)
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}
	fileName := fmt.Sprintf("review_%d_%s.xlsx", time.Now().Unix(), uuid.New().String())
	return buf, fileName, nil
}
