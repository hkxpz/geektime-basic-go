package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"geektime-basic-go/webook/pkg/migrator"
	"geektime-basic-go/webook/pkg/migrator/events"
	evtmocks "geektime-basic-go/webook/pkg/migrator/events/mocks"
	"geektime-basic-go/webook/pkg/migrator/integration/startup"
	"geektime-basic-go/webook/pkg/migrator/validator"
)

type InteractiveTestSuite struct {
	suite.Suite
	srcDB  *gorm.DB
	intrDB *gorm.DB
}

func (i *InteractiveTestSuite) TestValidator() {
	testCases := []struct {
		name    string
		before  func(t *testing.T)
		after   func(t *testing.T)
		mock    func(ctrl *gomock.Controller) events.Producer
		wantErr error
	}{
		{
			name: "src 有 intr 没有",
			before: func(t *testing.T) {
				err := i.srcDB.Create(&Interactives{ID: 1, BizId: 123, Biz: "test", ReadCnt: 1, LikeCnt: 1}).Error
				require.NoError(t, err)
			},
			after: func(t *testing.T) {
				i.TearDownTest()
			},
			mock: func(ctrl *gomock.Controller) events.Producer {
				p := evtmocks.NewMockProducer(ctrl)
				p.EXPECT().ProduceInconsistentEvent(gomock.Any(), events.InconsistentEvent{
					ID: 1, Direction: "src", Type: events.InconsistentEventTypeTargetMissing,
				}).Return(nil)
				return p
			},
		},
		{
			name: "src 和 intr 数据相同",
			before: func(t *testing.T) {
				intr := &Interactives{ID: 1, BizId: 123, Biz: "test", ReadCnt: 1, LikeCnt: 1}
				err := i.srcDB.Create(intr).Error
				require.NoError(t, err)
				err = i.intrDB.Create(intr).Error
				require.NoError(t, err)
			},
			after: func(t *testing.T) {
				i.TearDownTest()
			},
			mock: func(ctrl *gomock.Controller) events.Producer {
				return evtmocks.NewMockProducer(ctrl)
			},
		},
		{
			name: "src 和 intr 数据不相同",
			before: func(t *testing.T) {
				src := &Interactives{ID: 1, BizId: 123, Biz: "test", ReadCnt: 111, LikeCnt: 1}
				intr := &Interactives{ID: 1, BizId: 123, Biz: "test", ReadCnt: 1, LikeCnt: 1}
				err := i.srcDB.Create(src).Error
				require.NoError(t, err)
				err = i.intrDB.Create(intr).Error
				require.NoError(t, err)
			},
			after: func(t *testing.T) {
				i.TearDownTest()
			},
			mock: func(ctrl *gomock.Controller) events.Producer {
				p := evtmocks.NewMockProducer(ctrl)
				p.EXPECT().ProduceInconsistentEvent(gomock.Any(), events.InconsistentEvent{
					ID: 1, Direction: "src", Type: events.InconsistentEventTypeNotEqual,
				}).Return(nil)
				return p
			},
		},
		{
			name: "src 没有 intr 有",
			before: func(t *testing.T) {
				err := i.intrDB.Create(&Interactives{ID: 1, BizId: 123, Biz: "test", ReadCnt: 111, LikeCnt: 1}).Error
				require.NoError(t, err)
			},
			after: func(t *testing.T) {
				i.TearDownTest()
			},
			mock: func(ctrl *gomock.Controller) events.Producer {
				p := evtmocks.NewMockProducer(ctrl)
				p.EXPECT().ProduceInconsistentEvent(gomock.Any(), events.InconsistentEvent{
					ID: 1, Direction: "src", Type: events.InconsistentEventTypeBaseMissing,
				}).Return(nil)
				return p
			},
		},
	}

	l := startup.InitZapLogger()
	t := i.T()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			tc.before(t)
			v := validator.NewValidator[Interactives](i.srcDB, i.intrDB, "src", l, tc.mock(ctrl))
			require.Equal(t, tc.wantErr, v.Validate(context.Background()))
			tc.after(t)
		})
	}
}

func (i *InteractiveTestSuite) SetupSuite() {
	startup.InitViper()
	i.srcDB = startup.InitSrcDB()
	require.NoError(i.T(), i.srcDB.AutoMigrate(&Interactives{}))
	i.intrDB = startup.InitIntrDB()
	require.NoError(i.T(), i.intrDB.AutoMigrate(&Interactives{}))
}

func (i *InteractiveTestSuite) TearDownTest() {
	require.NoError(i.T(), i.srcDB.Exec("TRUNCATE TABLE interactives").Error)
	require.NoError(i.T(), i.intrDB.Exec("TRUNCATE TABLE interactives").Error)
}

func TestInteractive(t *testing.T) {
	suite.Run(t, &InteractiveTestSuite{})
}

type Interactives struct {
	ID         int64  `gorm:"primaryKey,autoIncrement"`
	BizId      int64  `gorm:"uniqueIndex:biz_type_id"`
	Biz        string `gorm:"type:varchar(128);uniqueIndex:biz_type_id"`
	ReadCnt    int64
	CollectCnt int64
	LikeCnt    int64
	CreateAt   int64
	UpdateAt   int64
}

func (i Interactives) Id() int64 {
	return i.ID
}

func (i Interactives) TableName() string {
	return "interactives"
}

func (i Interactives) CompareTo(entity migrator.Entity) bool {
	return i == entity.(migrator.Entity)
}
