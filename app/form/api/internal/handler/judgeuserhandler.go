package handler

import (
	"MuXiFresh-Be-2.0/app/form/api/internal/logic"
	"MuXiFresh-Be-2.0/app/form/api/internal/svc"
	"MuXiFresh-Be-2.0/app/form/api/internal/types"
	"MuXiFresh-Be-2.0/common/greet/response"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

func JudgeUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ClickReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := logic.NewJudgeUserLogic(r.Context(), svcCtx)
		resp, err := l.JudgeUser(&req)
		response.Response(w, resp, err)

	}
}
