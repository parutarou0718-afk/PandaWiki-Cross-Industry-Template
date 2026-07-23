package rag

import (
	"context"
)

// GetModelConfig 获取模型配置
func (c *Client) AddModelConfig(ctx context.Context, req AddModelConfigRequest) (*ModelConfig, error) {
	httpReq, err := c.newRequest(ctx, "POST", "models", req)
	if err != nil {
		return nil, err
	}
	var resp AddModelConfigResponse
	if err := c.do(httpReq, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) GetModelConfigList(ctx context.Context) ([]ModelConfig, error) {
	httpReq, err := c.newRequest(ctx, "GET", "models", nil)
	if err != nil {
		return nil, err
	}
	var resp ListModelConfigsResponse
	if err := c.do(httpReq, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) DeleteModelConfig(ctx context.Context, models []ModelItem) error {
	httpReq, err := c.newRequest(ctx, "DELETE", "models", DeleteModelConfigsRequest{Models: models})
	if err != nil {
		return err
	}
	var resp CommonResponse
	if err := c.do(httpReq, &resp); err != nil {
		return err
	}
	return nil
}
