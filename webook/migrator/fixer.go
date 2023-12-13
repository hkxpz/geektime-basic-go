package migrator

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"geektime-basic-go/webook/migrator/events"
)

type Fixer[T Entity] struct {
	srcDB   *gorm.DB
	dstDB   *gorm.DB
	columns []string
}

func (f *Fixer[T]) Fix(event events.InconsistentEvent) error {
	var src T
	err := f.srcDB.Where("id = ?", event.ID).First(&src).Error
	switch {
	default:
		return err
	case errors.Is(err, gorm.ErrRecordNotFound):
		return f.dstDB.Delete("id = ?", event.ID).Error
	case err == nil:
		return f.dstDB.Clauses(&clause.OnConflict{DoUpdates: clause.AssignmentColumns(f.columns)}).Create(&src).Error
	}
}
