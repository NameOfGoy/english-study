# 服务器版本备份 - 2026-04-02

> P0-P1 功能开发前的基线快照

## 当前版本

- **Makefile VERSION**: `v0.0.21`
- **后端镜像**: `crpi-r8vfvlukq94i69ai.cn-hangzhou.personal.cr.aliyuncs.com/oujy/english-study:v0.0.21`
- **前端镜像**: `crpi-r8vfvlukq94i69ai.cn-hangzhou.personal.cr.aliyuncs.com/oujy/english-study-ui:v0.0.21`

## 服务器 docker-compose (远程)

- 路径: `/root/english-study/docker-compose.yaml`
- 后端端口: `13896 -> 8888`
- 前端端口: `18080 -> 80`
- 后端配置: `/root/conf/englishstudy.yaml` 挂载到 `/app/config/config.yaml`
- 网络: postgres(external), minio(external), englishstudy(bridge)

## Git 基线

- **分支**: `develop`
- **最新 commit**: `e1c2a03` - feat: 图片搜索、导入任务系统、部署脚本
- **base branch**: `master`

## 数据库 (变更前)

word_statuses 表当前字段:
- id, word_id, word_type, status, times, weight, study_time, user_id, created_at, updated_at

SRS 开发后将新增:
- ease_factor, interval, next_review_at, repetitions

## 回滚方案

如需回滚:
1. 镜像回退到 `v0.0.21`
2. `ALTER TABLE word_statuses DROP COLUMN IF EXISTS ease_factor, interval, next_review_at, repetitions;`
3. 重启容器: `docker compose pull && docker compose up -d`
