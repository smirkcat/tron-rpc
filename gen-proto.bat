protoc -I=./protocol -I./third_party/googleapis --go_out=plugins=grpc:. ./protocol/api/*.proto
protoc -I=./protocol -I./third_party/googleapis --go_out=plugins=grpc:. ./protocol/core/*.proto