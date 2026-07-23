package wecom

import (
	"context"
	"encoding/json"

	"github.com/chaitin/panda-wiki/log"
)

// AIBotClient 微信智能机器人
// https://developer.work.weixin.qq.com/document/path/100719
type AIBotClient struct {
	ctx            context.Context
	logger         *log.Logger
	Token          string
	EncodingAESKey string
}

type UserReq struct {
	Msgid    string `json:"msgid"`
	Aibotid  string `json:"aibotid"`
	Chattype string `json:"chattype"`
	From     struct {
		Userid string `json:"userid"`
	} `json:"from"`
	Msgtype string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
	Stream struct {
		Id string `json:"id"`
	} `json:"stream"`
}
type UserResp struct {
	Msgtype string `json:"msgtype"`
	Stream  Stream `json:"stream"`
}

type Stream struct {
	Id      string `json:"id"`
	Finish  bool   `json:"finish"`
	Content string `json:"content"`
	MsgItem []struct {
		Msgtype string `json:"msgtype"`
		Image   struct {
			Base64 string `json:"base64"`
			Md5    string `json:"md5"`
		} `json:"image"`
	} `json:"msg_item"`
}

func NewAIBotClient(
	ctx context.Context,
	logger *log.Logger,
	Token string,
	EncodingAESKey string,
) (*AIBotClient, error) {
	return &AIBotClient{
		ctx:            ctx,
		logger:         logger,
		Token:          Token,
		EncodingAESKey: EncodingAESKey,
	}, nil
}

func (c *AIBotClient) VerifyUrlWecomService(signature, timestamp, nonce, echostr string) (string, error) {
	wx, _, err := NewWXBizJsonMsgCrypt(
		c.Token,
		c.EncodingAESKey,
		"",
	)
	if err != nil {
		return "", err
	}

	code, sReplyEchoStr := wx.VerifyURL(signature, timestamp, nonce, echostr)
	if code != 0 {
		c.logger.Error("VerifyUrlWecomService failed:", log.Any("code", code))
		return "", c.getErrorMessage(code)
	}

	return sReplyEchoStr, nil
}

func (c *AIBotClient) DecryptUserReq(signature, timestamp, nonce, msg string) (*UserReq, error) {

	wx, _, err := NewWXBizJsonMsgCrypt(
		c.Token,
		c.EncodingAESKey,
		"",
	)
	if err != nil {
		return nil, err
	}

	code, reqMsg := wx.DecryptMsg(msg, signature, timestamp, nonce)
	if code != 0 {
		return nil, c.getErrorMessage(code)
	}
	var data UserReq

	c.logger.Info("decrypt user req:", log.Any("reqMsg", reqMsg))
	err = json.Unmarshal([]byte(reqMsg), &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (c *AIBotClient) MakeStreamResp(nonce, id, content string, isFinish bool) (string, error) {
	c.logger.Debug("MakeStreamResp:", log.String("content", content), log.Any("isFinish", isFinish))
	wx, _, err := NewWXBizJsonMsgCrypt(
		c.Token,
		c.EncodingAESKey,
		"",
	)
	if err != nil {
		return "", err
	}

	resp := UserResp{
		Msgtype: "stream",
		Stream: Stream{
			Id:      id,
			Finish:  isFinish,
			Content: content,
			MsgItem: nil,
		},
	}

	b, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}

	code, msg := wx.EncryptMsg(string(b), nonce)
	if code != 0 {
		c.logger.Error("MakeStreamResp failed:", log.Any("code", code))
		return "", c.getErrorMessage(code)
	}

	return msg, nil
}
