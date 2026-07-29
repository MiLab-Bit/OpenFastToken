package system_setting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDiscordSettings(t *testing.T) {
	d := GetDiscordSettings()
	assert.Equal(t, &defaultDiscordSettings, d)
}

func TestGetLegalSettings(t *testing.T) {
	l := GetLegalSettings()
	assert.Equal(t, &defaultLegalSettings, l)
}

func TestGetOIDCSettings(t *testing.T) {
	o := GetOIDCSettings()
	assert.Equal(t, &defaultOIDCSettings, o)
}

func TestGetFetchSetting(t *testing.T) {
	f := GetFetchSetting()
	assert.Equal(t, &defaultFetchSetting, f)
	// defaults
	assert.True(t, f.EnableSSRFProtection)
	assert.Equal(t, []string{"80", "443", "8080", "8443"}, f.AllowedPorts)
}

func TestGetThemeSettings(t *testing.T) {
	th := GetThemeSettings()
	assert.Equal(t, &themeSettings, th)
	assert.Equal(t, "default", th.Frontend)
}

func TestUpdateAndSyncTheme(t *testing.T) {
	themeSettings.Frontend = "default"
	// must not panic; exercises syncThemeToCommon -> common.SetTheme
	UpdateAndSyncTheme()
}

func TestEnableWorker(t *testing.T) {
	old := WorkerUrl
	defer func() { WorkerUrl = old }()

	WorkerUrl = ""
	assert.False(t, EnableWorker())

	WorkerUrl = "https://worker.example.com"
	assert.True(t, EnableWorker())
}

func TestServerAddressDefault(t *testing.T) {
	// package-level default; just ensure it's a non-empty address
	assert.NotEmpty(t, ServerAddress)
}

func TestGetPasskeySettings_Default(t *testing.T) {
	defaultPasskeySettings.RPID = ""
	defaultPasskeySettings.Origins = ""
	p := GetPasskeySettings()
	assert.Equal(t, &defaultPasskeySettings, p)
	assert.False(t, p.Enabled)
	assert.Equal(t, "preferred", p.UserVerification)
}

func TestGetPasskeySettings_DerivesFromServerAddress(t *testing.T) {
	oldAddr := ServerAddress
	defer func() { ServerAddress = oldAddr }()

	ServerAddress = "https://fasttoken.example.com"
	defaultPasskeySettings.RPID = ""
	defaultPasskeySettings.Origins = "[]"
	p := GetPasskeySettings()
	assert.Equal(t, "fasttoken.example.com", p.RPID)
	assert.Equal(t, "https://fasttoken.example.com", p.Origins)
}

func TestGetPasskeySettings_OriginsEmptyDerives(t *testing.T) {
	oldAddr := ServerAddress
	defer func() { ServerAddress = oldAddr }()

	ServerAddress = "http://localhost:3000"
	defaultPasskeySettings.RPID = "already-set"
	defaultPasskeySettings.Origins = ""
	p := GetPasskeySettings()
	// RPID already set -> unchanged; Origins empty -> derives from ServerAddress
	assert.Equal(t, "already-set", p.RPID)
	assert.Equal(t, "http://localhost:3000", p.Origins)
}
