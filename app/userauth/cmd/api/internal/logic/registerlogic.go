package logic

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/common/code"
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/accountcenterclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"
	"MuXiFresh-Be-2.0/common/tool"
	"MuXiFresh-Be-2.0/common/xerr"
	"context"
	"github.com/golang-jwt/jwt/v4"
	"time"

	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.RegisterResp, err error) {
	//verify code
	if ok := code.VerifyEmailCode(globalKey.Register, req.Email, req.VerifyCode); !ok {
		return nil, xerr.ErrEmailVerificationFailed
	}

	//写入数据库
	registerDataResp, err := l.svcCtx.AccountCenterClient.Register(l.ctx, &accountcenterclient.RegisterDataReq{
		Email:    req.Email,
		Password: tool.EncryptedPasswordMD5(req.Password),
	})
	if err != nil {
		return nil, err
	}
	//gen token
	tokenStr, err := l.getJwtToken(l.svcCtx.Config.JwtAuth.AccessSecret, time.Now().Unix(), l.svcCtx.Config.JwtAuth.AccessExpire, registerDataResp.ID)
	if err != nil {
		return nil, xerr.ErrGenerateToken
	}

	return &types.RegisterResp{
		Token: tokenStr,
	}, nil
}

// @secretKey: JWT 加解密密钥
// @iat: 时间戳
// @seconds: 过期时间，单位秒
// @payload: 数据载体
func (l *RegisterLogic) getJwtToken(secretKey string, iat, seconds int64, userID string) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	claims[ctxData.CtxKeyJwtUserID] = userID
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}
