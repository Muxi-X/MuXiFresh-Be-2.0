package handler

import (
	"net/http"

	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/logic"
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/common/greet/response"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetInterviewCommentHistoryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetInterviewCommentHistoryReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := logic.NewGetInterviewCommentHistoryLogic(r.Context(), svcCtx)
		resp, err := l.GetInterviewCommentHistory(&req)
		response.Response(w, resp, err)
	}
}
