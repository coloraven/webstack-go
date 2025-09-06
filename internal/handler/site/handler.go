/**
 * @Author: chentong
 * @Date: 2024/05/26 上午12:35
 */

package site

import (
	"net/http"

	"github.com/gin-gonic/gin"

	v1 "github.com/ch3nnn/webstack-go/api/v1"
	"github.com/ch3nnn/webstack-go/internal/handler"
	"github.com/ch3nnn/webstack-go/internal/service/site"
)

type Handler struct {
	*handler.Handler
	siteService site.Service
}

func NewHandler(handler *handler.Handler, siteService site.Service) *Handler {
	return &Handler{
		Handler:     handler,
		siteService: siteService,
	}
}

// Import 导入网站
func (h *Handler) Import(ctx *gin.Context) {
	var req v1.SiteImportReq
	if err := ctx.ShouldBind(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	resp, err := h.siteService.Import(ctx, &req)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	v1.HandleSuccess(ctx, resp)
}