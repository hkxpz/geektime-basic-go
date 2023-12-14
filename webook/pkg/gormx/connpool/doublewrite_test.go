package connpool

import (
	"testing"

	"github.com/ecodeclub/ekit/syncx/atomicx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"geektime-basic-go/webook/pkg/gormx/connpool/integration/startup"
)

type DoubleWriteTestSuite struct {
	suite.Suite
	db *gorm.DB
}

func (d *DoubleWriteTestSuite) SetupSuite() {
	t := d.T()
	startup.InitViper()
	src := startup.InitSrcDB()
	err := src.AutoMigrate(&Interactive{})
	require.NoError(t, err)
	dst := startup.InitIntrDB()
	err = dst.AutoMigrate(&Interactive{})
	require.NoError(t, err)
	doubleWrite, err := gorm.Open(mysql.New(mysql.Config{
		Conn: &DoubleWritePool{
			src:     src.ConnPool,
			dst:     dst.ConnPool,
			pattern: atomicx.NewValueOf(PatternSrcFirst),
		},
	}))
	require.NoError(t, err)
	d.db = doubleWrite
}

func (d *DoubleWriteTestSuite) TearDownTest() {
	d.db.Exec("TRUNCATE TABLE interactives")
}

// 集成测试，需要启动数据库
func (d *DoubleWriteTestSuite) TestDoubleWriteTest() {
	t := d.T()
	err := d.db.Create(&Interactive{
		Biz:   "test",
		BizID: 10086,
	}).Error
	assert.NoError(t, err)
	// 查询数据库就可以看到对应的数据
}

func (d *DoubleWriteTestSuite) TestDoubleWriteTransaction() {
	t := d.T()
	err := d.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&Interactive{
			Biz:   "test",
			BizID: 10087,
		}).Error
	})
	require.NoError(t, err)
}

func TestDoubleWrite(t *testing.T) {
	suite.Run(t, new(DoubleWriteTestSuite))
}

type Interactive struct {
	ID         int64  `gorm:"primaryKey,autoIncrement"`
	BizID      int64  `gorm:"uniqueIndex:biz_type_id"`
	Biz        string `gorm:"type:varchar(128);uniqueIndex:biz_type_id"`
	ReadCnt    int64
	CollectCnt int64
	LikeCnt    int64
	CreateAt   int64
	UpdateAt   int64
}
