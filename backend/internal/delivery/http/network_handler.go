package http

import (
	"net/http"
	"strconv"

	"billing-backend/internal/domain"

	"github.com/gin-gonic/gin"
)

type MikrotikHandler struct {
	mikrotikUsecase domain.MikrotikUsecase
}

func NewMikrotikHandler(r *gin.RouterGroup, mu domain.MikrotikUsecase, authMiddleware gin.HandlerFunc) {
	handler := &MikrotikHandler{
		mikrotikUsecase: mu,
	}

	mikrotikGroup := r.Group("/mikrotik")
	mikrotikGroup.Use(authMiddleware)
	{
		mikrotikGroup.GET("", handler.FetchAll)
		mikrotikGroup.GET("/:id", handler.GetByID)
		mikrotikGroup.POST("", handler.Store)
		mikrotikGroup.PUT("/:id", handler.Update)
		mikrotikGroup.PATCH("/:id", handler.Update)
		mikrotikGroup.DELETE("/:id", handler.Delete)
		mikrotikGroup.POST("/:id/test_connection", handler.TestConnection)
		mikrotikGroup.POST("/:id/test-connection", handler.TestConnection)
	}

	mikrotikServersGroup := r.Group("/mikrotik_servers")
	mikrotikServersGroup.Use(authMiddleware)
	{
		mikrotikServersGroup.GET("", handler.FetchAll)
		mikrotikServersGroup.GET("/:id", handler.GetByID)
		mikrotikServersGroup.POST("", handler.Store)
		mikrotikServersGroup.PUT("/:id", handler.Update)
		mikrotikServersGroup.PATCH("/:id", handler.Update)
		mikrotikServersGroup.DELETE("/:id", handler.Delete)
		mikrotikServersGroup.POST("/:id/test_connection", handler.TestConnection)
		mikrotikServersGroup.POST("/:id/test-connection", handler.TestConnection)
	}
}

func (h *MikrotikHandler) FetchAll(c *gin.Context) {
	servers, err := h.mikrotikUsecase.FetchAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Remove encrypted passwords from response for security
	for i := range servers {
		servers[i].Password = ""
	}

	c.JSON(http.StatusOK, gin.H{"data": servers})
}

func (h *MikrotikHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	server, err := h.mikrotikUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Hide password
	server.Password = ""
	c.JSON(http.StatusOK, gin.H{"data": server})
}

func (h *MikrotikHandler) Store(c *gin.Context) {
	var server domain.MikrotikServer
	if err := c.ShouldBindJSON(&server); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.mikrotikUsecase.Store(c.Request.Context(), &server); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	server.Password = "" // Hide password in response
	c.JSON(http.StatusCreated, gin.H{"data": server})
}

func (h *MikrotikHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var server domain.MikrotikServer
	if err := c.ShouldBindJSON(&server); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.mikrotikUsecase.Update(c.Request.Context(), id, &server); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mikrotik server updated successfully"})
}

func (h *MikrotikHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := h.mikrotikUsecase.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mikrotik server deleted successfully"})
}

func (h *MikrotikHandler) TestConnection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	testResult, updatedServer, err := h.mikrotikUsecase.TestConnection(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"test_result":    testResult,
			"updated_server": updatedServer,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"test_result":    testResult,
		"updated_server": updatedServer,
	})
}

type OLTHandler struct {
	oltUsecase domain.OLTUsecase
}

func NewOLTHandler(r *gin.RouterGroup, ou domain.OLTUsecase, authMiddleware gin.HandlerFunc) {
	handler := &OLTHandler{
		oltUsecase: ou,
	}

	oltGroup := r.Group("/olt")
	oltGroup.Use(authMiddleware)
	{
		oltGroup.GET("", handler.FetchAll)
		oltGroup.GET("/:id", handler.GetByID)
		oltGroup.POST("", handler.Store)
		oltGroup.PATCH("/:id", handler.Update)
		oltGroup.PUT("/:id", handler.Update)
		oltGroup.DELETE("/:id", handler.Delete)
		oltGroup.POST("/:id/test-connection", handler.TestConnection)
		oltGroup.GET("/:id/uplinks", handler.GetUplinks)
		oltGroup.GET("/:id/board/:board/pon/:pon/onus", handler.GetONUs)
		oltGroup.GET("/:id/board/:board/pon/:pon/paginate", handler.GetPaginatedONUs)
		oltGroup.GET("/:id/board/:board/pon/:pon/onu/:onu_id", handler.GetONUDetail)
		oltGroup.GET("/:id/board/:board/pon/:pon/empty-onu-ids", handler.GetEmptyONUIDs)
		oltGroup.GET("/:id/board/:board/pon/:pon/onu-serials", handler.GetONUSerials)
		oltGroup.DELETE("/:id/board/:board/pon/:pon/cache", handler.ClearCache)
	}
}

func (h *OLTHandler) FetchAll(c *gin.Context) {
	olts, err := h.oltUsecase.FetchAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": olts})
}

func (h *OLTHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	olt, err := h.oltUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": olt})
}

func (h *OLTHandler) Store(c *gin.Context) {
	var olt domain.OLT
	if err := c.ShouldBindJSON(&olt); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.oltUsecase.Store(c.Request.Context(), &olt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": olt})
}

func (h *OLTHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var olt domain.OLT
	if err := c.ShouldBindJSON(&olt); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.oltUsecase.Update(c.Request.Context(), id, &olt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "OLT updated successfully"})
}

func (h *OLTHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := h.oltUsecase.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "OLT deleted successfully"})
}

func (h *OLTHandler) TestConnection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	msg, err := h.oltUsecase.TestConnection(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error(), "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": msg, "data": gin.H{"status": "connected", "details": msg}})
}

func (h *OLTHandler) GetUplinks(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	uplinks, err := h.oltUsecase.GetUplinks(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "status": "success", "data": uplinks})
}

func (h *OLTHandler) GetONUs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	board, err := strconv.Atoi(c.Param("board"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board parameter"})
		return
	}

	pon, err := strconv.Atoi(c.Param("pon"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pon parameter"})
		return
	}

	onus, err := h.oltUsecase.GetONUs(c.Request.Context(), id, board, pon)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "status": "success", "data": onus})
}

func (h *OLTHandler) GetPaginatedONUs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	board, err := strconv.Atoi(c.Param("board"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board parameter"})
		return
	}

	pon, err := strconv.Atoi(c.Param("pon"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pon parameter"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	result, err := h.oltUsecase.GetPaginatedONUs(c.Request.Context(), id, board, pon, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "status": "success", "data": result.Data, "meta": result.Meta})
}

func (h *OLTHandler) GetONUDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	board, err := strconv.Atoi(c.Param("board"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board parameter"})
		return
	}

	pon, err := strconv.Atoi(c.Param("pon"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pon parameter"})
		return
	}

	onuID, err := strconv.Atoi(c.Param("onu_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid onu_id parameter"})
		return
	}

	detail, err := h.oltUsecase.GetONUDetail(c.Request.Context(), id, board, pon, onuID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "status": "success", "data": detail})
}

func (h *OLTHandler) GetEmptyONUIDs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	board, err := strconv.Atoi(c.Param("board"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board parameter"})
		return
	}

	pon, err := strconv.Atoi(c.Param("pon"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pon parameter"})
		return
	}

	emptyIDs, err := h.oltUsecase.GetEmptyONUIDs(c.Request.Context(), id, board, pon)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "status": "success", "data": emptyIDs})
}

func (h *OLTHandler) GetONUSerials(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	board, err := strconv.Atoi(c.Param("board"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board parameter"})
		return
	}

	pon, err := strconv.Atoi(c.Param("pon"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pon parameter"})
		return
	}

	noCache := c.DefaultQuery("nocache", "false") == "true"

	serials, err := h.oltUsecase.GetONUSerials(c.Request.Context(), id, board, pon, noCache)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "status": "success", "data": serials})
}

func (h *OLTHandler) ClearCache(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	board, err := strconv.Atoi(c.Param("board"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board parameter"})
		return
	}

	pon, err := strconv.Atoi(c.Param("pon"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pon parameter"})
		return
	}

	if err := h.oltUsecase.ClearCache(c.Request.Context(), id, board, pon); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "status": "success", "message": "Cache cleared successfully"})
}

type ODPHandler struct {
	odpUsecase domain.ODPUsecase
}

func NewODPHandler(r *gin.RouterGroup, ou domain.ODPUsecase, authMiddleware gin.HandlerFunc) {
	handler := &ODPHandler{
		odpUsecase: ou,
	}

	odpGroup := r.Group("/odp")
	odpGroup.Use(authMiddleware)
	{
		odpGroup.GET("", handler.FetchAll)
		odpGroup.GET("/:id", handler.GetByID)
		odpGroup.POST("", handler.Store)
		odpGroup.PATCH("/:id", handler.Update)
		odpGroup.PUT("/:id", handler.Update)
		odpGroup.DELETE("/:id", handler.Delete)
	}
}

func (h *ODPHandler) FetchAll(c *gin.Context) {
	odps, err := h.odpUsecase.FetchAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": odps})
}

func (h *ODPHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	odp, err := h.odpUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": odp})
}

func (h *ODPHandler) Store(c *gin.Context) {
	var odp domain.ODP
	if err := c.ShouldBindJSON(&odp); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.odpUsecase.Store(c.Request.Context(), &odp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": odp})
}

func (h *ODPHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var odp domain.ODP
	if err := c.ShouldBindJSON(&odp); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.odpUsecase.Update(c.Request.Context(), id, &odp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ODP updated successfully"})
}

func (h *ODPHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := h.odpUsecase.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ODP deleted successfully"})
}

