package impl

import (
    "context"
    "testing"
    "english-study/internal/AI/llm"
)

// fakeLLM 是用于测试的LLM实现
type fakeLLM struct{}

func (f *fakeLLM) Chat(ctx context.Context, question string, opts ...llm.Options) (string, error) {
    return "翻译一", nil
}

func (f *fakeLLM) StreamChat(ctx context.Context, question string, opts ...llm.Options) (<-chan string, error) {
    ch := make(chan string)
    close(ch)
    return ch, nil
}

func TestGenerateTranslation(t *testing.T) {
    g := NewGenerator(&fakeLLM{})
    translation, err := g.Generate(context.Background(), "run")
    if err != nil {
        t.Fatalf("Generate error: %v", err)
    }
    if translation != "翻译一" {
        t.Fatalf("unexpected translation: %s", translation)
    }
}