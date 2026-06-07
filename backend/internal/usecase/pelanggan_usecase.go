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

	"github.com/xuri/excelize/v2"
)

type pelangganUsecase struct {
	pelangganRepo domain.PelangganRepository
}

func NewPelangganUsecase(p domain.PelangganRepository) domain.PelangganUsecase {
	return &pelangganUsecase{
		pelangganRepo: p,
	}
}

func (u *pelangganUsecase) FetchAll(ctx context.Context, page, pageSize int, connectionStatus string) ([]domain.Pelanggan, int64, error) {
	if page <= 0 { page = 1 }
	if pageSize <= 0 { pageSize = 10 }
	offset := (page - 1) * pageSize
	return u.pelangganRepo.GetAll(ctx, pageSize, offset, connectionStatus)
}

func (u *pelangganUsecase) GetByID(ctx context.Context, id uint64) (*domain.Pelanggan, error) {
	return u.pelangganRepo.GetByID(ctx, id)
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
	if !isDummyKtp(pelanggan.NoKtp) {
		existing, err := u.pelangganRepo.GetByNoKtp(ctx, pelanggan.NoKtp)
		if err == nil && existing != nil { return errors.New("NIK/No KTP sudah terdaftar") }
	}
	if err := u.pelangganRepo.Create(ctx, pelanggan); err != nil { return err }
	if websocket.GlobalHub != nil {
		websocket.GlobalHub.BroadcastNotification("new_customer", map[string]interface{}{"pelanggan_nama": pelanggan.Nama})
	}
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
	if !isDummyKtp(req.NoKtp) {
		dup, err := u.pelangganRepo.GetByNoKtp(ctx, req.NoKtp)
		if err == nil && dup != nil && dup.ID != id { return errors.New("NIK/No KTP sudah terdaftar oleh pelanggan lain") }
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
	return u.pelangganRepo.Update(ctx, existing)
}

func (u *pelangganUsecase) Delete(ctx context.Context, id uint64) error {
	_, err := u.pelangganRepo.GetByID(ctx, id)
	if err != nil { return err }
	return u.pelangganRepo.Delete(ctx, id)
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
		p := &domain.Pelanggan{Nama: nama, Email: email, NoKtp: getV("no ktp"), Alamat: getV("alamat"), Blok: getV("blok"), Unit: getV("unit"), NoTelp: getV("no telp")}
		lay := getV("layanan"); if lay != "" { p.Layanan = &lay }
		brand := getV("id brand"); if brand != "" { p.IDBrand = &brand }
		tglStr := getV("tgl instalasi"); if tglStr != "" {
			if t, err := time.Parse("2006-01-02", tglStr); err == nil { p.TglInstalasi = &t }
		}
		if err := u.pelangganRepo.Create(ctx, p); err == nil { successCount++ }
	}
	return successCount, nil
}

func (u *pelangganUsecase) Export(ctx context.Context, format string) ([]byte, string, error) {
	pelanggans, _, err := u.pelangganRepo.GetAll(ctx, 10000, 0, "")
	if err != nil { return nil, "", err }
	headers := []string{"ID", "No KTP", "Nama", "Alamat", "Blok", "Unit", "No Telp", "Email", "Layanan", "ID Brand", "Tgl Instalasi"}
	if format == "excel" {
		f := excelize.NewFile()
		sheet := "Pelanggan"
		f.SetSheetName("Sheet1", sheet)
		for i, h := range headers { cell, _ := excelize.CoordinatesToCellName(i+1, 1); f.SetCellValue(sheet, cell, h) }
		for r, p := range pelanggans {
			row := r + 2
			tgl, brand, lay := "", "", ""
			if p.TglInstalasi != nil { tgl = p.TglInstalasi.Format("2006-01-02") }
			if p.IDBrand != nil { brand = *p.IDBrand }
			if p.Layanan != nil { lay = *p.Layanan }
			vals := []interface{}{p.ID, p.NoKtp, p.Nama, p.Alamat, p.Blok, p.Unit, p.NoTelp, p.Email, lay, brand, tgl}
			for c, v := range vals { cell, _ := excelize.CoordinatesToCellName(c+1, row); f.SetCellValue(sheet, cell, v) }
		}
		buf, _ := f.WriteToBuffer()
		return buf.Bytes(), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", nil
	} else {
		buf := new(bytes.Buffer)
		w := csv.NewWriter(buf)
		w.Comma = ';'
		w.Write(headers)
		for _, p := range pelanggans {
			tgl, brand, lay := "", "", ""
			if p.TglInstalasi != nil { tgl = p.TglInstalasi.Format("2006-01-02") }
			if p.IDBrand != nil { brand = *p.IDBrand }
			if p.Layanan != nil { lay = *p.Layanan }
			w.Write([]string{fmt.Sprintf("%d", p.ID), p.NoKtp, p.Nama, p.Alamat, p.Blok, p.Unit, p.NoTelp, p.Email, lay, brand, tgl})
		}
		w.Flush()
		return buf.Bytes(), "text/csv", nil
	}
}
