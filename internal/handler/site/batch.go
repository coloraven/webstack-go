package site

import (
	"net/http"

	"github.com/gin-gonic/gin"

	v1 "github.com/ch3nnn/webstack-go/api/v1"
)

// SyncAll 批量同步网站
// @Summary 批量同步网站
// @Schemes
// @Description 批量同步网站信息
// @Tags 网站模块
// @Accept json
// @Produce json
// @Param request body v1.SiteBatchReq true "params"
// @Success 200 {object} v1.SiteBatchResp
// @Router /api/admin/site/sync-all [post]
func (h *Handler) SyncAll(ctx *gin.Context) {
	var req v1.SiteBatchReq
	if err := ctx.ShouldBind(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	resp, err := h.siteService.SyncAll(ctx, &req)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	v1.HandleSuccess(ctx, resp)
}

// ToggleAll 批量启用/禁用网站
// @Summary 批量启用/禁用网站
// @Schemes
// @Description 批量启用/禁用网站
// @Tags 网站模块
// @Accept json
// @Produce json
// @Param request body v1.SiteBatchReq true "params"
// @Success 200 {object} v1.SiteBatchResp
// @Router /api/admin/site/toggle-all [post]
func (h *Handler) ToggleAll(ctx *gin.Context) {
	var req v1.SiteBatchReq
	if err := ctx.ShouldBind(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	resp, err := h.siteService.ToggleAllSimple(ctx, &req)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	v1.HandleSuccess(ctx, resp)
}

// ClearAll 批量清空网站
// @Summary 批量清空网站
// @Schemes
// @Description 批量清空网站
// @Tags 网站模块
// @Accept json
// @Produce json
// @Param request body v1.SiteBatchReq true "params"
// @Success 200 {object} v1.SiteBatchResp
// @Router /api/admin/site/clear-all [post]
func (h *Handler) ClearAll(ctx *gin.Context) {
	var req v1.SiteBatchReq
	if err := ctx.ShouldBind(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	resp, err := h.siteService.ClearAllSimple(ctx, &req)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	v1.HandleSuccess(ctx, resp)
}