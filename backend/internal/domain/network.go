package domain

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// MikrotikServer represents a MikroTik router used for network management.
type MikrotikServer struct {
	ID                   uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name                 string         `gorm:"type:varchar(191);unique;not null;index" json:"name"`
	HostIP               string         `gorm:"type:varchar(191);not null;index" json:"host_ip"`
	Username             string         `gorm:"type:varchar(191);not null;index" json:"username"`
	Password             string         `gorm:"type:text;not null" json:"password"` // Encrypted password
	Port                 int            `gorm:"default:8728;index" json:"port"`
	RosVersion           *string        `gorm:"type:varchar(191);index" json:"ros_version"`
	IsActive             bool           `gorm:"default:true;index" json:"is_active"`
	Latitude             *float64       `json:"latitude"`
	Longitude            *float64       `json:"longitude"`
	LastConnectionStatus *string        `gorm:"type:varchar(191);index" json:"last_connection_status"`
	LastConnectedAt      *time.Time     `gorm:"type:datetime;index" json:"last_connected_at"`
	CreatedAt            *time.Time     `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt            *time.Time     `gorm:"type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	DataTeknisRecords []DataTeknis `gorm:"foreignKey:MikrotikServerID" json:"data_teknis_records"`
	Pelanggan         []Pelanggan  `gorm:"foreignKey:MikrotikServerID" json:"pelanggan"`
	Olts              []OLT        `gorm:"foreignKey:MikrotikServerID" json:"olts"`
}

// TableName overrides the default table name for MikrotikServer
func (MikrotikServer) TableName() string {
	return "mikrotik_servers"
}

// UnmarshalJSON custom unmarshaler to support both host_ip and ip_address fields.
func (m *MikrotikServer) UnmarshalJSON(data []byte) error {
	type Alias MikrotikServer
	aux := &struct {
		IPAddress string `json:"ip_address"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if m.HostIP == "" && aux.IPAddress != "" {
		m.HostIP = aux.IPAddress
	}
	return nil
}

// OLT represents Optical Line Terminal.
type OLT struct {
	ID               uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	NamaOlt          string         `gorm:"type:varchar(100);unique;not null" json:"nama_olt"`
	IPAddress        string         `gorm:"type:varchar(100);unique;not null" json:"ip_address"`
	TipeOlt          string         `gorm:"type:varchar(50);not null" json:"tipe_olt"`
	Username         *string        `gorm:"type:varchar(100)" json:"username"`
	Password         *string        `gorm:"type:varchar(100)" json:"password"`
	MikrotikServerID *uint64        `gorm:"index" json:"mikrotik_server_id"`
	Latitude         *float64       `json:"latitude"`
	Longitude        *float64       `json:"longitude"`
	CreatedAt        *time.Time     `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        *time.Time     `gorm:"type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	MikrotikServer *MikrotikServer `gorm:"foreignKey:MikrotikServerID" json:"mikrotik_server"`
	Odcs           []ODC           `gorm:"foreignKey:OltID" json:"odcs"`
	Odps           []ODP           `gorm:"foreignKey:OltID" json:"odps"`
}

// TableName overrides the default table name for OLT
func (OLT) TableName() string {
	return "olt"
}

// ODC represents Optical Distribution Cabinet (FDT / Kabinet Distribusi Fiber).
type ODC struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	KodeOdc       string         `gorm:"type:varchar(100);unique;not null" json:"kode_odc"`
	NamaOdc       string         `gorm:"type:varchar(100);not null" json:"nama_odc"`
	Alamat        *string        `gorm:"type:varchar(255)" json:"alamat"`
	KapasitasPort int            `gorm:"default:144" json:"kapasitas_port"`
	Latitude      *float64       `json:"latitude"`
	Longitude     *float64       `json:"longitude"`
	OltID         *uint64        `gorm:"index" json:"olt_id"`
	CreatedAt     *time.Time     `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     *time.Time     `gorm:"type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Olt  *OLT  `gorm:"foreignKey:OltID" json:"olt"`
	Odps []ODP `gorm:"foreignKey:OdcID" json:"odps"`
}

// TableName overrides the default table name for ODC
func (ODC) TableName() string {
	return "odc"
}

// ODP represents Optical Distribution Point (Tiang ODP).
type ODP struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	KodeOdp       string         `gorm:"type:varchar(100);unique;not null" json:"kode_odp"`
	Alamat        *string        `gorm:"type:varchar(255)" json:"alamat"`
	KapasitasPort int            `gorm:"default:8" json:"kapasitas_port"`
	Latitude      *float64       `json:"latitude"`
	Longitude     *float64       `json:"longitude"`
	ParentOdpID   *uint64        `gorm:"index" json:"parent_odp_id"`
	OltID         uint64         `gorm:"index;not null" json:"olt_id"`
	OdcID         *uint64        `gorm:"index" json:"odc_id"`
	PortTerpakai  int            `gorm:"-" json:"port_terpakai"`
	CreatedAt     *time.Time     `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     *time.Time     `gorm:"type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	ParentOdp  *ODP         `gorm:"foreignKey:ParentOdpID" json:"parent_odp"`
	ChildOdps  []ODP        `gorm:"foreignKey:ParentOdpID" json:"child_odps"`
	Olt        *OLT         `gorm:"foreignKey:OltID" json:"olt"`
	Odc        *ODC         `gorm:"foreignKey:OdcID" json:"odc"`
	DataTeknis []DataTeknis `gorm:"foreignKey:OdpID" json:"data_teknis"`
}

// TableName overrides the default table name for ODP
func (ODP) TableName() string {
	return "odp"
}

// TrafficHistory stores user bandwidth monitoring logs.
type TrafficHistory struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	DataTeknisID     uint64    `gorm:"index;not null" json:"data_teknis_id"`
	MikrotikServerID uint64    `gorm:"index;not null" json:"mikrotik_server_id"`
	UsernamePppoe    string    `gorm:"type:varchar(100);index;not null" json:"username_pppoe"`
	IPAddress        string    `gorm:"type:varchar(45);index;not null" json:"ip_address"`
	RxBytes          uint64    `gorm:"type:bigint;default:0" json:"rx_bytes"`
	TxBytes          uint64    `gorm:"type:bigint;default:0" json:"tx_bytes"`
	RxPackets        uint64    `gorm:"type:bigint;default:0" json:"rx_packets"`
	TxPackets        uint64    `gorm:"type:bigint;default:0" json:"tx_packets"`
	RxMbps           float64   `gorm:"default:0.0" json:"rx_mbps"`
	TxMbps           float64   `gorm:"default:0.0" json:"tx_mbps"`
	TotalMbps        float64   `gorm:"default:0.0" json:"total_mbps"`
	UptimeSeconds    uint64    `gorm:"type:bigint;default:0" json:"uptime_seconds"`
	IsActive         bool      `gorm:"default:true;index" json:"is_active"`
	Timestamp        time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP;index" json:"timestamp"`
	IsLatest         bool      `gorm:"default:true;index" json:"is_latest"`

	// Relationships
	DataTeknis     *DataTeknis     `gorm:"foreignKey:DataTeknisID" json:"data_teknis"`
	MikrotikServer *MikrotikServer `gorm:"foreignKey:MikrotikServerID" json:"mikrotik_server"`
}

// TableName overrides the default table name for TrafficHistory
func (TrafficHistory) TableName() string {
	return "traffic_history"
}

// NetworkMapPelanggan represents customer mapping data for GIS
type NetworkMapPelanggan struct {
	ID            uint64   `json:"id"`
	Nama          string   `json:"nama"`
	Alamat        string   `json:"alamat"`
	NoTelp        string   `json:"no_telp"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
	OdpID         *uint64  `json:"odp_id"`
	KodeOdp       string   `json:"kode_odp"`
	PortOdp       *int     `json:"port_odp"`
	UsernamePppoe string   `json:"username_pppoe"`
	Status        string   `json:"status"`
}

// NetworkMapData represents aggregated GIS network topology data
type NetworkMapData struct {
	Servers    []MikrotikServer      `json:"servers"`
	Olts       []OLT                 `json:"olts"`
	Odcs       []ODC                 `json:"odcs"`
	Odps       []ODP                 `json:"odps"`
	Pelanggans []NetworkMapPelanggan `json:"pelanggans"`
}

// UpdateCoordinateRequest represents payload for moving any network node on the map
type UpdateCoordinateRequest struct {
	Type      string  `json:"type" binding:"required"` // "server", "olt", "odc", "odp", "pelanggan"
	ID        uint64  `json:"id" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

// BatchImportODPRequest represents payload for bulk importing ODPs / nodes from KMZ or manual list
type BatchImportODPRequest struct {
	Items []BatchODPItem `json:"items" binding:"required"`
}

type BatchODPItem struct {
	KodeOdp       string   `json:"kode_odp"`
	Alamat        string   `json:"alamat"`
	KapasitasPort int      `json:"kapasitas_port"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
	OltID         uint64   `json:"olt_id"`
	OdcID         *uint64  `json:"odc_id"`
}
