package bigmodel

import (
	"context"
	"fmt"
	"log"

	"english-study/internal/AI/llm"
	"english-study/internal/thirdpart/bigmodel"
	"english-study/internal/utils"
)

const (
	defaultModel = bigmodel.ModelGLM4Flash250414
)

// BigModelLLM 智谱清言LLM实现
type BigModelLLM struct {
	client *bigmodel.SDK
}

// NewBigModelLLM 创建新的BigModelLLM实例
func NewBigModelLLM(apiKey string) *BigModelLLM {
	client := bigmodel.New(apiKey)
	return &BigModelLLM{
		client: client,
	}
}

// Chat 实现LLM接口的Chat方法
func (b *BigModelLLM) Chat(ctx context.Context, question string, opts ...llm.Options) (answer string, err error) {
	if question == "" {
		return "", fmt.Errorf("question cannot be empty")
	}

	// 解析选项
	opt := utils.GetOptionFromOptions[llm.Option, llm.Options](opts)
	model := opt.Model
	if model == "" {
		model = defaultModel // 默认模型
	}

	// 调用SDK的SimpleChat方法
	answer, err = b.client.SimpleChat(ctx, model, question)
	if err != nil {
		log.Printf("BigModel Chat error: %v", err)
		return "", fmt.Errorf("chat failed: %w", err)
	}

	return answer, nil
}

// StreamChat 实现LLM接口的StreamChat方法
func (b *BigModelLLM) StreamChat(ctx context.Context, question string, opts ...llm.Options) (answer <-chan string, err error) {
	if question == "" {
		return nil, fmt.Errorf("question cannot be empty")
	}

	// 解析选项
	opt := utils.GetOptionFromOptions[llm.Option, llm.Options](opts)
	model := opt.Model
	if model == "" {
		model = defaultModel // 默认模型
	}

	// 创建channel用于返回流式数据
	answerChan := make(chan string, 100)

	// 启动goroutine处理流式响应
	go func() {
		defer close(answerChan)

		// 定义流式回调函数
		callback := func(response bigmodel.ChatCompletionStreamResponse) error {
			// 提取内容并发送到channel
			for _, choice := range response.Choices {
				if choice.Delta.Content != "" {
					select {
					case answerChan <- choice.Delta.Content:
					case <-ctx.Done():
						return ctx.Err()
					}
				}
			}
			return nil
		}

		// 调用SDK的StreamChat方法
		err := b.client.StreamChat(ctx, model, question, callback)
		if err != nil {
			log.Printf("BigModel StreamChat error: %v", err)
			// 发送错误信息到channel
			select {
			case answerChan <- fmt.Sprintf("Error: %v", err):
			case <-ctx.Done():
			}
		}
	}()

	return answerChan, nil
}
