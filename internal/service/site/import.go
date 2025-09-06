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

		// 确保行数据完整（至少有9列：ID,Logo,标题,链接,描述,分类,创建日期,更新日期,状态）
		if len(row) < 9 {
			failCount++
			continue
		}

		// 解析行数据，使用字段名而不是序号
		logo := row[1]        // Logo
		title := row[2]       // 标题
		url := row[3]         // 链接
		description := row[4] // 描述
		categoryName := row[5] // 分类
		status := row[8]      // 状态

		// 根据分类名称查找或创建分类
		category, err := s.categoryRepository.WithContext(ctx).FindOne(s.categoryRepository.WhereByTitle(categoryName))
		if err != nil {
			// 如果分类不存在，创建新分类
			category, err = s.categoryRepository.WithContext(ctx).Create(&model.StCategory{
				Title: categoryName,
				Sort:  0,
				IsUsed: status == "true" || status == "1", // 根据网站状态设置分类状态
			})
			if err != nil {
				failCount++
				continue
			}
		} else {
			// 如果分类已存在，且当前网站状态为启用，则确保分类也是启用状态
			if (status == "true" || status == "1") && !category.IsUsed {
				_, err = s.categoryRepository.WithContext(ctx).Update(map[string]interface{}{"is_used": true}, s.categoryRepository.WhereByID(category.ID))
				if err != nil {
					failCount++
					continue
				}
			}
		}

		// 创建网站记录
		_, err = s.siteRepository.WithContext(ctx).Create(&model.StSite{
			Icon:        logo,
			Title:       title,
			URL:         url,
			CategoryID:  category.ID,
			Description: description,
			IsUsed:      status == "true" || status == "1",
			Sort:        0,
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
