package op

import (
	"encoding/json"
	"os"
	"sync"
	"urlAPI/internal/database"
	"urlAPI/internal/model"
)

/**
 * @brief 线程安全的任务缓存队列。
 */
type SafeTaskQueue struct {
	Mu    sync.RWMutex
	Queue map[TaskQueueFilter]TaskQueueItem
}

/**
 * @brief 线程安全的任务计数器。
 */
type SafeTaskCounter struct {
	Mu      sync.RWMutex
	Counter map[string]int
}

var (
	db        *database.SQLiteAdapter
	ImgPath   = "assets/img/"
	TaskQueue = SafeTaskQueue{
		Queue: make(map[TaskQueueFilter]TaskQueueItem),
	}
	TaskCounter = SafeTaskCounter{
		Counter: make(map[string]int),
	}
)

/**
 * @brief 初始化运行期目录和数据库引用。
 * @return error 初始化失败时返回错误。
 */
func Init() error {
	db = database.GetLocalDB()
	if err := os.RemoveAll(ImgPath); err != nil {
		return err
	}
	return os.MkdirAll(ImgPath, 0777)
}

/**
 * @brief 根据提供方名称读取其接口端点。
 * @param api 提供方标识。
 * @return string 目标端点地址，不存在时返回空字符串。
 */
func getEndpoint(api string) string {
	provider, ok := database.SettingsStore.Get().Providers.ByName(api)
	if !ok {
		return ""
	}
	return provider.Endpoint
}

/**
 * @brief 会话接口的数据传输对象。
 *
 * 同时承载前后端交互中用于读取和修改配置、任务及仓库的字段。
 */
type Session struct {
	// backend -> frontend
	SessionToken string          `json:"session_token"`
	SessionIP    string          `json:"session_ip"`
	SettingName  []string        `json:"setting_name"`
	SettingData  [][]string      `json:"setting_data"`
	SettingBody  json.RawMessage `json:"setting_body,omitempty"`
	TaskData     []model.Task    `json:"task_data"`
	TaskMaxPage  int             `json:"task_max_page"`
	RepoData     []model.Repo    `json:"repo_data"`

	// frontend -> backend
	Operation    string     `json:"operation"`
	LoginTerm    bool       `json:"login_term"`
	SettingEdit  [][]string `json:"setting_edit"`
	TaskCatagory string     `json:"task_catagory"`
	TaskBy       string     `json:"task_by"`
	TaskPage     int        `json:"task_page"`
	RepoAPI      string     `json:"repo_api"`
	RepoInfo     string     `json:"repo_info"`
	RepoUUID     string     `json:"repo_uuid"`

	//both
	SettingPart string `json:"setting_part"`
}

/**
 * @brief 生成类接口的统一返回结构。
 */
type GenerateResult struct {
	Prompt         string `json:"prompt"`
	OriginalPrompt string `json:"original_prompt"`
	ActualPrompt   string `json:"actual_prompt"`
	Response       string `json:"response"`
	URL            string `json:"url"`
}

/**
 * @brief 缓存中的任务执行状态。
 */
type TaskQueueItem struct {
	DB      model.Task `json:"db"`
	Return  GenerateResult
	Running bool `json:"running"`
}

/**
 * @brief 任务缓存去重键。
 */
type TaskQueueFilter struct {
	Type   string `json:"type"`
	Size   string `json:"size"`
	Target string `json:"target"`
	API    string `json:"api"`
}
