package site

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"gorm.io/gen"

	v1 "github.com/ch3nnn/webstack-go/api/v1"
	"github.com/ch3nnn/webstack-go/internal/dal/model"
	"github.com/ch3nnn/webstack-go/internal/dal/query"
)

// SyncAll 批量同步网站信息
func (s *service) SyncAll(ctx *gin.Context, req *v1.SiteBatchReq) (resp *v1.SiteBatchResp, err error) {
	// 构建查询条件
	var whereFunc []func(dao gen.Dao) gen.Dao
	if req.Search != "" {
		whereFunc = append(whereFunc, s.siteRepository.LikeInByTitleOrDescOrURL(req.Search))
	}
	if req.CategoryID != 0 {
		whereFunc = append(whereFunc, s.siteRepository.WhereByCategoryID(req.CategoryID))
	}

	// 查询符合条件的所有网站
	sites, err := s.siteRepository.WithContext(ctx).FindAll(whereFunc...)
	if err != nil {
		return nil, err
	}

	// 并发同步所有网站
	workerPool := make(chan struct{}, 10) // 限制并发数为10
	var eg errgroup.Group
	var successCount, failCount int

	for _, site := range sites {
		workerPool <- struct{}{} // 获取令牌
		site := site             // 避免循环变量问题
		eg.Go(func() error {
			defer func() { <-workerPool }() // 释放令牌

			// 调用单个网站同步函数
			_, syncErr := s.syncSite(ctx, site.ID)
			if syncErr != nil {
				failCount++
			} else {
				successCount++
			}
			return nil // 不返回错误，避免中断其他网站同步
		})
	}

	// 等待所有同步任务完成
	_ = eg.Wait()

	return &v1.SiteBatchResp{
		SuccessCount: successCount,
		FailCount:    failCount,
	}, nil
}

// buildQueryConditions 构建查询条件
func (s *service) buildQueryConditions(req *v1.SiteBatchReq) []func(dao gen.Dao) gen.Dao {
	var whereFunc []func(dao gen.Dao) gen.Dao
	if req.Search != "" {
		whereFunc = append(whereFunc, s.siteRepository.LikeInByTitleOrDescOrURL(req.Search))
	}
	if req.CategoryID != 0 {
		whereFunc = append(whereFunc, s.siteRepository.WhereByCategoryID(req.CategoryID))
	}
	return whereFunc
}

// ToggleAll 批量启用/禁用网站
func (s *service) ToggleAll(ctx *gin.Context, req *v1.SiteBatchReq) (resp *v1.SiteBatchResp, err error) {
	whereFunc := s.buildQueryConditions(req)

	// 批量更新网站状态
	updateData := map[string]interface{}{
		query.StSite.IsUsed.ColumnName().String(): req.IsUsed,
	}

	rowsAffected, err := s.siteRepository.WithContext(ctx).Update(updateData, whereFunc...)
	if err != nil {
		return nil, err
	}

	return newSuccessBatchResp(int(rowsAffected)), nil
}

// ClearAll 批量清空网站
func (s *service) ClearAll(ctx *gin.Context, req *v1.SiteBatchReq) (resp *v1.SiteBatchResp, err error) {
	// 直接清空 st_site 表中的所有记录
	err = s.siteRepository.WithContext(ctx).Delete()
	if err != nil {
		return nil, err
	}
	
	// 由于 Delete 方法不返回影响行数，我们返回一个默认的成功响应
	return &v1.SiteBatchResp{
		SuccessCount: -1, // 使用-1表示清空了所有记录
		FailCount:    0,
	}, nil
}

// newSuccessBatchResp 创建成功响应
func newSuccessBatchResp(count int) *v1.SiteBatchResp {
	return &v1.SiteBatchResp{
		SuccessCount: count,
		FailCount:    0,
	}
}

// syncSite 同步单个网站信息（内部函数）
func (s *service) syncSite(ctx *gin.Context, siteID int) (*model.StSite, error) {
	var (
		g                 errgroup.Group
		title, icon, desc string
	)

	site, err := s.siteRepository.WithContext(ctx).FindOne(s.siteRepository.WhereByID(siteID))
	if err != nil {
		return nil, err
	}

	url := site.URL

	g.Go(func() (err error) {
		title, err = getWebTitle(url)
		return
	})
	g.Go(func() (err error) {
		icon, err = getWebLogoIconBase64(url)
		return
	})
	g.Go(func() (err error) {
		desc, err = getWebDescription(url)
		return
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	_, err = s.siteRepository.WithContext(ctx).Update(&model.StSite{
		Title:       title,
		Icon:        icon,
		Description: desc,
	},
		s.siteRepository.WhereByID(siteID),
	)
	if err != nil {
		return nil, err
	}

	return site, nil
}
