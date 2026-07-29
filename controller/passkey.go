package controller

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/i18n"
	"github.com/MiLab-Bit/OpenFastToken/model"
	"github.com/MiLab-Bit/OpenFastToken/service/passkey"
	"github.com/MiLab-Bit/OpenFastToken/setting/system_setting"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	webauthn "github.com/go-webauthn/webauthn/webauthn"
)

// ==================== Registration ====================

// BeginPasskeyRegistration starts the WebAuthn registration ceremony
func BeginPasskeyRegistration(c *gin.Context) {
	settings := system_setting.GetPasskeySettings()
	if settings == nil || !settings.Enabled {
		common.ApiErrorI18n(c, i18n.MsgPasskeyCreateFailed)
		return
	}

	session := sessions.Default(c)
	id := session.Get("id")
	if id == nil {
		common.ApiErrorI18n(c, i18n.MsgPasskeyInvalidUserId)
		return
	}

	userId, ok := id.(int)
	if !ok {
		common.ApiErrorI18n(c, i18n.MsgPasskeyInvalidUserId)
		return
	}

	user, err := userRepo().GetByID(userId, false)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgPasskeyInvalidUserId)
		return
	}

	// Get existing credentials for the user to exclude them from registration
	existingCreds, _ := model.GetPasskeyCredentialsByUserId(userId)
	var existingWebAuthnCreds []webauthn.Credential
	var storedCredential *model.PasskeyCredential
	if len(existingCreds) > 0 {
		storedCredential = existingCreds[0]
		for _, cred := range existingCreds {
			existingWebAuthnCreds = append(existingWebAuthnCreds, cred.ToWebAuthnCredential())
		}
	}

	webAuthnUser := passkey.NewWebAuthnUser(user, storedCredential)

	wconfig, err := passkey.BuildWebAuthn(c.Request)
	if err != nil {
		common.SysLog(fmt.Sprintf("Passkey BuildWebAuthn error: %v", err))
		common.ApiErrorI18n(c, i18n.MsgPasskeyCreateFailed)
		return
	}

	// Build registration options with exclude list
	options, sessionData, err := wconfig.BeginRegistration(
		webAuthnUser,
		func(opt *protocol.PublicKeyCredentialCreationOptions) {
			opt.CredentialExcludeList = toExcludeList(existingWebAuthnCreds)
		},
	)
	if err != nil {
		common.SysLog(fmt.Sprintf("Passkey BeginRegistration error: %v", err))
		common.ApiErrorI18n(c, i18n.MsgPasskeyCreateFailed)
		return
	}

	// Save session data for the finish step
	if err := passkey.SaveSessionData(c, passkey.RegistrationSessionKey, sessionData); err != nil {
		common.SysLog(fmt.Sprintf("Passkey SaveSessionData error: %v", err))
		common.ApiErrorI18n(c, i18n.MsgPasskeyCreateFailed)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    options,
	})
}

// FinishPasskeyRegistration completes the WebAuthn registration ceremony
func FinishPasskeyRegistration(c *gin.Context) {
	settings := system_setting.GetPasskeySettings()
	if settings == nil || !settings.Enabled {
		common.ApiErrorI18n(c, i18n.MsgPasskeyCreateFailed)
		return
	}

	session := sessions.Default(c)
	id := session.Get("id")
	if id == nil {
		common.ApiErrorI18n(c, i18n.MsgPasskeyInvalidUserId)
		return
	}

	userId, ok := id.(int)
	if !ok {
		common.ApiErrorI18n(c, i18n.MsgPasskeyInvalidUserId)
		return
	}

	user, err := userRepo().GetByID(userId, false)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgPasskeyInvalidUserId)
		return
	}

	// Pop session data from the begin step
	sessionData, err := passkey.PopSessionData(c, passkey.RegistrationSessionKey)
	if err != nil {
		common.SysLog(fmt.Sprintf("Passkey PopSessionData error: %v", err))
		common.ApiErrorI18n(c, i18n.MsgPasskeyCreateFailed)
		return
	}

	existingCreds, _ := model.GetPasskeyCredentialsByUserId(userId)
	var storedCredential *model.PasskeyCredential
	if len(existingCreds) > 0 {
		storedCredential = existingCreds[0]
	}

	webAuthnUser := passkey.NewWebAuthnUser(user, storedCredential)

	wconfig, err := passkey.BuildWebAuthn(c.Request)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgPasskeyCreateFailed)
		return
	}

	credential, err := wconfig.FinishRegistration(webAuthnUser, *sessionData, c.Request)
	if err != nil {
		common.SysLog(fmt.Sprintf("Passkey FinishRegistration error: %v", err))
		common.ApiErrorI18n(c, i18n.MsgPasskeyCreateFailed)
		return
	}

	// Determine attachment type from transport
	attachment := "cross-platform"
	if len(credential.Transport) > 0 {
		for _, t := range credential.Transport {
			if t == protocol.Internal {
				attachment = "platform"
				break
			}
		}
	}

	// Store the credential in the database
	credID := base64.RawStdEncoding.EncodeToString(credential.ID)
	newCred := &model.PasskeyCredential{
		UserID:          userId,
		CredentialID:    credID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		AAGUID:          string(credential.Authenticator.AAGUID),
		SignCount:       credential.Authenticator.SignCount,
		BackupEligible:  credential.Flags.BackupEligible,
		BackupState:     credential.Flags.BackupState,
		Attachment:      attachment,
		DeviceName:      "", // User can update this later
	}

	if err := model.CreatePasskeyCredential(newCred); err != nil {
		common.SysLog(fmt.Sprintf("Passkey CreatePasskeyCredential error: %v", err))
		common.ApiErrorI18n(c, i18n.MsgPasskeyCreateFailed)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":              newCred.ID,
			"device_name":     newCred.DeviceName,
			"created_at":      newCred.CreatedAt,
			"backup_eligible": newCred.BackupEligible,
		},
	})
}

// ==================== Login ====================

// BeginPasskeyLogin starts the WebAuthn login ceremony (discovery-based / passwordless)
func BeginPasskeyLogin(c *gin.Context) {
	settings := system_setting.GetPasskeySettings()
	if settings == nil || !settings.Enabled {
		common.ApiErrorI18n(c, i18n.MsgPasskeyLoginAbnormal)
		return
	}

	wconfig, err := passkey.BuildWebAuthn(c.Request)
	if err != nil {
		common.SysLog(fmt.Sprintf("Passkey BuildWebAuthn error: %v", err))
		common.ApiErrorI18n(c, i18n.MsgPasskeyLoginAbnormal)
		return
	}

	// For passwordless login, we use discovery-based assertion (empty allowCredentials)
	options, sessionData, err := wconfig.BeginDiscoverableLogin(
		func(opt *protocol.PublicKeyCredentialRequestOptions) {
			opt.UserVerification = protocol.UserVerificationRequirement(settings.UserVerification)
		},
	)
	if err != nil {
		common.SysLog(fmt.Sprintf("Passkey BeginDiscoverableLogin error: %v", err))
		common.ApiErrorI18n(c, i18n.MsgPasskeyLoginAbnormal)
		return
	}

	// Save session data for the finish step
	if err := passkey.SaveSessionData(c, passkey.LoginSessionKey, sessionData); err != nil {
		common.SysLog(fmt.Sprintf("Passkey SaveSessionData error: %v", err))
		common.ApiErrorI18n(c, i18n.MsgPasskeyLoginAbnormal)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    options,
	})
}

// FinishPasskeyLogin completes the WebAuthn login ceremony
func FinishPasskeyLogin(c *gin.Context) {
	settings := system_setting.GetPasskeySettings()
	if settings == nil || !settings.Enabled {
		common.ApiErrorI18n(c, i18n.MsgPasskeyLoginAbnormal)
		return
	}

	// Pop session data from the begin step
	sessionData, err := passkey.PopSessionData(c, passkey.LoginSessionKey)
	if err != nil {
		common.SysLog(fmt.Sprintf("Passkey PopSessionData error: %v", err))
		common.ApiErrorI18n(c, i18n.MsgPasskeyLoginAbnormal)
		return
	}

	wconfig, err := passkey.BuildWebAuthn(c.Request)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgPasskeyLoginAbnormal)
		return
	}

	// Parse the assertion response to extract the credential ID
	parsedResponse, err := protocol.ParseCredentialRequestResponseBody(c.Request.Body)
	if err != nil {
		common.SysLog(fmt.Sprintf("Passkey ParseCredentialRequestResponseBody error: %v", err))
		common.ApiErrorI18n(c, i18n.MsgPasskeyLoginAbnormal)
		return
	}

	// Look up the user by credential ID from the assertion response
	user, storedCred, err := model.GetUserByPasskeyCredentialID(parsedResponse.RawID)
	if err != nil {
		common.SysLog(fmt.Sprintf("Passkey GetUserByCredentialID error: %v", err))
		common.ApiErrorI18n(c, i18n.MsgPasskeyLoginAbnormal)
		return
	}

	// Check user status
	if user.Status != common.UserStatusEnabled {
		common.ApiErrorI18n(c, i18n.MsgAuthUserBanned)
		return
	}

	webAuthnUser := passkey.NewWebAuthnUser(user, storedCred)

	// Use FinishDiscoverableLogin with a handler that returns the found user
	credential, err := wconfig.FinishDiscoverableLogin(
		func(rawID []byte, userHandle []byte) (webauthn.User, error) {
			// We already found the user, return it as webauthn.User interface
			return webAuthnUser, nil
		},
		*sessionData,
		c.Request,
	)
	if err != nil {
		common.SysLog(fmt.Sprintf("Passkey FinishDiscoverableLogin error: %v", err))
		common.ApiErrorI18n(c, i18n.MsgPasskeyLoginAbnormal)
		return
	}

	// Update the credential's sign count
	_ = model.UpdatePasskeyCredentialSignCount(storedCred.ID, credential.Authenticator.SignCount)

	// Set up the login session
	setupLogin(user, c)
}

// ==================== Verification (re-authentication) ====================

// BeginPasskeyVerification starts the WebAuthn verification ceremony for an already logged-in user
func BeginPasskeyVerification(c *gin.Context) {
	settings := system_setting.GetPasskeySettings()
	if settings == nil || !settings.Enabled {
		common.ApiErrorI18n(c, i18n.MsgPasskeyVerifyFailed)
		return
	}

	session := sessions.Default(c)
	id := session.Get("id")
	if id == nil {
		common.ApiErrorI18n(c, i18n.MsgPasskeyInvalidUserId)
		return
	}

	userId, ok := id.(int)
	if !ok {
		common.ApiErrorI18n(c, i18n.MsgPasskeyInvalidUserId)
		return
	}

	user, err := userRepo().GetByID(userId, false)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgPasskeyInvalidUserId)
		return
	}

	existingCreds, _ := model.GetPasskeyCredentialsByUserId(userId)
	if len(existingCreds) == 0 {
		common.ApiErrorI18n(c, i18n.MsgPasskeyVerifyFailed)
		return
	}

	storedCredential := existingCreds[0]

	webAuthnUser := passkey.NewWebAuthnUser(user, storedCredential)

	wconfig, err := passkey.BuildWebAuthn(c.Request)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgPasskeyVerifyFailed)
		return
	}

	options, sessionData, err := wconfig.BeginLogin(
		webAuthnUser,
		func(opt *protocol.PublicKeyCredentialRequestOptions) {
			opt.UserVerification = protocol.UserVerificationRequirement(settings.UserVerification)
		},
	)
	if err != nil {
		common.SysLog(fmt.Sprintf("Passkey BeginVerification error: %v", err))
		common.ApiErrorI18n(c, i18n.MsgPasskeyVerifyFailed)
		return
	}

	if err := passkey.SaveSessionData(c, passkey.VerifySessionKey, sessionData); err != nil {
		common.ApiErrorI18n(c, i18n.MsgPasskeyVerifyFailed)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    options,
	})
}

// FinishPasskeyVerification completes the WebAuthn verification ceremony
func FinishPasskeyVerification(c *gin.Context) {
	settings := system_setting.GetPasskeySettings()
	if settings == nil || !settings.Enabled {
		common.ApiErrorI18n(c, i18n.MsgPasskeyVerifyFailed)
		return
	}

	session := sessions.Default(c)
	id := session.Get("id")
	if id == nil {
		common.ApiErrorI18n(c, i18n.MsgPasskeyInvalidUserId)
		return
	}

	userId, ok := id.(int)
	if !ok {
		common.ApiErrorI18n(c, i18n.MsgPasskeyInvalidUserId)
		return
	}

	user, err := userRepo().GetByID(userId, false)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgPasskeyInvalidUserId)
		return
	}

	sessionData, err := passkey.PopSessionData(c, passkey.VerifySessionKey)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgPasskeyVerifyFailed)
		return
	}

	existingCreds, _ := model.GetPasskeyCredentialsByUserId(userId)
	var storedCredential *model.PasskeyCredential
	if len(existingCreds) > 0 {
		storedCredential = existingCreds[0]
	}

	webAuthnUser := passkey.NewWebAuthnUser(user, storedCredential)

	wconfig, err := passkey.BuildWebAuthn(c.Request)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgPasskeyVerifyFailed)
		return
	}

	credential, err := wconfig.FinishLogin(webAuthnUser, *sessionData, c.Request)
	if err != nil {
		common.SysLog(fmt.Sprintf("Passkey FinishVerification error: %v", err))
		common.ApiErrorI18n(c, i18n.MsgPasskeyVerifyFailed)
		return
	}

	// Update sign count
	if storedCredential != nil {
		_ = model.UpdatePasskeyCredentialSignCount(storedCredential.ID, credential.Authenticator.SignCount)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"verified":   true,
			"user_id":    userId,
			"sign_count": credential.Authenticator.SignCount,
		},
	})
}

// ==================== Management ====================

// GetPasskeyStatus returns the user's passkey status and credentials
func GetPasskeyStatus(c *gin.Context) {
	settings := system_setting.GetPasskeySettings()

	var rpDisplayName string
	var rpID string
	var enabled bool
	if settings != nil {
		enabled = settings.Enabled
		rpDisplayName = settings.RPDisplayName
		rpID = settings.RPID
	}

	session := sessions.Default(c)
	id := session.Get("id")
	if id == nil {
		// Not logged in — return system-level passkey status only
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"enabled":         enabled,
				"rp_display_name": rpDisplayName,
				"rp_id":           rpID,
			},
		})
		return
	}

	userId, ok := id.(int)
	if !ok {
		common.ApiErrorI18n(c, i18n.MsgPasskeyInvalidUserId)
		return
	}

	creds, err := model.GetPasskeyCredentialsByUserId(userId)
	if err != nil {
		common.SysLog(fmt.Sprintf("Passkey GetPasskeyStatus error: %v", err))
		common.ApiErrorI18n(c, i18n.MsgPasskeyCreateFailed)
		return
	}

	// Build credential list for the response
	credList := make([]gin.H, 0, len(creds))
	var lastUsedAt *int64
	for _, cred := range creds {
		credList = append(credList, gin.H{
			"id":              cred.ID,
			"device_name":     cred.DeviceName,
			"created_at":      cred.CreatedAt,
			"last_used_at":    cred.LastUsedAt,
			"backup_eligible": cred.BackupEligible,
			"backup_state":    cred.BackupState,
		})
		if cred.LastUsedAt > 0 {
			lastUsedAt = &cred.LastUsedAt
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"enabled":         enabled,
			"rp_display_name": rpDisplayName,
			"rp_id":           rpID,
			"credentials":     credList,
			"last_used_at":    lastUsedAt,
		},
	})
}

// DeletePasskey deletes a specific passkey credential for the authenticated user
func DeletePasskey(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("id")
	if id == nil {
		common.ApiErrorI18n(c, i18n.MsgPasskeyInvalidUserId)
		return
	}

	userId, ok := id.(int)
	if !ok {
		common.ApiErrorI18n(c, i18n.MsgPasskeyInvalidUserId)
		return
	}

	// Try to get credential ID from URL param
	rawID := c.Param("id")
	if rawID == "" {
		// If no specific ID, delete all passkeys for the user
		if err := model.DeletePasskeyCredentialsByUserId(userId); err != nil {
			common.SysLog(fmt.Sprintf("Passkey DeleteAll error: %v", err))
			common.ApiErrorI18n(c, i18n.MsgPasskeyUpdateFailed)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": i18n.Msg(c, "All passkeys deleted"),
		})
		return
	}

	// Parse specific credential ID
	credID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	if err := model.DeletePasskeyCredential(uint(credID), userId); err != nil {
		common.SysLog(fmt.Sprintf("Passkey DeleteCredential error: %v", err))
		common.ApiErrorI18n(c, i18n.MsgPasskeyUpdateFailed)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, "Passkey deleted"),
	})
}

// ==================== Helpers ====================

func toExcludeList(credentials []webauthn.Credential) []protocol.CredentialDescriptor {
	descriptors := make([]protocol.CredentialDescriptor, len(credentials))
	for i, cred := range credentials {
		descriptors[i] = protocol.CredentialDescriptor{
			CredentialID: cred.ID,
			Type:         protocol.PublicKeyCredentialType,
			Transport:    cred.Transport,
		}
	}
	return descriptors
}
