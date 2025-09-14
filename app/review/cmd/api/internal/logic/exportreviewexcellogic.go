package logic

import (
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/userclient"
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
	entryForms, err := l.svcCtx.EntryFormModel.FindByGroup(l.ctx, req.Group, req.School, req.Grade, startTime, endTime)
	if err != nil {
		return nil, "", err
	}
	var rows []types.Row
	for _, entryForm := range entryForms {

		userId := entryForm.UserId.String()[10:34]
		//录取进度
		schedule, err := l.svcCtx.ScheduleClient.FindOneByUserId(l.ctx, userId)
		if err != nil {
			return nil, "", err
		}

		if req.Status != "" && schedule.AdmissionStatus != req.Status {
			continue
		}
		//测验情况
		userInfo, err := l.svcCtx.UserInfoModel.FindOne(l.ctx, userId)
		if err != nil {
			return nil, "", err
		}

		examStatus := "已提交"

		rows = append(rows, types.Row{
			Name:            userInfo.Name,
			Grade:           entryForm.Grade,
			School:          userInfo.School,
			Group:           entryForm.Group,
			Gender:          entryForm.Gender,
			FormID:          entryForm.ID.String()[10:34],
			ExamStuatus:     examStatus,
			UserId:          userId,
			AdmissionStatus: schedule.AdmissionStatus,
			ScheduleID:      schedule.ID.String()[10:34],
			Understanding:   entryForm.Knowledge,
			Reason:          entryForm.Reason,
			SelfIntro:       entryForm.SelfIntro,
		})
	}
	// --- 生成 Excel ---
	f := excelize.NewFile()
	sheet := "AllGroups"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"Name", "Grade", "School", "Group", "Gender", "FormID", "ExamStatus", "UserId", "AdmissionStatus", "ScheduleID", "Understanding", "Reason", "SelfIntro"}
	for i, h := range headers {
		col := string(rune('A' + i))
		f.SetCellValue(sheet, col+"1", h)
	}

	for rowIdx, r := range rows {
		f.SetCellValue(sheet, "A"+strconv.Itoa(rowIdx+2), r.Name)
		f.SetCellValue(sheet, "B"+strconv.Itoa(rowIdx+2), r.Grade)
		f.SetCellValue(sheet, "C"+strconv.Itoa(rowIdx+2), r.School)
		f.SetCellValue(sheet, "D"+strconv.Itoa(rowIdx+2), r.Group)
		f.SetCellValue(sheet, "E"+strconv.Itoa(rowIdx+2), r.Gender)
		f.SetCellValue(sheet, "F"+strconv.Itoa(rowIdx+2), r.FormID)
		f.SetCellValue(sheet, "G"+strconv.Itoa(rowIdx+2), r.ExamStuatus)
		f.SetCellValue(sheet, "H"+strconv.Itoa(rowIdx+2), r.UserId)
		f.SetCellValue(sheet, "I"+strconv.Itoa(rowIdx+2), r.AdmissionStatus)
		f.SetCellValue(sheet, "J"+strconv.Itoa(rowIdx+2), r.ScheduleID)
		f.SetCellValue(sheet, "K"+strconv.Itoa(rowIdx+2), r.Understanding)
		f.SetCellValue(sheet, "L"+strconv.Itoa(rowIdx+2), r.Reason)
		f.SetCellValue(sheet, "M"+strconv.Itoa(rowIdx+2), r.SelfIntro)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}
	fileName := fmt.Sprintf("review_%d_%s.xlsx", time.Now().Unix(), uuid.New().String())
	return buf, fileName, nil
}
