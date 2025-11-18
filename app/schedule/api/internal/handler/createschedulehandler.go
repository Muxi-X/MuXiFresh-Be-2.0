package handler

import (
	"net/http"

	"MuXiFresh-Be-2.0/common/result"

	"MuXiFresh-Be-2.0/app/schedule/api/internal/logic"
	"MuXiFresh-Be-2.0/app/schedule/api/internal/svc"
	"MuXiFresh-Be-2.0/app/schedule/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func CreateScheduleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := logic.NewCreateScheduleLogic(r.Context(), svcCtx)
		resp, err := l.CreateSchedule(&req)
		result.HttpResult(r, w, resp, err)
	}
}
