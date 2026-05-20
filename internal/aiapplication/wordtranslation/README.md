# wordtranslation 模块

用于根据给定的英语词语/单词生成常见的中文翻译，参照同级目录 `wordexample` 的结构与实现。

## 目录结构

- `wordtranslation.go`：接口定义
- `option.go`：选项与设置方法
- `template.go`：提示模板与默认配置
- `impl/impl.go`：接口实现（基于 LLM）

## 使用示例

```go
package main

import (
    "context"
    "fmt"
    "english-study/internal/AI/llm/bigmodel"
    wti "english-study/internal/aiapplication/wordtranslation/impl"
    wt "english-study/internal/aiapplication/wordtranslation"
)

func main() {
    // 构造一个LLM实现（示例：智谱清言）
    bm := bigmodel.NewBigModelLLM("<YOUR_API_KEY>")

    // 创建生成器
    gen := wti.NewGenerator(bm)

    // 可选：自定义模板（纯中文文本返回）
    // gen.WithPromptTemplate(wt.PromptTemplate{Template: "..."})

    // 生成翻译（返回单个中文翻译字符串）
    translation, err := gen.Generate(context.Background(), "run", wt.WithPos("动词"))
    if err != nil {
        panic(err)
    }
    fmt.Println(translation)
}
```

## 说明

- 默认返回 `string`，为一个常见且准确的中文翻译。
- 通过 `WithPos` 可以限定词性（中文名称），帮助模型更精准地输出。
- 模板约束为返回纯中文释义文本；实现会尝试去除可能的 markdown 代码块修饰。