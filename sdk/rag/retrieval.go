package rag

import (
	"context"
)

// RetrieveChunks 检索分块（向量/关键词检索）
func (c *Client) RetrieveChunks(ctx context.Context, req RetrievalRequest) ([]RetrievalChunk, int, string, error) {
	httpReq, err := c.newRequest(ctx, "POST", "retrieval", req)
	if err != nil {
		return nil, 0, "", err
	}
	var resp RetrievalResponse
	if err := c.do(httpReq, &resp); err != nil {
		return nil, 0, "", err
	}
	return resp.Data.Chunks, resp.Data.Total, resp.Data.RewrittenQuery, nil
}

// RelatedQuestions 生成相关问题（多样化检索）
// 注意：该接口需要 Bearer Login Token，通常与API Key不同
func (c *Client) RelatedQuestions(ctx context.Context, loginToken string, req RelatedQuestionsRequest) ([]string, error) {
	httpReq, err := c.newRequest(ctx, "POST", "/v1/conversation/related_questions", req)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+loginToken)
	var resp RelatedQuestionsResponse
	if err := c.do(httpReq, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
