package handler

import (
	logic "MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/logic/ccnulogin"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/common/greet/response"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

func CcnuLoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CcnuLoginReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := logic.NewCcnuLoginLogic(r.Context(), svcCtx)
		resp, err := l.CcnuLogin(&req)
		response.Response(w, resp, err)

	}
}
