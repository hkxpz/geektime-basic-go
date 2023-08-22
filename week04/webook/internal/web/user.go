package web

import (
	"errors"
	"net/http"
	"time"
	"unicode/utf8"

	"geektime-basic-go/week04/webook/internal/domain"
	"geektime-basic-go/week04/webook/internal/service"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
)

const (
	emailRegexPattern    = `^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
)

type UserHandler struct {
	svc              *service.UserService
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		svc:              svc,
		emailRegexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
	}
}

func (uh *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.GET("/profile", uh.Profile)
	ug.POST("/signup", uh.SignUp)
	ug.POST("/login", uh.Login)
	ug.POST("/edit", uh.Edit)
}

func (uh *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	isEmail, err := uh.emailRegexExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "邮箱不正确")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入的密码不相同")
		return
	}

	isPasswd, err := uh.passwordRegexExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isPasswd {
		ctx.String(http.StatusOK, "密码必须包含数字、特殊字符，且长度不能小于 8 位")
		return
	}

	err = uh.svc.Signup(ctx.Request.Context(), domain.User{Email: req.Email, Password: req.ConfirmPassword})
	if errors.Is(err, service.ErrUserDuplicateEmail) {
		ctx.String(http.StatusOK, "重复邮箱，请换一个邮箱")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "服务器异常，注册失败")
		return
	}

	ctx.String(http.StatusOK, "你好，注册成功")
}

func (uh *UserHandler) Login(ctx *gin.Context) {
	type LogReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LogReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	u, err := uh.svc.Login(ctx.Request.Context(), req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidUserOrPassword) {
		ctx.String(http.StatusOK, "用户名或密码不正确，请重试")
		return
	}

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaims{
		Id:        u.Id,
		UserAgent: ctx.Request.UserAgent(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
		},
	})
	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	ctx.Header("x-jwt-token", tokenStr)
	ctx.String(http.StatusOK, "登录成功")
}

func (uh *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		Id           int64  `json:"id"`
		Nickname     string `json:"nickname"`
		Birthday     string `json:"birthday"`
		Introduction string `json:"introduction"`
	}

	req := EditReq{}
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	if utf8.RuneCountInString(req.Nickname) > 30 {
		ctx.String(http.StatusOK, "昵称过长")
		return
	}
	if _, err := time.Parse("2006-01-02", req.Birthday); err != nil {
		ctx.String(http.StatusOK, "生日错误")
		return
	}
	if utf8.RuneCountInString(req.Introduction) > 50 {
		ctx.String(http.StatusOK, "个人简介过长")
		return
	}

	err := uh.svc.Edit(ctx, domain.User{Id: req.Id, Nickname: req.Nickname, Birthday: req.Birthday, Introduction: req.Introduction})
	if errors.Is(err, service.ErrUserDuplicateNickname) {
		ctx.String(http.StatusOK, "重复昵称，请换一个昵称")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.String(http.StatusOK, "更新成功")
}

func (uh *UserHandler) Profile(ctx *gin.Context) {
	userClaims := ctx.MustGet("user").(UserClaims)
	user, err := uh.svc.Profile(ctx, userClaims.Id)
	if errors.Is(err, service.ErrUserNotFound) {
		ctx.String(http.StatusOK, "用户不存在")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, user)
}
