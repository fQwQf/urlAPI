package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"net/http"
	"time"
)

/**
 * @brief 调用文本生成接口（向后兼容，支持额外参数）。
 * @param endpoint 接口地址。
 * @param token 鉴权令牌。
 * @param model 模型名称。
 * @param context 系统上下文提示词。
 * @param prompt 用户输入提示词。
 * @return string 文本生成结果。
 * @return error 调用失败或响应非法时返回错误。
 */
func Txt(endpoint, token, model, context, prompt string) (string, error) {
	return TxtWithParams(endpoint, token, model, context, prompt, 0, 0, 0, 0)
}

/**
 * @brief 调用文本生成接口（支持完整参数）。
 * @param endpoint 接口地址。
 * @param token 鉴权令牌。
 * @param model 模型名称。
 * @param context 系统上下文提示词。
 * @param prompt 用户输入提示词。
 * @param temperature 温度参数。
 * @param maxTokens 最大token数。
 * @param topP Top-P参数。
 * @param presencePenalty 存在惩罚。
 * @return string 文本生成结果。
 * @return error 调用失败或响应非法时返回错误。
 */
func TxtWithParams(endpoint, token, model, context, prompt string, temperature float64, maxTokens int, topP, presencePenalty float64) (string, error) {
	if endpoint == "" || token == "" || model == "" || context == "" || prompt == "" {
		return "", errors.WithStack(errors.New("Util TxtAPI insufficient info"))
	}
	userMessage := TxtMessage{
		Role:    "user",
		Content: prompt,
	}
	developerMessage := TxtMessage{
		Role:    "system",
		Content: context,
	}
	txtPayload := TxtPayload{
		Model:    model,
		Messages: []TxtMessage{developerMessage, userMessage},
	}
	if temperature > 0 {
		txtPayload.Temperature = temperature
	}
	if maxTokens > 0 {
		txtPayload.MaxTokens = maxTokens
	}
	if topP > 0 {
		txtPayload.TopP = topP
	}
	if presencePenalty != 0 {
		txtPayload.PresencePenalty = presencePenalty
	}
	jsonPayload, err := json.Marshal(txtPayload)
	if err != nil {
		return "", errors.WithStack(err)
	}
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := GlobalHTTPClient.Do(req)
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer resp.Body.Close()
	var txtResp TxtResp
	jsonResponse, err := io.ReadAll(resp.Body)
	err = json.Unmarshal(jsonResponse, &txtResp)
	if err != nil || resp.StatusCode != http.StatusOK {
		return "", errors.WithMessage(err, resp.Status)
	} else {
		return txtResp.Choices[0].Message.Content, nil
	}
}

/**
 * @brief 调用阿里云文生图接口。
 * @param token 鉴权令牌。
 * @param prompt 用户提示词。
 * @param model 模型名称。
 * @param size 图像尺寸。
 * @return []byte 生成图像字节。
 * @return string 实际生效提示词。
 * @return error 调用失败时返回错误。
 */
func AlibabaImg(token, prompt, model, size string) ([]byte, string, error) {
	imgInput := AlibabaImgInput{
		Prompt: prompt,
	}
	imgParameter := AlibabaImgParameters{
		Size: size,
		N:    1,
	}
	imgPayload := AlibabaImgPayload{
		Model:      model,
		Input:      imgInput,
		Parameters: imgParameter,
	}
	jsonPayload, _ := json.Marshal(imgPayload)
	req, _ := http.NewRequest("POST", "https://dashscope.aliyuncs.com/api/v1/services/aigc/text2image/image-synthesis", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-DashScope-Async", "enable")
	resp, err := GlobalHTTPClient.Do(req)
	if err != nil {
		return nil, "", errors.WithStack(err)
	}
	defer resp.Body.Close()
	var response AlibabaImgResp
	jsonResponse, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(jsonResponse, &response)
	if err != nil {
		return nil, "", errors.WithStack(err)
	}
	id := response.Output.TaskID

	timer := time.NewTimer(time.Second * 30)
	timeout := make(chan bool)
	go func() {
		<-timer.C
		log.Println("Times up")
		timeout <- true
	}()

	for status := response.Output.TaskStatus; status == "PENDING" || status == "RUNNING"; status = response.Output.TaskStatus {
		time.Sleep(1 * time.Second)
		fmt.Println(status)
		if err = json.Unmarshal(alibabaFetchImgTask(id, token), &response); err != nil {
			timer.Stop()
			return nil, "", errors.WithStack(err)
		}
	}
	timer.Stop()

	if response.Output.TaskStatus != "SUCCEEDED" {
		return nil, "", errors.WithStack(err)
	}
	actualPrompt := response.Output.Results[0].ActualPrompt
	ret, err := Downloader(response.Output.Results[0].URL)
	return ret, actualPrompt, nil
}

/**
 * @brief 查询阿里云异步文生图任务状态。
 * @param id 任务 ID。
 * @param token 鉴权令牌。
 * @return []byte 原始响应字节，失败时返回 nil。
 */
func alibabaFetchImgTask(id, token string) []byte {
	req, _ := http.NewRequest("GET", "https://dashscope.aliyuncs.com/api/v1/tasks/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := GlobalHTTPClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	jsonResponse, _ := io.ReadAll(resp.Body)
	return jsonResponse
}

/**
 * @brief 调用 OpenAI 图像生成接口。
 * @param endpoint 接口地址。
 * @param token 鉴权令牌。
 * @param prompt 用户提示词。
 * @param model 模型名称。
 * @param size 图像尺寸。
 * @return []byte 生成图像字节。
 * @return error 调用失败时返回错误。
 */
func OpenaiImg(endpoint, token, prompt, model, size string) ([]byte, error) {
	imgPayload := OpenaiImgPayload{
		Model:  model,
		Prompt: prompt,
		Size:   size,
		N:      1,
	}
	jsonPayload, _ := json.Marshal(imgPayload)
	req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := GlobalHTTPClient.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer resp.Body.Close()
	jsonResponse, err := io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, errors.WithStack(err)
	}
	var response OpenaiImgResp
	if err = json.Unmarshal(jsonResponse, &response); err != nil {
		return nil, errors.WithStack(err)
	}
	ret, err := Downloader(response.Data[0].URL)
	return ret, errors.WithStack(err)
}
