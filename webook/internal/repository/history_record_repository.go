package repository

import (
	"context"

	"geektime-basic-go/webook/internal/domain"
)

type HistoryRecordRepository interface {
	AddRecord(ctx context.Context, r domain.HistoryRecord) error
}
