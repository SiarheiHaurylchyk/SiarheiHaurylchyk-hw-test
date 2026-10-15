package internalgrpc

//go:generate protoc -I ../../../api --go_out=./pb --go_opt=paths=source_relative --go-grpc_out=./pb --go-grpc_opt=paths=source_relative Event.proto EventService.proto
