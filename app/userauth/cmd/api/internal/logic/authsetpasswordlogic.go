package logic

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/common/code"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"
	"context"
	"fmt"
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

// signAuthSetPasswordToken is overridable so tests can exercise the failure
// path, where the consumed verification code must be restored.
var signAuthSetPasswordToken = getJwtToken

func (l *AuthSetPasswordLogic) AuthSetPassword(req *types.AuthSetPasswordReq) (resp *types.AuthSetPasswordResp, err error) {

	if ok := code.VerifyEmailCode(globalKey.SetPassword, req.Email, req.VerifyCode); !ok {
		return nil, fmt.Errorf("verify code failed")
	}
	//gen auth token
	AuthSetPasswordToken, err := signAuthSetPasswordToken(l.svcCtx.Config.JwtAuthChPass.AccessSecret, time.Now().Unix(), l.svcCtx.Config.JwtAuthChPass.AccessExpire, req.Email)
	if err != nil {
		l.restoreCode(req.Email, req.VerifyCode)
		return nil, err
	}
	return &types.AuthSetPasswordResp{
		AuthSetPasswordToken: AuthSetPasswordToken,
	}, nil
}

// restoreCode puts the consumed code back when the follow-up step failed, so a
// retry with the same code still works.
func (l *AuthSetPasswordLogic) restoreCode(email, verifyCode string) {
	if err := code.RestoreEmailCode(globalKey.SetPassword, email, verifyCode); err != nil {
		l.Errorf("restore set-password code for %s failed: %v", email, err)
	}
}

func getJwtToken(secretKey string, iat, seconds int64, email string) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	claims[ctxData.CtxKeyJwtEmail] = email
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}
