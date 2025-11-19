package handler

import (
	logic "MuXiFresh-Be-2.0/app/task/cmd/api/internal/logic/submitted"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/common/greet/response"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

func GetMySubmissionStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetMySubmissionStatusReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := logic.NewGetMySubmissionStatusLogic(r.Context(), svcCtx)
		resp, err := l.GetMySubmissionStatus(&req)
		response.Response(w, resp, err)

	}
}
