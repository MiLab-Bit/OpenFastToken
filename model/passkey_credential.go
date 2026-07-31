package model

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	webauthn "github.com/go-webauthn/webauthn/webauthn"
	"gorm.io/gorm"
)

// PasskeyCredential stores a WebAuthn credential for passwordless authentication
type PasskeyCredential struct {
	ID              uint   `gorm:"primaryKey" json:"id"`
	UserID          int    `json:"user_id" gorm:"not null;index:idx_passkey_user_id"`
	CredentialID    string `json:"credential_id" gorm:"type:varchar(1024);uniqueIndex:ux_passkey_cred_id"`
	PublicKey       []byte `json:"-" gorm:"type:bytea"`
	AttestationType string `json:"attestation_type" gorm:"type:varchar(100)"`
	AAGUID          string `json:"aaguid" gorm:"type:varchar(36)"`
	SignCount       uint32 `json:"sign_count" gorm:"default:0"`
	BackupEligible  bool   `json:"backup_eligible" gorm:"default:false"`
	BackupState     bool   `json:"backup_state" gorm:"default:false"`
	Attachment      string `json:"attachment" gorm:"type:varchar(50)"`
	CreatedAt       int64  `json:"created_at" gorm:"autoCreateTime"`
	LastUsedAt      int64  `json:"last_used_at" gorm:"default:0"`
	DeviceName      string `json:"device_name" gorm:"type:varchar(255);default:''"`
}

func (PasskeyCredential) TableName() string {
	return "passkey_credentials"
}

// ToWebAuthnCredential converts a PasskeyCredential to a webauthn.Credential
func (pc *PasskeyCredential) ToWebAuthnCredential() webauthn.Credential {
	credID, _ := base64.RawStdEncoding.DecodeString(pc.CredentialID)

	var transport []protocol.AuthenticatorTransport
	// Default transport; could be extended based on attachment type
	switch pc.Attachment {
	case "platform":
		transport = []protocol.AuthenticatorTransport{protocol.Internal}
	case "cross-platform":
		transport = []protocol.AuthenticatorTransport{protocol.USB, protocol.BLE, protocol.Hybrid}
	default:
		transport = []protocol.AuthenticatorTransport{protocol.Internal, protocol.Hybrid}
	}

	return webauthn.Credential{
		ID:              credID,
		PublicKey:       pc.PublicKey,
		AttestationType: pc.AttestationType,
		Transport:       transport,
		Flags: webauthn.CredentialFlags{
			BackupEligible: pc.BackupEligible,
			BackupState:    pc.BackupState,
		},
		Authenticator: webauthn.Authenticator{
			AAGUID:    []byte(pc.AAGUID),
			SignCount: pc.SignCount,
		},
		Attestation: webauthn.CredentialAttestation{}, // Not stored for privacy; we use none attestation
	}
}

// GetPasskeyCredentialsByUserId returns all passkey credentials for a user
func GetPasskeyCredentialsByUserId(userId int) ([]*PasskeyCredential, error) {
	var creds []*PasskeyCredential
	err := DB.Where("user_id = ?", userId).Order("created_at DESC").Find(&creds).Error
	return creds, err
}

// GetPasskeyCredentialByID returns a single passkey credential by its primary key
func GetPasskeyCredentialByID(id uint) (*PasskeyCredential, error) {
	var cred PasskeyCredential
	err := DB.First(&cred, id).Error
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

// GetPasskeyCredentialByCredentialID returns a credential by its WebAuthn credential ID
func GetPasskeyCredentialByCredentialID(credentialID string) (*PasskeyCredential, error) {
	var cred PasskeyCredential
	err := DB.Where("credential_id = ?", credentialID).First(&cred).Error
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

// GetUserByPasskeyCredentialID finds a user by their WebAuthn credential ID
func GetUserByPasskeyCredentialID(credentialID []byte) (*User, *PasskeyCredential, error) {
	encodedCredID := base64.RawStdEncoding.EncodeToString(credentialID)
	var cred PasskeyCredential
	err := DB.Where("credential_id = ?", encodedCredID).First(&cred).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("passkey credential not found")
		}
		return nil, nil, err
	}

	var user User
	err = DB.First(&user, cred.UserID).Error
	if err != nil {
		return nil, nil, err
	}
	return &user, &cred, nil
}

// CreatePasskeyCredential creates a new passkey credential
func CreatePasskeyCredential(cred *PasskeyCredential) error {
	if cred.UserID == 0 {
		return errors.New("user ID is required")
	}
	if len(cred.PublicKey) == 0 {
		return errors.New("public key is required")
	}
	if cred.CredentialID == "" {
		return errors.New("credential ID is required")
	}
	cred.CreatedAt = time.Now().Unix()
	return DB.Create(cred).Error
}

// UpdatePasskeyCredentialSignCount updates the sign count and last used time
func UpdatePasskeyCredentialSignCount(id uint, signCount uint32) error {
	return DB.Model(&PasskeyCredential{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"sign_count":  signCount,
			"last_used_at": time.Now().Unix(),
		}).Error
}

// DeletePasskeyCredential deletes a passkey credential by ID and user ID (ownership check)
func DeletePasskeyCredential(id uint, userId int) error {
	result := DB.Where("id = ? AND user_id = ?", id, userId).Delete(&PasskeyCredential{})
	if result.RowsAffected == 0 {
		return errors.New("passkey credential not found or not owned by user")
	}
	return result.Error
}

// DeletePasskeyCredentialsByUserId deletes all passkey credentials for a user
func DeletePasskeyCredentialsByUserId(userId int) error {
	return DB.Where("user_id = ?", userId).Delete(&PasskeyCredential{}).Error
}

// HasPasskeyCredential checks if a user has any passkey credential
func HasPasskeyCredential(userId int) (bool, error) {
	var count int64
	err := DB.Model(&PasskeyCredential{}).Where("user_id = ?", userId).Count(&count).Error
	return count > 0, err
}

// GetPasskeyCredentialCount returns the number of passkey credentials for a user
func GetPasskeyCredentialCount(userId int) (int64, error) {
	var count int64
	err := DB.Model(&PasskeyCredential{}).Where("user_id = ?", userId).Count(&count).Error
	return count, err
}
