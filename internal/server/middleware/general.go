package middleware

import (
	"fmt"
	"sync"
	"time"
	"urlAPI/internal/database"
	"urlAPI/util"
)

/**
 * @brief IP 频率限制过滤键。
 */
type FrequencyFilter struct {
	Type string `json:"type"`
	IP   string `json:"ip"`
}

/**
 * @brief 单个来源在时间窗口内的访问计数。
 */
type FrequencyData struct {
	Counter int       `json:"counter"`
	Time    time.Time `json:"time"`
}

/**
 * @brief 线程安全的 IP 访问频率缓存。
 */
type SafeIPFrequency struct {
	mu          sync.Mutex
	IPFrequency map[FrequencyFilter]FrequencyData
}

var IPFrequency = SafeIPFrequency{
	IPFrequency: make(map[FrequencyFilter]FrequencyData),
}

/**
 * @brief 执行通用安全检查流程。
 * @param general 请求安全上下文。
 */
func checkGeneralSecurity(general *General) {
	checkFrequency(general)
	checkException(general)
	checkInfo(general)
}

/**
 * @brief 检查单个 IP 的访问频率。
 * @param general 请求安全上下文。
 */
func checkFrequency(general *General) {
	// 上锁，解锁
	IPFrequency.mu.Lock()
	defer IPFrequency.mu.Unlock()

	filter := FrequencyFilter{
		Type: general.Type,
		IP:   general.IP,
	}
	value, exists := IPFrequency.IPFrequency[filter]
	if !exists {
		value = FrequencyData{}
		value.Counter = 1
		value.Time = time.Now()
		IPFrequency.IPFrequency[filter] = value
		return
	}
	switch {
	case general.Time.Sub(value.Time).Seconds() <= 0.25 && value.Counter >= 10:
		general.Unsafe = true
		general.Info = fmt.Sprintf("%s accessed too frequently", general.IP)
		return
	case general.Time.Sub(value.Time).Seconds() > 0.25:
		value.Counter = 1
		value.Time = time.Now()
	case value.Counter < 10:
		value.Counter++
	}
	IPFrequency.IPFrequency[filter] = value
	return
}

/**
 * @brief 校验请求目标和来源域名是否合法。
 * @param general 请求安全上下文。
 */
func checkInfo(general *General) {
	if general.Target == "" {
		general.Info = "Empty Target"
		general.Unsafe = true
	}
	allowedref := database.SettingsStore.Get().Security.AllowedReferers
	domain := util.GetDomain(general.Referer)
	if !util.WildcardChecker(&allowedref, &domain) || general.Referer == "" {
		general.Info = fmt.Sprintf("Referer %s not allowed", general.Referer)
		general.Unsafe = true
	}
	return
}

/**
 * @brief 检查当前请求是否命中跳过任务记录的例外规则。
 * @param general 请求安全上下文。
 */
func checkException(general *General) {
	settings := database.SettingsStore.Get()
	taskexceptdomain := settings.Task.ExceptDomains
	taskexceptinfo := settings.Task.ExceptInfos
	domain := util.GetDomain(general.Referer)
	auxInfo := general.Info
	if util.WildcardChecker(&taskexceptdomain, &domain) || util.WildcardChecker(&taskexceptinfo, &auxInfo) {
		general.SkipDB = true
		return
	}
	return
}
