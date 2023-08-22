package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"geektime-basic-go/week04/webook/internal/domain"
	"geektime-basic-go/week04/webook/internal/repository"
)

var (
	ErrUserDuplicateEmail    = repository.ErrUserDuplicateEmail
	ErrUserNotFound          = repository.ErrUserNotFound
	ErrDataTooLong           = repository.ErrDataTooLong
	ErrUserDuplicateNickname = repository.ErrUserDuplicateNickname

	ErrInvalidUserOrPassword = errors.New("邮箱或者密码不正确")
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (svc *UserService) Signup(ctx context.Context, user domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)

	return svc.repo.Create(ctx, user)
}

func (svc *UserService) Login(ctx context.Context, email, password string) (domain.User, error) {
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

func (svc *UserService) Edit(ctx context.Context, user domain.User) error {
	return svc.repo.Update(ctx, user)
}

func (svc *UserService) Profile(ctx context.Context, id int64) (domain.User, error) {
	return svc.repo.FindByID(ctx, id)
}
