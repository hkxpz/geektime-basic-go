package service

import (
	"context"
	"math"
	"time"

	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository"
)

type RankingService interface {
	RankTopN(ctx context.Context) error
	TopN(ctx context.Context) ([]domain.Article, error)
}

type batchRankingService struct {
	artSvc    ArticleService
	intrSvc   InteractiveService
	repo      repository.RankingRepository
	BatchSize int
	N         int
	soreFunc  func(likeCnt int64, updateAt time.Time) float64
}

func NewBatchRankingService(artSvc ArticleService, intrSvc InteractiveService, repo repository.RankingRepository) RankingService {
	svc := &batchRankingService{artSvc: artSvc, intrSvc: intrSvc, repo: repo, BatchSize: 100, N: 100}
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

	for {
		arts, err := svc.artSvc.ListPub(ctx, now, offset, svc.BatchSize)
		if err != nil {
			return nil, err
		}
		artIDs := slice.Map(arts, func(idx int, src domain.Article) int64 {
			return src.ID
		})
		intrMap, err := svc.intrSvc.GetByIDs(ctx, "article", artIDs)
		if err != nil {
			return nil, err
		}

		minScore := float64(0)
		for _, art := range arts {
			intr, ok := intrMap[art.ID]
			if !ok {
				continue
			}

			score := svc.soreFunc(intr.LikeCnt, art.UpdateAt)
			// 当前分数小于最小分数, 不做处理
			if score < minScore {
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

			// 队列最小值大于或等于当前值, 不做处理
			if queMinScore.score >= score {
				minScore = score
				continue
			}
			// 出队, 当前值入队
			minScore = queMinScore.score
			_, _ = que.Dequeue()
			_ = que.Enqueue(ele)
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
