package xerr

var (
	//全局
	ErrExistInvalidId   = &CodeError{errCode: 10002, errMsg: "存在无效Id"}
	ErrPermissionDenied = &CodeError{errCode: 10003, errMsg: "用户权限不足"}
	ErrNotFind          = &CodeError{errCode: 10004, errMsg: "未查询到指定内容"}
	//auth
	ErrGenerateToken              = &CodeError{errCode: 10101, errMsg: "身份令牌生成失败"}
	ErrEmailVerificationFailed    = &CodeError{errCode: 10102, errMsg: "邮箱验证失败"}
	ErrEmailHasBeenUsed           = &CodeError{errCode: 10103, errMsg: "邮箱已被注册"}
	ErrEmailHasNotBeenUsed        = &CodeError{errCode: 10104, errMsg: "邮箱尚未注册"}
	ErrEmailOrPasswordIsWrong     = &CodeError{errCode: 10105, errMsg: "邮箱或密码错误"}
	ErrGenerateCaptcha            = &CodeError{errCode: 10106, errMsg: "生成人机验证图失败"}
	ErrCaptchaVerificationFailed  = &CodeError{errCode: 10107, errMsg: "人机验证失败"}
	ErrStudentIdHasNotBingToEmail = &CodeError{errCode: 10108, errMsg: "学号尚未绑定"}
	ErrStudentIdHasBeenBind       = &CodeError{errCode: 10109, errMsg: "学号已被绑定"}
	ErrStudentIdOrPasswordIsWrong = &CodeError{errCode: 10110, errMsg: "学号或密码错误"}
)
