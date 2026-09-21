package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"billing-backend/config"
	"billing-backend/internal/domain"
)

type AIService interface {
	GenerateReply(ctx context.Context, room *domain.ChatRoom, history []domain.ChatMessage, userMessage string) (string, error)
	IsHumanHandoverRequested(message string) bool
}

type aiService struct {
	cfg        *config.Config
	paketRepo  domain.PaketLayananRepository
	httpClient *http.Client
}

func NewAIService(cfg *config.Config, paketRepo domain.PaketLayananRepository) AIService {
	return &aiService{
		cfg:       cfg,
		paketRepo: paketRepo,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

// IsHumanHandoverRequested memeriksa apakah pelanggan meminta berbicara dengan CS/agen manusia
func (s *aiService) IsHumanHandoverRequested(message string) bool {
	cleaned := strings.ToLower(strings.TrimSpace(message))

	exactMatches := []string{
		"2", "cs", "agent", "admin", "operator", "manusia", "orang",
	}
	for _, match := range exactMatches {
		if cleaned == match {
			return true
		}
	}

	phraseMatches := []string{
		"hubungkan ke cs", "hubungkan cs", "sambungkan ke cs", "sambungkan cs",
		"bicara dengan cs", "bicara dengan admin", "bicara dengan orang", "bicara dengan manusia",
		"mau cs", "minta cs", "panggil cs", "operator manusia", "komplain ke cs",
		"mau orang", "bukan bot", "jangan bot", "customer support",
	}
	for _, phrase := range phraseMatches {
		if strings.Contains(cleaned, phrase) {
			return true
		}
	}

	return false
}

// GenerateReply memanggil 9Router OpenAI-compatible endpoint dengan SOP CS ketat & data pelanggan
func (s *aiService) GenerateReply(ctx context.Context, room *domain.ChatRoom, history []domain.ChatMessage, userMessage string) (string, error) {
	if s.cfg == nil || !s.cfg.AIRouterEnabled || s.cfg.AIRouterURL == "" || s.cfg.AIRouterKey == "" {
		return "", fmt.Errorf("ai router tidak aktif atau konfigurasi belum lengkap")
	}

	brand := "Jakinet"
	if room != nil && room.Brand != "" {
		brand = room.Brand
	}

	// 1. Buat System Prompt terstruktur sesuai SOP Customer Service Resmi
	systemPrompt := s.buildSystemPrompt(ctx, room, brand)

	// 2. Siapkan riwayat percakapan (maksimal 6 pesan terakhir untuk efisiensi konteks)
	var messages []map[string]string
	messages = append(messages, map[string]string{
		"role":    "system",
		"content": systemPrompt,
	})

	// Tambahkan history percakapan
	startIndex := 0
	if len(history) > 6 {
		startIndex = len(history) - 6
	}
	for i := startIndex; i < len(history); i++ {
		msg := history[i]
		role := "user"
		if msg.SenderType == "system" || msg.SenderType == "admin" {
			role = "assistant"
		}
		if strings.TrimSpace(msg.Message) != "" {
			messages = append(messages, map[string]string{
				"role":    role,
				"content": msg.Message,
			})
		}
	}

	// Tambahkan pesan user terkini jika belum ada di history
	messages = append(messages, map[string]string{
		"role":    "user",
		"content": userMessage,
	})

	// 3. Payload Request ke 9Router
	reqBody := map[string]interface{}{
		"model":       s.cfg.AIRouterModel,
		"messages":    messages,
		"max_tokens":  220, // Membatasi output agar tidak bertele-tele dan to the point
		"temperature": 0.1, // Rendah agar deterministik, anti-halusinasi, dan konsisten
		"stream":      false,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	endpoint := strings.TrimRight(s.cfg.AIRouterURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.cfg.AIRouterKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("gagal menghubungi 9Router: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("gagal membaca respons 9Router: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[AIService] Error from 9Router HTTP %d: %s", resp.StatusCode, string(bodyBytes))
		return "", fmt.Errorf("9Router error HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return "", fmt.Errorf("gagal parsing JSON 9Router: %w", err)
	}

	if len(chatResp.Choices) == 0 || strings.TrimSpace(chatResp.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("balasan AI kosong")
	}

	replyContent := strings.TrimSpace(chatResp.Choices[0].Message.Content)

	// 4. Sematkan Penanda Resmi AI di bagian bawah pesan
	footer := fmt.Sprintf("\n\n---\n🤖 Dibalas otomatis oleh Asisten Virtual AI %s", brand)
	if !strings.Contains(replyContent, "Asisten Virtual AI") {
		replyContent += footer
	}

	return replyContent, nil
}

func formatRupiah(amount float64) string {
	intAmount := int64(amount)
	str := fmt.Sprintf("%d", intAmount)
	n := len(str)
	if n <= 3 {
		return "Rp " + str
	}
	var res []string
	rem := n % 3
	if rem > 0 {
		res = append(res, str[:rem])
	}
	for i := rem; i < n; i += 3 {
		res = append(res, str[i:i+3])
	}
	return "Rp " + strings.Join(res, ".")
}

// buildSystemPrompt menyusun prompt to the point, anti-halusinasi, anti-bias, dan sopan
func (s *aiService) buildSystemPrompt(ctx context.Context, room *domain.ChatRoom, brand string) string {
	customerName := "Pelanggan"
	packageName := "Internet FTTH"
	statusLangganan := "Aktif"

	if room != nil && room.Pelanggan != nil {
		if room.Pelanggan.Nama != "" {
			customerName = room.Pelanggan.Nama
		}
		if room.Pelanggan.Layanan != nil && *room.Pelanggan.Layanan != "" {
			packageName = *room.Pelanggan.Layanan
		}
		if len(room.Pelanggan.Langganan) > 0 {
			latestSub := room.Pelanggan.Langganan[0]
			if latestSub.Status != "" {
				statusLangganan = latestSub.Status
			}
			if latestSub.PaketLayanan != nil && latestSub.PaketLayanan.NamaPaket != "" {
				packageName = latestSub.PaketLayanan.NamaPaket
			}
		}
	}

	// Ambil daftar paket layanan resmi dari database sesuai brand
	var paketText string
	if s.paketRepo != nil {
		pakets, err := s.paketRepo.GetAll(ctx)
		if err == nil && len(pakets) > 0 {
			var lines []string
			for _, p := range pakets {
				if strings.EqualFold(p.IDBrand, brand) || p.IDBrand == "" {
					lines = append(lines, fmt.Sprintf("     * %s (%d Mbps): %s/bln", p.NamaPaket, p.Kecepatan, formatRupiah(p.Harga)))
				}
			}
			if len(lines) > 0 {
				paketText = strings.Join(lines, "\n")
			}
		}
	}

	if paketText == "" {
		paketText = "     * (Informasi paket dan harga resmi dapat dikonfirmasi langsung dengan CS kami)"
	}

	return fmt.Sprintf(`Anda adalah Asisten Virtual Customer Service resmi penyedia internet rumah (FTTH ISP) %[1]s.
Melayani pelanggan: Kak %[2]s (Paket: %[3]s, Status: %[4]s).

PRINSIP UTAMA: TO THE POINT, PADAT, ANTI-BIAS, ANTI-HALUSINASI, DAN TANPA BASA-BASI.

ATURAN STRUKTUR JAWABAN (WAJIB DIIKUTI):
1. Sapaan awal: Maksimal 1 kalimat singkat (Contoh: "Halo Kak %[2]s, berikut langkah penanganannya:").
2. Solusi teknis: Langsung ke poin bernomor (1, 2, 3), maksimal 1 kalimat per poin.
3. Penutup: 1 kalimat singkat mengarahkan ke CS jika kendala belum selesai.
4. Total panjang: Maksimal 3 hingga 4 kalimat padat. DILARANG membuat paragraf panjang, berulang-ulang, atau bertele-tele.
5. DILARANG KERAS:
   - DILARANG menulis basa-basi panjang (seperti "Terima kasih telah menghubungi...", "Kami memahami betapa pentingnya...").
   - DILARANG mengarang harga atau paket baru yang tidak ada pada daftar paket resmi di bawah.
   - DILARANG menjanjikan waktu kedatangan teknisi fiktif (misal: "teknisi akan datang 15 menit lagi").
   - DILARANG menjawab pertanyaan di luar layanan internet %[1]s (politik, resep, jokes, topik umum). Jika ada, tolak sopan dalam 1 kalimat: "Mohon maaf Kak %[2]s, saya hanya dapat membantu seputar kendala dan layanan internet %[1]s. Ada yang bisa dibantu terkait koneksi Anda?"

BASIS PENGETAHUAN RESMI:
- WIFI LAMBAT: (1) Cabut adaptor listrik modem selama 30 detik lalu colokkan kembali, (2) Sambungkan ke frekuensi WiFi 5 GHz jika tersedia, (3) Uji kecepatan di speedtest.net dekat modem.
- LAMPU LOS MERAH: Modem tidak menerima sinyal optik tiang. (1) Pastikan kabel optik kecil kuning di belakang modem tertancap rapat dan tidak tertekuk tajam/terjepit, (2) Jangan cabut paksa kabel atau tusuk lubang reset. Jika kabel fisik aman namun tetap merah, arahkan ke CS untuk tiket kunjungan teknisi.
- LARANGAN RESET: Dilarang menusuk lubang reset modem karena menghapus konfigurasi akun internet (PPPoE). Cukup restart via cabut colokan listrik.
- GANTI PASSWORD WIFI: Buka browser ke 192.168.1.1 (atau 192.168.100.1), login dengan username & password di stiker bawah modem, masuk menu Network/WLAN > WPA Passphrase.
- INFORMASI TAGIHAN & PEMBAYARAN:
  1. Rincian tagihan dapat dicek langsung di menu Tagihan pada aplikasi.
  2. Invoice terbit tgl 1 setiap bulan, batas jatuh tempo tgl 10 atau 20.
  3. Pembayaran via Virtual Account (BCA, Mandiri, BRI, BNI) atau scan QRIS resmi di aplikasi.
  4. Konfirmasi sudah bayar tapi masih isolir: Sistem membuka isolir otomatis dalam 1-5 menit. Silakan cabut adaptor modem 30 detik lalu colokkan lagi agar modem mengambil IP baru.
  5. Jika bayar transfer manual atau belum aktif >10 menit, kirim bukti transfer di sini untuk diverifikasi admin keuangan.
- PILIHAN PAKET RESMI %[1]s:
%[5]s
- JAM OPERASIONAL: Chat CS 08:00 - 22:00 WIB, Kunjungan teknisi 08:30 - 17:00 WIB.

CONTOH FORMAT JAWABAN (IKUTI POLA INI):
User: "wifi saya lemot"
Assistant: Halo Kak %[2]s, berikut langkah cepat mengatasi WiFi lambat:
1. Cabut adaptor listrik modem selama 30 detik lalu colokkan kembali.
2. Sambungkan perangkat ke frekuensi WiFi 5 GHz jika tersedia.
3. Uji kembali kecepatan di dekat modem melalui speedtest.net.
Jika kendala berlanjut, silakan ketik 'CS' atau tekan [Hubungkan ke CS Manusia].

User: "lampu los kedip merah"
Assistant: Halo Kak %[2]s, lampu LOS merah menandakan modem tidak menerima sinyal optik:
1. Pastikan kabel optik kecil di belakang modem tertancap rapat dan tidak tertekuk tajam.
2. Jangan mencabut paksa kabel atau menusuk lubang reset modem.
Jika kabel fisik sudah aman namun tetap merah, silakan ketik 'CS' atau tekan [Hubungkan ke CS Manusia] untuk tiket penanganan teknisi.

User: "bagaimana cara bayar tagihan?"
Assistant: Halo Kak %[2]s, berikut cara pembayaran tagihan %[1]s:
1. Buka menu Tagihan pada aplikasi ini untuk melihat rincian tagihan.
2. Lakukan pembayaran melalui Virtual Account atau scan QRIS yang tersedia.
Jika butuh bantuan lebih lanjut, silakan ketik 'CS' atau tekan [Hubungkan ke CS Manusia].

User: "sudah bayar tapi masih isolir"
Assistant: Halo Kak %[2]s, berikut langkah jika internet masih terisolir setelah bayar:
1. Tunggu 1-5 menit karena sistem membuka isolir otomatis setelah pembayaran via VA/QRIS.
2. Cabut adaptor listrik modem selama 30 detik lalu colokkan kembali agar mendapat IP baru.
Jika masih belum aktif, silakan ketik 'CS' atau tekan [Hubungkan ke CS Manusia].
`, brand, customerName, packageName, statusLangganan, paketText)
}
