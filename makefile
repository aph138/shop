grpc_path = api
grpc_prefix= github.com/aph138/shop
grpc_files = user.proto shop.proto stock.proto 
rpc:
	rm -fr $(grpc_path)
	mkdir -p $(grpc_path)
	protoc -I=./protos --go_out=$(grpc_path) --go_opt=module=$(grpc_prefix) --go-grpc_out=$(grpc_path) --go-grpc_opt=module=$(grpc_prefix) $(grpc_files)
	protoc -I=./protos --go_out=$(grpc_path) \
	--go_opt=module=github.com/aph138/shop/api \
	--go-grpc_out=$(grpc_path) \
	--go-grpc_opt=module=github.com/aph138/shop/api common.proto
docker:
	npx tailwind -i ./server/web/css/main.css -o ./server/public/css/tailwind.css
	templ generate
	docker compose up --build
templ:
	templ generate --watch
css:
	npx tailwind -i ./server/web/css/main.css -o ./server/public/css/tailwind.css --watch

run: temp css

