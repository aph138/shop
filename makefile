grpc_path = api
grpc_prefix= github.com/aph138/shop
grpc_files = protos/user.proto protos/shop.proto
rpc:
	rm -fr $(grpc_path)
	mkdir -p $(grpc_path)
	protoc --go_out=$(grpc_path) --go_opt=module=$(grpc_prefix) --go-grpc_out=$(grpc_path) --go-grpc_opt=module=$(grpc_prefix) $(grpc_files)
docker:
	npx tailwind -i ./server/web/css/main.css -o ./server/public/css/tailwind.css
	templ generate
	docker compose up --build
temp:
	templ generate --watch
css:
	npx tailwind -i ./server/web/css/main.css -o ./server/public/css/tailwind.css --watch

run: temp css

