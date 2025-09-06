/**
 * @Author: your name
 * @Date: 2025/04/05 19:06
 */

package site

import (
	"bytes"
	"context"
	"io"

	v1 "github.com/ch3nnn/webstack-go/api/v1"
	"github.com/ch3nnn/webstack-go/internal/dal/model"
	// "github.com/ch3nnn/webstack-go/internal/dal/repository"
	"github.com/xuri/excelize/v2"
)

func (s *service) Import(ctx context.Context, req *v1.SiteImportReq) (resp *v1.SiteImportResp, err error) {
	// 打开上传的Excel文件
	file, err := req.File.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 读取文件内容
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// 使用Excelize打开Excel文件
	excelFile, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := excelFile.Close(); err != nil {
			// 记录日志但不返回错误
		}
	}()

	// 获取所有行数据
	rows, err := excelFile.GetRows("Sheet1")
	if err != nil {
		return nil, err
	}

	// 跳过标题行，从第二行开始处理
	var successCount, failCount int
	for i, row := range rows {
		// 跳过标题行
		if i == 0 {
			continue
		}

		// 确保行数据完整（至少有8列）
		if len(row) < 8 {
			failCount++
			continue
		}

		// 解析行数据
		categoryName := row[4] // 分类名称在第5列（索引4）

		// 根据分类名称查找或创建分类
		category, err := s.categoryRepository.WithContext(ctx).FindOne(s.categoryRepository.WhereByTitle(categoryName))
		if err != nil {
			// 如果分类不存在，创建新分类
			category, err = s.categoryRepository.WithContext(ctx).Create(&model.StCategory{
				Title: categoryName,
				Sort:  0,
			})
			if err != nil {
				failCount++
				continue
			}
		}

		// 创建网站记录
		_, err = s.siteRepository.WithContext(ctx).Create(&model.StSite{
			Icon:        row[1], // Logo在第2列
			Title:       row[2], // 名称简介在第3列
			URL:         row[3], // 链接在第4列
			CategoryID:  category.ID,
			Description: "", // Excel中没有描述字段
			IsUsed:      row[7] == "true" || row[7] == "1", // 状态在第8列
			Sort:        0,                                 // Excel中没有排序字段
		})

		if err != nil {
			failCount++
			continue
		}

		successCount++
	}

	return &v1.SiteImportResp{
		SuccessCount: successCount,
		FailCount:    failCount,
	}, nil
}
