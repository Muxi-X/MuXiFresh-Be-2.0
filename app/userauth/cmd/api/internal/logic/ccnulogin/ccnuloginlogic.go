package ccnulogin

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/accountcenterclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"context"
	"github.com/golang-jwt/jwt/v4"
	"time"

	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CcnuLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCcnuLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CcnuLoginLogic {
	return &CcnuLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CcnuLoginLogic) CcnuLogin(req *types.CcnuLoginReq) (resp *types.CcnuLoginResp, err error) {
	ccnuLoginResp, err := l.svcCtx.AccountCenterClient.CcnuLogin(l.ctx, &accountcenterclient.CcnuLoginReq{
		StudentID: req.StdudentID,
		Password:  req.Password,
	})
	if err != nil {
		return nil, err
	}
	token, err := l.getJwtToken(l.svcCtx.Config.JwtAuth.AccessSecret, time.Now().Unix(), l.svcCtx.Config.JwtAuth.AccessExpire, ccnuLoginResp.UserinfoID)
	return &types.CcnuLoginResp{
		Token: token,
	}, nil
}

func (l *CcnuLoginLogic) getJwtToken(secretKey string, iat, seconds int64, userID string) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	claims[ctxData.CtxKeyJwtUserID] = userID
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}
