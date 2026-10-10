// Package service 实现 product 领域的商品业务逻辑。
package service

import (
	"context"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/errs"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/model"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/repository"
)

const (
	maxNameLen        = 64
	maxDescriptionLen = 255
	defaultPageSize   = 20
	maxPageSize       = 100
)

// CreateProductInput 是创建商品的业务入参。
type CreateProductInput struct {
	Name        string
	Description string
	Price       uint32
	Status      model.Status
}

// ProductPage 是分页查询结果。
type ProductPage struct {
	Items []*model.Product
	Total int64
	Page  int
	Size  int
}

// ProductService 提供商品主数据读写业务。
type ProductService struct {
	repo repository.ProductRepository
}

// NewProductService 构造 ProductService，注入数据访问依赖。
func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// Create 校验入参后创建商品，返回新商品 ID。
func (s *ProductService) Create(ctx context.Context, in CreateProductInput) (uint64, error) {
	if err := validateCreate(in); err != nil {
		return 0, err
	}
	return s.repo.Create(ctx, &model.Product{
		Name:        strings.TrimSpace(in.Name),
		Description: in.Description,
		Price:       in.Price,
		Status:      in.Status,
	})
}

// GetByID 按 ID 查询商品；不存在返回 CodeProductNotFound(404)。
func (s *ProductService) GetByID(ctx context.Context, id uint64) (*model.Product, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errs.New(errs.CodeProductNotFound, "product not found",
			http.StatusNotFound)
	}
	return p, nil
}

// List 归一化分页参数后查询商品。
func (s *ProductService) List(ctx context.Context, page, size int) (*ProductPage, error) {
	page, size = normalizePage(page, size)
	items, err := s.repo.List(ctx, size, (page-1)*size)
	if err != nil {
		return nil, err
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, err
	}
	return &ProductPage{Items: items, Total: total, Page: page, Size: size}, nil
}

// validateCreate 校验创建商品入参，非法返回 CodeProductInvalid(400)。
func validateCreate(in CreateProductInput) error {
	name := strings.TrimSpace(in.Name)
	if name == "" || utf8.RuneCountInString(name) > maxNameLen {
		return errs.New(errs.CodeProductInvalid, "invalid product name", http.StatusBadRequest)
	}
	if utf8.RuneCountInString(in.Description) > maxDescriptionLen {
		return errs.New(errs.CodeProductInvalid, "invalid product description", http.StatusBadRequest)
	}
	if in.Price == 0 {
		return errs.New(errs.CodeProductInvalid, "invalid product price", http.StatusBadRequest)
	}
	if in.Status != model.StatusOnSale && in.Status != model.StatusOffShelf {
		return errs.New(errs.CodeProductInvalid, "invalid product status", http.StatusBadRequest)
	}
	return nil
}

// normalizePage 归一化分页参数：page<1→1，size<=0→默认，size>上限→钳制。
func normalizePage(page, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size <= 0 {
		size = defaultPageSize
	}
	if size > maxPageSize {
		size = maxPageSize
	}
	return page, size
}
