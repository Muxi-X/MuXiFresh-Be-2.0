package ccnulogin

import (
	"net/http"

	"MuXiFresh-Be-2.0/common/result"

	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/logic/ccnulogin"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func SetStudentIDHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SetStudentIDReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := ccnulogin.NewSetStudentIDLogic(r.Context(), svcCtx)
		resp, err := l.SetStudentID(&req)
		result.HttpResult(r, w, resp, err)
	}
}
