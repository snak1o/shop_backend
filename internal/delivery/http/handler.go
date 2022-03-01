package delivery

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"shop_backend/internal/config"
	v1 "shop_backend/internal/delivery/http/v1"
	"shop_backend/internal/service"
)

type Handler struct {
	services *service.Services
	cfg      *config.Config
}

func NewHandler(services *service.Services, cfg *config.Config) *Handler {
	return &Handler{
		services: services,
		cfg:      cfg,
	}
}

func (h *Handler) Init(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	h.InitApi(r)

	return r
}

func (h *Handler) InitApi(r *gin.Engine) {
	handlerV1 := v1.NewHandler(h.services, h.cfg)
	api := r.Group("/api")
	{
		handlerV1.Init(api)
	}
}
