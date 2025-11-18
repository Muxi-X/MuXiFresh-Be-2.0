package handler

import (
	"net/http"

	"MuXiFresh-Be-2.0/common/result"

	"MuXiFresh-Be-2.0/app/test/api/internal/logic"
	"MuXiFresh-Be-2.0/app/test/api/internal/svc"
	"MuXiFresh-Be-2.0/app/test/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func CheckTestResultHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TestInfoReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := logic.NewCheckTestResultLogic(r.Context(), svcCtx)
		resp, err := l.CheckTestResult(&req)
		result.HttpResult(r, w, resp, err)
	}
}
