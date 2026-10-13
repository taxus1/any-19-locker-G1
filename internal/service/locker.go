package service

import (
	"context"
	"errors"
	"sync"

	"somepro/internal/model"
	"somepro/internal/repository"
)

// LockerService 快递柜服务
type LockerService struct {
	cellRepo   *repository.LockerCellRepository
	parcelRepo *repository.LockerParcelRepository
}

// cellMutex 用于格口级别的并发控制
var cellMutex sync.Map

func NewLocker(cellRepo *repository.LockerCellRepository, parcelRepo *repository.LockerParcelRepository) *LockerService {
	return &LockerService{
		cellRepo:   cellRepo,
		parcelRepo: parcelRepo,
	}
}

// getCellMutex 获取格口的互斥锁
func (s *LockerService) getCellMutex(cellID uint64) *sync.Mutex {
	key := cellID
	if v, ok := cellMutex.Load(key); ok {
		return v.(*sync.Mutex)
	}
	m := &sync.Mutex{}
	_, _ = cellMutex.LoadOrStore(key, m)
	return m
}

// CreateCell 创建格口
func (s *LockerService) CreateCell(ctx context.Context, cellNo, size string) (*model.LockerCell, error) {
	// 检查格口是否存在（包括已软删除的）
	existing, err := s.cellRepo.FindByCellNo(ctx, cellNo)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		// 如果已被软删除，不允许重新创建
		return nil, errors.New("格口编号已存在")
	}

	cell := &model.LockerCell{
		CellNo: cellNo,
		Size:   size,
		Status: model.LockerCellStatusFree,
	}
	if err := s.cellRepo.Create(ctx, cell); err != nil {
		return nil, err
	}
	return cell, nil
}

// DisableCell 停用格口
func (s *LockerService) DisableCell(ctx context.Context, cellNo string) error {
	cell, err := s.cellRepo.FindByCellNo(ctx, cellNo)
	if err != nil {
		return err
	}
	if cell == nil {
		return errors.New("格口不存在")
	}
	if cell.Status == model.LockerCellStatusDisabled {
		// 已停用，直接返回成功
		return nil
	}
	cell.Status = model.LockerCellStatusDisabled
	return s.cellRepo.Update(ctx, cell)
}

// StoreParcel 存件
func (s *LockerService) StoreParcel(ctx context.Context, parcelNo, cellNo, phone, size string) (*model.Parcel, error) {
	// 查找格口
	cell, err := s.cellRepo.FindByCellNo(ctx, cellNo)
	if err != nil {
		return nil, err
	}
	if cell == nil {
		return nil, errors.New("格口不存在")
	}

	// 检查格口是否停用
	if cell.Status == model.LockerCellStatusDisabled {
		return nil, repository.ErrCellDisabled
	}

	// 检查格口是否被占用
	parcelMutex := s.getCellMutex(cell.ID)
	parcelMutex.Lock()
	defer parcelMutex.Unlock()

	// 再次检查格口状态（防止并发问题）
	if cell.Status == model.LockerCellStatusDisabled {
		return nil, repository.ErrCellDisabled
	}

	// 检查格口是否为空
	existingParcel, err := s.parcelRepo.FindByCellID(ctx, cell.ID)
	if err != nil {
		return nil, err
	}
	if existingParcel != nil {
		return nil, repository.ErrCellNotFree
	}

	// 检查尺寸是否匹配
	if !repository.CanFit(size, cell.Size) {
		return nil, repository.ErrParcelTooLarge
	}

	// 检查包裹单号是否重复
	existing, err := s.parcelRepo.FindByParcelNo(ctx, parcelNo)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("包裹单号已存在")
	}

	// 创建包裹
	now := model.Now()
	parcel := &model.Parcel{
		ParcelNo: parcelNo,
		CellID:   cell.ID,
		Phone:    phone,
		Size:     size,
		Status:   model.ParcelStatusStored,
		StoredAt: &now,
	}
	if err := s.parcelRepo.Create(ctx, parcel); err != nil {
		return nil, err
	}

	return parcel, nil
}

// PickupParcel 取件
func (s *LockerService) PickupParcel(ctx context.Context, parcelNo string) error {
	// 查找包裹
	parcel, err := s.parcelRepo.FindByParcelNo(ctx, parcelNo)
	if err != nil {
		return err
	}
	if parcel == nil {
		// 重复取件：包裹不存在，按需求"重复取别报错"
		return nil
	}

	// 检查包裹是否在柜状态
	if parcel.Status != model.ParcelStatusStored {
		// 已取走，按需求"取件能重复按，按几下都行，状态别一会儿一变"
		return nil
	}

	// 更新包裹状态为已取走
	parcel.Status = model.ParcelStatusPicked
	now := model.Now()
	parcel.PickedAt = &now

	return s.parcelRepo.Update(ctx, parcel)
}

// ListParcelsByPhone 按手机号查询在柜包裹
func (s *LockerService) ListParcelsByPhone(ctx context.Context, phone string, pageNum, pageSize int) (map[string]interface{}, error) {
	rows, total, err := s.parcelRepo.PageByPhone(ctx, phone, pageNum, pageSize)
	if err != nil {
		return nil, err
	}

	// 构造返回数据，包含 parcelNo 和 cellNo
	result := make([]map[string]interface{}, 0, len(rows))
	for _, p := range rows {
		// 通过 cellID 查格口获取 cellNo
		cell, err := s.cellRepo.FindByID(ctx, p.CellID)
		if err != nil {
			return nil, err
		}

		item := map[string]interface{}{
			"parcelNo": p.ParcelNo,
			"cellNo":   "",
			"phone":    p.Phone,
			"size":     p.Size,
		}
		if cell != nil {
			item["cellNo"] = cell.CellNo
		}
		result = append(result, item)
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return map[string]interface{}{
		"content":    result,
		"total":      total,
		"pageNum":    pageNum,
		"pageSize":   pageSize,
		"totalPages": totalPages,
	}, nil
}