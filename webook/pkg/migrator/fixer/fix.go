package fixer

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"geektime-basic-go/webook/pkg/migrator"
	"geektime-basic-go/webook/pkg/migrator/events"
)

type OverrideFixer[T migrator.Entity] struct {
	base    *gorm.DB
	target  *gorm.DB
	columns []string
}

func NewOverrideFixer[T migrator.Entity](base *gorm.DB, target *gorm.DB, columns ...string) (*OverrideFixer[T], error) {
	if len(columns) < 1 {
		rows, err := target.Model(new(T)).Limit(1).Rows()
		if err != nil {
			return nil, err
		}
		if columns, err = rows.Columns(); err != nil {
			return nil, err
		}
	}
	return &OverrideFixer[T]{base: base, target: target, columns: columns}, nil
}

func (f *OverrideFixer[T]) Fix(event events.InconsistentEvent) error {
	var src T
	err := f.base.Where("id = ?", event.ID).First(&src).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return f.target.Delete("id = ?", event.ID).Error
	case err == nil:
		return f.target.Clauses(&clause.OnConflict{DoUpdates: clause.AssignmentColumns(f.columns)}).Create(&src).Error
	default:
		return err
	}
}
