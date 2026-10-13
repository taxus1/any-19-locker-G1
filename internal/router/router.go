package router

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"somepro/internal/handler"
	"somepro/internal/middleware"
	"somepro/internal/repository"
	"somepro/internal/service"
)

// New 组装路由。验收测试契约无关：运行时用 gin 的 Routes() 扫路由，不写死类名/方法名。
func New(db *gorm.DB, rdb *redis.Client) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), middleware.OperatorAuth())

	// Item 服务（原有 demo）
	svc := service.NewItem(repository.NewItem(db), rdb)
	itemHandler := handler.NewItem(svc)

	// Locker 服务（快递柜）
	cellRepo := repository.NewLockerCellRepository(db)
	parcelRepo := repository.NewParcelRepository(db)
	lockerSvc := service.NewLocker(cellRepo, parcelRepo)
	lockerHandler := handler.NewLocker(lockerSvc)

	api := r.Group("/api/demo")
	{
		api.GET("/ping", itemHandler.Ping)
		api.POST("/item", itemHandler.CreateItem)
		api.GET("/list", itemHandler.List)
		api.GET("/cache", itemHandler.Cache)
	}

	// 快递柜 API
	locker := r.Group("/api/locker")
	{
		// 格口管理
		locker.POST("/cell", lockerHandler.RegisterCell)
		locker.POST("/cell/disable", lockerHandler.DisableCell)

		// 包裹管理
		locker.POST("/parcel", lockerHandler.StoreParcel)
		locker.POST("/parcel/pickup", lockerHandler.PickupParcel)

		// 查询
		locker.GET("/parcel/list", lockerHandler.ListParcelsByPhone)
	}
	return r
}
