protoc --go_out=pb/ --go_opt=paths=source_relative --go-grpc_out=pb/ --go-grpc_opt=paths=source_relative  pb/auth.proto


docker run -p 5432:5432 --name user_db_postgresql -e POSTGRES_PASSWORD=pass -e POSTGRES_USER=auth_user -d postgres