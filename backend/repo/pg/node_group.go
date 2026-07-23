package pg

import (
	"context"

	"github.com/chaitin/panda-wiki/consts"
	"github.com/chaitin/panda-wiki/domain"
)

func (r *NodeRepository) GetNodeGroupsByGroupIdsPerm(ctx context.Context, authGroupIds []uint, perm consts.NodePermName) ([]domain.NodeAuthGroup, error) {
	nodeGroups := make([]domain.NodeAuthGroup, 0)

	if err := r.db.WithContext(ctx).
		Model(&domain.NodeAuthGroup{}).
		Where("auth_group_id in (?) and perm = ?", authGroupIds, perm).Find(&nodeGroups).Error; err != nil {
		return nil, err
	}
	return nodeGroups, nil
}

// GetNodeAuthGroupIdsByNodeId 查询该node下的用户组（非部分开放的情况下无返回）
func (r *NodeRepository) GetNodeAuthGroupIdsByNodeId(ctx context.Context, nodeId string, perm consts.NodePermName) ([]int, error) {

	node, err := r.GetNodeByID(ctx, nodeId)
	if err != nil {
		return nil, err
	}
	switch node.Permissions.Answerable {
	case consts.NodeAccessPermOpen:
		return nil, nil
	case consts.NodeAccessPermPartial:
		authGroupIds := make([]int, 0)

		if err := r.db.WithContext(ctx).
			Model(&domain.NodeAuthGroup{}).
			Joins("left join nodes on nodes.id = node_auth_groups.node_id").
			Where("nodes.permissions->>'answerable' = ?", consts.NodeAccessPermPartial).
			Where("node_auth_groups.node_id = ? and node_auth_groups.perm = ?", nodeId, perm).
			Pluck("node_auth_groups.auth_group_id", &authGroupIds).Error; err != nil {
			return nil, err
		}
		return authGroupIds, nil

	case consts.NodeAccessPermClosed:
		return make([]int, 0), nil
	}
	return nil, nil
}
