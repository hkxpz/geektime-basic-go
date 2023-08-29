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
)

const (
	emailRegexPattern  = `^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)+$`
	passwdRegexPattern = `^^(?=.*[0-9])(?=.*[a-zA-Z])[0-9A-Za-z~!@#$%^&*._?]{8,15}$`
	phoneRegexPattern  = `^(13[0-9]|14[01456879]|15[0-35-9]|16[2567]|17[0-8]|18[0-9]|19[0-35-9])\d{8}$`
)

const bizLogin = "login"

var _ handler = &UserHandler{}

type UserHandler struct {
	svc              service.UserService
	codeSvc          service.CodeService
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
	phoneRegexExp    *regexp.Regexp
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService) *UserHandler {
	return &UserHandler{
		svc:              svc,
		codeSvc:          codeSvc,
		emailRegexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexExp: regexp.MustCompile(passwdRegexPattern, regexp.None),
		phoneRegexExp:    regexp.MustCompile(phoneRegexPattern, regexp.None),
	}
}

func (uh *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", uh.SignUp)
	ug.POST("/login", uh.Login)
	ug.POST("/edit", uh.Edit)
	ug.POST("/login_sms/code/send", uh.SendSMSLoginCode)
	ug.POST("/login_sms", uh.LoginSMS)

	ug.GET("/profile", uh.Profile)

}

func (uh *UserHandler) SignUp(ctx *gin.Context) {
	req := struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}{}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}

	isEmail, err := uh.emailRegexExp.MatchString(req.Email)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	if !isEmail {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "邮箱不正确"})
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "两次输入的密码不相同"})
		return
	}

	isPasswd, err := uh.passwordRegexExp.MatchString(req.Password)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	if !isPasswd {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "密码必须包括数字、字母两种字符，长度在8-15位之间"})
		return
	}

	err = uh.svc.Signup(ctx.Request.Context(), domain.User{Email: req.Email, Password: req.ConfirmPassword})
	if errors.Is(err, service.ErrUserDuplicate) {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "重复邮箱，请换一个邮箱"})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}

	ctx.JSON(http.StatusOK, Result{Msg: "你好，注册成功"})
}

func (uh *UserHandler) Login(ctx *gin.Context) {
	req := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	u, err := uh.svc.Login(ctx.Request.Context(), req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidUserOrPassword) {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "用户名或密码不正确，请重试"})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}

	if err = setJWTToken(ctx, u.Id); err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	ctx.JSON(http.StatusOK, Result{Msg: "登录成功"})
}

func setJWTToken(ctx *gin.Context, uid int64) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaims{
		Id:        uid,
		UserAgent: ctx.Request.UserAgent(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
		},
	})
	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (uh *UserHandler) Edit(ctx *gin.Context) {
	req := struct {
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}{}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}

	if req.Nickname == "" {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "昵称不能为空"})
		return
	}
	if utf8.RuneCountInString(req.Nickname) > 30 {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "昵称过长"})
		return
	}
	if len(req.AboutMe) > 1024 {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "关于我过长"})
		return
	}
	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "日期格式不对"})
		return
	}

	id := ctx.MustGet("user").(UserClaims).Id
	err = uh.svc.Edit(ctx, domain.User{Id: id, Nickname: req.Nickname, Birthday: birthday, AboutMe: req.AboutMe})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	ctx.JSON(http.StatusOK, Result{Msg: "OK"})
}

func (uh *UserHandler) Profile(ctx *gin.Context) {
	userClaims := ctx.MustGet("user").(UserClaims)
	user, err := uh.svc.Profile(ctx, userClaims.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func (uh *UserHandler) SendSMSLoginCode(ctx *gin.Context) {
	req := struct {
		Phone string `json:"phone"`
	}{}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}

	isPhone, err := uh.phoneRegexExp.MatchString(req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	if !isPhone {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "手机号码错误"})
		return
	}

	switch err = uh.codeSvc.Send(ctx, bizLogin, req.Phone); {
	default:
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
	case err == nil:
		ctx.JSON(http.StatusOK, Result{Msg: "发送成功"})
	case errors.Is(err, service.ErrCodeSendTooMany):
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "短信发送太频繁，请稍后再试"})
	}
}

func (uh *UserHandler) LoginSMS(ctx *gin.Context) {
	req := struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}{}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	ok, err := uh.codeSvc.Verify(ctx, bizLogin, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "验证码错误"})
		return
	}

	u, err := uh.svc.FindOrCreate(ctx, req.Phone)
	if errors.Is(err, service.ErrUserDuplicate) {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "重复手机号，请换一个手机号"})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}

	if err = setJWTToken(ctx, u.Id); err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	ctx.JSON(http.StatusOK, Result{Msg: "登录成功"})
}
