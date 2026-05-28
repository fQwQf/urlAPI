package util

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"image/png"
	"io"
	"net/http"
	"strconv"
	"strings"
	"urlAPI/file"
)

/**
 * @brief 生成 Bilibili 视频信息图。
 * @param ABV 视频 av/BV 标识。
 * @return []byte PNG 图片字节。
 * @return error 请求、解析或绘制失败时返回错误。
 */
func Bili(ABV string) ([]byte, error) {
	var url string
	if ABV[0] == 'a' {
		ABV = ABV[2:]
		url = "https://api.bilibili.com/x/web-interface/view?aid=" + ABV
	} else if ABV[0] == 'B' {
		url = "https://api.bilibili.com/x/web-interface/view?bvid=" + ABV
	} else {
		return nil, errors.New("Util Bili Invalid ABV")
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	resp, err := GlobalHTTPClient.Do(req)
	switch {
	case err != nil:
		return nil, errors.WithStack(err)
	case resp.StatusCode != http.StatusOK:
		return nil, errors.WithStack(errors.New(resp.Status))
	}
	defer resp.Body.Close()
	jsonResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var info BiliResp
	err = json.Unmarshal(jsonResp, &info)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	picURL := info.Data.Pic
	name := info.Data.Title
	author := info.Data.Owner.Name
	description := info.Data.Desc
	view := biliGetStr(info.Data.Stat.View)
	favorite := biliGetStr(info.Data.Stat.Favorite)
	like := biliGetStr(info.Data.Stat.Like)
	coin := biliGetStr(info.Data.Stat.Coin)
	ret, err := DrawVideo(picURL, name, author, description, view, favorite, like, coin)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return ret, nil
}

/**
 * @brief 生成 YouTube 视频信息图。
 * @param ID 视频 ID。
 * @param Token YouTube API Token。
 * @return []byte PNG 图片字节。
 * @return error 请求、解析或绘制失败时返回错误。
 */
func Ytb(ID, Token string) ([]byte, error) {
	url := "https://www.googleapis.com/youtube/v3/videos?part=snippet,statistics&id=" + ID + "&key=" + Token
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	resp, err := GlobalHTTPClient.Do(req)
	switch {
	case err != nil:
		return nil, errors.WithStack(err)
	case resp.StatusCode != http.StatusOK:
		return nil, errors.WithStack(errors.New(resp.Status))
	}
	defer resp.Body.Close()
	jsonResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var info YtbResp
	err = json.Unmarshal(jsonResp, &info)
	name := info.Items[0].Snippet.Title
	author := info.Items[0].Snippet.ChannelTitle
	description := info.Items[0].Snippet.Description
	picURL := info.Items[0].Snippet.Thumbnails.Standard.URL
	view := info.Items[0].Statistics.ViewCount
	like := info.Items[0].Statistics.LikeCount
	favorite := "N/A"
	coin := "N/A"
	ret, err := DrawVideo(picURL, name, author, description, view, favorite, like, coin)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return ret, nil
}

/**
 * @brief 生成 arXiv 论文信息图。
 * @param URL 论文页面地址。
 * @return []byte PNG 图片字节。
 * @return error 请求、解析或绘制失败时返回错误。
 */
func Arxiv(URL string) ([]byte, error) {
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	resp, err := GlobalHTTPClient.Do(req)
	switch {
	case err != nil:
		return nil, errors.WithStack(err)
	case resp.StatusCode != http.StatusOK:
		return nil, errors.WithStack(errors.New(resp.Status))
	}
	defer resp.Body.Close()
	rawResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithMessage(err, resp.Status)
	}
	doc, err := html.Parse(bytes.NewReader(rawResp))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	id := URL[22:]
	title, author, description := traverseArxiv(doc, "", "", "")
	logoFile, err := file.Logos.Open("logo/arxiv_logo.png")
	logoImg, err := png.Decode(logoFile)
	ret, err := DrawArticle(logoImg, id, title, author, description, "")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return ret, nil
}

/**
 * @brief 生成 IT 之家文章信息图。
 * @param URL 文章地址。
 * @param endpoint 文本总结接口地址。
 * @param token 鉴权令牌。
 * @param model 文本模型名称。
 * @param context 总结上下文提示词。
 * @return []byte PNG 图片字节。
 * @return error 请求、总结或绘制失败时返回错误。
 */
func ITHome(URL, endpoint, token, model, context string) ([]byte, error) {
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	resp, err := GlobalHTTPClient.Do(req)
	switch {
	case err != nil:
		return nil, errors.WithStack(err)
	case resp.StatusCode != http.StatusOK:
		return nil, errors.WithStack(errors.New(resp.Status))
	}
	defer resp.Body.Close()
	rawResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	doc, err := html.Parse(bytes.NewReader(rawResp))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	title, tim, content := traverseITHome(doc, "", "", "")
	description, err := Txt(endpoint, token, model, context, content)
	logoFile, err := file.Logos.Open("assets/logo/ithome_logo.png")
	logoImg, err := png.Decode(logoFile)
	ret, err := DrawArticle(logoImg, "", title, "", description, tim)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return ret, nil
}

/**
 * @brief 生成 GitHub 或 Gitee 仓库信息图。
 * @param URL 仓库地址。
 * @param Token GitHub API Token。
 * @return []byte PNG 图片字节。
 * @return error 请求、解析或绘制失败时返回错误。
 */
func Repo(URL string, Token string) ([]byte, error) {
	var logoURL string
	switch {
	case strings.HasPrefix(URL, "https://github.com"):
		URL = strings.ReplaceAll(URL, "https://github.com", "https://api.github.com/repos")
		logoURL = "logo/github_logo.png"
	case strings.HasPrefix(URL, "https://gitee.com"):
		URL = strings.ReplaceAll(URL, "https://gitee.com", "https://gitee.com/api/v5/repos")
		logoURL = "logo/gitee_logo.png"
	}
	req, err := http.NewRequest("GET", URL, nil)
	if Token != "" && strings.HasPrefix(URL, "https://api.github.com/repos") {
		req.Header.Set("Authorization", "Bearer "+Token)
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}
	resp, err := GlobalHTTPClient.Do(req)
	switch {
	case err != nil:
		return nil, errors.WithStack(err)
	case resp.StatusCode != http.StatusOK:
		return nil, errors.WithStack(errors.New(resp.Status))
	}
	defer resp.Body.Close()
	jsonResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var repo RepoResp
	if err = json.Unmarshal(jsonResp, &repo); err != nil {
		return nil, errors.WithStack(err)
	}
	author := repo.Owner.Login
	name := repo.Name
	description := repo.Description
	forkCount := getRepoCount(repo.ForksCount)
	starCount := getRepoCount(repo.StargazersCount)
	bgFile, err := file.Logos.Open(logoURL)
	bgImg, err := png.Decode(bgFile)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	ret, err := DrawRepo(bgImg, name, author, description, starCount, forkCount)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return ret, nil
}

/**
 * @brief 将 Bilibili 数值格式化为展示文本。
 * @param x 原始数值。
 * @return string 格式化后的展示字符串。
 */
func biliGetStr(x float64) string {
	if x >= 10000 {
		return strconv.FormatFloat(x/10000.0, 'f', 1, 64) + "w"
	} else {
		return strconv.FormatFloat(x, 'f', -1, 64)
	}
}

/**
 * @brief 遍历 arXiv HTML 节点并提取标题、作者和摘要。
 * @param n 当前 HTML 节点。
 * @param title 当前提取到的标题。
 * @param author 当前提取到的作者。
 * @param description 当前提取到的摘要。
 * @return string 标题。
 * @return string 作者。
 * @return string 摘要。
 */
func traverseArxiv(n *html.Node, title, author, description string) (string, string, string) {
	if n.Type == html.ElementNode {
		for _, attr := range n.Attr {
			if n.Data == "h1" && attr.Key == "class" && attr.Val == "title mathjax" {
				title = findItem(n)
			} else if n.Data == "div" && attr.Key == "class" && attr.Val == "authors" {
				author = findItem(n)
			} else if n.Data == "blockquote" && attr.Key == "class" && attr.Val == "abstract mathjax" {
				description = findItem(n)
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		title, author, description = traverseArxiv(c, title, author, description)
	}
	return title, author, description
}

/**
 * @brief 提取节点下的纯文本内容。
 * @param n 当前 HTML 节点。
 * @return string 提取出的文本。
 */
func findItem(n *html.Node) string {
	var ret string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			ret += strings.TrimSpace(c.Data)
		} else if c.Type == html.ElementNode {
			ret += findItem(c)
		}
	}
	return ret
}

/**
 * @brief 遍历 IT 之家 HTML 节点并提取标题、时间和正文。
 * @param n 当前 HTML 节点。
 * @param title 当前提取到的标题。
 * @param tim 当前提取到的时间。
 * @param content 当前提取到的正文。
 * @return string 标题。
 * @return string 时间。
 * @return string 正文。
 */
func traverseITHome(n *html.Node, title, tim, content string) (string, string, string) {
	if n.Type == html.ElementNode {
		for _, attr := range n.Attr {
			if n.Data == "img" && attr.Key == "title" {
				title = attr.Val
			} else if n.Data == "div" && attr.Key == "class" && attr.Val == "post_content" {
				content = findItem(n)
			} else if n.Data == "span" && attr.Key == "id" && attr.Val == "pubtime_baidu" {
				tim = findItem(n)
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		title, tim, content = traverseITHome(c, title, tim, content)
	}
	return title, tim, content
}

/**
 * @brief 将仓库统计数值格式化为简短文本。
 * @param x 原始数值。
 * @return string 格式化后的展示字符串。
 */
func getRepoCount(x float64) string {
	if x >= 1000 {
		return strconv.FormatFloat(x/1000.0, 'f', 1, 64) + "k"
	} else {
		return strconv.FormatFloat(x, 'f', -1, 64)
	}
}
