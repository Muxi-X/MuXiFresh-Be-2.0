package handler

import (
	"MuXiFresh-Be-2.0/app/test/api/internal/logic"
	"MuXiFresh-Be-2.0/app/test/api/internal/svc"
	"MuXiFresh-Be-2.0/app/test/api/internal/types"
	"MuXiFresh-Be-2.0/common/greet/response"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

func CheckTestResultHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TestInfoReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := logic.NewCheckTestResultLogic(r.Context(), svcCtx)
		resp, err := l.CheckTestResult(&req)
		response.Response(w, resp, err)

	}
}
