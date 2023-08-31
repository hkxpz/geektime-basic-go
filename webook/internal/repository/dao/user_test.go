package dao

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func initDB(db *sql.DB) (*gorm.DB, error) {
	return gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		DisableAutomaticPing:   true,
		SkipDefaultTransaction: true,
	})
}

func TestGormUserDAO_Insert(t *testing.T) {
	userDao := User{
		Email:    sql.NullString{String: "123@qq.com", Valid: true},
		Password: "$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi",
		Phone:    sql.NullString{String: "13888888888", Valid: true},
	}
	testCases := []struct {
		name    string
		sqlmock func(t *testing.T) *sql.DB

		ctx  context.Context
		user User

		wantErr error
	}{
		{
			name: "插入成功",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mockRes := sqlmock.NewResult(1, 1)
				mock.ExpectExec("INSERT INTO `users` .*").WillReturnResult(mockRes)
				return db
			},
			ctx:  context.Background(),
			user: userDao,
		},
		{
			name: "插入失败-邮箱冲突",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec("INSERT INTO `users` .*").WillReturnError(&mysqlDriver.MySQLError{Number: 1062})
				return db
			},
			ctx:     context.Background(),
			user:    userDao,
			wantErr: ErrUserDuplicate,
		},
		{
			name: "插入失败",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec("INSERT INTO `users` .*").WillReturnError(errors.New("模拟插入失败"))
				return db
			},
			ctx:     context.Background(),
			user:    userDao,
			wantErr: errors.New("模拟插入失败"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sqlmockDB := tc.sqlmock(t)
			db, err := initDB(sqlmockDB)
			require.NoError(t, err)
			dao := NewUserDAO(db)
			err = dao.Insert(tc.ctx, tc.user)
			require.Equal(t, tc.wantErr, err)
		})
	}
}

func TestGormUserDAO_Update(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	userDao := User{
		Id:       1,
		Email:    sql.NullString{String: "123@qq.com", Valid: true},
		Password: "$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi",
		Phone:    sql.NullString{String: "13888888888", Valid: true},
		Nickname: sql.NullString{String: "泰裤辣", Valid: true},
		AboutMe:  sql.NullString{String: "泰裤辣", Valid: true},
		Birthday: sql.NullInt64{Int64: nowMs, Valid: true},
	}
	testCases := []struct {
		name    string
		sqlmock func(t *testing.T) *sql.DB

		ctx  context.Context
		user User

		wantErr error
	}{
		{
			name: "更新成功",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mockRes := sqlmock.NewResult(1, 1)
				mock.ExpectExec("UPDATE `users` SET .*").WillReturnResult(mockRes)
				return db
			},
			ctx:  context.Background(),
			user: userDao,
		},
		{
			name: "更新失败",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec("UPDATE `users` SET .*").WillReturnError(errors.New("模拟更新失败"))
				return db
			},
			ctx:     context.Background(),
			user:    userDao,
			wantErr: errors.New("模拟更新失败"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sqlmockDB := tc.sqlmock(t)
			db, err := initDB(sqlmockDB)
			require.NoError(t, err)
			dao := NewUserDAO(db)
			err = dao.Update(tc.ctx, tc.user)
			require.Equal(t, tc.wantErr, err)
		})
	}
}

func TestGormUserDAO_FindByID(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	userDao := User{
		Id:       1,
		Email:    sql.NullString{String: "123@qq.com", Valid: true},
		Password: "$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi",
		Phone:    sql.NullString{String: "13888888888", Valid: true},
		Nickname: sql.NullString{String: "泰裤辣", Valid: true},
		AboutMe:  sql.NullString{String: "泰裤辣", Valid: true},
		Birthday: sql.NullInt64{Int64: nowMs, Valid: true},
		CreateAt: nowMs,
		UpdateAt: nowMs,
	}
	userColumns := []string{"id", "email", "password", "phone", "nickname", "birthday", "about_me", "create_at", "update_at"}
	userRows := [][]driver.Value{
		{
			1,
			"123@qq.com",
			"$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi",
			"13888888888",
			"泰裤辣",
			nowMs,
			"泰裤辣",
			nowMs,
			nowMs,
		},
	}
	testCases := []struct {
		name    string
		sqlmock func(t *testing.T) *sql.DB

		ctx context.Context
		id  int64

		wantUser User
		wantErr  error
	}{
		{
			name: "找到用户",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				rows := sqlmock.NewRows(userColumns).AddRow(userRows[0]...)
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
				return db
			},
			ctx:      context.Background(),
			id:       1,
			wantUser: userDao,
		},
		{
			name: "未找到用户",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectQuery("SELECT").WillReturnError(ErrDataNotFound)
				return db
			},
			ctx:     context.Background(),
			id:      1,
			wantErr: ErrDataNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sqlmockDB := tc.sqlmock(t)
			db, err := initDB(sqlmockDB)
			require.NoError(t, err)
			dao := NewUserDAO(db)
			user, err := dao.FindByID(tc.ctx, tc.id)
			require.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}

func TestGormUserDAO_FindByEmail(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	userDao := User{
		Id:       1,
		Email:    sql.NullString{String: "123@qq.com", Valid: true},
		Password: "$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi",
		Phone:    sql.NullString{String: "13888888888", Valid: true},
		Nickname: sql.NullString{String: "泰裤辣", Valid: true},
		AboutMe:  sql.NullString{String: "泰裤辣", Valid: true},
		Birthday: sql.NullInt64{Int64: nowMs, Valid: true},
		CreateAt: nowMs,
		UpdateAt: nowMs,
	}
	userColumns := []string{"id", "email", "password", "phone", "nickname", "birthday", "about_me", "create_at", "update_at"}
	userRows := [][]driver.Value{
		{
			1,
			"123@qq.com",
			"$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi",
			"13888888888",
			"泰裤辣",
			nowMs,
			"泰裤辣",
			nowMs,
			nowMs,
		},
	}
	testCases := []struct {
		name    string
		sqlmock func(t *testing.T) *sql.DB

		ctx   context.Context
		email string

		wantUser User
		wantErr  error
	}{
		{
			name: "找到用户",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				rows := sqlmock.NewRows(userColumns).AddRow(userRows[0]...)
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
				return db
			},
			ctx:      context.Background(),
			email:    "123@qq.com",
			wantUser: userDao,
		},
		{
			name: "未找到用户",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectQuery("SELECT").WillReturnError(ErrDataNotFound)
				return db
			},
			ctx:     context.Background(),
			email:   "123@qq.com",
			wantErr: ErrDataNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sqlmockDB := tc.sqlmock(t)
			db, err := initDB(sqlmockDB)
			require.NoError(t, err)
			dao := NewUserDAO(db)
			user, err := dao.FindByEmail(tc.ctx, tc.email)
			require.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}

func TestGormUserDAO_FindByPhone(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	userDao := User{
		Id:       1,
		Email:    sql.NullString{String: "123@qq.com", Valid: true},
		Password: "$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi",
		Phone:    sql.NullString{String: "13888888888", Valid: true},
		Nickname: sql.NullString{String: "泰裤辣", Valid: true},
		AboutMe:  sql.NullString{String: "泰裤辣", Valid: true},
		Birthday: sql.NullInt64{Int64: nowMs, Valid: true},
		CreateAt: nowMs,
		UpdateAt: nowMs,
	}
	userColumns := []string{"id", "email", "password", "phone", "nickname", "birthday", "about_me", "create_at", "update_at"}
	userRows := [][]driver.Value{
		{
			1,
			"123@qq.com",
			"$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi",
			"13888888888",
			"泰裤辣",
			nowMs,
			"泰裤辣",
			nowMs,
			nowMs,
		},
	}
	testCases := []struct {
		name    string
		sqlmock func(t *testing.T) *sql.DB

		ctx   context.Context
		phone string

		wantUser User
		wantErr  error
	}{
		{
			name: "找到用户",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				rows := sqlmock.NewRows(userColumns).AddRow(userRows[0]...)
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
				return db
			},
			ctx:      context.Background(),
			phone:    "13888888888",
			wantUser: userDao,
		},
		{
			name: "未找到用户",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectQuery("SELECT").WillReturnError(ErrDataNotFound)
				return db
			},
			ctx:     context.Background(),
			phone:   "13888888888",
			wantErr: ErrDataNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sqlmockDB := tc.sqlmock(t)
			db, err := initDB(sqlmockDB)
			require.NoError(t, err)
			dao := NewUserDAO(db)
			user, err := dao.FindByPhone(tc.ctx, tc.phone)
			require.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}
