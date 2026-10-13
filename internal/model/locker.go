package model

import (
	"time"

	"gorm.io/gorm"
)

// LockerCellSize 格口尺寸档
const (
	LockerCellSizeSmall    = "SMALL"
	LockerCellSizeMedium   = "MEDIUM"
	LockerCellSizeLarge    = "LARGE"
)

// LockerCellStatus 格口状态
const (
	LockerCellStatusFree    = "FREE"
	LockerCellStatusDisabled = "DISABLED"
)

// ParcelStatus 包裹状态
const (
	ParcelStatusStored = "STORED"
	ParcelStatusPicked = "PICKED"
)

// LockerCell 快递柜格口
// 审计字段由 GORM 钩子自动填充；软删除走 gorm.DeletedAt
type LockerCell struct {
	ID         uint64       `gorm:"primaryKey;column:id" json:"id"`
	CellNo     string       `gorm:"column:cell_no;size:32;not null;uniqueIndex" json:"cellNo"`
	Size       string       `gorm:"column:size;size:16;not null" json:"size"`
	Status     string       `gorm:"column:status;size:16;not null;default:'FREE'" json:"status"`
	CreateBy   string       `gorm:"column:create_by;size:64" json:"-"`
	CreateTime time.Time    `gorm:"column:create_time" json:"-"`
	UpdateBy   string       `gorm:"column:update_by;size:64" json:"-"`
	UpdateTime time.Time    `gorm:"column:update_time" json:"-"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (LockerCell) TableName() string { return "t_locker_cell" }

func (c *LockerCell) BeforeCreate(tx *gorm.DB) error {
	op := OperatorFrom(tx.Statement.Context)
	now := time.Now()
	c.CreateBy, c.CreateTime = op, now
	c.UpdateBy, c.UpdateTime = op, now
	return nil
}

func (c *LockerCell) BeforeUpdate(tx *gorm.DB) error {
	c.UpdateBy, c.UpdateTime = OperatorFrom(tx.Statement.Context), time.Now()
	return nil
}

// Parcel 快递包裹
// 审计字段由 GORM 钩子自动填充；软删除走 gorm.DeletedAt
type Parcel struct {
	ID         uint64       `gorm:"primaryKey;column:id" json:"id"`
	ParcelNo   string       `gorm:"column:parcel_no;size:32;not null;uniqueIndex" json:"parcelNo"`
	CellID     uint64       `gorm:"column:cell_id;not null" json:"-"`
	Phone      string       `gorm:"column:phone;size:32;not null" json:"phone"`
	Size       string       `gorm:"column:size;size:16;not null" json:"size"`
	Status     string       `gorm:"column:status;size:16;not null;default:'STORED'" json:"status"`
	StoredAt   *time.Time   `gorm:"column:stored_at" json:"storedAt,omitempty"`
	PickedAt   *time.Time   `gorm:"column:picked_at" json:"pickedAt,omitempty"`
	CreateBy   string       `gorm:"column:create_by;size:64" json:"-"`
	CreateTime time.Time    `gorm:"column:create_time" json:"-"`
	UpdateBy   string       `gorm:"column:update_by;size:64" json:"-"`
	UpdateTime time.Time    `gorm:"column:update_time" json:"-"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Parcel) TableName() string { return "t_parcel" }

func (p *Parcel) BeforeCreate(tx *gorm.DB) error {
	op := OperatorFrom(tx.Statement.Context)
	now := time.Now()
	p.CreateBy, p.CreateTime = op, now
	p.UpdateBy, p.UpdateTime = op, now
	return nil
}

func (p *Parcel) BeforeUpdate(tx *gorm.DB) error {
	p.UpdateBy, p.UpdateTime = OperatorFrom(tx.Statement.Context), time.Now()
	return nil
}

// Now 返回当前时间（用于在非 GORM 钩子场景创建时间戳）
func Now() time.Time {
	return time.Now()
}