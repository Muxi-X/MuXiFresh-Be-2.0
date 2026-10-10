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

	// 届次窗口与 form 的 CycleOf 分界对齐（含春招 6/30 边界），与 GetReview 共用
	startTime, endTime := seasonWindow(req.Year, req.Season)

	rows, err := buildReviewRows(l.ctx, l.svcCtx, groupFilter(req.Group), req.School, req.Grade, req.Status, startTime, endTime)
	if err != nil {
		return nil, "", err
	}
	// --- 生成 Excel ---
	f := excelize.NewFile()

	sheets := buildExportSheets(req.Group, rows)

	headers := []string{"姓名", "年级", "学校", "组别", "性别", "专业", "电话", "QQ", "报名表ID", "录取状态", "知识储备", "报名理由", "自我简介", "附加问题", "面试评价"}

	for idx, s := range sheets {
		if idx == 0 {
			f.SetSheetName("Sheet1", s.name)
		} else {
			f.NewSheet(s.name)
		}
		for i, h := range headers {
			f.SetCellValue(s.name, string(rune('A'+i))+"1", h)
		}
		for rowIdx, r := range s.rows {
			row := strconv.Itoa(rowIdx + 2)
			f.SetCellValue(s.name, "A"+row, r.Name)
			f.SetCellValue(s.name, "B"+row, r.Grade)
			f.SetCellValue(s.name, "C"+row, r.School)
			f.SetCellValue(s.name, "D"+row, convert.GroupCvtChinese(r.Group))
			f.SetCellValue(s.name, "E"+row, r.Gender)
			f.SetCellValue(s.name, "F"+row, r.Major)
			f.SetCellValue(s.name, "G"+row, r.Phone)
			f.SetCellValue(s.name, "H"+row, r.QQ)
			f.SetCellValue(s.name, "I"+row, r.FormID)
			f.SetCellValue(s.name, "J"+row, r.AdmissionStatus)
			f.SetCellValue(s.name, "K"+row, r.Understanding)
			f.SetCellValue(s.name, "L"+row, r.Reason)
			f.SetCellValue(s.name, "M"+row, r.SelfIntro)
			f.SetCellValue(s.name, "N"+row, r.ExtraQuestion)
			f.SetCellValue(s.name, "O"+row, r.InterviewComment)
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}
	fileName := fmt.Sprintf("review_%d_%s.xlsx", time.Now().Unix(), uuid.New().String())
	return buf, fileName, nil
}

// groupNames 是导出按组拆 sheet 时的固定顺序与中文表名。
var groupNames = []struct{ en, cn string }{
	{"Product", "产品组"},
	{"Design", "设计组"},
	{"Frontend", "前端组"},
	{"Backend", "后端组"},
	{"Android", "安卓组"},
	{"Operation", "运营组"},
}

// exportSheet 是导出的一个工作表：name 为表名，rows 为该表数据。
type exportSheet struct {
	name string
	rows []types.Row
}

// buildExportSheets 按请求的 group 产出有序工作表列表。
// 全量导出（group 为空或 All）时，首个工作表为包含全部记录的"全部"表，
// 其后按 groupNames 顺序为各组拆表；指定具体组时只含该组；未知组回退为各组拆表。
func buildExportSheets(group string, rows []types.Row) []exportSheet {
	byGroup := make(map[string][]types.Row)
	for _, r := range rows {
		byGroup[r.Group] = append(byGroup[r.Group], r)
	}

	if group == "" || isGroupAll(group) {
		sheets := make([]exportSheet, 0, len(groupNames)+1)
		sheets = append(sheets, exportSheet{name: "全部", rows: rows})
		for _, g := range groupNames {
			sheets = append(sheets, exportSheet{name: g.cn, rows: byGroup[g.en]})
		}
		return sheets
	}

	for _, g := range groupNames {
		if g.en == group {
			return []exportSheet{{name: g.cn, rows: byGroup[g.en]}}
		}
	}

	sheets := make([]exportSheet, 0, len(groupNames))
	for _, g := range groupNames {
		sheets = append(sheets, exportSheet{name: g.cn, rows: byGroup[g.en]})
	}
	return sheets
}
