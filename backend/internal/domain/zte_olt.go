package domain

// ZTEONUInfo represents summary information of an ONU connected to ZTE OLT.
type ZTEONUInfo struct {
	Board        int    `json:"board"`
	PON          int    `json:"pon"`
	ONUID        int    `json:"onu_id"`
	Name         string `json:"name"`
	ONUType      string `json:"onu_type"`
	SerialNumber string `json:"serial_number"`
	RxPower      string `json:"rx_power"`
	Status       string `json:"status"`
}

// ZTEONUDetail represents comprehensive telemetry data for a specific ONU.
type ZTEONUDetail struct {
	Board                int    `json:"board"`
	PON                  int    `json:"pon"`
	ONUID                int    `json:"onu_id"`
	Name                 string `json:"name"`
	Description          string `json:"description"`
	ONUType              string `json:"onu_type"`
	SerialNumber         string `json:"serial_number"`
	RxPower              string `json:"rx_power"`
	TxPower              string `json:"tx_power"`
	Status               string `json:"status"`
	IPAddress            string `json:"ip_address"`
	LastOnline           string `json:"last_online"`
	LastOffline          string `json:"last_offline"`
	Uptime               string `json:"uptime"`
	LastDownTimeDuration string `json:"last_down_time_duration"`
	OfflineReason        string `json:"offline_reason"`
	GponOpticalDistance  string `json:"gpon_optical_distance"`
}

// ZTECardInfo represents a line card or control board in ZTE OLT.
type ZTECardInfo struct {
	EntIndex int    `json:"ent_index"`
	Slot     int    `json:"slot"`
	Type     string `json:"type"`
	Role     string `json:"role"`
}

// ZTEPortInfo represents an uplink port in ZTE OLT.
type ZTEPortInfo struct {
	Name        string `json:"name"`
	Shelf       int    `json:"shelf"`
	Slot        int    `json:"slot"`
	Port        int    `json:"port"`
	Kind        string `json:"kind"`
	AdminStatus string `json:"admin_status"`
	OperStatus  string `json:"oper_status"`
	SpeedMbps   int    `json:"speed_mbps"`
}

// ZTEUplinksData represents detected cards and uplink ports in ZTE OLT.
type ZTEUplinksData struct {
	Cards []ZTECardInfo `json:"cards"`
	Ports []ZTEPortInfo `json:"ports"`
}

// ZTEPaginationMeta holds pagination metadata for ONU listings.
type ZTEPaginationMeta struct {
	Page      int `json:"page"`
	Limit     int `json:"limit"`
	PageCount int `json:"page_count"`
	TotalRows int `json:"total_rows"`
}

// ZTEPaginatedONUs represents paginated response of ONUs.
type ZTEPaginatedONUs struct {
	Data []ZTEONUInfo      `json:"data"`
	Meta ZTEPaginationMeta `json:"meta"`
}

// ONUSerialInfo represents an ONU ID to Serial Number mapping.
type ONUSerialInfo struct {
	ONUID        int    `json:"onu_id"`
	SerialNumber string `json:"serial_number"`
}
