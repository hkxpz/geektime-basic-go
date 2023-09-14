package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

const uniqueIndexErrNo uint16 = 1062

// ErrDataNotFound 通用的数据没找到
var ErrDataNotFound = gorm.ErrRecordNotFound

var ErrUserDuplicate = errors.New("用户邮箱或者手机号冲突")

type UserDAO interface {
	Insert(ctx context.Context, u User) error
	Update(ctx context.Context, u User) error
	FindByID(ctx context.Context, id int64) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	FindByWechat(ctx context.Context, openID string) (User, error)
}

type gormUserDAO struct {
	db *gorm.DB
}

type User struct {
	ID       int64          `gorm:"primaryKey,autoIncrement"`
	Email    sql.NullString `gorm:"unique;size:50;comment:邮箱"`
	Password string         `gorm:"comment:密码"`
	Phone    sql.NullString `gorm:"unique;comment:电话"`
	Nickname sql.NullString `gorm:"size:30;comment:昵称"`
	Birthday sql.NullInt64  `gorm:"comment:生日"`
	AboutMe  sql.NullString `gorm:"type=varchar(1024);comment:个人介绍"`

	WechatOpenID  sql.NullString `gorm:"type=varchar(1024),unique"`
	WechatUnionID sql.NullString `gorm:"type=varchar(1024)"`

	CreateAt int64 `gorm:"comment:创建时间"`
	UpdateAt int64 `gorm:"comment:更新时间"`
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &gormUserDAO{db: db}
}

func (ud *gormUserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.CreateAt, u.UpdateAt = now, now

	err := ud.db.WithContext(ctx).Create(&u).Error
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		if me.Number == uniqueIndexErrNo {
			return ErrUserDuplicate
		}
	}
	return err
}

func (ud *gormUserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).Find(&u, "email = ?", email).Error
	return u, err
}

func (ud *gormUserDAO) Update(ctx context.Context, u User) error {
	u.UpdateAt = time.Now().UnixMilli()
	return ud.db.WithContext(ctx).Updates(&u).Error
}

func (ud *gormUserDAO) FindByID(ctx context.Context, id int64) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).Find(&u, "id = ?", id).Error
	return u, err
}

func (ud *gormUserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).First(&u, "phone = ?", phone).Error
	return u, err
}

func (ud *gormUserDAO) FindByWechat(ctx context.Context, openID string) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).First(&u, "wechat_open_id = ?", openID).Error
	return u, err
}
