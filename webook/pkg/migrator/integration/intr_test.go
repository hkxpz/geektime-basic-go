package integration

import (
	"context"
	"testing"
	"time"

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

func (i *InteractiveTestSuite) TestValidator() {
	now := time.Now().UnixMilli()
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
				data := []Interactives{
					{ID: 1, BizId: 123, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
					{ID: 2, BizId: 456, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
					{ID: 3, BizId: 789, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
				}
				err := i.srcDB.CreateInBatches(data, 3).Error
				require.NoError(t, err)
			},
			after: func(t *testing.T) { i.TearDownTest() },
			mock: func(ctrl *gomock.Controller) events.Producer {
				p := evtmocks.NewMockProducer(ctrl)
				p.EXPECT().ProduceInconsistentEvent(gomock.Any(), events.InconsistentEvent{ID: 1, Direction: "src", Type: events.InconsistentEventTypeTargetMissing}).Return(nil)
				p.EXPECT().ProduceInconsistentEvent(gomock.Any(), events.InconsistentEvent{ID: 2, Direction: "src", Type: events.InconsistentEventTypeTargetMissing}).Return(nil)
				p.EXPECT().ProduceInconsistentEvent(gomock.Any(), events.InconsistentEvent{ID: 3, Direction: "src", Type: events.InconsistentEventTypeTargetMissing}).Return(nil)
				return p
			},
		},
		{
			name: "src 和 intr 数据相同",
			before: func(t *testing.T) {
				data := []Interactives{
					{ID: 1, BizId: 123, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
					{ID: 2, BizId: 456, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
					{ID: 3, BizId: 789, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
				}
				err := i.srcDB.CreateInBatches(data, 3).Error
				require.NoError(t, err)
				err = i.intrDB.CreateInBatches(data, 3).Error
				require.NoError(t, err)
			},
			after: func(t *testing.T) { i.TearDownTest() },
			mock: func(ctrl *gomock.Controller) events.Producer {
				return evtmocks.NewMockProducer(ctrl)
			},
		},
		{
			name: "src 和 intr 数据不相同",
			before: func(t *testing.T) {
				src := []Interactives{
					{ID: 1, BizId: 123, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
					{ID: 2, BizId: 456, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
					{ID: 3, BizId: 789, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
				}
				dst := []Interactives{
					{ID: 1, BizId: 123, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
					{ID: 2, BizId: 456, Biz: "test", ReadCnt: 122, LikeCnt: 1, CreateAt: now, UpdateAt: now},
					{ID: 3, BizId: 789, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
				}
				err := i.srcDB.CreateInBatches(src, 3).Error
				require.NoError(t, err)
				err = i.intrDB.CreateInBatches(dst, 3).Error
				require.NoError(t, err)
			},
			after: func(t *testing.T) { i.TearDownTest() },
			mock: func(ctrl *gomock.Controller) events.Producer {
				p := evtmocks.NewMockProducer(ctrl)
				p.EXPECT().ProduceInconsistentEvent(gomock.Any(), events.InconsistentEvent{ID: 2, Direction: "src", Type: events.InconsistentEventTypeNotEqual}).Return(nil)
				return p
			},
		},
		{
			name: "src 没有 intr 有",
			before: func(t *testing.T) {
				data := []Interactives{
					{ID: 1, BizId: 123, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
					{ID: 2, BizId: 456, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
					{ID: 3, BizId: 789, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
				}
				err := i.intrDB.CreateInBatches(data, 3).Error
				require.NoError(t, err)
			},
			after: func(t *testing.T) { i.TearDownTest() },
			mock: func(ctrl *gomock.Controller) events.Producer {
				p := evtmocks.NewMockProducer(ctrl)
				p.EXPECT().ProduceInconsistentEvent(gomock.Any(), events.InconsistentEvent{ID: 1, Direction: "src", Type: events.InconsistentEventTypeBaseMissing}).Return(nil)
				p.EXPECT().ProduceInconsistentEvent(gomock.Any(), events.InconsistentEvent{ID: 2, Direction: "src", Type: events.InconsistentEventTypeBaseMissing}).Return(nil)
				p.EXPECT().ProduceInconsistentEvent(gomock.Any(), events.InconsistentEvent{ID: 3, Direction: "src", Type: events.InconsistentEventTypeBaseMissing}).Return(nil)
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

func (i *InteractiveTestSuite) TestValidatorBatch() {
	now := time.Now().UnixMilli()
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
				data := []Interactives{
					{ID: 1, BizId: 123, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
					{ID: 2, BizId: 456, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
					{ID: 3, BizId: 789, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
				}
				err := i.srcDB.CreateInBatches(data, 3).Error
				require.NoError(t, err)
			},
			after: func(t *testing.T) { i.TearDownTest() },
			mock: func(ctrl *gomock.Controller) events.Producer {
				p := evtmocks.NewMockProducer(ctrl)
				p.EXPECT().ProduceInconsistentEvent(gomock.Any(), events.InconsistentEvent{ID: 1, Direction: "src", Type: events.InconsistentEventTypeTargetMissing}).Return(nil)
				p.EXPECT().ProduceInconsistentEvent(gomock.Any(), events.InconsistentEvent{ID: 2, Direction: "src", Type: events.InconsistentEventTypeTargetMissing}).Return(nil)
				p.EXPECT().ProduceInconsistentEvent(gomock.Any(), events.InconsistentEvent{ID: 3, Direction: "src", Type: events.InconsistentEventTypeTargetMissing}).Return(nil)
				return p
			},
		},
		{
			name: "src 和 intr 数据相同",
			before: func(t *testing.T) {
				data := []Interactives{
					{ID: 1, BizId: 123, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
					{ID: 2, BizId: 456, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
					{ID: 3, BizId: 789, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
				}
				err := i.srcDB.CreateInBatches(data, 3).Error
				require.NoError(t, err)
				err = i.intrDB.CreateInBatches(data, 3).Error
				require.NoError(t, err)
			},
			after: func(t *testing.T) { i.TearDownTest() },
			mock: func(ctrl *gomock.Controller) events.Producer {
				return evtmocks.NewMockProducer(ctrl)
			},
		},
		{
			name: "src 和 intr 数据不相同",
			before: func(t *testing.T) {
				src := []Interactives{
					{ID: 1, BizId: 123, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
					{ID: 2, BizId: 456, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
					{ID: 3, BizId: 789, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
				}
				dst := []Interactives{
					{ID: 1, BizId: 123, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
					{ID: 2, BizId: 456, Biz: "test", ReadCnt: 122, LikeCnt: 1, CreateAt: now, UpdateAt: now},
					{ID: 3, BizId: 789, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
				}
				err := i.srcDB.CreateInBatches(src, 3).Error
				require.NoError(t, err)
				err = i.intrDB.CreateInBatches(dst, 3).Error
				require.NoError(t, err)
			},
			after: func(t *testing.T) { i.TearDownTest() },
			mock: func(ctrl *gomock.Controller) events.Producer {
				p := evtmocks.NewMockProducer(ctrl)
				p.EXPECT().ProduceInconsistentEvent(gomock.Any(), events.InconsistentEvent{ID: 2, Direction: "src", Type: events.InconsistentEventTypeNotEqual}).Return(nil)
				return p
			},
		},
		{
			name: "src 没有 intr 有",
			before: func(t *testing.T) {
				data := []Interactives{
					{ID: 1, BizId: 123, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
					{ID: 2, BizId: 456, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
					{ID: 3, BizId: 789, Biz: "test", ReadCnt: 1, LikeCnt: 1, CreateAt: now, UpdateAt: now},
				}
				err := i.intrDB.CreateInBatches(data, 3).Error
				require.NoError(t, err)
			},
			after: func(t *testing.T) { i.TearDownTest() },
			mock: func(ctrl *gomock.Controller) events.Producer {
				p := evtmocks.NewMockProducer(ctrl)
				p.EXPECT().ProduceInconsistentEvent(gomock.Any(), events.InconsistentEvent{ID: 1, Direction: "src", Type: events.InconsistentEventTypeBaseMissing}).Return(nil)
				p.EXPECT().ProduceInconsistentEvent(gomock.Any(), events.InconsistentEvent{ID: 2, Direction: "src", Type: events.InconsistentEventTypeBaseMissing}).Return(nil)
				p.EXPECT().ProduceInconsistentEvent(gomock.Any(), events.InconsistentEvent{ID: 3, Direction: "src", Type: events.InconsistentEventTypeBaseMissing}).Return(nil)
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
			v.SetBatchSize(100)
			require.Equal(t, tc.wantErr, v.Validate(context.Background()))
			tc.after(t)
		})
	}
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
