package service

import (
	"context"
	"math"
	"time"

	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"

	intr "geektime-basic-go/webook/api/proto/gen/interactive"
	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository"
)

//go:generate mockgen -source=ranking.go -package=svcmocks -destination=mocks/ranking_mock_gen.go RankingService
type RankingService interface {
	RankTopN(ctx context.Context) error
	TopN(ctx context.Context) ([]domain.Article, error)
}

type batchRankingService struct {
	artSvc     ArticleService
	intrClient intr.InteractiveServiceClient
	repo       repository.RankingRepository
	BatchSize  int
	N          int
	soreFunc   func(likeCnt int64, updateAt time.Time) float64
}

func NewBatchRankingService(artSvc ArticleService, intrClient intr.InteractiveServiceClient, repo repository.RankingRepository) RankingService {
	svc := &batchRankingService{artSvc: artSvc, intrClient: intrClient, repo: repo, BatchSize: 100, N: 100}
	svc.soreFunc = svc.score
	return svc
}

func (svc *batchRankingService) RankTopN(ctx context.Context) error {
	arts, err := svc.rankTopN(ctx)
	if err != nil {
		return err
	}
	return svc.repo.ReplaceTopN(ctx, arts)
}

func (svc *batchRankingService) TopN(ctx context.Context) ([]domain.Article, error) {
	return svc.repo.GetTopN(ctx)
}

func (svc *batchRankingService) rankTopN(ctx context.Context) ([]domain.Article, error) {
	now := time.Now()
	// 只计算7天内, 超过7天我们可以认为不能成为热榜
	ddl := now.Add(-7 * 24 * time.Hour)
	offset := 0
	type Score struct {
		art   domain.Article
		score float64
	}
	que := queue.NewConcurrentPriorityQueue(svc.N, func(src Score, dst Score) int {
		if src.score > dst.score {
			return 1
		}
		if src.score < dst.score {
			return -1
		}
		return 0
	})

	minScore := float64(0)
	for {
		arts, err := svc.artSvc.ListPub(ctx, now, offset, svc.BatchSize)
		if err != nil {
			return nil, err
		}
		if len(arts) < 1 {
			break
		}

		artIDs := slice.Map(arts, func(idx int, src domain.Article) int64 {
			return src.ID
		})
		res, err := svc.intrClient.GetByIDs(ctx, &intr.GetByIDsRequest{Biz: "article", Ids: artIDs})
		if err != nil {
			return nil, err
		}

		intrMap := res.GetIntrs()
		for _, art := range arts {
			interactive, ok := intrMap[art.ID]
			if !ok {
				continue
			}

			score := svc.soreFunc(interactive.LikeCnt, art.UpdateAt)
			// 当前分数小于最小分数, 不做处理
			if score < minScore && que.Len() > 0 {
				continue
			}

			ele := Score{art: art, score: score}
			queMinScore, e := que.Peek()
			// 队列为空或队列不满, 直接入队
			if e != nil || que.Len() < svc.N {
				_ = que.Enqueue(ele)
				minScore = score
				if queMinScore.score < score {
					minScore = queMinScore.score
				}
				continue
			}

			// 出队, 当前值入队
			_, _ = que.Dequeue()
			_ = que.Enqueue(ele)
			val, _ := que.Peek()
			minScore = val.score
		}

		length := len(arts)
		if length == 0 || length < svc.BatchSize || arts[length-1].UpdateAt.Before(ddl) {
			break
		}
		offset += length
	}

	ql := que.Len()
	res := make([]domain.Article, ql)
	for i := ql - 1; i >= 0; i-- {
		val, _ := que.Dequeue()
		res[i] = val.art
	}
	return res, nil
}

func (svc *batchRankingService) score(likeCnt int64, UpdateAt time.Time) float64 {
	const factor = 1.5
	return float64(likeCnt-1) / math.Pow(time.Since(UpdateAt).Hours()+2, factor)
}
