package repository

import (
	"context"

	"gorm.io/gorm"

	"somepro/internal/model"
)

// LockerCellRepository 格口仓储
type LockerCellRepository struct {
	db *gorm.DB
}

func NewLockerCellRepository(db *gorm.DB) *LockerCellRepository {
	return &LockerCellRepository{db: db}
}

// Create 创建格口
func (r *LockerCellRepository) Create(ctx context.Context, cell *model.LockerCell) error {
	return r.db.WithContext(ctx).Create(cell).Error
}

// FindByCellNo 通过编号查询格口
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

// FindByID 通过ID查询格口
func (r *LockerCellRepository) FindByID(ctx context.Context, id uint) (*model.LockerCell, error) {
	var cell model.LockerCell
	if err := r.db.WithContext(ctx).First(&cell, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cell, nil
}

// FindAllFreeOrOccupied 查询所有未停用的格口（FREE或OCCUPIED状态）
func (r *LockerCellRepository) FindAll(ctx context.Context) ([]model.LockerCell, error) {
	var cells []model.LockerCell
	if err := r.db.WithContext(ctx).Where("status != ?", "DISABLED").Find(&cells).Error; err != nil {
		return nil, err
	}
	return cells, nil
}

// FindDisabled 查询所有停用的格口
func (r *LockerCellRepository) FindDisabled(ctx context.Context) ([]model.LockerCell, error) {
	var cells []model.LockerCell
	if err := r.db.WithContext(ctx).Where("status = ?", "DISABLED").Find(&cells).Error; err != nil {
		return nil, err
	}
	return cells, nil
}

// UpdateStatus 更新格口状态
func (r *LockerCellRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	return r.db.WithContext(ctx).Model(&model.LockerCell{}).Where("id = ?", id).
		Update("status", status).Error
}

// ParcelRepository 包裹仓储
type ParcelRepository struct {
	db *gorm.DB
}

func NewParcelRepository(db *gorm.DB) *ParcelRepository {
	return &ParcelRepository{db: db}
}

// Create 创建包裹
func (r *ParcelRepository) Create(ctx context.Context, parcel *model.Parcel) error {
	return r.db.WithContext(ctx).Create(parcel).Error
}

// FindByParcelNo 通过单号查询包裹
func (r *ParcelRepository) FindByParcelNo(ctx context.Context, parcelNo string) (*model.Parcel, error) {
	var parcel model.Parcel
	if err := r.db.WithContext(ctx).Where("parcel_no = ?", parcelNo).First(&parcel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &parcel, nil
}

// FindByCellID 查询格口内的包裹（未取走的）
func (r *ParcelRepository) FindByCellID(ctx context.Context, cellID uint) (*model.Parcel, error) {
	var parcel model.Parcel
	if err := r.db.WithContext(ctx).
		Where("cell_id = ? AND status = ?", cellID, "STORED").
		First(&parcel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &parcel, nil
}

// UpdateStatus 更新包裹状态
func (r *ParcelRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	return r.db.WithContext(ctx).Model(&model.Parcel{}).Where("id = ?", id).
		Update("status", status).Error
}

// Pick parcel by ID and set picked_at
func (r *ParcelRepository) Pick(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.Parcel{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":    "PICKED",
			"picked_at": gorm.Expr("NOW()"),
		}).Error
}

// PageByPhone 按手机号分页查询在柜包裹
func (r *ParcelRepository) PageByPhone(ctx context.Context, phone string, pageNum, pageSize int) ([]model.Parcel, int64, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Parcel{}).
		Where("phone = ? AND status = ?", phone, "STORED").
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var parcels []model.Parcel
	if err := r.db.WithContext(ctx).
		Where("phone = ? AND status = ?", phone, "STORED").
		Order("id DESC").
		Offset((pageNum - 1) * pageSize).
		Limit(pageSize).
		Find(&parcels).Error; err != nil {
		return nil, 0, err
	}

	return parcels, total, nil
}

// FindUnpickedByCellID 查找未取走的包裹（用于检查格口是否被占用）
func (r *ParcelRepository) FindUnpickedByCellID(ctx context.Context, cellID uint) (*model.Parcel, error) {
	var parcel model.Parcel
	if err := r.db.WithContext(ctx).
		Where("cell_id = ? AND status = ?", cellID, "STORED").
		First(&parcel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &parcel, nil
}