package site

import (
	"context"
	"github.com/gin-gonic/gin"
	"gorm.io/gen"
	"gorm.io/gorm"

	v1 "github.com/ch3nnn/webstack-go/api/v1"
	"github.com/ch3nnn/webstack-go/internal/dal/model"
	"github.com/ch3nnn/webstack-go/internal/dal/query"
)

// ToggleAllSimple 全部启用/禁用网站（简化版）
func (s *service) ToggleAllSimple(ctx *gin.Context, req *v1.SiteBatchReq) (resp *v1.SiteBatchResp, err error) {
	// 构建查询条件
	var whereFunc []func(dao gen.Dao) gen.Dao
	if req.Search != "" {
		whereFunc = append(whereFunc, s.siteRepository.LikeInByTitleOrDescOrURL(req.Search))
	}
	if req.CategoryID != 0 {
		whereFunc = append(whereFunc, s.siteRepository.WhereByCategoryID(req.CategoryID))
	}
	if req.Status != nil {
		whereFunc = append(whereFunc, s.siteRepository.WhereByIsUsed(*req.Status == 1))
	}

	// 批量更新网站状态
	updateData := map[string]interface{}{
		query.StSite.IsUsed.ColumnName().String(): req.IsUsed,
	}

	var rowsAffected int64
	if len(whereFunc) == 0 {
		// 如果没有提供任何条件，则更新所有记录
		q := query.StSite
		resultInfo, err := q.WithContext(context.Background()).Session(&gorm.Session{AllowGlobalUpdate: true}).Updates(updateData)
		if err != nil {
			return nil, err
		}
		rowsAffected = resultInfo.RowsAffected
	} else {
		// 根据条件更新网站状态
		rowsAffected, err = s.siteRepository.WithContext(ctx).Update(updateData, whereFunc...)
		if err != nil {
			return nil, err
		}
	}

	// 同步更新对应分类的状态
	// 先查询所有受影响的网站，获取它们的分类ID
	var sites []*model.StSite
	if len(whereFunc) == 0 {
		// 查询所有网站
		sites, err = s.siteRepository.WithContext(ctx).FindAll()
	} else {
		// 根据条件查询网站
		sites, err = s.siteRepository.WithContext(ctx).FindAll(whereFunc...)
	}
	
	if err != nil {
		return nil, err
	}

	// 收集所有涉及的分类ID
	categoryIDs := make(map[int]bool)
	for _, site := range sites {
		categoryIDs[site.CategoryID] = true
	}

	// 提取分类ID到切片中
	categoryIDList := make([]int, 0, len(categoryIDs))
	for id := range categoryIDs {
		categoryIDList = append(categoryIDList, id)
	}

	// 更新所有涉及的分类状态
	if len(categoryIDList) > 0 {
		categoryUpdateData := map[string]interface{}{
			query.StCategory.IsUsed.ColumnName().String(): req.IsUsed,
		}
		
		categoryQuery := query.StCategory
		_, err = categoryQuery.WithContext(context.Background()).
			Where(categoryQuery.ID.In(categoryIDList...)).
			Updates(categoryUpdateData)
		if err != nil {
			return nil, err
		}
	}

	return &v1.SiteBatchResp{
		SuccessCount: int(rowsAffected),
		FailCount:    0,
	}, nil
}

// ClearAllSimple 全部清空网站（简化版）
func (s *service) ClearAllSimple(ctx *gin.Context, req *v1.SiteBatchReq) (resp *v1.SiteBatchResp, err error) {
	// 构建查询条件
	var whereFunc []func(dao gen.Dao) gen.Dao
	if req.Search != "" {
		whereFunc = append(whereFunc, s.siteRepository.LikeInByTitleOrDescOrURL(req.Search))
	}
	if req.CategoryID != 0 {
		whereFunc = append(whereFunc, s.siteRepository.WhereByCategoryID(req.CategoryID))
	}
	if req.Status != nil {
		whereFunc = append(whereFunc, s.siteRepository.WhereByIsUsed(*req.Status == 1))
	}

	// 如果没有提供任何条件，则清空所有记录
	if len(whereFunc) == 0 {
		// 使用 Session(&gorm.Session{AllowGlobalUpdate: true}) 允许全局更新/删除
		q := query.StSite
		_, err = q.WithContext(context.Background()).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete()
	} else {
		// 批量删除网站
		err = s.siteRepository.WithContext(ctx).Delete(whereFunc...)
	}
	
	if err != nil {
		return nil, err
	}

	// 由于 Delete 方法不返回影响行数，我们无法准确获取删除的记录数
	// 这里简单地查询一下符合条件的记录数作为删除成功的数量
	count, err := s.siteRepository.WithContext(ctx).FindCount(whereFunc...)
	if err != nil {
		return nil, err
	}

	return &v1.SiteBatchResp{
		SuccessCount: int(count),
		FailCount:    0,
	}, nil
}