package anydoc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	dingtalkListPath   = "/api/docs/dingtalk/list"
	dingtalkExportPath = "/api/docs/dingtalk/export"
)

// DingtalkListDocs 获取 dingtalk 文档列表
func (c *Client) DingtalkListDocs(ctx context.Context, uuid string, dingtalkSetting DingtalkSetting) (*ListDocResponse, error) {
	u, err := url.Parse(crawlerServiceHost)
	if err != nil {
		return nil, err
	}
	u.Path = dingtalkListPath
	requestURL := u.String()

	bodyMap := map[string]interface{}{
		"uuid":       uuid,
		"app_id":     dingtalkSetting.AppID,
		"app_secret": dingtalkSetting.AppSecret,
		"unionid":    dingtalkSetting.UnionID,
		"space_id":   dingtalkSetting.SpaceID,
		"phone":      dingtalkSetting.Phone,
	}

	jsonData, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	c.logger.Info("dingtalkListDocs", "requestURL:", requestURL, "resp", string(respBody))

	var dingtalkResp ListDocResponse
	err = json.Unmarshal(respBody, &dingtalkResp)
	if err != nil {
		return nil, err
	}

	if !dingtalkResp.Success {
		return nil, errors.New(dingtalkResp.Msg)
	}

	return &dingtalkResp, nil
}
