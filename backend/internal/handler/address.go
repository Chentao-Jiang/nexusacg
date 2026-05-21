package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"github.com/planforever/nexusacg/internal/service"
)

type AddressHandler struct{ svc *service.AddressService }

func NewAddressHandler(r *gin.RouterGroup, svc *service.AddressService, authMW gin.HandlerFunc) {
	h := &AddressHandler{svc: svc}
	private := r.Group("/addresses")
	private.Use(authMW)
	private.GET("", h.List)
	private.POST("", h.Create)
	private.PUT("/:id", h.Update)
	private.DELETE("/:id", h.Delete)
}

func (h *AddressHandler) List(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))
	addrs, err := h.svc.List(uid)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	if addrs == nil { addrs = []model.Address{} }
	Success(c, gin.H{"items": addrs})
}

func (h *AddressHandler) Create(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))
	var addr model.Address
	if err := c.ShouldBindJSON(&addr); err != nil {
		BadRequest(c, "invalid data")
		return
	}
	addr.ID = uuid.New()
	addr.UserID = uid
	if err := h.svc.Create(&addr); err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, addr)
}

func (h *AddressHandler) Update(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))
	var addr model.Address
	if err := c.ShouldBindJSON(&addr); err != nil {
		BadRequest(c, "invalid data")
		return
	}
	addr.ID = id
	addr.UserID = uid
	if err := h.svc.Update(&addr); err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, addr)
}

func (h *AddressHandler) Delete(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))
	if err := h.svc.Delete(id, uid); err != nil {
		BadRequest(c, err.Error())
		return
	}
	Success(c, gin.H{"deleted": true})
}
