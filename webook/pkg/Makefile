.PHONY:mockgen
mockgen:mockgen.ratelimit

.PHONY:mockgen.ratelimit
mockgen.ratelimit:
	@rm -rf ./ratelimit/mocks
	@mockgen -source=./ratelimit/types.go -package=mocks -destination=./ratelimit/mocks/types.mock_gen.go

.PHONY:clear
	@rm -rf ./ratelimit/mocks