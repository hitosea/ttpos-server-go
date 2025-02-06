新建一个迁移文件

go run migration/main.go make --name create_user

迁移到某个版本。如迁移到1738765726版本

go run migration/main.go run --version 1738765726
