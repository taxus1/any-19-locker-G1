package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"somepro/internal/model"
	"somepro/internal/repository"
)

// Size 比较辅助函数
// 返回: -1 if s1 < s2, 0 if s1 == s2, 1 if s1 > s2
// 大小关系: SMALL < MEDIUM < LARGE
func compareSize(s1, s2 string) int {
	order := map[string]int{
		"SMALL":   1,
		"MEDIUM":  2,
		"LARGE":   3,
	}
	v1 := order[strings.ToUpper(s1)]
	v2 := order[strings.ToUpper(s2)]
	if v1 < v2 {
		return -1
	} else if v1 > v2 {
		return 1
	}
	return 0
}

// LockerService 快递柜服务
type LockerService struct {
	cellRepo    *repository.LockerCellRepository
	parcelRepo  *repository.ParcelRepository
	cellMux     sync.Mutex
	parcelMux   sync.Mutex
}

func NewLocker(cellRepo *repository.LockerCellRepository, parcelRepo *repository.ParcelRepository) *LockerService {
	return &LockerService{
		cellRepo:   cellRepo,
		parcelRepo: parcelRepo,
	}
}

// Size 比较辅助函数（用于公开的 service 方法）
func (s *LockerService) canFit(parcelSize, cellSize string) bool {
	return compareSize(parcelSize, cellSize) <= 0
}

// LockerCell 管理格口

// RegisterCell 登记格口
func (s *LockerService) RegisterCell(ctx context.Context, cellNo, size string) (*model.LockerCell, error) {
	// 校验尺寸
	validSizes := map[string]bool{"SMALL": true, "MEDIUM": true, "LARGE": true}
	if !validSizes[strings.ToUpper(size)] {
		return nil, errors.New("尺寸档必须为 SMALL、MEDIUM 或 LARGE")
	}

	// 加锁防止并发注册同一格口
	s.cellMux.Lock()
	defer s.cellMux.Unlock()

	// 检查是否已存在
	existing, err := s.cellRepo.FindByCellNo(ctx, cellNo)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("格口编号已存在")
	}

	cell := &model.LockerCell{
		CellNo: cellNo,
		Size:   strings.ToUpper(size),
		Status: "FREE",
	}
	if err := s.cellRepo.Create(ctx, cell); err != nil {
		return nil, err
	}
	return cell, nil
}

// DisableCell 停用格口
func (s *LockerService) DisableCell(ctx context.Context, cellNo string) error {
	s.cellMux.Lock()
	defer s.cellMux.Unlock()

	cell, err := s.cellRepo.FindByCellNo(ctx, cellNo)
	if err != nil {
		return err
	}
	if cell == nil {
		return errors.New("格口不存在")
	}

	// 如果已经是停用状态，直接返回成功
	if cell.Status == "DISABLED" {
		return nil
	}

	// 检查格口内是否有已存但未取走的包裹
	parcel, err := s.parcelRepo.FindUnpickedByCellID(ctx, cell.ID)
	if err != nil {
		return err
	}

	// 如果有包裹，不能停用（或者可以停用但不接新件，已存件可取）
	// 根据需求：停用后已存放的件照旧能取
	// 所以可以直接停用，但需要确保在柜的件仍可取
	if parcel != nil {
		// 有在柜包裹，需要先释放包裹？不，需求说已放进去的件照旧能取
		// 停用只是不再接新存件，已存件继续正常取
	}

	return s.cellRepo.UpdateStatus(ctx, cell.ID, "DISABLED")
}

// Parcel 存件/取件

// StoreParcel 存件
func (s *LockerService) StoreParcel(ctx context.Context, parcelNo, cellNo, phone, size string) (*model.Parcel, error) {
	// 校验参数
	if parcelNo == "" || cellNo == "" || phone == "" || size == "" {
		return nil, errors.New("参数不能为空")
	}

	validSizes := map[string]bool{"SMALL": true, "MEDIUM": true, "LARGE": true}
	if !validSizes[strings.ToUpper(size)] {
		return nil, errors.New("尺寸档必须为 SMALL、MEDIUM 或 LARGE")
	}

	// 并发控制：先锁定整个格口的操作
	s.cellMux.Lock()
	defer s.cellMux.Unlock()

	// 查找格口
	cell, err := s.cellRepo.FindByCellNo(ctx, cellNo)
	if err != nil {
		return nil, err
	}
	if cell == nil {
		return nil, errors.New("格口不存在")
	}

	// 检查格口是否停用
	if cell.Status == "DISABLED" {
		return nil, errors.New("格口已停用")
	}

	// 检查格口是否已占
	existingParcel, err := s.parcelRepo.FindUnpickedByCellID(ctx, cell.ID)
	if err != nil {
		return nil, err
	}
	if existingParcel != nil {
		return nil, errors.New("格口已被占用")
	}

	// 检查尺寸是否匹配
	if !s.canFit(size, cell.Size) {
		return nil, fmt.Errorf("包裹尺寸超过格口尺寸档（包裹: %s, 格口: %s）", size, cell.Size)
	}

	// 创建包裹
	parcel := &model.Parcel{
		ParcelNo: parcelNo,
		CellID:   cell.ID,
		Phone:    phone,
		Size:     strings.ToUpper(size),
		Status:   "STORED",
		StoredAt: time.Now(), // 设置当前时间
	}

	if err := s.parcelRepo.Create(ctx, parcel); err != nil {
		return nil, err
	}

	// 更新格口状态为占用（可选，用于查询）
	// 不更新格口状态，因为状态只在 FREE/DISABLED 之间切换
	// 格口占用通过检查是否有 STORED 状态的包裹来判断

	return parcel, nil
}

// PickupParcel 取件
func (s *LockerService) PickupParcel(ctx context.Context, parcelNo string) error {
	if parcelNo == "" {
		return errors.New("包裹单号不能为空")
	}

	// 加锁防止并发取件问题
	s.parcelMux.Lock()
	defer s.parcelMux.Unlock()

	// 查找包裹
	parcel, err := s.parcelRepo.FindByParcelNo(ctx, parcelNo)
	if err != nil {
		return err
	}
	if parcel == nil {
		return errors.New("包裹不存在")
	}

	// 如果已经取走，直接返回成功（重复取件不报错）
	if parcel.Status == "PICKED" {
		return nil
	}

	// 更新包裹状态
	if err := s.parcelRepo.Pick(ctx, parcel.ID); err != nil {
		return err
	}

	// 更新格口状态为空闲（可选）
	cell, err := s.cellRepo.FindByID(ctx, parcel.CellID)
	if err != nil {
		return err
	}
	if cell != nil && cell.Status != "FREE" && cell.Status != "DISABLED" {
		_ = s.cellRepo.UpdateStatus(ctx, cell.ID, "FREE")
	}

	return nil
}

// ListParcelsByPhone 查询用户在柜包裹
func (s *LockerService) ListParcelsByPhone(ctx context.Context, phone string, pageNum, pageSize int) (map[string]interface{}, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	parcels, total, err := s.parcelRepo.PageByPhone(ctx, phone, pageNum, pageSize)
	if err != nil {
		return nil, err
	}

	// 构造返回数据
	var result []map[string]interface{}
	for _, p := range parcels {
		result = append(result, map[string]interface{}{
			"parcelNo": p.ParcelNo,
			"cellNo":   "", // 需要查格口
			"phone":    p.Phone,
			"size":     p.Size,
		})
	}

	// 补充格口编号
	for i := range result {
		if parcels[i].CellID > 0 {
			cell, err := s.cellRepo.FindByID(ctx, parcels[i].CellID)
			if err == nil && cell != nil {
				result[i]["cellNo"] = cell.CellNo
			}
		}
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return map[string]interface{}{
		"content": result,
		"total":   total,
		"pageNum": pageNum,
		"pageSize": pageSize,
		"totalPages": totalPages,
	}, nil
}

// CellInfo 格口详细信息（用于注册时返回）
type CellInfo struct {
	CellNo string `json:"cellNo"`
	Size   string `json:"size"`
	Status string `json:"status"`
}

// CellDisableResponse 停用格口返回
type CellDisableResponse struct {
	CellNo string `json:"cellNo"`
	Status string `json:"status"`
}

// ListAllCells 查询所有格口
func (s *LockerService) ListAllCells(ctx context.Context) ([]model.LockerCell, error) {
	return s.cellRepo.FindAll(ctx)
}

// ListDisabledCells 查询停用的格口
func (s *LockerService) ListDisabledCells(ctx context.Context) ([]model.LockerCell, error) {
	return s.cellRepo.FindDisabled(ctx)
}