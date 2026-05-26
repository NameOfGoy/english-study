.PHONY: api fmt doc error model ui build-backend build-ui push-backend push-ui deploy

# 镜像仓库
REGISTRY := crpi-r8vfvlukq94i69ai.cn-hangzhou.personal.cr.aliyuncs.com/oujy
BACKEND_IMAGE := $(REGISTRY)/english-study
UI_IMAGE := $(REGISTRY)/english-study-ui
VERSION ?= v0.0.48

# 代码生成
api:
	goctl api go --home api/template/goctl --api api/englishstudy.api --dir .
fmt:
	goctl api format --dir ./api
doc:
	goctl api swagger --api api/englishstudy.api --dir ./api/swagger --yaml
error:
	protoc -I ./internal/errors/third_party \
			-I ./internal/errors \
			--go_out=paths=source_relative:./internal/errors \
			--go-errors_out=paths=source_relative:./internal/errors \
			./internal/errors/*.proto
model:
	go test .\internal\model\bean

# 本地开发
ui:
	cd English-Study-UI && cmd /c "npm run dev"

# 构建镜像
build-backend:
	docker build -t $(BACKEND_IMAGE):$(VERSION) .
build-ui:
	cd English-Study-UI && docker build -t $(UI_IMAGE):$(VERSION) .
build: build-backend build-ui

# 推送镜像
push-backend: build-backend
	docker push $(BACKEND_IMAGE):$(VERSION)
push-ui: build-ui
	docker push $(UI_IMAGE):$(VERSION)
push: push-backend push-ui

# 一键部署：构建+推送+远程更新
deploy: push
	@echo "镜像推送完成: $(VERSION)"
	@echo "请登录服务器执行部署"
