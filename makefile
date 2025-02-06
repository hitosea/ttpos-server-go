run:
	go run cmd/server/main.go cmd/server/swagger_enabled.go

# 添加新的数据库迁移文件
# 如：创建一个名为add_age的迁移文件，执行：
# make make_file name=add_age
name ?= default_value
make_file:
	go run migration/main.go make --name=$(name)

# 迁移数据到某个版本
# 如：迁移到版本10，执行：
# make run_migrate version=10
version ?= 1738797767
run_migrate:
	go run migration/main.go run --version=$(version)
