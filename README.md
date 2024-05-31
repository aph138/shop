# shop
Simple implementation of an online shop, written in Go, using gRPC.

## Directories map
* api: gRPC files
* pkg: 
    * auth: JWT helper
    * db: MongoDB helper
    * logger: logger helper for gRPCs services
    * myredis: Redis helper
* protos: proto files
* server: main server
    * handler: web server handlers
    * public: static files that serve as public
    * web: templ files and tailwind input css file
* shared: shared functions and structres 
* user_service: gRPC user service
---
## Docker compose services
* mongodb
* redis
* user_service
* server: main server

## Frontend stack
- templ
- Htmx
- tailwind