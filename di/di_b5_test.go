package di

import (
	"testing"

	"github.com/MiLab-Bit/OpenFastToken/repository/impl"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitAndDefault(t *testing.T) {
	require.NoError(t, Init())
	mgr := Default()
	require.NotNil(t, mgr)
	require.NotNil(t, mgr.User)
	require.NotNil(t, mgr.Channel)
	require.NotNil(t, mgr.Token)
	require.NotNil(t, MustUser())
	require.NotNil(t, MustChannel())
	require.NotNil(t, MustToken())
}

func TestInitIdempotent(t *testing.T) {
	require.NoError(t, Init())
	require.NoError(t, Init())
}

func TestSetDefault(t *testing.T) {
	custom := &Manager{
		User:    impl.NewUserRepository(),
		Channel: impl.NewChannelRepository(),
		Token:   impl.NewTokenRepository(),
	}
	SetDefault(custom)
	defer ResetDefault()
	require.Same(t, custom, Default())
}

func TestResetDefault(t *testing.T) {
	ResetDefault()
	mgr := Default()
	require.NotNil(t, mgr)
	require.NotNil(t, mgr.User)
	require.NotNil(t, mgr.Channel)
	require.NotNil(t, mgr.Token)
}

func TestMustReturnsNilWhenUnset(t *testing.T) {
	SetDefault(&Manager{})
	defer ResetDefault()
	assert.NotPanics(t, func() { _ = MustUser() })
	assert.Nil(t, MustUser())
	assert.Nil(t, MustChannel())
	assert.Nil(t, MustToken())
}
