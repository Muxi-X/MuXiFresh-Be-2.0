package handler

import (
	"MuXiFresh-Be-2.0/app/schedule/api/internal/logic"
	"MuXiFresh-Be-2.0/app/schedule/api/internal/svc"
	"MuXiFresh-Be-2.0/app/schedule/api/internal/types"
	"MuXiFresh-Be-2.0/common/greet/response"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

func CreateScheduleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := logic.NewCreateScheduleLogic(r.Context(), svcCtx)
		resp, err := l.CreateSchedule(&req)
		response.Response(w, resp, err)

	}
}
