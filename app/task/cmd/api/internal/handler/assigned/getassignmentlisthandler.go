package assigned

import (
	"net/http"

	"MuXiFresh-Be-2.0/common/result"

	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/logic/assigned"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetAssignmentListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetAssignmentListReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := assigned.NewGetAssignmentListLogic(r.Context(), svcCtx)
		resp, err := l.GetAssignmentList(&req)
		result.HttpResult(r, w, resp, err)
	}
}
