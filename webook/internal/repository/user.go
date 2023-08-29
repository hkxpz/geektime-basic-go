package repository

import (
	"context"
	"database/sql"
	"time"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository/cache"
	"geektime-basic-go/webook/internal/repository/dao"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrDataNotFound
)

type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	Update(ctx context.Context, u domain.User) error
	FindByID(ctx context.Context, id int64) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
}

type userRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewUserRepository(d dao.UserDAO, cache cache.UserCache) UserRepository {
	return &userRepository{dao: d, cache: cache}
}

func (ur *userRepository) Create(ctx context.Context, u domain.User) error {
	return ur.dao.Insert(ctx, dao.User{
		Email:    sql.NullString{String: u.Email, Valid: u.Email != ""},
		Phone:    sql.NullString{String: u.Phone, Valid: u.Phone != ""},
		Password: u.Password,
	})
}

func (ur *userRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := ur.dao.FindByEmail(ctx, email)
	return ur.entityToDomain(u), err
}

func (ur *userRepository) Update(ctx context.Context, u domain.User) error {
	if err := ur.dao.Update(ctx, ur.domainToEntity(u)); err != nil {
		return err
	}
	return ur.cache.Delete(ctx, u.Id)
}

func (ur *userRepository) FindByID(ctx context.Context, id int64) (domain.User, error) {
	u, err := ur.cache.Get(ctx, id)
	if err == nil {
		return u, err
	}
	ue, err := ur.dao.FindByID(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	u = ur.entityToDomain(ue)
	_ = ur.cache.Set(ctx, u)
	return u, nil
}

func (ur *userRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := ur.dao.FindByPhone(ctx, phone)
	return ur.entityToDomain(u), err
}

func (ur *userRepository) entityToDomain(ue dao.User) domain.User {
	var birthday time.Time
	if ue.Birthday.Valid {
		birthday = time.UnixMilli(ue.Birthday.Int64)
	}
	return domain.User{
		Id:       ue.Id,
		Email:    ue.Email.String,
		Password: ue.Password,
		Phone:    ue.Phone.String,
		Nickname: ue.Nickname.String,
		AboutMe:  ue.AboutMe.String,
		Birthday: birthday,
		CreateAt: time.UnixMilli(ue.CreateAt),
	}
}

func (ur *userRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id:       u.Id,
		Email:    sql.NullString{String: u.Email, Valid: u.Email != ""},
		Phone:    sql.NullString{String: u.Phone, Valid: u.Phone != ""},
		Birthday: sql.NullInt64{Int64: u.Birthday.UnixMilli(), Valid: !u.Birthday.IsZero()},
		Nickname: sql.NullString{String: u.Nickname, Valid: u.Nickname != ""},
		AboutMe:  sql.NullString{String: u.AboutMe, Valid: u.AboutMe != ""},
		Password: u.Password,
	}
}
