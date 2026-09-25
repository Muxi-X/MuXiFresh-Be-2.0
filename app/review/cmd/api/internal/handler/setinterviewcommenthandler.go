package handler

import (
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/logic"
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/common/greet/response"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

func SetInterviewCommentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SetInterviewCommentReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := logic.NewSetInterviewCommentLogic(r.Context(), svcCtx)
		resp, err := l.SetInterviewComment(&req)
		response.Response(w, resp, err)

	}
}
