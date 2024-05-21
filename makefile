grpc_path = grpc
grpc_prefix= github.com/aph138/shop
grpc_files = protos/user.proto protos/shop.proto
rpc:
	rm -fr $(grpc_path)
	mkdir -p $(grpc_path)
	protoc --go_out=$(grpc_path) --go_opt=module=$(grpc_prefix) --go-grpc_out=$(grpc_path) --go-grpc_opt=module=$(grpc_prefix) $(grpc_files)
docker:
	templ generate
	docker compose up --build

temp:
	templ generate