package article

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"geektime-basic-go/webook/internal/domain"
)

func nweMockDB(db *sql.DB) (*gorm.DB, error) {
	return gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		DisableAutomaticPing:   true,
		SkipDefaultTransaction: true,
	})
}

func TestGormDAO_Insert(t *testing.T) {
	testCases := []struct {
		name    string
		sqlmock func(t *testing.T) *sql.DB
		user    Article

		wantID  int64
		wantErr error
	}{
		{
			name: "插入成功",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mockRes := sqlmock.NewResult(1, 1)
				mock.ExpectExec("INSERT INTO `articles` .*").WillReturnResult(mockRes)
				return db
			},
			user: Article{
				Title:    "新的文章",
				Content:  "文章内容",
				AuthorID: 123,
				Status:   domain.ArticleStatusUnpublished.ToUint8(),
			},
		},
		{
			name: "插入失败",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec("INSERT INTO `articles` .*").WillReturnError(errors.New("模拟失败"))
				return db
			},
			user: Article{
				Title:    "新的文章",
				Content:  "文章内容",
				AuthorID: 123,
				Status:   domain.ArticleStatusUnpublished.ToUint8(),
			},
			wantErr: errors.New("模拟失败"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sqlmockDB := tc.sqlmock(t)
			db, err := nweMockDB(sqlmockDB)
			require.NoError(t, err)
			dao := NewGormArticleDAO(db)
			_, err = dao.Insert(context.Background(), tc.user)
			require.Equal(t, tc.wantErr, err)
		})
	}
}

func TestGormDAO_UpdateById(t *testing.T) {
	testCases := []struct {
		name    string
		sqlmock func(t *testing.T) *sql.DB
		user    Article

		wantID  int64
		wantErr error
	}{
		{
			name: "更新成功",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mockRes := sqlmock.NewResult(1, 1)
				mock.ExpectExec("UPDATE").WillReturnResult(mockRes)
				return db
			},
			user: Article{
				Title:    "新的文章",
				Content:  "文章内容",
				AuthorID: 123,
				Status:   domain.ArticleStatusUnpublished.ToUint8(),
			},
		},
		{
			name: "更新别人的文章",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mockRes := sqlmock.NewResult(1, 0)
				mock.ExpectExec("UPDATE").WillReturnResult(mockRes)
				return db
			},
			user: Article{
				Title:    "新的文章",
				Content:  "文章内容",
				AuthorID: 456,
				Status:   domain.ArticleStatusUnpublished.ToUint8(),
			},
			wantErr: errors.New("更新数据失败"),
		},
		{
			name: "更新失败",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec("UPDATE").WillReturnError(errors.New("模拟失败"))
				return db
			},
			user: Article{
				Title:    "新的文章",
				Content:  "文章内容",
				AuthorID: 123,
				Status:   domain.ArticleStatusUnpublished.ToUint8(),
			},
			wantErr: errors.New("模拟失败"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sqlmockDB := tc.sqlmock(t)
			db, err := nweMockDB(sqlmockDB)
			require.NoError(t, err)
			dao := NewGormArticleDAO(db)
			err = dao.UpdateById(context.Background(), tc.user)
			require.Equal(t, tc.wantErr, err)
		})
	}
}

func TestGormDAO_Sync(t *testing.T) {
	testCases := []struct {
		name    string
		sqlmock func(t *testing.T) *sql.DB
		user    Article

		wantID  int64
		wantErr error
	}{
		{
			name: "新建并发布成功",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mockRes := sqlmock.NewResult(1, 1)
				mock.ExpectBegin()
				mock.ExpectExec("INSERT").WillReturnResult(mockRes)
				mock.ExpectExec("INSERT").WillReturnResult(mockRes)
				mock.ExpectCommit()
				return db
			},
			user: Article{
				Title:    "新的文章",
				Content:  "文章内容",
				AuthorID: 123,
				Status:   domain.ArticleStatusUnpublished.ToUint8(),
			},
		},
		{
			name: "更新并发布成功",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mockRes := sqlmock.NewResult(1, 1)
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE").WillReturnResult(mockRes)
				mock.ExpectExec("UPDATE").WillReturnResult(mockRes)
				mock.ExpectCommit()
				return db
			},
			user: Article{
				ID:       1,
				Title:    "新的文章",
				Content:  "文章内容",
				AuthorID: 123,
				Status:   domain.ArticleStatusUnpublished.ToUint8(),
			},
		},
		{
			name: "新建失败",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectBegin()
				mock.ExpectExec("INSERT").WillReturnError(errors.New("模拟新建失败"))
				mock.ExpectRollback()
				return db
			},
			user: Article{
				Title:    "新的文章",
				Content:  "文章内容",
				AuthorID: 123,
				Status:   domain.ArticleStatusUnpublished.ToUint8(),
			},
			wantErr: errors.New("模拟新建失败"),
		},
		{
			name: "更新失败",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE").WillReturnError(errors.New("模拟更新失败"))
				mock.ExpectRollback()
				return db
			},
			user: Article{
				ID:       1,
				Title:    "新的文章",
				Content:  "文章内容",
				AuthorID: 123,
				Status:   domain.ArticleStatusUnpublished.ToUint8(),
			},
			wantErr: errors.New("模拟更新失败"),
		},
		{
			name: "新建成功但发布失败",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mockRes := sqlmock.NewResult(1, 1)
				mock.ExpectBegin()
				mock.ExpectExec("INSERT").WillReturnResult(mockRes)
				mock.ExpectExec("INSERT").WillReturnError(errors.New("模拟发布失败"))
				mock.ExpectRollback()
				return db
			},
			user: Article{
				Title:    "新的文章",
				Content:  "文章内容",
				AuthorID: 123,
				Status:   domain.ArticleStatusUnpublished.ToUint8(),
			},
			wantErr: errors.New("模拟发布失败"),
		},
		{
			name: "更新成功但发布失败",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mockRes := sqlmock.NewResult(0, 1)
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE").WillReturnResult(mockRes)
				mock.ExpectExec("UPDATE").WillReturnError(errors.New("模拟发布失败"))
				mock.ExpectRollback()
				return db
			},
			user: Article{
				ID:       1,
				Title:    "新的文章",
				Content:  "文章内容",
				AuthorID: 123,
				Status:   domain.ArticleStatusUnpublished.ToUint8(),
			},
			wantErr: errors.New("模拟发布失败"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sqlmockDB := tc.sqlmock(t)
			db, err := nweMockDB(sqlmockDB)
			require.NoError(t, err)
			dao := NewGormArticleDAO(db)
			_, err = dao.Sync(context.Background(), tc.user)
			require.Equal(t, tc.wantErr, err)
		})
	}
}
