package handler

import (
	"net/http"

	"github.com/foggyeclipse/dns-manager-go/internal/dnsmanager"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	manager *dnsmanager.Manager
}

func NewHandler(manager *dnsmanager.Manager) *Handler {
	return &Handler{manager: manager}
}

type DNSRequest struct {
	Nameserver string `json:"nameserver" binding:"required,ip"`
}

func (h *Handler) GetDNS(c *gin.Context) {
	cfg, err := h.manager.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func (h *Handler) AddDNS(c *gin.Context) {
	var req DNSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.manager.Add(req.Nameserver); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Nameserver added successfully"})
}

func (h *Handler) RemoveDNS(c *gin.Context) {
	var req DNSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.manager.Remove(req.Nameserver); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Nameserver removed successfully"})
}
