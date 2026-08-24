package usecase

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/csv"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"billing-backend/internal/domain"
	"billing-backend/internal/websocket"
	"billing-backend/pkg/database"
	"billing-backend/pkg/utils"

	"github.com/xuri/excelize/v2"
)

type pelangganUsecase struct {
	pelangganRepo domain.PelangganRepository
	systemRepo    domain.SystemRepository
}

func NewPelangganUsecase(p domain.PelangganRepository, sr ...domain.SystemRepository) domain.PelangganUsecase {
	var systemRepo domain.SystemRepository
	if len(sr) > 0 {
		systemRepo = sr[0]
	}
	return &pelangganUsecase{
		pelangganRepo: p,
		systemRepo:    systemRepo,
	}
}

func (u *pelangganUsecase) logActivity(ctx context.Context, action string, details string) {
	if u.systemRepo == nil {
		return
	}
	log := &domain.ActivityLog{
		UserID:    utils.GetUserIDFromCtx(ctx),
		Action:    action,
		Details:   &details,
		Timestamp: time.Now(),
	}
	_ = u.systemRepo.CreateActivityLog(ctx, log)
}

func (u *pelangganUsecase) FetchAll(ctx context.Context, skip, limit int, filters domain.PelangganFilterParams) ([]domain.Pelanggan, int64, error) {
	pelanggans, total, err := u.pelangganRepo.GetAll(ctx, limit, skip, filters)
	if err == nil {
		for i := range pelanggans {
			pelanggans[i].NoKtp = utils.Decrypt(pelanggans[i].NoKtp)
		}
	}
	return pelanggans, total, err
}


func (u *pelangganUsecase) GetByID(ctx context.Context, id uint64) (*domain.Pelanggan, error) {
	pelanggan, err := u.pelangganRepo.GetByID(ctx, id)
	if err == nil && pelanggan != nil {
		pelanggan.NoKtp = utils.Decrypt(pelanggan.NoKtp)
	}
	return pelanggan, err
}

func isDummyKtp(ktp string) bool {
	if ktp == "" { return true }
	for _, char := range ktp {
		if char != '0' { return false }
	}
	return true
}

func generateRandomCustomerID() string {
	// First digit: 1-9
	nBig, err := rand.Int(rand.Reader, big.NewInt(9))
	firstDigit := int64(1)
	if err == nil {
		firstDigit = nBig.Int64() + 1
	}
	// Remaining 10 digits: 0000000000 to 9999999999
	nBig2, err2 := rand.Int(rand.Reader, big.NewInt(10000000000))
	rem := int64(0)
	if err2 == nil {
		rem = nBig2.Int64()
	}
	return fmt.Sprintf("%d%010d", firstDigit, rem)
}

func (u *pelangganUsecase) generateUniqueCustomerID(ctx context.Context) (string, error) {
	for i := 0; i < 10; i++ {
		cid := generateRandomCustomerID()
		existing, err := u.pelangganRepo.GetByCustomerID(ctx, cid)
		if err == nil && existing == nil {
			return cid, nil
		}
	}
	return "", errors.New("failed to generate unique customer id after 10 attempts")
}

func (u *pelangganUsecase) Store(ctx context.Context, pelanggan *domain.Pelanggan) error {
	if pelanggan.Email == "" { return errors.New("email is required") }
	existingEmail, err := u.pelangganRepo.GetByEmail(ctx, pelanggan.Email)
	if err == nil && existingEmail != nil { return errors.New("Email sudah terdaftar") }

	if pelanggan.NoTelp != "" {
		existingPhone, err := u.pelangganRepo.GetByNoTelp(ctx, pelanggan.NoTelp)
		if err == nil && existingPhone != nil { return errors.New("No. Telepon sudah terdaftar") }
	}

	// Auto-generate unique 11-digit CustomerID if not set
	if pelanggan.CustomerID == nil || *pelanggan.CustomerID == "" {
		cid, err := u.generateUniqueCustomerID(ctx)
		if err != nil {
			return fmt.Errorf("gagal membuat ID pelanggan unik: %w", err)
		}
		pelanggan.CustomerID = &cid
	}

	// Encrypt NIK if it's not already encrypted and not dummy
	if pelanggan.NoKtp != "" && !utils.GlobalEncryptionService.IsEncrypted(pelanggan.NoKtp) {
		pelanggan.NoKtp = utils.Encrypt(pelanggan.NoKtp)
	}

	if !isDummyKtp(pelanggan.NoKtp) {
		// Decrypt temporarily for uniqueness check if necessary
	}
	if err := u.pelangganRepo.Create(ctx, pelanggan); err != nil { return err }
	cidStr := ""
	if pelanggan.CustomerID != nil {
		cidStr = *pelanggan.CustomerID
	}
	u.logActivity(ctx, "Create Pelanggan", fmt.Sprintf("Created pelanggan: %s (ID: %d, CustomerID: %s)", pelanggan.Nama, pelanggan.ID, cidStr))
	if websocket.GlobalHub != nil {
		websocket.GlobalHub.BroadcastNotification("new_customer", map[string]interface{}{"pelanggan_nama": pelanggan.Nama})
	}
	websocket.InvalidateDashboardCache(ctx)
	return nil
}

func (u *pelangganUsecase) Update(ctx context.Context, id uint64, req *domain.Pelanggan) error {
	existing, err := u.pelangganRepo.GetByID(ctx, id)
	if err != nil { return err }
	if existing == nil { return errors.New("pelanggan not found") }

	if req.Email != "" {
		dupEmail, err := u.pelangganRepo.GetByEmail(ctx, req.Email)
		if err == nil && dupEmail != nil && dupEmail.ID != id { return errors.New("Email sudah terdaftar oleh pelanggan lain") }
	}

	if req.NoTelp != "" {
		dupPhone, err := u.pelangganRepo.GetByNoTelp(ctx, req.NoTelp)
		if err == nil && dupPhone != nil && dupPhone.ID != id { return errors.New("No. Telepon sudah terdaftar oleh pelanggan lain") }
	}
	
	// Encrypt NIK if updated
	if req.NoKtp != "" && !utils.GlobalEncryptionService.IsEncrypted(req.NoKtp) {
		req.NoKtp = utils.Encrypt(req.NoKtp)
	}

	existing.Nama = req.Nama
	existing.NoKtp = req.NoKtp
	existing.Alamat = req.Alamat
	existing.AlamatCustom = req.AlamatCustom
	existing.TglInstalasi = req.TglInstalasi
	existing.Blok = req.Blok
	existing.Unit = req.Unit
	existing.NoTelp = req.NoTelp
	existing.Email = req.Email
	existing.IDBrand = req.IDBrand
	existing.Layanan = req.Layanan
	existing.BrandDefault = req.BrandDefault
	existing.MikrotikServerID = req.MikrotikServerID
	err = u.pelangganRepo.Update(ctx, existing)
	if err == nil {
		u.logActivity(ctx, "Update Pelanggan", fmt.Sprintf("Updated pelanggan: %s (ID: %d)", existing.Nama, id))
		websocket.InvalidateDashboardCache(ctx)
	}
	return err
}

func (u *pelangganUsecase) Delete(ctx context.Context, id uint64) error {
	pelanggan, err := u.pelangganRepo.GetByID(ctx, id)
	if err != nil { return err }
	if pelanggan == nil { return errors.New("pelanggan not found") }
	err = u.pelangganRepo.Delete(ctx, id)
	if err == nil {
		u.logActivity(ctx, "Delete Pelanggan", fmt.Sprintf("Deleted pelanggan: %s (ID: %d)", pelanggan.Nama, id))
		websocket.InvalidateDashboardCache(ctx)
	}
	return err
}

func (u *pelangganUsecase) GetUniqueLocations(ctx context.Context) ([]string, error) {
	return u.pelangganRepo.GetUniqueLocations(ctx)
}

func (u *pelangganUsecase) ImportFromCSV(ctx context.Context, csvContent string) (int, error) {
	rows, err := utils.ParseCSV(csvContent)
	if err != nil || len(rows) < 2 {
		return 0, errors.New("invalid csv")
	}
	header := rows[0]
	colMap := make(map[string]int)
	for i, name := range header {
		colMap[utils.NormalizeCSVHeader(name)] = i
		colMap[strings.ToLower(strings.TrimSpace(name))] = i
	}
	successCount := 0
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 {
			continue
		}
		getV := func(k string) string {
			norm := utils.NormalizeCSVHeader(k)
			if idx, ok := colMap[norm]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			if idx, ok := colMap[strings.ToLower(strings.TrimSpace(k))]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}
		nama, email := getV("nama"), getV("email")
		if nama == "" || email == "" { continue }
		if ex, _ := u.pelangganRepo.GetByEmail(ctx, email); ex != nil { continue }

		noTelp := getV("no telp")
		if noTelp != "" {
			if exPhone, _ := u.pelangganRepo.GetByNoTelp(ctx, noTelp); exPhone != nil { continue }
		}
		
		nik := getV("no ktp")
		if nik != "" && !utils.GlobalEncryptionService.IsEncrypted(nik) {
			nik = utils.Encrypt(nik)
		}

		p := &domain.Pelanggan{Nama: nama, Email: email, NoKtp: nik, Alamat: getV("alamat"), AlamatCustom: &[]string{getV("alamat tambahan")}[0], Blok: getV("blok"), Unit: getV("unit"), NoTelp: getV("no telp")}
		if *p.AlamatCustom == "" { p.AlamatCustom = nil }
		lay := getV("layanan"); if lay != "" { p.Layanan = &lay }
		brand := getV("id brand"); if brand != "" { p.IDBrand = &brand }
		tglStr := getV("tgl instalasi"); if tglStr != "" {
			if t, err := time.Parse("2006-01-02", tglStr); err == nil { p.TglInstalasi = &t }
		}
		if err := u.pelangganRepo.Create(ctx, p); err == nil { successCount++ }
	}
	if successCount > 0 {
		u.logActivity(ctx, "Import Pelanggan", fmt.Sprintf("Imported %d pelanggan from CSV", successCount))
		websocket.InvalidateDashboardCache(ctx)
	}
	return successCount, nil
}

func (u *pelangganUsecase) Export(ctx context.Context, format string) ([]byte, string, error) {
	headers := []string{"ID", "No KTP", "Nama", "Alamat", "Alamat Tambahan", "Blok", "Unit", "No Telp", "Email", "Layanan", "Brand", "Tgl Instalasi"}
	limit := 1000
	offset := 0

	brandMap := make(map[string]string)
	db := database.GetDB()
	if db != nil {
		var brands []domain.HargaLayanan
		if err := db.WithContext(ctx).Find(&brands).Error; err == nil {
			for _, b := range brands {
				if b.IDBrand != "" && b.Brand != "" {
					brandMap[strings.ToLower(b.IDBrand)] = b.Brand
					brandMap[strings.ToLower(b.Brand)] = b.Brand
				}
			}
		}
	}

	getBrandName := func(raw *string) string {
		if raw == nil || *raw == "" {
			return ""
		}
		if name, ok := brandMap[strings.ToLower(*raw)]; ok {
			return name
		}
		return *raw
	}

	if format == "excel" {
		f := excelize.NewFile()
		sheet := "Pelanggan"
		f.SetSheetName("Sheet1", sheet)
		for i, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(sheet, cell, h)
		}

		row := 2
		for {
			pelanggans, _, err := u.pelangganRepo.GetAll(ctx, limit, offset, domain.PelangganFilterParams{})
			if err != nil {
				return nil, "", err
			}
			if len(pelanggans) == 0 {
				break
			}

			for _, p := range pelanggans {
				noKtpDec := utils.Decrypt(p.NoKtp)
				tgl, lay, al2 := "", "", ""
				if p.TglInstalasi != nil {
					tgl = p.TglInstalasi.Format("2006-01-02")
				}
				brand := getBrandName(p.IDBrand)
				if p.Layanan != nil {
					lay = *p.Layanan
				}
				if p.AlamatCustom != nil {
					al2 = *p.AlamatCustom
				}
				vals := []interface{}{p.ID, noKtpDec, p.Nama, p.Alamat, al2, p.Blok, p.Unit, p.NoTelp, p.Email, lay, brand, tgl}
				for c, v := range vals {
					cell, _ := excelize.CoordinatesToCellName(c+1, row)
					f.SetCellValue(sheet, cell, v)
				}
				row++
			}

			offset += limit
			if len(pelanggans) < limit {
				break
			}
		}

		buf, _ := f.WriteToBuffer()
		return buf.Bytes(), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", nil
	} else {
		buf := new(bytes.Buffer)
		w := csv.NewWriter(buf)
		w.Comma = ';'
		w.Write(headers)

		for {
			pelanggans, _, err := u.pelangganRepo.GetAll(ctx, limit, offset, domain.PelangganFilterParams{})
			if err != nil {
				return nil, "", err
			}
			if len(pelanggans) == 0 {
				break
			}

			for _, p := range pelanggans {
				noKtpDec := utils.Decrypt(p.NoKtp)
				tgl, lay, al2 := "", "", ""
				if p.TglInstalasi != nil {
					tgl = p.TglInstalasi.Format("2006-01-02")
				}
				brand := getBrandName(p.IDBrand)
				if p.Layanan != nil {
					lay = *p.Layanan
				}
				if p.AlamatCustom != nil {
					al2 = *p.AlamatCustom
				}
				w.Write([]string{
					fmt.Sprintf("%d", p.ID),
					noKtpDec,
					p.Nama,
					p.Alamat,
					al2,
					p.Blok,
					p.Unit,
					p.NoTelp,
					p.Email,
					lay,
					brand,
					tgl,
				})
			}

			offset += limit
			if len(pelanggans) < limit {
				break
			}
		}

		w.Flush()
		return buf.Bytes(), "text/csv", nil
	}
}

func (u *pelangganUsecase) BackfillCustomerIDs(ctx context.Context) error {
	list, err := u.pelangganRepo.GetWithoutCustomerID(ctx)
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return nil
	}
	count := 0
	for i := range list {
		p := &list[i]
		cid, err := u.generateUniqueCustomerID(ctx)
		if err != nil {
			continue
		}
		p.CustomerID = &cid
		if err := u.pelangganRepo.Update(ctx, p); err == nil {
			count++
		}
	}
	if count > 0 {
		u.logActivity(ctx, "Backfill Customer IDs", fmt.Sprintf("Berhasil meng-generate Customer ID untuk %d pelanggan lama", count))
	}
	return nil
}
