package oauth

import (
	"sync"

	"github.com/MiLab-Bit/OpenFastToken/model"
)

var (
	providers = make(map[string]Provider)
	mu        sync.RWMutex
)

// Register registers an OAuth provider with the given name
func Register(name string, provider Provider) {
	mu.Lock()
	defer mu.Unlock()
	providers[name] = provider
}

// GetProvider returns the OAuth provider for the given name
func GetProvider(name string) Provider {
	mu.RLock()
	defer mu.RUnlock()
	return providers[name]
}

// GetAllProviders returns all registered OAuth providers
func GetAllProviders() map[string]Provider {
	mu.RLock()
	defer mu.RUnlock()
	result := make(map[string]Provider, len(providers))
	for k, v := range providers {
		result[k] = v
	}
	return result
}

// IsProviderRegistered checks if a provider is registered
func IsProviderRegistered(name string) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, ok := providers[name]
	return ok
}

// GetEnabledCustomProviders returns all registered custom OAuth providers that are enabled
func GetEnabledCustomProviders() []Provider {
	mu.RLock()
	defer mu.RUnlock()

	var enabled []Provider
	for _, p := range providers {
		if p.IsEnabled() {
			enabled = append(enabled, p)
		}
	}
	return enabled
}

// LoadCustomProviders loads custom OAuth providers from the database and registers them
func LoadCustomProviders() error {
	var customProviders []model.CustomOAuthProvider
	err := model.DB.Find(&customProviders).Error
	if err != nil {
		return err
	}

	for _, p := range customProviders {
		provider := &GenericOAuthProvider{
			Name:         p.DisplayName,
			ClientID:     p.ClientID,
			ClientSecret: p.ClientSecret,
			Enabled:      p.Enabled,
			ProviderID:   int(p.ID),
		}
		Register(p.Name, provider)
	}

	return nil
}
