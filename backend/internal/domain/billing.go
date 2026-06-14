package domain

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Langganan represents an active customer subscription to a package.
type Langganan struct {
	ID                 uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	PelangganID        uint64     `gorm:"index;not null" json:"pelanggan_id"`
	PaketLayananID     uint64     `gorm:"index;not null" json:"paket_layanan_id"`
	Status             string     `gorm:"type:varchar(100);index;not null" json:"status"`
	TglJatuhTempo              *time.Time `gorm:"type:date;index" json:"tgl_jatuh_tempo"`
	TglJatuhTempoPembayaran    *time.Time `gorm:"type:date;index" json:"tgl_jatuh_tempo_pembayaran"`
	TglInvoiceTerakhir         *time.Time `gorm:"type:date;index" json:"tgl_invoice_terakhir"`
	TglMulaiLangganan  *time.Time `gorm:"type:date;index" json:"tgl_mulai_langganan"`
	TglBerhenti        *time.Time `gorm:"type:date;index" json:"tgl_berhenti"`
	MetodePembayaran   string     `gorm:"type:varchar(50);default:'Otomatis';index" json:"metode_pembayaran"`
	HargaAwal          *float64   `gorm:"type:decimal(15,2);index" json:"harga_awal"`
	AlasanBerhenti     *string    `gorm:"type:varchar(500)" json:"alasan_berhenti"`
	StatusModem        *string    `gorm:"type:varchar(50);index" json:"status_modem"`
	WhatsappStatus     *string    `gorm:"type:varchar(50);index" json:"whatsapp_status"`
	LastWhatsappSent   *time.Time `gorm:"type:datetime;index" json:"last_whatsapp_sent"`
	RiwayatTglBerhenti *string    `gorm:"type:varchar(2000)" json:"riwayat_tgl_berhenti"` // Stored as JSON string
	CreatedAt          *time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP;index" json:"created_at"`
	UpdatedAt          *time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime;index" json:"updated_at"`

	// Relationships
	Pelanggan    *Pelanggan    `gorm:"foreignKey:PelangganID" json:"-"`
	PaketLayanan *PaketLayanan `gorm:"foreignKey:PaketLayananID" json:"paket_layanan"`

	// Helper JSON-only fields (not persisted)
	SertakanBulanDepan bool `gorm:"-" json:"sertakan_bulan_depan"`
}

// TableName overrides the default table name for Langganan
func (Langganan) TableName() string {
	return "langganan"
}

// UnmarshalJSON custom unmarshaler for Langganan to handle various date string formats.
func (l *Langganan) UnmarshalJSON(data []byte) error {
	type Alias Langganan
	aux := &struct {
		TglJatuhTempo           interface{} `json:"tgl_jatuh_tempo"`
		TglJatuhTempoPembayaran interface{} `json:"tgl_jatuh_tempo_pembayaran"`
		TglInvoiceTerakhir      interface{} `json:"tgl_invoice_terakhir"`
		TglMulaiLangganan       interface{} `json:"tgl_mulai_langganan"`
		TglBerhenti             interface{} `json:"tgl_berhenti"`
		LastWhatsappSent        interface{} `json:"last_whatsapp_sent"`
		TglMulai                interface{} `json:"tgl_mulai"`
		SertakanBulanDepan      bool        `json:"sertakan_bulan_depan"`
		*Alias
	}{
		Alias: (*Alias)(l),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	parseTime := func(val interface{}) (*time.Time, error) {
		if val == nil {
			return nil, nil
		}
		str, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("invalid type for date field, expected string")
		}
		if str == "" {
			return nil, nil
		}

		formats := []string{
			"2006-01-02",
			time.RFC3339,
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05.999Z07:00",
		}
		var parsed time.Time
		var err error
		for _, f := range formats {
			parsed, err = time.Parse(f, str)
			if err == nil {
				return &parsed, nil
			}
		}
		return nil, fmt.Errorf("cannot parse %q as date/time", str)
	}

	var err error
	l.TglJatuhTempo, err = parseTime(aux.TglJatuhTempo)
	if err != nil {
		return fmt.Errorf("failed to parse tgl_jatuh_tempo: %w", err)
	}
	l.TglJatuhTempoPembayaran, err = parseTime(aux.TglJatuhTempoPembayaran)
	if err != nil {
		return fmt.Errorf("failed to parse tgl_jatuh_tempo_pembayaran: %w", err)
	}
	l.TglInvoiceTerakhir, err = parseTime(aux.TglInvoiceTerakhir)
	if err != nil {
		return fmt.Errorf("failed to parse tgl_invoice_terakhir: %w", err)
	}
	l.TglMulaiLangganan, err = parseTime(aux.TglMulaiLangganan)
	if err != nil {
		return fmt.Errorf("failed to parse tgl_mulai_langganan: %w", err)
	}
	l.TglBerhenti, err = parseTime(aux.TglBerhenti)
	if err != nil {
		return fmt.Errorf("failed to parse tgl_berhenti: %w", err)
	}
	l.LastWhatsappSent, err = parseTime(aux.LastWhatsappSent)
	if err != nil {
		return fmt.Errorf("failed to parse last_whatsapp_sent: %w", err)
	}

	// If tgl_mulai was supplied, map it to TglMulaiLangganan if not already set
	if l.TglMulaiLangganan == nil && aux.TglMulai != nil {
		l.TglMulaiLangganan, err = parseTime(aux.TglMulai)
		if err != nil {
			return fmt.Errorf("failed to parse tgl_mulai: %w", err)
		}
	}

	l.SertakanBulanDepan = aux.SertakanBulanDepan

	return nil
}

// Invoice represents a billing invoice.
type Invoice struct {
	ID                 uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	InvoiceNumber      string         `gorm:"type:varchar(191);uniqueIndex;not null" json:"invoice_number"`
	PelangganID        uint64         `gorm:"index;not null" json:"pelanggan_id"`
	IDPelanggan        string         `gorm:"type:varchar(255)" json:"id_pelanggan"` // Historical string ID
	Brand              string         `gorm:"type:varchar(191)" json:"brand"`
	NoTelp             string         `gorm:"type:varchar(191)" json:"no_telp"`
	Email              string         `gorm:"type:varchar(191)" json:"email"`
	TotalHarga         float64        `gorm:"type:decimal(15,2);not null" json:"total_harga"`
	TglInvoice         time.Time      `gorm:"type:date;not null" json:"tgl_invoice"`
	TglJatuhTempo      time.Time      `gorm:"type:date;not null;index" json:"tgl_jatuh_tempo"`
	StatusInvoice      string         `gorm:"type:varchar(50);not null" json:"status_invoice"`
	DiskonID           *uint64        `gorm:"index" json:"diskon_id"`
	DiskonPersen       *float64       `gorm:"type:decimal(5,2)" json:"diskon_persen"`
	DiskonAmount       *float64       `gorm:"type:decimal(15,2)" json:"diskon_amount"`
	HargaSebelumDiskon *float64       `gorm:"type:decimal(15,2)" json:"harga_sebelum_diskon"`
	PaymentLink        *string        `gorm:"type:text" json:"payment_link"`
	MetodePembayaran   *string        `gorm:"type:varchar(50)" json:"metode_pembayaran"`
	ExpiryDate         *time.Time     `gorm:"type:datetime" json:"expiry_date"`
	PaidAmount         *float64       `gorm:"type:decimal(15,2)" json:"paid_amount"`
	PaidAt             *time.Time     `gorm:"type:timestamp" json:"paid_at"`
	XenditID           *string        `gorm:"type:varchar(191);index" json:"xendit_id"`
	XenditExternalID   *string        `gorm:"type:varchar(191)" json:"xendit_external_id"`
	IsProcessing       bool           `gorm:"default:false" json:"is_processing"`
	XenditRetryCount   int64          `gorm:"default:0" json:"xendit_retry_count"`
	XenditLastRetry    *time.Time     `gorm:"type:datetime" json:"xendit_last_retry"`
	XenditErrorMessage *string        `gorm:"type:text" json:"xendit_error_message"`
	XenditStatus       string         `gorm:"type:varchar(50);default:'pending'" json:"xendit_status"`
	InvoiceType        string         `gorm:"type:varchar(50);default:'manual'" json:"invoice_type"`
	IsReinvoice        bool           `gorm:"default:false;index" json:"is_reinvoice"`
	OriginalInvoiceID  *uint64        `gorm:"index" json:"original_invoice_id"`
	ReinvoiceReason    *string        `gorm:"type:varchar(255)" json:"reinvoice_reason"`
	CreatedAt          *time.Time     `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt          *time.Time     `gorm:"type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Pelanggan *Pelanggan `gorm:"foreignKey:PelangganID" json:"-"`

	// Helper JSON-only fields (not persisted)
	PelangganNama string `gorm:"-" json:"pelanggan_nama"`
}

// TableName overrides the default table name for Invoice
func (Invoice) TableName() string {
	return "invoices"
}

// InvoiceArchive represents archived historical invoices.
type InvoiceArchive struct {
	ID                 uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	InvoiceNumber      string     `gorm:"type:varchar(191);uniqueIndex;not null" json:"invoice_number"`
	PelangganID        uint64     `gorm:"index;not null" json:"pelanggan_id"`
	IDPelanggan        string     `gorm:"type:varchar(255)" json:"id_pelanggan"`
	Brand              string     `gorm:"type:varchar(191)" json:"brand"`
	NoTelp             string     `gorm:"type:varchar(191)" json:"no_telp"`
	Email              string     `gorm:"type:varchar(191)" json:"email"`
	TotalHarga         float64    `gorm:"type:decimal(15,2);not null" json:"total_harga"`
	TglInvoice         time.Time  `gorm:"type:date;not null" json:"tgl_invoice"`
	TglJatuhTempo      time.Time  `gorm:"type:date;not null" json:"tgl_jatuh_tempo"`
	StatusInvoice      string     `gorm:"type:varchar(50);not null" json:"status_invoice"`
	PaymentLink        *string    `gorm:"type:text" json:"payment_link"`
	MetodePembayaran   *string    `gorm:"type:varchar(50)" json:"metode_pembayaran"`
	ExpiryDate         *time.Time `gorm:"type:datetime" json:"expiry_date"`
	PaidAmount         *float64   `gorm:"type:decimal(15,2)" json:"paid_amount"`
	PaidAt             *time.Time `gorm:"type:timestamp" json:"paid_at"`
	XenditID           *string    `gorm:"type:varchar(191)" json:"xendit_id"`
	XenditExternalID   *string    `gorm:"type:varchar(191)" json:"xendit_external_id"`
	IsProcessing       bool       `gorm:"default:false" json:"is_processing"`
	XenditRetryCount   int64      `gorm:"default:0" json:"xendit_retry_count"`
	XenditLastRetry    *time.Time `gorm:"type:datetime" json:"xendit_last_retry"`
	XenditErrorMessage *string    `gorm:"type:text" json:"xendit_error_message"`
	XenditStatus       string     `gorm:"type:varchar(50);default:'pending'" json:"xendit_status"`
	InvoiceType        string     `gorm:"type:varchar(50);default:'manual'" json:"invoice_type"`
	IsReinvoice        bool       `gorm:"default:false" json:"is_reinvoice"`
	OriginalInvoiceID  *uint64    `gorm:"index" json:"original_invoice_id"`
	ReinvoiceReason    *string    `gorm:"type:varchar(255)" json:"reinvoice_reason"`
	CreatedAt          *time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt          *time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`

	// Relationships
	Pelanggan *Pelanggan `gorm:"foreignKey:PelangganID" json:"-"`
}

// TableName overrides the default table name for InvoiceArchive
func (InvoiceArchive) TableName() string {
	return "invoices_archive"
}

// PaymentCallbackLog stores payment gateway callback transactions.
type PaymentCallbackLog struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	IdempotencyKey *string   `gorm:"type:varchar(255);index" json:"idempotency_key"`
	XenditID       string    `gorm:"type:varchar(255);index;not null" json:"xendit_id"`
	ExternalID     string    `gorm:"type:varchar(255);index;not null" json:"external_id"`
	CallbackData   *string   `gorm:"type:varchar(1000)" json:"callback_data"`
	Status         string    `gorm:"type:varchar(50);index;not null" json:"status"`
	ProcessedAt    time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"processed_at"`
	CreatedAt      time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName overrides the default table name for PaymentCallbackLog
func (PaymentCallbackLog) TableName() string {
	return "payment_callback_logs"
}

// LanggananCalculateRequest represents the parameters for calculating subscription pricing.
type LanggananCalculateRequest struct {
	PaketLayananID   uint64     `json:"paket_layanan_id"`
	MetodePembayaran string     `json:"metode_pembayaran"`
	PelangganID      uint64     `json:"pelanggan_id"`
	TglMulai         *time.Time `json:"tgl_mulai"`
}

// UnmarshalJSON custom unmarshaler for LanggananCalculateRequest to handle various date string formats.
func (r *LanggananCalculateRequest) UnmarshalJSON(data []byte) error {
	type Alias LanggananCalculateRequest
	aux := &struct {
		TglMulai interface{} `json:"tgl_mulai"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.TglMulai == nil {
		return nil
	}

	str, ok := aux.TglMulai.(string)
	if !ok {
		return fmt.Errorf("invalid type for tgl_mulai, expected string")
	}
	if str == "" {
		return nil
	}

	formats := []string{
		"2006-01-02",
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05.999Z07:00",
	}
	var parsed time.Time
	var err error
	for _, f := range formats {
		parsed, err = time.Parse(f, str)
		if err == nil {
			r.TglMulai = &parsed
			return nil
		}
	}
	return fmt.Errorf("cannot parse %q as date/time", str)
}

// LanggananCalculateResponse represents the result of calculating a standard subscription price.
type LanggananCalculateResponse struct {
	HargaAwal               float64    `json:"harga_awal"`
	TglJatuhTempo           time.Time  `json:"tgl_jatuh_tempo"`
	TglJatuhTempoPembayaran *time.Time `json:"tgl_jatuh_tempo_pembayaran"`
	TglMulaiLangganan       *time.Time `json:"tgl_mulai_langganan"`
}

// LanggananCalculateProratePlusFullResponse represents the result of calculating a combined prorate + full-month subscription price.
type LanggananCalculateProratePlusFullResponse struct {
	HargaProrate            float64    `json:"harga_prorate"`
	HargaNormal             float64    `json:"harga_normal"`
	HargaTotalAwal          float64    `json:"harga_total_awal"`
	TglJatuhTempo           time.Time  `json:"tgl_jatuh_tempo"`
	TglJatuhTempoPembayaran *time.Time `json:"tgl_jatuh_tempo_pembayaran"`
	TglMulaiLangganan       *time.Time `json:"tgl_mulai_langganan"`
}

// ProrateCalculationRequest represents general calculator parameters.
type ProrateCalculationRequest struct {
	PaketLayananID      uint64     `json:"paket_layanan_id"`
	IDBrand             string     `json:"id_brand"`
	TglMulai            *time.Time `json:"tgl_mulai"`
	IncludePpnNextMonth bool       `json:"include_ppn_next_month"`
}

// UnmarshalJSON custom unmarshaler for ProrateCalculationRequest to handle various date string formats.
func (r *ProrateCalculationRequest) UnmarshalJSON(data []byte) error {
	type Alias ProrateCalculationRequest
	aux := &struct {
		TglMulai interface{} `json:"tgl_mulai"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.TglMulai == nil {
		return nil
	}

	str, ok := aux.TglMulai.(string)
	if !ok {
		return fmt.Errorf("invalid type for tgl_mulai, expected string")
	}
	if str == "" {
		return nil
	}

	formats := []string{
		"2006-01-02",
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05.999Z07:00",
	}
	var parsed time.Time
	var err error
	for _, f := range formats {
		parsed, err = time.Parse(f, str)
		if err == nil {
			r.TglMulai = &parsed
			return nil
		}
	}
	return fmt.Errorf("cannot parse %q as date/time", str)
}

// ProrateCalculationResponse represents the calculator result for prorate.
type ProrateCalculationResponse struct {
	HargaDasarProrate        float64  `json:"harga_dasar_prorate"`
	Pajak                    float64  `json:"pajak"`
	TotalHargaProrate        float64  `json:"total_harga_prorate"`
	PeriodeHari              int      `json:"periode_hari"`
	HargaBulanDepan          *float64 `json:"harga_bulan_depan,omitempty"`
	PpnBulanDepan            *float64 `json:"ppn_bulan_depan,omitempty"`
	TotalBulanDepanDenganPpn *float64 `json:"total_bulan_depan_dengan_ppn,omitempty"`
	TotalKeseluruhan         *float64 `json:"total_keseluruhan,omitempty"`
}

// DiskonCalculationRequest represents discount calculator parameters.
type DiskonCalculationRequest struct {
	PaketLayananID   uint64  `json:"paket_layanan_id"`
	IDBrand          string  `json:"id_brand"`
	PersentaseDiskon float64 `json:"persentase_diskon"`
}

// DiskonCalculationResponse represents the result of discount calculator.
type DiskonCalculationResponse struct {
	NamaPaket             string  `json:"nama_paket"`
	NamaBrand             string  `json:"nama_brand"`
	HargaPaket            float64 `json:"harga_paket"`
	PajakPersen           float64 `json:"pajak_persen"`
	PajakAmount           float64 `json:"pajak_amount"`
	SubtotalSebelumDiskon float64 `json:"subtotal_sebelum_diskon"`
	PersentaseDiskon      float64 `json:"persentase_diskon"`
	DiskonAmount          float64 `json:"diskon_amount"`
	HargaFinal            float64 `json:"harga_final"`
	DetailPerhitungan     string  `json:"detail_perhitungan"`
}

// InvoiceSummaryStats represents invoice dashboard statistics.
type InvoiceSummaryStats struct {
	InvoiceTypes struct {
		Automatic int64 `json:"automatic"`
		Manual    int64 `json:"manual"`
	} `json:"invoice_types"`
	TotalReinvoice int64 `json:"total_reinvoice"`
}

// LanggananPelangganResponse represents customer info nested within LanggananResponse
type LanggananPelangganResponse struct {
	ID     uint64 `json:"id"`
	Nama   string `json:"nama"`
	NoTelp string `json:"no_telp"`
	Alamat string `json:"alamat"`
	Email  string `json:"email"`
}

// LanggananResponse represents a Langganan with denormalized fields from relationships
type LanggananResponse struct {
	ID                 uint64                      `json:"id"`
	PelangganID        uint64                      `json:"pelanggan_id"`
	NamaPelanggan      string                      `json:"nama_pelanggan"`
	NoTelp             string                      `json:"no_telp"`
	Alamat             string                      `json:"alamat"`
	IDBrand            string                      `json:"id_brand"`
	PaketLayananID     uint64                      `json:"paket_layanan_id"`
	NamaPaket          string                      `json:"nama_paket"`
	Brand              string                      `json:"brand"`
	Harga              float64                     `json:"harga"`
	HargaFinal         float64                     `json:"harga_final"`
	Status             string                      `json:"status"`
	TglJatuhTempo              *time.Time                  `json:"tgl_jatuh_tempo"`
	TglJatuhTempoPembayaran    *time.Time                  `json:"tgl_jatuh_tempo_pembayaran"`
	TglInvoiceTerakhir         *time.Time                  `json:"tgl_invoice_terakhir"`
	TglMulaiLangganan          *time.Time                  `json:"tgl_mulai_langganan"`
	TglBerhenti                *time.Time                  `json:"tgl_berhenti"`
	MetodePembayaran   string                      `json:"metode_pembayaran"`
	HargaAwal          *float64                    `json:"harga_awal"`
	AlasanBerhenti     *string                     `json:"alasan_berhenti"`
	StatusModem        *string                     `json:"status_modem"`
	WhatsappStatus     *string                     `json:"whatsapp_status"`
	LastWhatsappSent   *time.Time                  `json:"last_whatsapp_sent"`
	CreatedAt          *time.Time                  `json:"created_at"`
	UpdatedAt          *time.Time                  `json:"updated_at"`
	Pelanggan          *LanggananPelangganResponse `json:"pelanggan"`
}
