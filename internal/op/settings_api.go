package op

import (
	"context"
	"encoding/json"
	"net/url"
	"sort"
	"strings"
	"time"
	"urlAPI/internal/database"
	"urlAPI/internal/llm"
	"urlAPI/util"

	"github.com/pkg/errors"
)

/** @brief 单个提供方设置的接口传输结构。 */
type providerSettingsDTO struct {
	APIKeySet        bool              `json:"api_key_set"`
	APIKey           string            `json:"api_key,omitempty"`
	TextModel        string            `json:"text_model"`
	SummaryModel     string            `json:"summary_model"`
	ImageModel       string            `json:"image_model,omitempty"`
	ImageSize        string            `json:"image_size,omitempty"`
	EmbeddingModel   string            `json:"embedding_model,omitempty"`
	Endpoint         string            `json:"endpoint"`
	APIType          string            `json:"api_type"`
	Temperature      float64           `json:"temperature"`
	MaxTokens        int               `json:"max_tokens"`
	TopP             float64           `json:"top_p"`
	PresencePenalty  float64           `json:"presence_penalty"`
	FrequencyPenalty float64           `json:"frequency_penalty"`
	CustomHeaders    map[string]string `json:"custom_headers,omitempty"`
	Enabled          bool              `json:"enabled"`
}

/** @brief 提供方模型列表响应结构。 */
type providerModelsDTO struct {
	Models []string `json:"models"`
}

/** @brief 文本功能设置的接口传输结构。 */
type textSettingsDTO struct {
	Enabled            bool     `json:"enabled"`
	GenerationAPI      string   `json:"generation_api"`
	SummaryAPI         string   `json:"summary_api"`
	CacheMinutes       int      `json:"cache_minutes"`
	FallbackImageURL   string   `json:"fallback_image_url"`
	EnabledPromptKeys  []string `json:"enabled_prompt_keys"`
	AcceptedPromptGlob []string `json:"accepted_prompt_glob"`
}

/** @brief 图像功能设置的接口传输结构。 */
type imageSettingsDTO struct {
	Enabled            bool     `json:"enabled"`
	API                string   `json:"api"`
	CacheMinutes       int      `json:"cache_minutes"`
	FallbackImageURL   string   `json:"fallback_image_url"`
	AcceptedPromptGlob []string `json:"accepted_prompt_glob"`
}

/** @brief 网页功能设置的接口传输结构。 */
type webSettingsDTO struct {
	Enabled          bool     `json:"enabled"`
	SummaryAPI       string   `json:"summary_api"`
	CacheMinutes     int      `json:"cache_minutes"`
	FallbackImageURL string   `json:"fallback_image_url"`
	RepoTokenSet     bool     `json:"repo_token_set"`
	RepoToken        string   `json:"repo_token,omitempty"`
	YouTubeTokenSet  bool     `json:"youtube_token_set"`
	YouTubeToken     string   `json:"youtube_token,omitempty"`
	AllowedHosts     []string `json:"allowed_hosts"`
}

/** @brief 随机图片功能设置的接口传输结构。 */
type randomSettingsDTO struct {
	Enabled           bool   `json:"enabled"`
	SourceRewriteFrom string `json:"source_rewrite_from"`
	FallbackImageURL  string `json:"fallback_image_url"`
	DefaultAPI        string `json:"default_api"`
}

/** @brief 后台安全设置的接口传输结构。 */
type dashboardSecurityDTO struct {
	PasswordHash        string   `json:"password_hash,omitempty"`
	DashboardAllowedIPs []string `json:"dashboard_allowed_ips"`
	AllowedReferers     []string `json:"allowed_referers"`
}

/** @brief 提示词设置的接口传输结构。 */
type promptSettingsDTO struct {
	GenerationContext string            `json:"generation_context"`
	SummaryContext    string            `json:"summary_context"`
	Templates         map[string]string `json:"templates"`
}

/** @brief 任务行为设置的接口传输结构。 */
type taskBehaviorDTO struct {
	ExceptDomains []string `json:"except_domains"`
	ExceptInfos   []string `json:"except_infos"`
}

/** @brief 文本提示词安全设置传输结构。 */
type textPromptSecurityDTO struct {
	AcceptedPromptGlob []string `json:"accepted_prompt_glob"`
}

/** @brief 图像提示词安全设置传输结构。 */
type imagePromptSecurityDTO struct {
	AcceptedPromptGlob []string `json:"accepted_prompt_glob"`
}

/**
 * @brief 读取指定分区的应用设置。
 * @param info 会话请求与响应对象。
 * @return error 读取或编码失败时返回错误。
 */
func fetchSettings(info *Session) error {
	body, err := settingsBody(info.SettingPart, database.SettingsStore.Get())
	if err != nil {
		return errors.WithStack(err)
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return errors.WithStack(err)
	}
	info.SettingBody = encoded
	return nil
}

/**
 * @brief 修改指定分区的应用设置。
 * @param info 会话请求与响应对象。
 * @return error 反序列化、校验或保存失败时返回错误。
 */
func editSettings(info *Session) error {
	settings := database.SettingsStore.Get()
	updated, err := applySettingsBody(info.SettingPart, settings, info.SettingBody)
	if err != nil {
		return errors.WithStack(err)
	}
	if err := validateSettings(updated); err != nil {
		return errors.WithStack(err)
	}
	return database.SaveAppSettings(updated)
}

/**
 * @brief 读取指定提供方可用模型列表。
 * @param info 会话请求与响应对象。
 * @return error 提供方不存在或远端请求失败时返回错误。
 */
func fetchProviderModels(info *Session) error {
	settings := database.SettingsStore.Get()
	providerName := providerNameFromPart(info.SettingPart)
	provider, ok := settings.Providers.ByName(providerName)
	if !ok {
		return errors.New("provider not found")
	}
	if !provider.Enabled {
		return errors.New("provider is disabled")
	}
	if provider.APIKey == "" {
		return errors.New("provider API key is not configured")
	}

	client, err := llm.NewProvider(provider)
	if err != nil {
		return errors.WithStack(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	resp, err := client.Models(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	if resp.Error != nil {
		return errors.New(resp.Error.Message)
	}

	models := make([]string, 0, len(resp.Data))
	seen := make(map[string]struct{}, len(resp.Data))
	for _, model := range resp.Data {
		id := strings.TrimSpace(model.ID)
		if id == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		models = append(models, id)
	}
	sort.Strings(models)

	encoded, err := json.Marshal(providerModelsDTO{Models: models})
	if err != nil {
		return errors.WithStack(err)
	}
	info.SettingBody = encoded
	return nil
}

/**
 * @brief 根据分区名称构造设置响应体。
 * @param part 设置分区标识。
 * @param settings 当前完整应用设置。
 * @return any 对应分区的响应结构。
 * @return error 分区不存在时返回错误。
 */
func settingsBody(part string, settings util.AppSettings) (any, error) {
	switch part {
	case "provider.openai", "openai":
		return providerDTO(settings.Providers.OpenAI), nil
	case "provider.deepseek", "deepseek":
		return providerDTO(settings.Providers.DeepSeek), nil
	case "provider.alibaba", "alibaba":
		return providerDTO(settings.Providers.Alibaba), nil
	case "provider.anthropic", "anthropic":
		return providerDTO(settings.Providers.Anthropic), nil
	case "provider.gemini", "gemini":
		return providerDTO(settings.Providers.Gemini), nil
	case "provider.azure", "azure":
		return providerDTO(settings.Providers.Azure), nil
	case "provider.moonshot", "moonshot":
		return providerDTO(settings.Providers.Moonshot), nil
	case "provider.otherapi", "otherapi":
		return providerDTO(settings.Providers.OtherAPI), nil
	case "feature.text", "txt":
		return textSettingsDTO{
			Enabled:            settings.Features.TextEnabled,
			GenerationAPI:      settings.Text.GenerationAPI,
			SummaryAPI:         settings.Text.SummaryAPI,
			CacheMinutes:       settings.Text.CacheMinutes,
			FallbackImageURL:   settings.Text.FallbackImageURL,
			EnabledPromptKeys:  settings.Text.EnabledPromptKeys,
			AcceptedPromptGlob: settings.Text.AcceptedPromptGlob,
		}, nil
	case "feature.image", "img":
		return imageSettingsDTO{
			Enabled:            settings.Features.ImageEnabled,
			API:                settings.Image.API,
			CacheMinutes:       settings.Image.CacheMinutes,
			FallbackImageURL:   settings.Image.FallbackImageURL,
			AcceptedPromptGlob: settings.Image.AcceptedPromptGlob,
		}, nil
	case "feature.web", "web":
		return webSettingsDTO{
			Enabled:          settings.Features.WebImgEnabled,
			SummaryAPI:       settings.Web.SummaryAPI,
			CacheMinutes:     settings.Web.CacheMinutes,
			FallbackImageURL: settings.Web.FallbackImageURL,
			RepoTokenSet:     settings.Web.RepoToken != "",
			YouTubeTokenSet:  settings.Web.YouTubeToken != "",
			AllowedHosts:     settings.Web.AllowedHosts,
		}, nil
	case "feature.random", "rand":
		return randomSettingsDTO{
			Enabled:           settings.Features.RandomEnabled,
			SourceRewriteFrom: settings.Random.SourceRewriteFrom,
			FallbackImageURL:  settings.Random.FallbackImageURL,
			DefaultAPI:        settings.Random.DefaultAPI,
		}, nil
	case "security.dashboard", "security":
		return dashboardSecurityDTO{
			DashboardAllowedIPs: settings.Security.DashboardAllowedIPs,
			AllowedReferers:     settings.Security.AllowedReferers,
		}, nil
	case "prompt", "contxt":
		return promptSettingsDTO{
			GenerationContext: settings.Prompts.GenerationContext,
			SummaryContext:    settings.Prompts.SummaryContext,
			Templates:         settings.Prompts.Templates,
		}, nil
	case "security.task_behavior", "taskBehavior":
		return taskBehaviorDTO{
			ExceptDomains: settings.Task.ExceptDomains,
			ExceptInfos:   settings.Task.ExceptInfos,
		}, nil
	case "security.text_prompt", "txtSecurity":
		return textPromptSecurityDTO{AcceptedPromptGlob: settings.Text.AcceptedPromptGlob}, nil
	case "security.image_prompt", "imgSecurity":
		return imagePromptSecurityDTO{AcceptedPromptGlob: settings.Image.AcceptedPromptGlob}, nil
	default:
		return nil, errors.New("Setting part not found")
	}
}

/**
 * @brief 从设置分区名提取提供方名称。
 * @param part 设置分区标识。
 * @return string 提供方名称。
 */
func providerNameFromPart(part string) string {
	part = strings.TrimSpace(part)
	if strings.HasPrefix(part, "provider.") {
		return strings.TrimPrefix(part, "provider.")
	}
	return part
}

/**
 * @brief 将请求体应用到指定分区设置中。
 * @param part 设置分区标识。
 * @param settings 当前完整应用设置。
 * @param body 前端提交的原始 JSON。
 * @return util.AppSettings 更新后的应用设置。
 * @return error 反序列化失败或分区不存在时返回错误。
 */
func applySettingsBody(part string, settings util.AppSettings, body json.RawMessage) (util.AppSettings, error) {
	switch part {
	case "provider.openai", "openai":
		var dto providerSettingsDTO
		if err := json.Unmarshal(body, &dto); err != nil {
			return settings, err
		}
		settings.Providers.OpenAI = applyProviderDTO(settings.Providers.OpenAI, dto)
	case "provider.deepseek", "deepseek":
		var dto providerSettingsDTO
		if err := json.Unmarshal(body, &dto); err != nil {
			return settings, err
		}
		settings.Providers.DeepSeek = applyProviderDTO(settings.Providers.DeepSeek, dto)
	case "provider.alibaba", "alibaba":
		var dto providerSettingsDTO
		if err := json.Unmarshal(body, &dto); err != nil {
			return settings, err
		}
		settings.Providers.Alibaba = applyProviderDTO(settings.Providers.Alibaba, dto)
	case "provider.anthropic", "anthropic":
		var dto providerSettingsDTO
		if err := json.Unmarshal(body, &dto); err != nil {
			return settings, err
		}
		settings.Providers.Anthropic = applyProviderDTO(settings.Providers.Anthropic, dto)
	case "provider.gemini", "gemini":
		var dto providerSettingsDTO
		if err := json.Unmarshal(body, &dto); err != nil {
			return settings, err
		}
		settings.Providers.Gemini = applyProviderDTO(settings.Providers.Gemini, dto)
	case "provider.azure", "azure":
		var dto providerSettingsDTO
		if err := json.Unmarshal(body, &dto); err != nil {
			return settings, err
		}
		settings.Providers.Azure = applyProviderDTO(settings.Providers.Azure, dto)
	case "provider.moonshot", "moonshot":
		var dto providerSettingsDTO
		if err := json.Unmarshal(body, &dto); err != nil {
			return settings, err
		}
		settings.Providers.Moonshot = applyProviderDTO(settings.Providers.Moonshot, dto)
	case "provider.otherapi", "otherapi":
		var dto providerSettingsDTO
		if err := json.Unmarshal(body, &dto); err != nil {
			return settings, err
		}
		settings.Providers.OtherAPI = applyProviderDTO(settings.Providers.OtherAPI, dto)
	case "feature.text", "txt":
		var dto textSettingsDTO
		if err := json.Unmarshal(body, &dto); err != nil {
			return settings, err
		}
		settings.Features.TextEnabled = dto.Enabled
		settings.Text.GenerationAPI = dto.GenerationAPI
		settings.Text.SummaryAPI = dto.SummaryAPI
		settings.Text.CacheMinutes = dto.CacheMinutes
		settings.Text.FallbackImageURL = dto.FallbackImageURL
		settings.Text.EnabledPromptKeys = dto.EnabledPromptKeys
		settings.Text.AcceptedPromptGlob = dto.AcceptedPromptGlob
	case "feature.image", "img":
		var dto imageSettingsDTO
		if err := json.Unmarshal(body, &dto); err != nil {
			return settings, err
		}
		settings.Features.ImageEnabled = dto.Enabled
		settings.Image.API = dto.API
		settings.Image.CacheMinutes = dto.CacheMinutes
		settings.Image.FallbackImageURL = dto.FallbackImageURL
		settings.Image.AcceptedPromptGlob = dto.AcceptedPromptGlob
	case "feature.web", "web":
		var dto webSettingsDTO
		if err := json.Unmarshal(body, &dto); err != nil {
			return settings, err
		}
		settings.Features.WebImgEnabled = dto.Enabled
		settings.Web.SummaryAPI = dto.SummaryAPI
		settings.Web.CacheMinutes = dto.CacheMinutes
		settings.Web.FallbackImageURL = dto.FallbackImageURL
		settings.Web.AllowedHosts = dto.AllowedHosts
		if dto.RepoToken != "" {
			settings.Web.RepoToken = dto.RepoToken
		}
		if dto.YouTubeToken != "" {
			settings.Web.YouTubeToken = dto.YouTubeToken
		}
	case "feature.random", "rand":
		var dto randomSettingsDTO
		if err := json.Unmarshal(body, &dto); err != nil {
			return settings, err
		}
		settings.Features.RandomEnabled = dto.Enabled
		settings.Random.SourceRewriteFrom = dto.SourceRewriteFrom
		settings.Random.FallbackImageURL = dto.FallbackImageURL
		settings.Random.DefaultAPI = dto.DefaultAPI
	case "security.dashboard", "security":
		var dto dashboardSecurityDTO
		if err := json.Unmarshal(body, &dto); err != nil {
			return settings, err
		}
		if dto.PasswordHash != "" {
			settings.Security.DashboardPasswordHash = dto.PasswordHash
		}
		settings.Security.DashboardAllowedIPs = dto.DashboardAllowedIPs
		settings.Security.AllowedReferers = dto.AllowedReferers
	case "prompt", "contxt":
		var dto promptSettingsDTO
		if err := json.Unmarshal(body, &dto); err != nil {
			return settings, err
		}
		settings.Prompts.GenerationContext = dto.GenerationContext
		settings.Prompts.SummaryContext = dto.SummaryContext
		settings.Prompts.Templates = dto.Templates
	case "security.task_behavior", "taskBehavior":
		var dto taskBehaviorDTO
		if err := json.Unmarshal(body, &dto); err != nil {
			return settings, err
		}
		settings.Task.ExceptDomains = dto.ExceptDomains
		settings.Task.ExceptInfos = dto.ExceptInfos
	case "security.text_prompt", "txtSecurity":
		var dto textPromptSecurityDTO
		if err := json.Unmarshal(body, &dto); err != nil {
			return settings, err
		}
		settings.Text.AcceptedPromptGlob = dto.AcceptedPromptGlob
	case "security.image_prompt", "imgSecurity":
		var dto imagePromptSecurityDTO
		if err := json.Unmarshal(body, &dto); err != nil {
			return settings, err
		}
		settings.Image.AcceptedPromptGlob = dto.AcceptedPromptGlob
	default:
		return settings, errors.New("Setting part not found")
	}
	return util.NormalizeSettings(settings), nil
}

/**
 * @brief 将提供方配置转换为接口 DTO。
 * @param provider 内部提供方配置。
 * @return providerSettingsDTO 前端可消费的设置结构。
 */
func providerDTO(provider util.ProviderConfig) providerSettingsDTO {
	return providerSettingsDTO{
		APIKeySet:        provider.APIKey != "",
		TextModel:        provider.TextModel,
		SummaryModel:     provider.SummaryModel,
		ImageModel:       provider.ImageModel,
		ImageSize:        provider.ImageSize,
		EmbeddingModel:   provider.EmbeddingModel,
		Endpoint:         provider.Endpoint,
		APIType:          provider.APIType,
		Temperature:      provider.Temperature,
		MaxTokens:        provider.MaxTokens,
		TopP:             provider.TopP,
		PresencePenalty:  provider.PresencePenalty,
		FrequencyPenalty: provider.FrequencyPenalty,
		CustomHeaders:    provider.CustomHeaders,
		Enabled:          provider.Enabled,
	}
}

/**
 * @brief 将接口 DTO 回写到提供方配置中。
 * @param provider 原始提供方配置。
 * @param dto 前端提交的配置结构。
 * @return util.ProviderConfig 更新后的提供方配置。
 */
func applyProviderDTO(provider util.ProviderConfig, dto providerSettingsDTO) util.ProviderConfig {
	if dto.APIKey != "" {
		provider.APIKey = dto.APIKey
	}
	provider.TextModel = dto.TextModel
	provider.SummaryModel = dto.SummaryModel
	provider.ImageModel = dto.ImageModel
	provider.ImageSize = dto.ImageSize
	provider.EmbeddingModel = dto.EmbeddingModel
	provider.Endpoint = dto.Endpoint
	provider.APIType = dto.APIType
	provider.Temperature = dto.Temperature
	provider.MaxTokens = dto.MaxTokens
	provider.TopP = dto.TopP
	provider.PresencePenalty = dto.PresencePenalty
	provider.FrequencyPenalty = dto.FrequencyPenalty
	if dto.CustomHeaders != nil {
		provider.CustomHeaders = dto.CustomHeaders
	}
	provider.Enabled = dto.Enabled
	return provider
}

/**
 * @brief 对完整应用设置做一致性校验。
 * @param settings 待校验的应用设置。
 * @return error 校验失败时返回错误。
 */
func validateSettings(settings util.AppSettings) error {
	if err := validateProviderAPI(settings.Text.GenerationAPI, true); err != nil {
		return err
	}
	if err := validateProviderAPI(settings.Text.SummaryAPI, true); err != nil {
		return err
	}
	if err := validateProviderAPI(settings.Image.API, false); err != nil {
		return err
	}
	if err := validateProviderAPI(settings.Web.SummaryAPI, true); err != nil {
		return err
	}
	if settings.Text.CacheMinutes < 0 || settings.Image.CacheMinutes < 0 || settings.Web.CacheMinutes < 0 {
		return errors.New("cache_minutes must be greater than or equal to 0")
	}
	for _, rawURL := range []string{settings.Text.FallbackImageURL, settings.Image.FallbackImageURL, settings.Web.FallbackImageURL, settings.Random.FallbackImageURL} {
		if err := validateOptionalURL(rawURL); err != nil {
			return err
		}
	}
	for _, endpoint := range []string{
		settings.Providers.OpenAI.Endpoint,
		settings.Providers.DeepSeek.Endpoint,
		settings.Providers.Alibaba.Endpoint,
		settings.Providers.Anthropic.Endpoint,
		settings.Providers.Gemini.Endpoint,
		settings.Providers.Azure.Endpoint,
		settings.Providers.Moonshot.Endpoint,
		settings.Providers.OtherAPI.Endpoint,
	} {
		if err := validateOptionalURL(endpoint); err != nil {
			return err
		}
	}
	return nil
}

/**
 * @brief 校验提供方 API 名称是否合法。
 * @param api 提供方标识。
 * @param allowOther 是否允许 `otherapi`。
 * @return error 非法时返回错误。
 */
func validateProviderAPI(api string, allowOther bool) error {
	switch api {
	case "openai", "deepseek", "alibaba", "anthropic", "gemini", "azure", "moonshot":
		return nil
	case "otherapi":
		if allowOther {
			return nil
		}
	}
	return errors.New("invalid provider api")
}

/**
 * @brief 校验可选 URL 字段是否合法。
 * @param rawURL 原始 URL 字符串。
 * @return error URL 非法时返回错误。
 */
func validateOptionalURL(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return nil
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("invalid url")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("invalid url scheme")
	}
	return nil
}
