package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"somepro/internal/model"
)

// LockerCellRepository 格口仓储
type LockerCellRepository struct {
	db *gorm.DB
}

func NewLockerCell(db *gorm.DB) *LockerCellRepository {
	return &LockerCellRepository{db: db}
}

// FindByCellNo 根据格口编号查找（含软删除）
func (r *LockerCellRepository) FindByCellNo(ctx context.Context, cellNo string) (*model.LockerCell, error) {
	var cell model.LockerCell
	if err := r.db.WithContext(ctx).Where("cell_no = ?", cellNo).First(&cell).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cell, nil
}

// FindAllFree 查询所有空闲格口
func (r *LockerCellRepository) FindAllFree(ctx context.Context) ([]model.LockerCell, error) {
	var cells []model.LockerCell
	if err := r.db.WithContext(ctx).Where("status = ?", model.LockerCellStatusFree).Find(&cells).Error; err != nil {
		return nil, err
	}
	return cells, nil
}

// FindByID 根据 ID 查找格口（含软删除）
func (r *LockerCellRepository) FindByID(ctx context.Context, id uint64) (*model.LockerCell, error) {
	var cell model.LockerCell
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&cell).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cell, nil
}

// Create 创建格口
func (r *LockerCellRepository) Create(ctx context.Context, cell *model.LockerCell) error {
	return r.db.WithContext(ctx).Create(cell).Error
}

// Update 更新格口
func (r *LockerCellRepository) Update(ctx context.Context, cell *model.LockerCell) error {
	return r.db.WithContext(ctx).Save(cell).Error
}

// LockerParcelRepository 包裹仓储
type LockerParcelRepository struct {
	db *gorm.DB
}

func NewLockerParcel(db *gorm.DB) *LockerParcelRepository {
	return &LockerParcelRepository{db: db}
}

// FindByParcelNo 根据包裹单号查找（含软删除）
func (r *LockerParcelRepository) FindByParcelNo(ctx context.Context, parcelNo string) (*model.Parcel, error) {
	var parcel model.Parcel
	if err := r.db.WithContext(ctx).Where("parcel_no = ?", parcelNo).First(&parcel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &parcel, nil
}

// FindByCellID 查询格口当前的包裹（用于存件时检查格口是否为空）
func (r *LockerParcelRepository) FindByCellID(ctx context.Context, cellID uint64) (*model.Parcel, error) {
	var parcel model.Parcel
	if err := r.db.WithContext(ctx).Where("cell_id = ? AND status = ?", cellID, model.ParcelStatusStored).First(&parcel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &parcel, nil
}

// Create 创建包裹
func (r *LockerParcelRepository) Create(ctx context.Context, parcel *model.Parcel) error {
	return r.db.WithContext(ctx).Create(parcel).Error
}

// Update 更新包裹
func (r *LockerParcelRepository) Update(ctx context.Context, parcel *model.Parcel) error {
	return r.db.WithContext(ctx).Save(parcel).Error
}

// PageByPhone 按手机号分页查询在柜包裹
func (r *LockerParcelRepository) PageByPhone(ctx context.Context, phone string, pageNum, pageSize int) ([]model.Parcel, int64, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var total int64
	cq := r.db.WithContext(ctx).Model(&model.Parcel{}).Where("phone = ? AND status = ?", phone, model.ParcelStatusStored)
	if err := cq.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []model.Parcel
	q := r.db.WithContext(ctx).Model(&model.Parcel{}).Where("phone = ? AND status = ?", phone, model.ParcelStatusStored)
	if err := q.Order("id DESC").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

// LockerCellSizeRule 尺寸档规则
var LockerCellSizeOrder = map[string]int{
	model.LockerCellSizeSmall:    1,
	model.LockerCellSizeMedium:   2,
	model.LockerCellSizeLarge:    3,
}

// CanFit 判断包裹尺寸是否能装入格口
func CanFit(parcelSize, cellSize string) bool {
	parcelOrder, ok1 := LockerCellSizeOrder[parcelSize]
	cellOrder, ok2 := LockerCellSizeOrder[cellSize]
	if !ok1 || !ok2 {
		return false
	}
	return parcelOrder <= cellOrder
}

// IsCellDisabled 判断格口是否已停用
func IsCellDisabled(cell *model.LockerCell) bool {
	return cell != nil && cell.Status == model.LockerCellStatusDisabled
}

// IsCellFree 判断格口是否空闲
func IsCellFree(cell *model.LockerCell) bool {
	return cell != nil && cell.Status == model.LockerCellStatusFree
}

// IsParcelStored 判断包裹是否在柜状态
func IsParcelStored(parcel *model.Parcel) bool {
	return parcel != nil && parcel.Status == model.ParcelStatusStored
}

// ErrCellNotFree 格口非空闲
var ErrCellNotFree = errors.New("格口已被占用")

// ErrCellDisabled 格口已停用
var ErrCellDisabled = errors.New("格口已停用")

// ErrParcelTooLarge 包裹尺寸过大
var ErrParcelTooLarge = errors.New("包裹尺寸超过格口容量")

// ErrParcelNotFound 包裹未找到
var ErrParcelNotFound = errors.New("包裹不存在")

// ErrParcelNotStored 包裹不在柜状态
var ErrParcelNotStored = errors.New("包裹不在柜状态")