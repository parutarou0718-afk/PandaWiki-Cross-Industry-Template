package usecase

import (
	"testing"

	"github.com/stretchr/testify/require"

	v1node "github.com/chaitin/panda-wiki/api/node/v1"
	"github.com/chaitin/panda-wiki/consts"
	"github.com/chaitin/panda-wiki/domain"
)

func TestNodePermissionEditingIsNotLimitedByLegacyEdition(t *testing.T) {
	usecase := &NodeUsecase{}
	req := v1node.NodePermissionEditReq{
		Permissions: &domain.NodePermissions{
			Answerable: consts.NodeAccessPermPartial,
		},
		AnswerableGroups: &[]int{1},
	}

	require.NoError(t, usecase.ValidateNodePermissionsEdit(req, consts.LicenseEditionFree))
}

func TestStatRangesAreNotLimitedByLegacyEdition(t *testing.T) {
	usecase := &StatUseCase{}

	require.NoError(t, usecase.ValidateStatDay(consts.StatDay7, consts.LicenseEditionFree))
	require.NoError(t, usecase.ValidateStatDay(consts.StatDay30, consts.LicenseEditionFree))
	require.NoError(t, usecase.ValidateStatDay(consts.StatDay90, consts.LicenseEditionFree))
}
