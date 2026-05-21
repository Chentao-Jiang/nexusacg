package handler

import (
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"github.com/planforever/nexusacg/internal/service"
)

type GroupHandler struct{ svc *service.GroupService }

func NewGroupHandler(r *gin.RouterGroup, svc *service.GroupService, authMW gin.HandlerFunc) {
	h := &GroupHandler{svc: svc}
	public := r.Group("/groups")
	private := r.Group("/groups")
	private.Use(authMW)

	public.GET("", h.List)
	public.GET("/:id", h.Get)
	public.GET("/:id/members", h.Members)
	public.GET("/:id/posts", h.Posts)
	private.POST("", h.Create)
	private.PUT("/:id", h.Update)
	private.POST("/:id/join", h.Join)
	private.POST("/:id/leave", h.Leave)
	private.GET("/my", h.MyGroups)
}

func (h *GroupHandler) List(c *gin.Context) {
	var input service.GroupListInput
	c.ShouldBindQuery(&input)
	result, _ := h.svc.List(input)
	Success(c, result)
}

func (h *GroupHandler) Get(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	g, err := h.svc.Get(id)
	if err != nil { NotFound(c, err.Error()); return }
	Success(c, g)
}

func (h *GroupHandler) Create(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))
	var g model.Group
	if err := c.ShouldBindJSON(&g); err != nil { BadRequest(c, "invalid data"); return }
	g.ID = uuid.New()
	g.OwnerID = uid
	g.MemberCount = 1
	g.Status = "active"
	if err := h.svc.Create(&g); err != nil { BadRequest(c, err.Error()); return }
	// Auto-join creator as owner
	h.svc.CreateMember(g.ID, uid, "owner")
	Success(c, g)
}

func (h *GroupHandler) Update(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))
	g, err := h.svc.Get(id)
	if err != nil { NotFound(c, err.Error()); return }
	if g.OwnerID != uid { BadRequest(c, "only owner can edit"); return }
	c.ShouldBindJSON(&g)
	g.ID = id
	h.svc.Update(g)
	Success(c, g)
}

func (h *GroupHandler) Join(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))
	if err := h.svc.Join(id, uid); err != nil { BadRequest(c, err.Error()); return }
	Success(c, gin.H{"joined": true})
}

func (h *GroupHandler) Leave(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))
	if err := h.svc.Leave(id, uid); err != nil { BadRequest(c, err.Error()); return }
	Success(c, gin.H{"left": true})
}

func (h *GroupHandler) Members(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	members, total, _ := h.svc.GetMembers(id, page, ps)
	Success(c, gin.H{"items": members, "total": total})
}

func (h *GroupHandler) Posts(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	posts, total, _ := h.svc.GetGroupPosts(id, page, ps)
	Success(c, gin.H{"items": posts, "total": total})
}

func (h *GroupHandler) MyGroups(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))
	groups, _ := h.svc.GetMyGroups(uid)
	if groups == nil { groups = []model.Group{} }
	Success(c, gin.H{"items": groups})
}
