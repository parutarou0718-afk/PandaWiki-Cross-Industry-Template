package pg

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/utils"
)

func (r *NodeRepository) GetNodeStatsByNodeId(ctx context.Context, nodeId string) (*domain.NodeStats, error) {
	var nodeStats *domain.NodeStats
	if err := r.db.WithContext(ctx).
		Model(&domain.NodeStats{}).
		Where("node_id = ?", nodeId).
		First(&nodeStats).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			nodeStats = &domain.NodeStats{
				ID:     0,
				NodeID: nodeId,
				PV:     0,
			}
		} else {
			return nil, err
		}
	}

	var todayStats int64
	if err := r.db.WithContext(ctx).Model(&domain.StatPage{}).
		Where("created_at >= ?", utils.GetTimeHourOffset(-24)).
		Where("node_id = ?", nodeId).Count(&todayStats).Error; err != nil {
		return nil, err
	}
	nodeStats.PV += todayStats

	return nodeStats, nil
}
