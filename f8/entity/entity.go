package entity

import (
	"github.com/google/uuid"
	"github.com/phanirithvij/fate/f8/xfs"
	"gorm.io/gorm"
)

// BaseEntity a base entity model
type BaseEntity struct {
	// ID entity id autogenerated
	UID string
	// https://gorm.io/docs/has_many.html
	Buckets []xfs.Bucket `gorm:"foreignKey:EntityID" json:"buckets"`
}

// NewBase a new base
func NewBase() *BaseEntity {
	id := uuid.New()
	return &BaseEntity{
		UID:     id.String(),
		Buckets: []xfs.Bucket{{ID: "default"}},
	}
}

// AutoMigrate auto migrations required for the database
func (b *BaseEntity) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&xfs.Bucket{})
}
