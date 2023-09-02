package handler

import (
	"net/http"

	"MuXiFresh-Be-2.0/common/result"

	"MuXiFresh-Be-2.0/app/form/api/internal/logic"
	"MuXiFresh-Be-2.0/app/form/api/internal/svc"
	"MuXiFresh-Be-2.0/app/form/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func JudgeUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ClickReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := logic.NewJudgeUserLogic(r.Context(), svcCtx)
		resp, err := l.JudgeUser(&req)
		result.HttpResult(r, w, resp, err)
	}
}
