package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// PaketLayanan represents an internet package sold to customers.
type PaketLayanan struct {
	ID        uint64  `gorm:"primaryKey;autoIncrement" json:"id"`
	IDBrand   string  `gorm:"type:varchar(191);not null;index" json:"id_brand"`
	NamaPaket string  `gorm:"type:varchar(191);not null" json:"nama_paket"`
	Kecepatan int     `gorm:"not null" json:"kecepatan"`
	Harga     float64 `gorm:"type:decimal(15,2);not null" json:"harga"`

	// Relationships
	Langganan    []Langganan   `gorm:"foreignKey:PaketLayananID" json:"langganan"`
	HargaLayanan *HargaLayanan `gorm:"foreignKey:IDBrand;references:IDBrand;constraint:-" json:"harga_layanan"`
}

// TableName overrides the default table name for PaketLayanan
func (PaketLayanan) TableName() string {
	return "paket_layanan"
}

// HargaLayanan represents service provider brands and tax settings.
type HargaLayanan struct {
	IDBrand       string  `gorm:"type:varchar(191);primaryKey" json:"id_brand"`
	Brand         string  `gorm:"type:varchar(191);not null" json:"brand"`
	Pajak         float64 `gorm:"type:decimal(5,2);not null" json:"pajak"`
	XenditKeyName string  `gorm:"type:varchar(50);default:'JAKINET';not null" json:"xendit_key_name"`
}

// TableName overrides the default table name for HargaLayanan
func (HargaLayanan) TableName() string {
	return "harga_layanan"
}

// Diskon represents discounts active for specific customer clusters.
type Diskon struct {
	ID               uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	NamaDiskon       string     `gorm:"type:varchar(191);not null" json:"nama_diskon"`
	PersentaseDiskon float64    `gorm:"type:decimal(5,2);not null" json:"persentase_diskon"`
	Cluster          string     `gorm:"type:varchar(191);index;not null" json:"cluster"`
	IsActive         bool       `gorm:"default:true;index;not null" json:"is_active"`
	TglMulai         *time.Time `gorm:"type:date;index" json:"tgl_mulai"`
	TglSelesai       *time.Time `gorm:"type:date;index" json:"tgl_selesai"`
	CreatedAt        *time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        *time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
}

// TableName overrides the default table name for Diskon
func (Diskon) TableName() string {
	return "diskon"
}

// UnmarshalJSON custom unmarshaler for Diskon to handle date string unmarshaling for TglMulai and TglSelesai.
func (d *Diskon) UnmarshalJSON(data []byte) error {
	type Alias Diskon
	aux := &struct {
		TglMulai   *string `json:"tgl_mulai"`
		TglSelesai *string `json:"tgl_selesai"`
		*Alias
	}{
		Alias: (*Alias)(d),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	parseDate := func(str *string) (*time.Time, error) {
		if str == nil || *str == "" {
			return nil, nil
		}
		t, err := time.Parse(time.RFC3339, *str)
		if err == nil {
			return &t, nil
		}
		t2, err2 := time.Parse("2006-01-02", *str)
		if err2 == nil {
			return &t2, nil
		}
		t3, err3 := time.Parse("2006-01-02 15:04:05", *str)
		if err3 == nil {
			return &t3, nil
		}
		return nil, err2
	}

	var err error
	d.TglMulai, err = parseDate(aux.TglMulai)
	if err != nil {
		return fmt.Errorf("failed to parse tgl_mulai: %w", err)
	}

	d.TglSelesai, err = parseDate(aux.TglSelesai)
	if err != nil {
		return fmt.Errorf("failed to parse tgl_selesai: %w", err)
	}

	return nil
}


