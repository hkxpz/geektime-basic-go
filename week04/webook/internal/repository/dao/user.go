package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

const (
	uniqueIndexErrNo uint16 = 1062
	dataTooLongErrNo uint16 = 1406
)

// ErrDataNotFound 通用的数据没找到
var ErrDataNotFound = gorm.ErrRecordNotFound

var (
	ErrUserDuplicateEmail    = errors.New("邮箱冲突")
	ErrUserDuplicateNickname = errors.New("昵称冲突")
	ErrDataTooLong           = errors.New("数据太长")
)

type UserDAO struct {
	db *gorm.DB
}

type User struct {
	Id           int64         `gorm:"primaryKey,autoIncrement"`
	Email        string        `gorm:"unique;size:50;comment:邮箱"`
	Password     string        `gorm:"comment:密码"`
	Nickname     string        `gorm:"size:30;comment:昵称"`
	Birthday     sql.NullInt64 `gorm:"size:10;comment:生日"`
	Introduction string        `gorm:"size:150;comment:个人介绍"`
	CreateAt     int64         `gorm:"comment:创建时间"`
	UpdateAt     int64         `gorm:"comment:更新时间"`
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{db: db}
}

func (ud *UserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.CreateAt, u.UpdateAt = now, now

	err := ud.db.WithContext(ctx).Create(&u).Error
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		if me.Number == uniqueIndexErrNo {
			return ErrUserDuplicateEmail
		}
	}
	return err
}

func (ud *UserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).Find(&u, "email = ?", email).Error
	return u, err
}

func (ud *UserDAO) Update(ctx context.Context, user User) error {
	user.UpdateAt = time.Now().UnixMilli()
	err := ud.db.WithContext(ctx).Updates(&user).Error
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		switch me.Number {
		case uniqueIndexErrNo:
			return ErrUserDuplicateNickname
		case dataTooLongErrNo:
			return ErrDataTooLong
		}
	}

	return err
}

func (ud *UserDAO) FindByID(ctx context.Context, id int64) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).Find(&u, "id = ?", id).Error
	if err == nil && u.Id == 0 {
		return u, ErrDataNotFound
	}
	return u, err
}
