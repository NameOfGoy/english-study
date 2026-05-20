package llm

import "context"

// LLM 大语言模型接口
type LLM interface {
	// Chat 单问题同步接口
	Chat(ctx context.Context, question string, opts ...Options) (answer string, err error)
	// StreamChat 单问题流式接口
	StreamChat(ctx context.Context, question string, opts ...Options) (answer <-chan string, err error)
}