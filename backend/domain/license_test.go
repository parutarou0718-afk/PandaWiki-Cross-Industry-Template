package domain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetBaseEditionLimitationDefaultsToUnlimitedSelfHostedCapabilities(t *testing.T) {
	limitation := GetBaseEditionLimitation(context.Background())

	require.True(t, limitation.AllowAdminPerm)
	require.True(t, limitation.AllowAdvancedBot)
	require.True(t, limitation.AllowCommentAudit)
	require.True(t, limitation.AllowCopyProtection)
	require.True(t, limitation.AllowCustomCopyright)
	require.True(t, limitation.AllowMCPServer)
	require.True(t, limitation.AllowNodeStats)
	require.True(t, limitation.AllowOpenAIBotSettings)
	require.True(t, limitation.AllowWatermark)
	require.Greater(t, limitation.MaxKb, 1)
	require.Greater(t, limitation.MaxNode, 300)
	require.Greater(t, limitation.MaxSSOUser, 1)
	require.Greater(t, limitation.MaxAdmin, int64(1))
}

func TestGetBaseEditionLimitationIgnoresMalformedLegacyContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), ContextKeyEditionLimitation, []byte("{"))

	limitation := GetBaseEditionLimitation(ctx)

	require.True(t, limitation.AllowOpenAIBotSettings)
	require.True(t, limitation.AllowMCPServer)
}
