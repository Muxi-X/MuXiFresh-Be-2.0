package handler

import (
	"context"
	"net/http"
	"time"

	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/logic"
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 一键导出 Excel（包含每个组的名单）
func ExportReviewExcelHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Minute)
		defer cancel()
		var req types.ExportReviewExcelReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewExportReviewExcelLogic(ctx, svcCtx)
		buf, fileName, err := l.ExportReviewExcel(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		w.WriteHeader(http.StatusOK)
		w.Write(buf.Bytes())
	}
}
