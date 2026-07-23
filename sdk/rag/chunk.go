package rag

import (
	"context"
	"fmt"
)

// AddChunk 向指定文档添加分块
func (c *Client) AddChunk(ctx context.Context, datasetID, documentID string, req AddChunkRequest) (*Chunk, error) {
	path := fmt.Sprintf("datasets/%s/documents/%s/chunks", datasetID, documentID)
	httpReq, err := c.newRequest(ctx, "POST", path, req)
	if err != nil {
		return nil, err
	}
	var resp AddChunkResponse
	if err := c.do(httpReq, &resp); err != nil {
		return nil, err
	}
	return &resp.Data.Chunk, nil
}

// ListChunks 列出指定文档的分块
func (c *Client) ListChunks(ctx context.Context, datasetID, documentID string, params map[string]string) ([]Chunk, int, error) {
	path := fmt.Sprintf("datasets/%s/documents/%s/chunks", datasetID, documentID)
	httpReq, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, 0, err
	}
	q := httpReq.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	httpReq.URL.RawQuery = q.Encode()
	var resp ListChunksResponse
	if err := c.do(httpReq, &resp); err != nil {
		return nil, 0, err
	}
	return resp.Data.Chunks, resp.Data.Total, nil
}

// DeleteChunks 删除指定文档的分块（支持批量）
func (c *Client) DeleteChunks(ctx context.Context, datasetID, documentID string, chunkIDs []string) error {
	path := fmt.Sprintf("datasets/%s/documents/%s/chunks", datasetID, documentID)
	body := DeleteChunksRequest{ChunkIDs: chunkIDs}
	httpReq, err := c.newRequest(ctx, "DELETE", path, body)
	if err != nil {
		return err
	}
	var resp DeleteChunksResponse
	return c.do(httpReq, &resp)
}

// UpdateChunk 更新指定分块内容
func (c *Client) UpdateChunk(ctx context.Context, datasetID, documentID, chunkID string, req UpdateChunkRequest) error {
	path := fmt.Sprintf("datasets/%s/documents/%s/chunks/%s", datasetID, documentID, chunkID)
	httpReq, err := c.newRequest(ctx, "PUT", path, req)
	if err != nil {
		return err
	}
	var resp UpdateChunkResponse
	return c.do(httpReq, &resp)
}

// ParseDocuments 解析指定文档（批量）
func (c *Client) ParseDocuments(ctx context.Context, datasetID string, documentIDs []string) error {
	path := fmt.Sprintf("datasets/%s/chunks", datasetID)
	body := ParseDocumentsRequest{DocumentIDs: documentIDs}
	httpReq, err := c.newRequest(ctx, "POST", path, body)
	if err != nil {
		return err
	}
	var resp ParseDocumentsResponse
	return c.do(httpReq, &resp)
}

// StopParseDocuments 停止解析指定文档（批量）
func (c *Client) StopParseDocuments(ctx context.Context, datasetID string, documentIDs []string) error {
	path := fmt.Sprintf("datasets/%s/chunks", datasetID)
	body := StopParseDocumentsRequest{DocumentIDs: documentIDs}
	httpReq, err := c.newRequest(ctx, "DELETE", path, body)
	if err != nil {
		return err
	}
	var resp StopParseDocumentsResponse
	return c.do(httpReq, &resp)
}
