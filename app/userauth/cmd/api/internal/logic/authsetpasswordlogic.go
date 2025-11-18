package logic

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/common/code"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"
	"MuXiFresh-Be-2.0/common/xerr"
	"context"
	"github.com/golang-jwt/jwt/v4"
	"time"

	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthSetPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthSetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthSetPasswordLogic {
	return &AuthSetPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AuthSetPasswordLogic) AuthSetPassword(req *types.AuthSetPasswordReq) (resp *types.AuthSetPasswordResp, err error) {

	if ok := code.VerifyEmailCode(globalKey.SetPassword, req.Email, req.VerifyCode); !ok {
		return nil, xerr.ErrEmailVerificationFailed
	}
	//gen auth token
	AuthSetPasswordToken, err := l.getJwtToken(l.svcCtx.Config.JwtAuthChPass.AccessSecret, time.Now().Unix(), l.svcCtx.Config.JwtAuthChPass.AccessExpire, req.Email)
	if err != nil {
		return nil, xerr.ErrGenerateToken
	}
	return &types.AuthSetPasswordResp{
		AuthSetPasswordToken: AuthSetPasswordToken,
	}, nil
}
func (l *AuthSetPasswordLogic) getJwtToken(secretKey string, iat, seconds int64, email string) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	claims[ctxData.CtxKeyJwtEmail] = email
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}
