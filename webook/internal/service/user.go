package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository"
)

var (
	ErrUserDuplicateEmail    = repository.ErrUserDuplicateEmail
	ErrInvalidUserOrPassword = errors.New("邮箱或者密码不正确")
)

type UserService interface {
	Signup(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email, passwd string) (domain.User, error)
	Profile(ctx context.Context, id int64) (domain.User, error)
	Edit(ctx context.Context, user domain.User) error
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (svc *userService) Signup(ctx context.Context, user domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)

	return svc.repo.Create(ctx, user)
}

func (svc *userService) Login(ctx context.Context, email, password string) (domain.User, error) {
	u, err := svc.repo.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, err
}

func (svc *userService) Edit(ctx context.Context, user domain.User) error {
	return svc.repo.Update(ctx, user)
}

func (svc *userService) Profile(ctx context.Context, id int64) (domain.User, error) {
	return svc.repo.FindByID(ctx, id)
}

// FindOrCreate 如果手机号不存在，那么会初始化一个用户
func (svc *userService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	u, err := svc.repo.FindByPhone(ctx, phone)
	if !errors.Is(err, repository.ErrUserNotFound) {
		return u, err
	}

	err = svc.repo.Create(ctx, domain.User{Phone: phone})
	if err != nil && !errors.Is(err, repository.ErrUserDuplicateEmail) {
		return domain.User{}, err
	}
	return svc.repo.FindByPhone(ctx, phone)
}
