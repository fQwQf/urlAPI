package model

import "time"

/**
 * @brief 仓库缓存记录。
 *
 * 用于保存仓库接口、标识信息及解析后的内容列表。
 */
type Repo struct {
	UUID    string `json:"uuid" gorm:"primaryKey"`
	API     string `json:"api"`
	Info    string `json:"info"`
	Content string `json:"content"`
}

/**
 * @brief 后台登录会话记录。
 */
type Session struct {
	Token  string    `json:"token" gorm:"primaryKey"`
	Expire time.Time `json:"expire"`
	Term   bool      `json:"term"`
}

/**
 * @brief 任务记录模型。
 *
 * 保存一次生成或下载请求的来源、参数和处理结果。
 */
type Task struct {
	UUID     string    `json:"uuid" gorm:"primaryKey"`
	Time     time.Time `json:"time"`
	IP       string    `json:"ip"`
	Type     string    `json:"type"`
	Status   string    `json:"status"`
	Target   string    `json:"target"`
	Return   string    `json:"return"`
	Region   string    `json:"region"`
	Referer  string    `json:"referer"`
	Device   string    `json:"device"`
	MoreInfo string    `json:"more_info" gorm:"more_info"`
	API      string    `json:"api"`
	Model    string    `json:"model"`
	Temp     string    `json:"temp"`
	Size     string    `json:"size"`
}

/**
 * @brief 标量应用设置项。
 */
type AppSetting struct {
	Key       string    `json:"key" gorm:"primaryKey"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at" gorm:"-"`
	UpdatedAt time.Time `json:"updated_at" gorm:"-"`
	Version   int       `json:"version" gorm:"-"`
}

/**
 * @brief AI 服务提供方配置模型。
 */
type Provider struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	Name             string    `json:"name" gorm:"unique;not null"`
	APIKeyEnc        string    `json:"api_key_enc"`
	TextModel        string    `json:"text_model"`
	SummaryModel     string    `json:"summary_model"`
	ImageModel       *string   `json:"image_model"`
	ImageSize        *string   `json:"image_size"`
	EmbeddingModel   string    `json:"embedding_model"`
	Endpoint         string    `json:"endpoint"`
	APIType          string    `json:"api_type"`
	Temperature      float64   `json:"temperature" gorm:"default:1"`
	MaxTokens        int       `json:"max_tokens" gorm:"default:0"`
	TopP             float64   `json:"top_p" gorm:"default:1"`
	PresencePenalty  float64   `json:"presence_penalty" gorm:"default:0"`
	FrequencyPenalty float64   `json:"frequency_penalty" gorm:"default:0"`
	CustomHeaders    string    `json:"custom_headers" gorm:"type:json"`
	Enabled          bool      `json:"enabled" gorm:"default:1"`
	UpdatedAt        time.Time `json:"updated_at"`
}

/**
 * @brief 业务服务配置模型。
 */
type ServiceConfig struct {
	Service          string `json:"service" gorm:"primaryKey"`
	CacheMinutes     int    `json:"cache_minutes"`
	FallbackImageURL string `json:"fallback_image_url"`
	Settings         string `json:"settings" gorm:"type:json"`
}

/**
 * @brief 提示词模板配置模型。
 */
type Prompt struct {
	Key      string `json:"key" gorm:"primaryKey"`
	Template string `json:"template" gorm:"not null"`
}

/**
 * @brief 返回提示词表名。
 * @return string 提示词在数据库中的表名。
 */
func (Prompt) TableName() string {
	return "prompts"
}

/**
 * @brief 通用配置列表项模型。
 */
type ConfigListItem struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Scope     string `json:"scope" gorm:"not null;uniqueIndex:idx_config_list_scope_value"`
	Value     string `json:"value" gorm:"not null;uniqueIndex:idx_config_list_scope_value"`
	SortOrder int    `json:"sort_order" gorm:"default:0"`
}

/**
 * @brief 数据库批量查询结果集合。
 */
type DBList struct {
	RepoList       []Repo
	TaskList       []Task
	SessionList    []Session
	AppSettingList []AppSetting
}
