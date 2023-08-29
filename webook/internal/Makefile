.PHONY:mockgen
mockgen:clear mockgen.service mockgen.repository mockgen.dao mockgen.cache mockgen.redis

.PHONY:mockgen.service
mockgen.service:
	@rm -rf ./service/mocks
	@mockgen -source=./service/user.go -package=mocks -destination=./service/mocks/user.mock.go
	@mockgen -source=./service/code.go -package=mocks -destination=./service/mocks/code.mock.go

.PHONY:mockgen.repository
mockgen.repository:
	@rm -rf ./repository/mocks
	@mockgen -source=./repository/user.go -package=mocks -destination=./repository/mocks/user.mock.go
	@mockgen -source=./repository/code.go -package=mocks -destination=./repository/mocks/code.mock.go

.PHONY:mockgen.dao
mockgen.dao:
	@rm -rf ./repository/dao/mocks
	@mockgen -source=./repository/dao/user.go -package=mocks -destination=./repository/dao/mocks/user.mock.go

.PHONY:mockgen.cache
mockgen.cache:
	@rm -rf ./repository/cache/mocks
	@mockgen -source=./repository/cache/types.go -package=mocks -destination=./repository/cache/mocks/types.mock.go

.PHONY:mockgen.redis
mockgen.redis:
	@rm -rf ./repository/cache/redis/mock
	@@mockgen -package=mocks -destination=./repository/cache/redis/mocks/cmd.mock.go github.com/redis/go-redis/v9 Cmdable

.PHONY:clear
clear:
	@rm -rf ./service/mocks ./repository/mocks ./repository/dao/mocks ./repository/cache/mocks ./repository/cache/redis/mock