package domain

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// Pelanggan represents the customer table.
type Pelanggan struct {
	ID               uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	NoKtp            string         `gorm:"type:varchar(191)" json:"no_ktp"`
	Nama             string         `gorm:"type:varchar(191)" json:"nama"`
	Alamat           string         `gorm:"type:varchar(191)" json:"alamat"`
	AlamatCustom     *string        `gorm:"type:varchar(191)" json:"alamat_custom"`
	TglInstalasi     *time.Time     `gorm:"type:date" json:"tgl_instalasi"`
	Blok             string         `gorm:"type:varchar(191)" json:"blok"`
	Unit             string         `gorm:"type:varchar(191)" json:"unit"`
	NoTelp           string         `gorm:"type:varchar(191)" json:"no_telp"`
	Email            string         `gorm:"type:varchar(191);uniqueIndex" json:"email"`
	IDBrand          *string        `gorm:"type:varchar(191)" json:"id_brand"`
	Layanan          *string        `gorm:"type:varchar(191)" json:"layanan"`
	BrandDefault     *string        `gorm:"type:varchar(191)" json:"brand_default"`
	MikrotikServerID *uint64        `gorm:"index" json:"mikrotik_server_id"`
	CreatedAt        *time.Time     `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        *time.Time     `gorm:"type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Index tags could be complex in GORM. Usually, we define complex composite indexes on struct level,
	// but GORM 2 allows them on fields: `gorm:"index:idx_name,priority:1"`.
	// For simplicity, we keep them simple, GORM will auto-migrate basic indexes.

	// Relationships
	DataTeknis     *DataTeknis     `gorm:"foreignKey:PelangganID" json:"data_teknis"`
	Langganan      []Langganan     `gorm:"foreignKey:PelangganID" json:"langganan"`
	Invoices       []Invoice       `gorm:"foreignKey:PelangganID" json:"invoices"`
	MikrotikServer *MikrotikServer `gorm:"foreignKey:MikrotikServerID" json:"mikrotik_server"`
	HargaLayanan   *HargaLayanan   `gorm:"foreignKey:IDBrand;references:IDBrand;constraint:-" json:"harga_layanan"`
	TroubleTickets []TroubleTicket `gorm:"foreignKey:PelangganID" json:"trouble_tickets"`
	InventoryItems []InventoryItem `gorm:"foreignKey:PelangganID" json:"inventory_items"`
}

// DataTeknis represents the technical connection data for a customer.
type DataTeknis struct {
	ID                  uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	PelangganID         uint64         `gorm:"uniqueIndex;not null" json:"pelanggan_id"`
	IDPelanggan         string         `gorm:"type:varchar(191);index" json:"id_pelanggan"` // PPPoE username
	PasswordPppoe       string         `gorm:"type:varchar(191);index" json:"password_pppoe"`
	ProfilePppoe        *string        `gorm:"type:varchar(191);index" json:"profile_pppoe"`
	IPPelanggan         *string        `gorm:"type:varchar(191);index" json:"ip_pelanggan"`
	IDVlan              *string        `gorm:"type:varchar(191);index" json:"id_vlan"`
	Olt                 *string        `gorm:"type:varchar(191);index" json:"olt"`
	OltCustom           *string        `gorm:"type:varchar(191);index" json:"olt_custom"`
	Pon                 *int           `gorm:"index" json:"pon"`
	Otb                 *int           `gorm:"index" json:"otb"`
	Odc                 *int           `gorm:"index" json:"odc"`
	OdpID               *uint64        `gorm:"index" json:"odp_id"`
	PortOdp             *int           `gorm:"index" json:"port_odp"`
	Sn                  *string        `gorm:"type:varchar(191);index" json:"sn"`
	OnuPower            *int           `gorm:"index" json:"onu_power"`
	SpeedtestProof      *string        `gorm:"type:varchar(191);index" json:"speedtest_proof"`
	MikrotikSyncPending bool           `gorm:"default:false;index" json:"mikrotik_sync_pending"`
	MikrotikServerID    *uint64        `gorm:"index" json:"mikrotik_server_id"`
	CreatedAt           *time.Time     `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt           *time.Time     `gorm:"type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`

	// Relationships
	Pelanggan      *Pelanggan      `gorm:"foreignKey:PelangganID" json:"pelanggan"`
	MikrotikServer *MikrotikServer `gorm:"foreignKey:MikrotikServerID" json:"mikrotik_server"`
	Odp            *ODP            `gorm:"foreignKey:OdpID" json:"odp"`
	TroubleTickets []TroubleTicket `gorm:"foreignKey:DataTeknisID" json:"trouble_tickets"`
}

// UnmarshalJSON custom unmarshaler to support binding numbers that might be passed as strings.
func (dt *DataTeknis) UnmarshalJSON(data []byte) error {
	type Alias DataTeknis
	aux := &struct {
		Pon     interface{} `json:"pon"`
		Otb     interface{} `json:"otb"`
		Odc     interface{} `json:"odc"`
		PortOdp interface{} `json:"port_odp"`
		*Alias
	}{
		Alias: (*Alias)(dt),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	parseInt := func(val interface{}) (*int, error) {
		if val == nil {
			return nil, nil
		}
		switch v := val.(type) {
		case float64:
			i := int(v)
			return &i, nil
		case string:
			if v == "" {
				return nil, nil
			}
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, err
			}
			return &i, nil
		default:
			return nil, fmt.Errorf("invalid type for integer field")
		}
	}

	var err error
	dt.Pon, err = parseInt(aux.Pon)
	if err != nil {
		return fmt.Errorf("failed to parse pon: %w", err)
	}

	dt.Otb, err = parseInt(aux.Otb)
	if err != nil {
		return fmt.Errorf("failed to parse otb: %w", err)
	}

	dt.Odc, err = parseInt(aux.Odc)
	if err != nil {
		return fmt.Errorf("failed to parse odc: %w", err)
	}

	dt.PortOdp, err = parseInt(aux.PortOdp)
	if err != nil {
		return fmt.Errorf("failed to parse port_odp: %w", err)
	}

	return nil
}

// UnmarshalJSON custom unmarshaler for Pelanggan to handle date string unmarshaling for TglInstalasi.
func (p *Pelanggan) UnmarshalJSON(data []byte) error {
	type Alias Pelanggan
	aux := &struct {
		TglInstalasi *string `json:"tgl_instalasi"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.TglInstalasi != nil && *aux.TglInstalasi != "" {
		// Try parsing as RFC3339 first
		t, err := time.Parse(time.RFC3339, *aux.TglInstalasi)
		if err == nil {
			p.TglInstalasi = &t
		} else {
			// Try parsing as YYYY-MM-DD
			t2, err2 := time.Parse("2006-01-02", *aux.TglInstalasi)
			if err2 == nil {
				p.TglInstalasi = &t2
			} else {
				// Try YYYY-MM-DD with timezone/time suffix if any
				t3, err3 := time.Parse("2006-01-02 15:04:05", *aux.TglInstalasi)
				if err3 == nil {
					p.TglInstalasi = &t3
				} else {
					return fmt.Errorf("failed to parse tgl_instalasi: %v", err2)
				}
			}
		}
	} else {
		p.TglInstalasi = nil
	}

	return nil
}


