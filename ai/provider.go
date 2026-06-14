package ai

import "context"

// Provider defines the interface for AI generation services.
type Provider interface {
	// Name returns the provider's identifier (e.g., "gemini", "kling").
	Name() string

	// GetModels returns the list of available model names for this provider.
	GetModels() []string

	// GenerateImage creates an image from the given prompt.
	GenerateImage(ctx context.Context, prompt string, imageData []byte, mimeType string, opts GenerateOptions) ([]byte, error)

	// GenerateVideo creates a video from the given prompt.
	GenerateVideo(ctx context.Context, prompt string, imageData []byte, mimeType string, opts GenerateOptions) ([]byte, error)
}

// GenerateOptions holds per-request generation options.
type GenerateOptions struct {
	ModelName string
}

// Registry holds registered provider instances.
type Registry struct {
	providers map[string]Provider
}

// NewRegistry creates a new provider registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register adds a provider to the registry.
func (r *Registry) Register(p Provider) {
	r.providers[p.Name()] = p
}

// Get returns a provider by name, or nil if not found.
func (r *Registry) Get(name string) Provider {
	return r.providers[name]
}

// List returns all registered providers.
func (r *Registry) List() []Provider {
	list := make([]Provider, 0, len(r.providers))
	for _, p := range r.providers {
		list = append(list, p)
	}
	return list
}

// GetDefault returns the default provider name.
func (r *Registry) GetDefault() string {
	if _, ok := r.providers["gemini"]; ok {
		return "gemini"
	}
	for name := range r.providers {
		return name
	}
	return ""
}

// ProviderInfo holds metadata for the API response.
type ProviderInfo struct {
	Name   string   `json:"name"`
	Models []string `json:"models"`
}

// GetProviderInfo returns a map of provider info for API responses.
func (r *Registry) GetProviderInfo() map[string]ProviderInfo {
	info := make(map[string]ProviderInfo)
	for name, p := range r.providers {
		info[name] = ProviderInfo{
			Name:   name,
			Models: p.GetModels(),
		}
	}
	return info
}
