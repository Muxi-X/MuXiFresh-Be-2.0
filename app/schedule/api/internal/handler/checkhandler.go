package handler

import (
	"MuXiFresh-Be-2.0/app/schedule/api/internal/logic"
	"MuXiFresh-Be-2.0/app/schedule/api/internal/svc"
	"MuXiFresh-Be-2.0/app/schedule/api/internal/types"
	"MuXiFresh-Be-2.0/common/greet/response"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

func CheckHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CheckReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := logic.NewCheckLogic(r.Context(), svcCtx)
		resp, err := l.Check(&req)
		response.Response(w, resp, err)

	}
}
