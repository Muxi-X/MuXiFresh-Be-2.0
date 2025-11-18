package comment

import (
	"net/http"

	"MuXiFresh-Be-2.0/common/result"

	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/logic/comment"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func DelSubmissionCommentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DelSubmissionCommentReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := comment.NewDelSubmissionCommentLogic(r.Context(), svcCtx)
		resp, err := l.DelSubmissionComment(&req)
		result.HttpResult(r, w, resp, err)
	}
}
