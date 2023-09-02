package handler

import (
	"net/http"

	"MuXiFresh-Be-2.0/common/result"

	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/logic"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func AuthSetPasswordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AuthSetPasswordReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := logic.NewAuthSetPasswordLogic(r.Context(), svcCtx)
		resp, err := l.AuthSetPassword(&req)
		result.HttpResult(r, w, resp, err)
	}
}
