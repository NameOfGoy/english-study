# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go backend REST API for an English learning platform, built with the **go-zero** framework. Features dictionary management, AI-powered content generation, multiple learning/practice modes, and WeChat authentication.

## Common Commands

```bash
# Run the service
go run englishstudy.go -f etc/englishstudy-api.yaml

# Generate API handlers/types from .api definitions
make api

# Format .api files
make fmt

# Generate Swagger docs to api/swagger/englishstudy.yaml
make doc

# Generate error types from protobuf definitions
make error

# Build for production (Docker uses this pattern)
go build -gcflags '-N -l' -ldflags '-s -w' -o ./bin/englishstudy
```

Code generation uses `goctl` (go-zero's CLI). After modifying `.api` files in `api/`, run `make api` to regenerate handlers and types.

## Architecture

**Layered structure:**
```
api/*.api  (API definitions, goctl format)
  → internal/handler/    (generated HTTP handlers)
  → internal/logic/      (business logic - manually written)
  → internal/svc/        (service context / dependency injection)
  → internal/model/      (GORM models & generated DTO queries)
  → internal/dictionary/  (dictionary interface + impl/)
```

**Entry point:** `englishstudy.go` — loads config, initializes DB/MinIO/WeChat/AI services, registers routes, starts server on port 8888.

**Service context** (`internal/svc/servicecontext.go`): Central DI container holding Config, DB Model, OSS, WeChat client, Dictionary, and AI services.

**Key internal packages:**
- `internal/AI/llm/` — BigModel (智谱AI/ChatGLM) LLM integration
- `internal/AI/tts/` — Aliyun NLS text-to-speech
- `internal/AI/view/` — BigModel image generation
- `internal/aiapplication/` — AI application layer (word examples, pictures, pronunciation, translation)
- `internal/oss/` — MinIO object storage
- `internal/wx/` — WeChat login/registration

**Database:** PostgreSQL with GORM. Per-user tables are created dynamically (`word_user_{userId}`, `word_pos_user_{userId}`, `word_phrase_user_{userId}`). Schema auto-migrates on startup via `internal/model/db.go`.

**GORM Gen:** Generated query code lives in `internal/model/dto/`. Entity beans in `internal/model/bean/`.

## API Routes

All routes prefixed with `/api/v1/`. JWT auth required on all except login/register endpoints.

Main modules: `dictionary` (word/phrase CRUD, import/export, status, tags), `user` (auth, profile), `practise` (study/review/strength/spot modes), `tag`, `file-service` (upload/download via MinIO).

## Config

`etc/englishstudy-api.yaml` contains all service configuration (DB, MinIO, WeChat, Aliyun, BigModel, JWT). This file is gitignored as it contains credentials. Config structs defined in `internal/config/config.go` with Viper for dynamic reloading.
