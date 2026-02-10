package delivery

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/moroshma/proxy_checker/proxy_checker/internal/models"
)

type ProxyUseCase interface {
	CreateTaskProxy(ctx context.Context, resumeObject models.ProxyCheckApiModelRes) (models.ProxyCheckServiceResponse, error)
	GetStatusProxy(ctx context.Context, resumeObject models.ProxyResultServiceReq) ([]models.ProxyResultServiceResponse, error)
	GetHistory(ctx context.Context) ([]models.HistoryItem, error)
}

type ProxyHandler struct {
	proxyService ProxyUseCase
}

func NewProxyHandler(resumeUseCase ProxyUseCase) *ProxyHandler {
	return &ProxyHandler{
		proxyService: resumeUseCase,
	}
}

func (handler *ProxyHandler) Create(con *gin.Context) {
	var statistic models.ProxyCheckApiModelRes
	if err := con.ShouldBindJSON(&statistic); err != nil {
		con.JSON(http.StatusBadRequest, gin.H{"error": "Invalid statistic data"})
		return
	}

	id, err := handler.proxyService.CreateTaskProxy(context.Background(), statistic)
	if err != nil {
		con.JSON(http.StatusConflict, err)
		return
	}
	con.JSON(http.StatusCreated, id)
}

func (handler *ProxyHandler) GetStatus(con *gin.Context) {
	id := con.Param("id")
	if id == "" {
		con.JSON(http.StatusBadRequest, gin.H{"error": "missing id parameter"})
		return
	}

	result, err := handler.proxyService.GetStatusProxy(context.Background(), models.ProxyResultServiceReq{TaskUUID: id})
	if err != nil {
		con.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	con.JSON(http.StatusOK, result)
}

func (handler *ProxyHandler) GetHistory(con *gin.Context) {
	result, err := handler.proxyService.GetHistory(context.Background())
	if err != nil {
		con.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	con.JSON(http.StatusOK, result)
}
