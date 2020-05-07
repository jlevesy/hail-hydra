VERSION = "0.0.1"
DIST_DIR ?= dist

.PHONY: all
all: build

#
# Build targets
#

.PHONY: download
download:
	go mod download

.PHONY: build
build: clean dist binary

.PHONY: binary
binary:
	CGO_ENABLED=0 go build -ldflags='-s -w' -o $(DIST_DIR)/hail-hydra ./

.PHONY: image
image: binary
	docker build -t jlevesy/hail-hydra:v$(VERSION) .

.PHONY: dist
dist:
	mkdir -p $(DIST_DIR)

.PHONY: clean
clean:
	rm -rf $(DIST_DIR)

#
# Hydra helpful commands
#

HYDRA=docker-compose exec hydra hydra --endpoint http://127.0.0.1:4445/

.PHONY: token
token:
	-${HYDRA} clients create \
    --endpoint http://127.0.0.1:4445 \
    --id auth-code-client \
    --secret secret \
    --grant-types authorization_code,refresh_token \
    --response-types code,id_token \
    --scope openid,offline \
    --callbacks http://127.0.0.1:5555/callback
	${HYDRA} token user \
    --client-id auth-code-client \
    --client-secret secret \
    --endpoint http://127.0.0.1:4444/ \
    --port 5555 \
    --scope openid,offline
