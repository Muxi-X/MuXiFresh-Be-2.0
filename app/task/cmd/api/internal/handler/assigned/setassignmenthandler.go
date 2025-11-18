package assigned

import (
	"net/http"

	"MuXiFresh-Be-2.0/common/result"

	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/logic/assigned"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func SetAssignmentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SetAssignmentReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := assigned.NewSetAssignmentLogic(r.Context(), svcCtx)
		resp, err := l.SetAssignment(&req)
		result.HttpResult(r, w, resp, err)
	}
}
