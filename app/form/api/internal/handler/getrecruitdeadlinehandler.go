package handler

import (
	"net/http"

	"MuXiFresh-Be-2.0/app/form/api/internal/logic"
	"MuXiFresh-Be-2.0/app/form/api/internal/svc"
	"MuXiFresh-Be-2.0/app/form/api/internal/types"
	"MuXiFresh-Be-2.0/common/greet/response"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetRecruitDeadlineHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetRecruitDeadlineReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := logic.NewGetRecruitDeadlineLogic(r.Context(), svcCtx)
		resp, err := l.GetRecruitDeadline(&req)
		response.Response(w, resp, err)
	}
}
