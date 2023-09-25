package web

import (
	"errors"
	"net/http"
	"time"
	"unicode/utf8"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/service"
	myjwt "geektime-basic-go/webook/internal/web/jwt"
	"geektime-basic-go/webook/internal/web/middleware/handlefunc"
	"geektime-basic-go/webook/pkg/logger"
)

const bizLogin = "login"

var _ handler = (*UserHandler)(nil)

type UserHandler struct {
	svc              service.UserService
	codeSvc          service.CodeService
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
	phoneRegexExp    *regexp.Regexp
	logFunc          handlefunc.LogFunc
	myjwt.Handler
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService, jwtHandler myjwt.Handler, l logger.Logger) *UserHandler {
	const (
		emailRegexPattern  = `^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)+$`
		passwdRegexPattern = `^^(?=.*[0-9])(?=.*[a-zA-Z])[0-9A-Za-z~!@#$%^&*._?]{8,15}$`
		phoneRegexPattern  = `^(13[0-9]|14[01456879]|15[0-35-9]|16[2567]|17[0-8]|18[0-9]|19[0-35-9])\d{8}$`
	)

	return &UserHandler{
		svc:              svc,
		codeSvc:          codeSvc,
		emailRegexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexExp: regexp.MustCompile(passwdRegexPattern, regexp.None),
		phoneRegexExp:    regexp.MustCompile(phoneRegexPattern, regexp.None),
		Handler:          jwtHandler,
		logFunc:          handlefunc.DefaultLogFunc(l),
	}
}

func (uh *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")

	ug.GET("/profile", uh.Profile)

	ug.POST("/signup", uh.SignUp)
	ug.POST("/login", uh.Login)
	ug.POST("/edit", handlefunc.WrapReqWithLog[EditReq](uh.Edit, uh.logFunc))
	ug.POST("/login_sms/code/send", uh.SendSMSLoginCode)
	ug.POST("/login_sms", uh.LoginSMS)
	ug.POST("/refresh_token", uh.RefreshToken)
	ug.POST("/logout", uh.Logout)
}

func (uh *UserHandler) SignUp(ctx *gin.Context) {
	req := struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}{}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, InternalServerError())
		return
	}

	isEmail, err := uh.emailRegexExp.MatchString(req.Email)
	if err != nil {
		ctx.JSON(http.StatusOK, InternalServerError())
		return
	}
	if !isEmail {
		ctx.JSON(http.StatusOK, handlefunc.Response{Code: 4, Msg: "邮箱不正确"})
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.JSON(http.StatusOK, handlefunc.Response{Code: 4, Msg: "两次输入的密码不相同"})
		return
	}

	isPasswd, err := uh.passwordRegexExp.MatchString(req.Password)
	if err != nil {
		ctx.JSON(http.StatusOK, InternalServerError())
		return
	}
	if !isPasswd {
		ctx.JSON(http.StatusOK, handlefunc.Response{Code: 4, Msg: "密码必须包括数字、字母两种字符，长度在8-15位之间"})
		return
	}

	err = uh.svc.Signup(ctx.Request.Context(), domain.User{Email: req.Email, Password: req.ConfirmPassword})
	if errors.Is(err, service.ErrUserDuplicate) {
		ctx.JSON(http.StatusOK, handlefunc.Response{Code: 4, Msg: "重复邮箱，请换一个邮箱"})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, handlefunc.Response{Code: 5, Msg: "服务器异常，注册失败"})
		return
	}

	ctx.JSON(http.StatusOK, handlefunc.Response{Msg: "你好，注册成功"})
}

func (uh *UserHandler) Login(ctx *gin.Context) {
	req := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, InternalServerError())
		return
	}
	u, err := uh.svc.Login(ctx.Request.Context(), req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidUserOrPassword) {
		ctx.JSON(http.StatusOK, handlefunc.Response{Code: 4, Msg: "用户名或密码不正确，请重试"})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, InternalServerError())
		return
	}

	if err = uh.SetLoginToken(ctx, u.ID); err != nil {
		ctx.JSON(http.StatusOK, InternalServerError())
		return
	}
	ctx.JSON(http.StatusOK, handlefunc.Response{Msg: "登录成功"})
}

type EditReq struct {
	Nickname string `json:"nickname"`
	Birthday string `json:"birthday"`
	AboutMe  string `json:"aboutMe"`
}

func (uh *UserHandler) Edit(ctx *gin.Context, req EditReq, uc myjwt.UserClaims) (handlefunc.Response, error) {
	if req.Nickname == "" {
		return handlefunc.Response{Code: 4, Msg: "昵称不能为空"}, errors.New("昵称不能为空")
	}
	if utf8.RuneCountInString(req.Nickname) > 30 {
		return handlefunc.Response{Code: 4, Msg: "昵称过长"}, errors.New("昵称过长")
	}
	if utf8.RuneCountInString(req.AboutMe) > 50 {
		return handlefunc.Response{Code: 4, Msg: "关于我过长"}, errors.New("关于我过长")
	}
	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		return handlefunc.Response{Code: 4, Msg: "日期格式不对"}, errors.New("日期格式不对")
	}

	err = uh.svc.Edit(ctx, domain.User{ID: uc.ID, Nickname: req.Nickname, Birthday: birthday, AboutMe: req.AboutMe})
	if err != nil {
		return handlefunc.InternalServerError(), err
	}

	return handlefunc.Response{Msg: "OK"}, nil
}

func (uh *UserHandler) Profile(ctx *gin.Context) {
	type Profile struct {
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}
	uc := ctx.MustGet("user").(myjwt.UserClaims)
	user, err := uh.svc.Profile(ctx, uc.ID)
	if err != nil {
		ctx.JSON(http.StatusOK, InternalServerError())
		return
	}
	ctx.JSON(http.StatusOK, handlefunc.Response{Msg: "OK", Data: Profile{
		Email:    user.Email,
		Phone:    user.Phone,
		Nickname: user.Nickname,
		Birthday: user.Birthday.Format(time.DateOnly),
		AboutMe:  user.AboutMe,
	}})
}

func (uh *UserHandler) SendSMSLoginCode(ctx *gin.Context) {
	req := struct {
		Phone string `json:"phone"`
	}{}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, InternalServerError())
		return
	}

	isPhone, err := uh.phoneRegexExp.MatchString(req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, InternalServerError())
		return
	}
	if !isPhone {
		ctx.JSON(http.StatusOK, handlefunc.Response{Code: 4, Msg: "手机号码错误"})
		return
	}

	switch err = uh.codeSvc.Send(ctx, bizLogin, req.Phone); {
	default:
		ctx.JSON(http.StatusOK, InternalServerError())
	case err == nil:
		ctx.JSON(http.StatusOK, handlefunc.Response{Msg: "发送成功"})
	case errors.Is(err, service.ErrCodeSendTooMany):
		ctx.JSON(http.StatusOK, handlefunc.Response{Code: 4, Msg: "短信发送太频繁，请稍后再试"})
	}
}

func (uh *UserHandler) LoginSMS(ctx *gin.Context) {
	req := struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}{}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, InternalServerError())
		return
	}
	ok, err := uh.codeSvc.Verify(ctx, bizLogin, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, InternalServerError())
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, handlefunc.Response{Code: 4, Msg: "验证码错误"})
		return
	}

	u, err := uh.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, InternalServerError())
		return
	}

	if err = uh.SetLoginToken(ctx, u.ID); err != nil {
		ctx.JSON(http.StatusOK, InternalServerError())
		return
	}
	ctx.JSON(http.StatusOK, handlefunc.Response{Msg: "登录成功"})
}

func (uh *UserHandler) RefreshToken(ctx *gin.Context) {
	tokenStr := uh.ExtractTokenString(ctx)
	var rc myjwt.RefreshClaims
	token, err := jwt.ParseWithClaims(tokenStr, &rc, func(token *jwt.Token) (interface{}, error) {
		return myjwt.RefreshTokenKey, nil
	})
	if err != nil || token == nil || !token.Valid {
		ctx.JSON(http.StatusUnauthorized, handlefunc.Response{Code: 4, Msg: "请登录"})
		return
	}
	expireTime, err := rc.GetExpirationTime()
	if err != nil || expireTime == nil {
		// 拿不到过期时间或者token过期
		ctx.JSON(http.StatusUnauthorized, handlefunc.Response{Code: 4, Msg: "请登录"})
		return
	}
	if err = uh.CheckSession(ctx, rc.SSID); err != nil {
		// 系统错误或者用户已经主动退出登录了
		ctx.JSON(http.StatusUnauthorized, handlefunc.Response{Code: 4, Msg: "请登录"})
		return
	}

	if err = uh.SetJWTToken(ctx, rc.SSID, rc.ID); err != nil {
		ctx.JSON(http.StatusOK, InternalServerError())
		return
	}
	ctx.JSON(http.StatusOK, handlefunc.Response{Msg: "刷新成功"})
}

func (uh *UserHandler) Logout(ctx *gin.Context) {
	if err := uh.ClearToken(ctx); err != nil {
		ctx.JSON(http.StatusOK, InternalServerError())
		return
	}

	ctx.JSON(http.StatusOK, handlefunc.Response{Msg: "OK"})
}
