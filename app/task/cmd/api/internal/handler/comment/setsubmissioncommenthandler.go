package handler

import (
	logic "MuXiFresh-Be-2.0/app/task/cmd/api/internal/logic/comment"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/common/greet/response"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

func SetSubmissionCommentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SetSubmissionCommentReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := logic.NewSetSubmissionCommentLogic(r.Context(), svcCtx)
		resp, err := l.SetSubmissionComment(&req)
		response.Response(w, resp, err)

	}
}
