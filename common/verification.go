package common

import (
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type verificationValue struct {
	code string
	time time.Time
}

const (
	EmailVerificationPurpose = "v"
	PasswordResetPurpose     = "r"
)

var verificationMutex sync.Mutex
var verificationMap map[string]verificationValue
var verificationMapMaxSize = 10
var VerificationValidMinutes = 10

func GenerateVerificationCode(length int) string {
	code := uuid.New().String()
	code = strings.Replace(code, "-", "", -1)
	if length == 0 {
		return code
	}
	return code[:length]
}

func RegisterVerificationCodeWithKey(key string, code string, purpose string) {
	verificationMutex.Lock()
	defer verificationMutex.Unlock()
	verificationMap[purpose+key] = verificationValue{
		code: code,
		time: time.Now(),
	}
	if len(verificationMap) > verificationMapMaxSize {
		removeExpiredPairs()
	}
}

func VerifyCodeWithKey(key string, code string, purpose string) bool {
	verificationMutex.Lock()
	defer verificationMutex.Unlock()
	value, okay := verificationMap[purpose+key]
	now := time.Now()
	if !okay || int(now.Sub(value.time).Seconds()) >= VerificationValidMinutes*60 {
		return false
	}
	return code == value.code
}

func DeleteKey(key string, purpose string) {
	verificationMutex.Lock()
	defer verificationMutex.Unlock()
	delete(verificationMap, purpose+key)
}

// no lock inside, so the caller must lock the verificationMap before calling!
func removeExpiredPairs() {
	now := time.Now()
	for key := range verificationMap {
		if int(now.Sub(verificationMap[key].time).Seconds()) >= VerificationValidMinutes*60 {
			delete(verificationMap, key)
		}
	}
}

func init() {
	verificationMutex.Lock()
	defer verificationMutex.Unlock()
	verificationMap = make(map[string]verificationValue)
}

// IsVerificationCodeActive checks whether there is an active (non-expired) verification
// code for the given key+pid. Returns true if a cooldown marker or an active code exists.
func IsVerificationCodeActive(key string, pid string) bool {
	verificationMutex.Lock()
	defer verificationMutex.Unlock()
	value, ok := verificationMap[pid+key]
	if !ok {
		return false
	}
	if strings.HasPrefix(value.code, "__cooldown__") {
		return true
	}
	return int(time.Since(value.time).Seconds()) < VerificationValidMinutes*60
}

// SetVerificationCodeCooldown stores a short-lived cooldown marker so that subsequent
// calls to IsVerificationCodeActive for the same key+pid return true for the given duration.
func SetVerificationCodeCooldown(key string, pid string, duration time.Duration) error {
	verificationMutex.Lock()
	defer verificationMutex.Unlock()
	verificationMap[pid+key] = verificationValue{
		code: "__cooldown__" + key,
		time: time.Now().Add(duration - time.Duration(VerificationValidMinutes)*60*time.Second),
	}
	return nil
}