//go:build swagger

package main

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"ttpos-server-go/docs"
)

func setupSwagger(r *gin.Engine) {
	// 允许自定义Swagger文档链接
	docs.SwaggerInfo.BasePath = "/api/v1"
	// Swagger API 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
