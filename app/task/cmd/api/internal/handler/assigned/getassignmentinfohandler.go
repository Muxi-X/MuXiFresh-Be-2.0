package assigned

import (
	"net/http"

	"MuXiFresh-Be-2.0/common/result"

	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/logic/assigned"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetAssignmentInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetAssignmentInfoReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := assigned.NewGetAssignmentInfoLogic(r.Context(), svcCtx)
		resp, err := l.GetAssignmentInfo(&req)
		result.HttpResult(r, w, resp, err)
	}
}
