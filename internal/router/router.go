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

	// item 相关
	itemSvc := service.NewItem(repository.NewItem(db), rdb)
	itemH := handler.NewItem(itemSvc)

	// locker 相关
	lockerCellRepo := repository.NewLockerCell(db)
	lockerParcelRepo := repository.NewLockerParcel(db)
	lockerSvc := service.NewLocker(lockerCellRepo, lockerParcelRepo)
	lockerH := handler.NewLocker(lockerSvc)

	api := r.Group("/api")
	{
		// demo 接口
		demo := api.Group("/demo")
		{
			demo.GET("/ping", itemH.Ping)
			demo.POST("/item", itemH.CreateItem)
			demo.GET("/list", itemH.List)
			demo.GET("/cache", itemH.Cache)
		}

		// 快递柜格口接口
		locker := api.Group("/locker/cell")
		{
			locker.POST("", lockerH.CreateCell)
			locker.POST("/disable", lockerH.DisableCell)
		}

		// 快递柜包裹接口
		parcel := api.Group("/locker/parcel")
		{
			parcel.POST("", lockerH.StoreParcel)
			parcel.POST("/pickup", lockerH.PickupParcel)
			parcel.GET("/list", lockerH.ListParcelsByPhone)
		}
	}

	return r
}