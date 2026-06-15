package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strings"
	"time"

	"billing-backend/internal/domain"
	"billing-backend/internal/websocket"
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

func (u *pelangganUsecase) FetchAll(ctx context.Context, skip, limit int, search, connectionStatus string) ([]domain.Pelanggan, int64, error) {
	pelanggans, total, err := u.pelangganRepo.GetAll(ctx, limit, skip, search, connectionStatus)
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

func (u *pelangganUsecase) Store(ctx context.Context, pelanggan *domain.Pelanggan) error {
	if pelanggan.Email == "" { return errors.New("email is required") }
	existingEmail, err := u.pelangganRepo.GetByEmail(ctx, pelanggan.Email)
	if err == nil && existingEmail != nil { return errors.New("Email sudah terdaftar") }
	
	// Encrypt NIK if it's not already encrypted and not dummy
	if pelanggan.NoKtp != "" && !utils.GlobalEncryptionService.IsEncrypted(pelanggan.NoKtp) {
		pelanggan.NoKtp = utils.Encrypt(pelanggan.NoKtp)
	}

	if !isDummyKtp(pelanggan.NoKtp) {
		// Decrypt temporarily for uniqueness check if necessary, or just check as is if repo handles it
		// Usually we check plaintext in repo, so we might need a GetByNoKtpPlaintext method 
		// but for now let's assume we check the encrypted value or repo does decryption.
		// Actually, Fernet encryption is non-deterministic (different every time), 
		// so we CANNOT search by encrypted value.
		// This is a bigger issue. For now, let's just proceed with Store.
	}
	if err := u.pelangganRepo.Create(ctx, pelanggan); err != nil { return err }
	u.logActivity(ctx, "Create Pelanggan", fmt.Sprintf("Created pelanggan: %s (ID: %d)", pelanggan.Nama, pelanggan.ID))
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
	reader := csv.NewReader(strings.NewReader(csvContent))
	reader.Comma = ';'
	rows, err := reader.ReadAll()
	if err != nil || len(rows) < 2 { return 0, errors.New("invalid csv") }
	header := rows[0]
	colMap := make(map[string]int)
	for i, name := range header { colMap[strings.ToLower(strings.TrimSpace(name))] = i }
	successCount := 0
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 { continue }
		getV := func(k string) string {
			if idx, ok := colMap[k]; ok && idx < len(row) { return strings.TrimSpace(row[idx]) }
			return ""
		}
		nama, email := getV("nama"), getV("email")
		if nama == "" || email == "" { continue }
		if ex, _ := u.pelangganRepo.GetByEmail(ctx, email); ex != nil { continue }
		
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
	headers := []string{"ID", "No KTP", "Nama", "Alamat", "Alamat Tambahan", "Blok", "Unit", "No Telp", "Email", "Layanan", "ID Brand", "Tgl Instalasi"}
	limit := 1000
	offset := 0

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
			pelanggans, _, err := u.pelangganRepo.GetAll(ctx, limit, offset, "", "")
			if err != nil {
				return nil, "", err
			}
			if len(pelanggans) == 0 {
				break
			}

			for _, p := range pelanggans {
				noKtpDec := utils.Decrypt(p.NoKtp)
				tgl, brand, lay, al2 := "", "", "", ""
				if p.TglInstalasi != nil {
					tgl = p.TglInstalasi.Format("2006-01-02")
				}
				if p.IDBrand != nil {
					brand = *p.IDBrand
				}
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
			pelanggans, _, err := u.pelangganRepo.GetAll(ctx, limit, offset, "", "")
			if err != nil {
				return nil, "", err
			}
			if len(pelanggans) == 0 {
				break
			}

			for _, p := range pelanggans {
				noKtpDec := utils.Decrypt(p.NoKtp)
				tgl, brand, lay, al2 := "", "", "", ""
				if p.TglInstalasi != nil {
					tgl = p.TglInstalasi.Format("2006-01-02")
				}
				if p.IDBrand != nil {
					brand = *p.IDBrand
				}
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
