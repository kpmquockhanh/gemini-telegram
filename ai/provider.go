package ai

import "context"

// ModelParamDef describes a configurable parameter for a model.
type ModelParamDef struct {
	Name    string           `json:"name"`
	Label   string           `json:"label"`
	Type    string           `json:"type"` // "slider", "number", "select"
	Default any              `json:"default"`
	Min     *float64         `json:"min,omitempty"`
	Max     *float64         `json:"max,omitempty"`
	Step    *float64         `json:"step,omitempty"`
	Options []ModelParamOption `json:"options,omitempty"`
}

// ModelParamOption represents a selectable value for select-type params.
type ModelParamOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// ModelInfo holds a single model's name and its configurable parameters.
type ModelInfo struct {
	Name   string         `json:"name"`
	Params []ModelParamDef `json:"params,omitempty"`
}

// Provider defines the interface for AI generation services.
type Provider interface {
	// Name returns the provider's identifier (e.g., "gemini", "kling").
	Name() string

	// GetModels returns the list of available model names for this provider.
	GetModels() []string

	// GetModelParams returns the configurable parameters for a specific model.
	GetModelParams(modelName string) []ModelParamDef

	// GenerateImage creates an image from the given prompt.
	GenerateImage(ctx context.Context, prompt string, imageData []byte, mimeType string, opts GenerateOptions) ([]byte, error)

	// GenerateVideo creates a video from the given prompt.
	GenerateVideo(ctx context.Context, prompt string, imageData []byte, mimeType string, opts GenerateOptions) ([]byte, error)
}

// GenerateOptions holds per-request generation options.
type GenerateOptions struct {
	ModelName string
	Params    map[string]any
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
	Name   string      `json:"name"`
	Models []ModelInfo `json:"models"`
}

// GetProviderInfo returns a map of provider info for API responses.
func (r *Registry) GetProviderInfo() map[string]ProviderInfo {
	info := make(map[string]ProviderInfo)
	for name, p := range r.providers {
		models := make([]ModelInfo, 0, len(p.GetModels()))
		for _, m := range p.GetModels() {
			models = append(models, ModelInfo{
				Name:   m,
				Params: p.GetModelParams(m),
			})
		}
		info[name] = ProviderInfo{
			Name:   name,
			Models: models,
		}
	}
	return info
}
