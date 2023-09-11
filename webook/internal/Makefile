.PHONY:mockgen
mockgen:mockgen.jwt mockgen.service mockgen.service.sms mockgen.repository mockgen.dao mockgen.cache mockgen.redis

.PHONY:mockgen.jwt
mockgen.jwt:
	@rm -rf ./web/jwt/mocks
	@mockgen -source=./web/jwt/types.go -package=mocks -destination=./web/jwt/mocks/types.mock_gen.go

.PHONY:mockgen.service
mockgen.service:
	@rm -rf ./service/mocks
	@mockgen -source=./service/user.go -package=mocks -destination=./service/mocks/user.mock_gen.go
	@mockgen -source=./service/code.go -package=mocks -destination=./service/mocks/code.mock_gen.go

.PHONY:mockgen.service.sms
mockgen.service.sms:
	@rm -rf ./service/sms/mocks
	@mockgen -source=./service/sms/types.go -package=mocks -destination=./service/sms/mocks/types.mock_gen.go


.PHONY:mockgen.repository
mockgen.repository:
	@rm -rf ./repository/mocks
	@mockgen -source=./repository/user.go -package=mocks -destination=./repository/mocks/user.mock_gen.go
	@mockgen -source=./repository/code.go -package=mocks -destination=./repository/mocks/code.mock_gen.go

.PHONY:mockgen.dao
mockgen.dao:
	@rm -rf ./repository/dao/mocks
	@mockgen -source=./repository/dao/user.go -package=mocks -destination=./repository/dao/mocks/user.mock_gen.go

.PHONY:mockgen.cache
mockgen.cache:
	@rm -rf ./repository/cache/mocks
	@mockgen -source=./repository/cache/types.go -package=mocks -destination=./repository/cache/mocks/types.mock_gen.go

.PHONY:mockgen.redis
mockgen.redis:
	@rm -rf ./repository/cache/redis/mocks
	@@mockgen -package=mocks -destination=./repository/cache/redis/mocks/cmd.mock_gen.go github.com/redis/go-redis/v9 Cmdable

.PHONY:clear
clear:
	@rm -rf ./service/mocks \
	./service/sms/mocks \
	./repository/mocks \
	./repository/dao/mocks \
	./repository/cache/mocks \
	./repository/cache/redis/mocks \
	./web/jwt/mocks