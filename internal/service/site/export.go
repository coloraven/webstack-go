/**
 * @Author: chentong
 * @Date: 2025/01/29 19:10
 */

package site

import (
	"strconv"

	"github.com/gin-gonic/gin"
	excelize "github.com/xuri/excelize/v2"
	"gorm.io/gen"
	"gorm.io/gen/field"

	v1 "github.com/ch3nnn/webstack-go/api/v1"
	"github.com/ch3nnn/webstack-go/internal/dal/query"
	"github.com/ch3nnn/webstack-go/internal/dal/repository"
)

var (
	sheetName = "Sheet1"
	headers   = []string{"ID", "Logo", "标题", "链接", "描述", "分类", "创建日期", "更新日期", "状态"}
)

func (s *service) Export(ctx *gin.Context, req *v1.SiteExportReq) (resp *v1.SiteExportResp, err error) {
	var orderColumns []field.Expr
	orderColumns = append(orderColumns, query.StSite.CreatedAt.Desc())

	var whereFunc []func(dao gen.Dao) gen.Dao
	if req.Search != "" {
		whereFunc = append(whereFunc, s.siteRepository.LikeInByTitleOrDescOrURL(req.Search))
	}
	if req.CategoryID != 0 {
		whereFunc = append(whereFunc, s.siteRepository.WhereByCategoryID(req.CategoryID))
		orderColumns = []field.Expr{query.StSite.Sort.Asc()}
	}

	var siteCategories []repository.SiteCategory
	_, err = s.siteRepository.WithContext(ctx).FindSiteCategoryWithPage(1, 10000, &siteCategories, orderColumns, whereFunc...)
	if err != nil {
		return nil, err
	}

	excelFile := excelize.NewFile()

	// 设置表头
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := excelFile.SetCellValue(sheetName, cell, header); err != nil {
			continue
		}
	}

	// 填充数据
	for i, siteCategory := range siteCategories {
		rowNum := i + 2
		row := strconv.Itoa(rowNum)

		// 使用字段名明确对应，避免使用序号
		excelFile.SetCellValue(sheetName, "A"+row, siteCategory.StSite.ID)           // ID
		excelFile.SetCellValue(sheetName, "B"+row, siteCategory.StSite.Icon)         // Logo
		excelFile.SetCellValue(sheetName, "C"+row, siteCategory.StSite.Title)        // 标题
		excelFile.SetCellValue(sheetName, "D"+row, siteCategory.StSite.URL)          // 链接
		excelFile.SetCellValue(sheetName, "E"+row, siteCategory.StSite.Description)  // 描述
		excelFile.SetCellValue(sheetName, "F"+row, siteCategory.StCategory.Title)    // 分类
		excelFile.SetCellValue(sheetName, "G"+row, siteCategory.StSite.CreatedAt)    // 创建日期
		excelFile.SetCellValue(sheetName, "H"+row, siteCategory.StSite.UpdatedAt)    // 更新日期
		excelFile.SetCellValue(sheetName, "I"+row, siteCategory.StSite.IsUsed)       // 状态
	}

	index, err := excelFile.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	excelFile.SetActiveSheet(index)

	return &v1.SiteExportResp{File: excelFile}, nil
}