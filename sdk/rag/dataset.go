package rag

import (
	"context"
	"fmt"
)

// CreateDataset 创建数据集
func (c *Client) CreateDataset(ctx context.Context, req CreateDatasetRequest) (*Dataset, error) {
	httpReq, err := c.newRequest(ctx, "POST", "datasets", req)
	if err != nil {
		return nil, err
	}
	var resp CreateDatasetResponse
	if err := c.do(httpReq, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// DeleteDatasets 删除数据集（支持批量）
func (c *Client) DeleteDatasets(ctx context.Context, ids []string) error {
	reqBody := DeleteDatasetsRequest{IDs: ids}
	httpReq, err := c.newRequest(ctx, "DELETE", "datasets", reqBody)
	if err != nil {
		return err
	}
	var resp DeleteDatasetsResponse
	return c.do(httpReq, &resp)
}

// UpdateDataset 更新数据集
func (c *Client) UpdateDataset(ctx context.Context, datasetID string, req UpdateDatasetRequest) error {
	path := fmt.Sprintf("datasets/%s", datasetID)
	httpReq, err := c.newRequest(ctx, "PUT", path, req)
	if err != nil {
		return err
	}
	var resp UpdateDatasetResponse
	return c.do(httpReq, &resp)
}

// ListDatasets 列出数据集
func (c *Client) ListDatasets(ctx context.Context, req ListDatasetsRequest) ([]Dataset, error) {
	httpReq, err := c.newRequest(ctx, "GET", "datasets", nil)
	if err != nil {
		return nil, err
	}
	q := httpReq.URL.Query()
	if req.Page > 0 {
		q.Add("page", fmt.Sprintf("%d", req.Page))
	}
	if req.PageSize > 0 {
		q.Add("page_size", fmt.Sprintf("%d", req.PageSize))
	}
	if req.OrderBy != "" {
		q.Add("orderby", req.OrderBy)
	}
	q.Add("desc", fmt.Sprintf("%t", req.Desc))
	if req.Name != "" {
		q.Add("name", req.Name)
	}
	if req.ID != "" {
		q.Add("id", req.ID)
	}
	httpReq.URL.RawQuery = q.Encode()
	var resp ListDatasetsResponse
	if err := c.do(httpReq, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
