package pg

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/chaitin/panda-wiki/domain"
	store "github.com/chaitin/panda-wiki/store/pg"
)

var ErrInvalidReportTransition = errors.New("invalid report status transition")

type ReportRepo struct{ db *store.DB }

func NewReportRepo(db *store.DB) *ReportRepo { return &ReportRepo{db: db} }

func (r *ReportRepo) CreatePending(ctx context.Context, report *domain.Report) error {
	if report.ID == "" {
		report.ID = uuid.NewString()
	}
	report.Status = domain.ReportStatusPending
	report.ResultCode = ""
	return r.db.WithContext(ctx).Create(report).Error
}

func (r *ReportRepo) MarkRunning(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Model(&domain.Report{}).Where("id = ? AND status = ?", id, domain.ReportStatusPending).Updates(map[string]any{
		"status": domain.ReportStatusRunning, "updated_at": time.Now().UTC(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrInvalidReportTransition
	}
	return nil
}

func (r *ReportRepo) CompleteWithCitations(ctx context.Context, id, content string, resultCode domain.ReportResultCode, citations []domain.ReportCitation) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		result := tx.Model(&domain.Report{}).Where("id = ? AND status = ?", id, domain.ReportStatusRunning).Updates(map[string]any{
			"status": domain.ReportStatusCompleted, "content": content, "result_code": resultCode, "error_message": "", "completed_at": now, "updated_at": now,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrInvalidReportTransition
		}
		for i := range citations {
			if citations[i].ID == "" {
				citations[i].ID = uuid.NewString()
			}
			citations[i].ReportID = id
			if err := tx.Create(&citations[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ReportRepo) MarkFailed(ctx context.Context, id string, code domain.ReportResultCode, message string) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).Model(&domain.Report{}).Where("id = ? AND status = ?", id, domain.ReportStatusRunning).Updates(map[string]any{
		"status": domain.ReportStatusFailed, "result_code": code, "error_message": message, "completed_at": now, "updated_at": now,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrInvalidReportTransition
	}
	return nil
}

func (r *ReportRepo) FailStaleRunning(ctx context.Context, before time.Time) (int64, error) {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).Model(&domain.Report{}).
		Where("status = ? AND updated_at < ?", domain.ReportStatusRunning, before).
		Updates(map[string]any{
			"status": domain.ReportStatusFailed, "result_code": domain.ReportResultTimeout,
			"error_message": "report generation timed out", "completed_at": now, "updated_at": now,
		})
	return result.RowsAffected, result.Error
}

func (r *ReportRepo) ListByCreator(ctx context.Context, creator uint) ([]domain.Report, error) {
	var reports []domain.Report
	err := r.db.WithContext(ctx).Model(&domain.Report{}).
		Select("reports.*, (SELECT COUNT(*) FROM report_citations WHERE report_citations.report_id = reports.id) AS citation_count").
		Where("created_by = ?", creator).Order("created_at DESC").Find(&reports).Error
	return reports, err
}

func (r *ReportRepo) GetByIDAndCreator(ctx context.Context, id string, creator uint) (*domain.Report, []domain.ReportCitation, error) {
	var report domain.Report
	if err := r.db.WithContext(ctx).Where("id = ? AND created_by = ?", id, creator).First(&report).Error; err != nil {
		return nil, nil, err
	}
	var citations []domain.ReportCitation
	if err := r.db.WithContext(ctx).Where("report_id = ?", id).Order("citation_index ASC").Find(&citations).Error; err != nil {
		return nil, nil, err
	}
	report.CitationCount = len(citations)
	return &report, citations, nil
}
