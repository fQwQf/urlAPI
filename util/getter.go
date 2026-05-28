package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

/**
 * @brief 根据 User-Agent 判断设备类型。
 * @param ua 用户代理字符串。
 * @return string 设备类型，可能为 `Mobile`、`Desktop`、`Bot` 或空字符串。
 */
func GetDeviceType(ua string) string {
	mobileRegexp := `(?i)(Mobile|Tablet|Android|iOS|iPhone|iPad|iPod)`
	desktopRegexp := `(?i)(Desktop|Windows|Macintosh|Linux|PC)`
	botRegexp := `(?i)(Bot)`
	matched, _ := regexp.MatchString(mobileRegexp, ua)
	if matched {
		return "Mobile"
	}
	matched, _ = regexp.MatchString(desktopRegexp, ua)
	if matched {
		return "Desktop"
	}
	matched, _ = regexp.MatchString(botRegexp, ua)
	if matched {
		return "Bot"
	}
	return ""
}

/**
 * @brief 根据 IP 获取地区信息，并使用内存缓存减少重复请求。
 * @param ip 待查询的 IP 地址。
 * @return string 地区名称，失败时返回 `Unknown`。
 */
func GetRegion(ip string) string {
	if value, ok := IPTmp[ip]; ok {
		return value
	}
	url := "https://api.live.bilibili.com/ip_service/v1/ip_service/get_ip_addr?ip=" + ip
	resp, err := GlobalHTTPClient.Get(url)
	if err != nil {
		return "Unknown"
	}
	defer resp.Body.Close()
	jsonResp, err := io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != http.StatusOK {
		return "Unknown"
	}
	var response RegionResp
	err = json.Unmarshal(jsonResp, &response)
	if err != nil {
		return "Unknown"
	}

	var region string
	if response.Data.Country == "中国" {
		region = response.Data.Province
	} else {
		region = response.Data.Country
	}
	if len(IPTmp) >= 1000 {
		IPTmp = make(map[string]string)
	}

	IPTmp[ip] = region
	return region
}

/**
 * @brief 下载指定 URL 的原始内容。
 * @param url 目标地址。
 * @return []byte 下载到的字节数据。
 * @return error 下载失败时返回错误。
 */
func Downloader(url string) ([]byte, error) {
	resp, err := GlobalHTTPClient.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, errors.WithMessage(err, resp.Status)
	}
	defer resp.Body.Close()
	ret, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	} else {
		return ret, nil
	}
}

/**
 * @brief 获取仓库目录接口中的下载链接列表。
 * @param url 仓库内容接口地址。
 * @return []string 下载地址列表。
 * @return error 请求或解析失败时返回错误。
 */
func GetRepo(url string) ([]string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	resp, err := GlobalHTTPClient.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer resp.Body.Close()
	jsonResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var response []RepoContentResp
	if err = json.Unmarshal(jsonResponse, &response); err != nil {
		return nil, errors.WithStack(err)
	}
	var ret []string
	for _, repo := range response {
		ret = append(ret, repo.DownloadURL)
	}
	return ret, nil
}

/**
 * @brief 从 URL 中提取域名。
 * @param URL 原始 URL 字符串。
 * @return string 域名，解析失败时返回空字符串。
 */
func GetDomain(URL string) string {
	domainParse, err := url.Parse(URL)
	if err != nil {
		return ""
	}
	return domainParse.Hostname()
}

/**
 * @brief 将 `yyyy.mm` 格式的字符串转换为时间对象。
 * @param ori 原始日期字符串。
 * @return time.Time 转换后的 UTC 时间。
 */
func GetDate(ori string) time.Time {
	parts := strings.Split(ori, ".")
	year, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])
	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
}

/**
 * @brief 生成 64 位十六进制随机字符串。
 * @return string 随机字符串。
 */
func GetRandomString() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	randomNumber := n.String()
	hash := sha256.Sum256([]byte(randomNumber))
	hashStr := hex.EncodeToString(hash[:])
	return hashStr
}

/**
 * @brief 生成指定长度的随机字符串。
 * @param len 目标长度，超过 64 时返回完整随机串。
 * @return string 随机字符串。
 */
func GetShortRandomString(len int) string {
	if len >= 64 {
		return GetRandomString()
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	randomNumber := n.String()
	hash := sha256.Sum256([]byte(randomNumber))
	hashStr := hex.EncodeToString(hash[:])
	return hashStr[:len]
}
