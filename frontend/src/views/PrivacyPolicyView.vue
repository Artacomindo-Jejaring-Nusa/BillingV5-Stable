<template>
  <div class="privacy-policy-page">
    <!-- Top Navbar -->
    <v-app-bar flat class="border-b px-4 bg-white" height="64">
      <div class="d-flex align-center cursor-pointer" @click="goHome">
        <v-img
          src="/portal_logo.svg"
          alt="Portal Ajnusa Logo"
          width="36"
          height="36"
          class="me-3"
          @error="handleLogoError"
        />
        <div>
          <div class="text-subtitle-1 font-weight-bold text-primary leading-tight">
            Portal Ajnusa
          </div>
          <div class="text-caption text-medium-emphasis">
            PT Artacomindo Jejaring Nusa
          </div>
        </div>
      </div>

      <v-spacer></v-spacer>

      <v-chip color="success" size="small" variant="flat" class="me-3 d-none d-sm-flex font-weight-medium">
        <v-icon start size="14">mdi-shield-check</v-icon>
        Dokumen Publik Resmi
      </v-chip>

      <v-btn
        color="primary"
        variant="outlined"
        size="small"
        prepend-icon="mdi-login"
        class="text-none font-weight-bold"
        to="/login"
      >
        Portal Login
      </v-btn>
    </v-app-bar>

    <!-- Main Content Container -->
    <v-main class="bg-grey-lighten-4 py-8">
      <v-container class="max-w-4xl">
        <!-- Hero Header -->
        <v-card class="mb-6 rounded-xl elevation-2 overflow-hidden border">
          <div class="header-hero pa-6 pa-md-8 text-white">
            <v-chip color="white" variant="flat" size="small" class="text-primary font-weight-bold mb-3">
              KEBIJAKAN PRIVASI & DATA PENGGUNA
            </v-chip>
            <h1 class="text-h4 text-md-h3 font-weight-bold mb-2">
              Kebijakan Privasi Portal Ajnusa
            </h1>
            <p class="text-body-1 opacity-90 mb-3">
              PT Artacomindo Jejaring Nusa &mdash; Sistem Manajemen FTTH & Billing Jelantik
            </p>
            <div class="d-flex flex-wrap align-center gap-4 text-caption opacity-80">
              <span class="d-flex align-center me-4">
                <v-icon size="14" class="me-1">mdi-calendar-clock</v-icon>
                Pembaruan Terakhir: 2 Oktober 2026
              </span>
              <span class="d-flex align-center">
                <v-icon size="14" class="me-1">mdi-earth</v-icon>
                Akses Terbuka & Publik (Bebas Login)
              </span>
            </div>
          </div>
        </v-card>

        <!-- Quick Summary Notice -->
        <v-alert
          type="info"
          variant="tonal"
          class="mb-6 rounded-xl"
          icon="mdi-information-outline"
        >
          <strong>Keterbukaan Akses Publik:</strong> Dokumen kebijakan privasi ini disediakan secara terbuka untuk pengguna aplikasi seluler <strong>Portal Ajnusa</strong>, pelanggan layanan FTTH, dan tim peninjau aplikasi (Google Play Review Team) tanpa memerlukan kredensial login atau autentikasi akun.
        </v-alert>

        <!-- Section 1 -->
        <v-card class="mb-5 rounded-xl border elevation-1">
          <v-card-text class="pa-6 pa-md-8">
            <h2 class="text-h6 font-weight-bold text-primary mb-3 d-flex align-center">
              <v-icon color="primary" class="me-2">mdi-office-building</v-icon>
              1. Identitas Pengembang & Ruang Lingkup Layanan
            </h2>
            <p class="text-body-2 text-high-emphasis leading-relaxed mb-3">
              Kebijakan Privasi ini mengatur bagaimana <strong>PT Artacomindo Jejaring Nusa</strong> ("Kami", "Perusahaan", atau "Jelantik") mengumpulkan, menggunakan, memproses, menyimpan, dan melindungi informasi data pribadi pengguna aplikasi <strong>Portal Ajnusa</strong> serta situs portal web resmi di <strong>https://jpo.jelantik.com</strong>.
            </p>
            <p class="text-body-2 text-high-emphasis leading-relaxed">
              <strong>Portal Ajnusa</strong> merupakan sistem informasi operasional dan portal pelanggan terpadu untuk layanan Internet Serat Optik (FTTH - Fiber To The Home) dan Penagihan (Billing). Aplikasi ini memudahkan pelanggan untuk memantau status langganan, melihat tagihan/invoice bulanan, melakukan konfirmasi pembayaran, serta melaporkan dan memantau status kendala teknis (Trouble Ticket).
            </p>
          </v-card-text>
        </v-card>

        <!-- Section 2 -->
        <v-card class="mb-5 rounded-xl border elevation-1">
          <v-card-text class="pa-6 pa-md-8">
            <h2 class="text-h6 font-weight-bold text-primary mb-3 d-flex align-center">
              <v-icon color="primary" class="me-2">mdi-database-eye</v-icon>
              2. Data Pribadi yang Kami Kumpulkan
            </h2>
            <p class="text-body-2 text-high-emphasis leading-relaxed mb-4">
              Untuk menyediakan fungsionalitas aplikasi dan layanan konektivitas internet secara optimal, Kami mengumpulkan kategori informasi berikut:
            </p>

            <v-expansion-panels variant="accordion" class="rounded-lg mb-4">
              <v-expansion-panel title="A. Data Akun & Identitas Pelanggan">
                <v-expansion-panel-text>
                  <ul class="ps-4 text-body-2 text-medium-emphasis">
                    <li class="mb-2"><strong>Nomor Pelanggan (Customer ID):</strong> Identitas unik langganan FTTH Anda.</li>
                    <li class="mb-2"><strong>Nama Lengkap:</strong> Untuk validasi data pelanggan dan pencetakan faktur tagihan resmi.</li>
                    <li class="mb-2"><strong>Nomor Telepon & WhatsApp:</strong> Untuk autentikasi, penyampaian notifikasi faktur, dan koordinasi teknis penanganan gangguan.</li>
                    <li class="mb-2"><strong>Alamat Email:</strong> Untuk pengiriman bukti pembayaran elektronik dan login akun.</li>
                    <li><strong>Alamat Pemasangan FTTH:</strong> Koordinat/alamat fisik lokasi modem untuk penugasan tim teknisi dan penanganan kendala fisik kabel.</li>
                  </ul>
                </v-expansion-panel-text>
              </v-expansion-panel>

              <v-expansion-panel title="B. Data Perangkat & Diagnostik Sistem">
                <v-expansion-panel-text>
                  <ul class="ps-4 text-body-2 text-medium-emphasis">
                    <li class="mb-2"><strong>FCM Token (Firebase Cloud Messaging):</strong> Token unik perangkat yang diterbitkan oleh Google Firebase guna menyalurkan push notification saat ada tagihan baru atau pembaruan status tiket gangguan.</li>
                    <li class="mb-2"><strong>Informasi Perangkat:</strong> Model perangkat, versi sistem operasi (Android/iOS), dan ID perangkat unik guna mendeteksi ancaman keamanan.</li>
                    <li><strong>Alamat IP dan Log Aktivitas:</strong> Untuk pencatatan audit koneksi keamanan dan verifikasi sesi.</li>
                  </ul>
                </v-expansion-panel-text>
              </v-expansion-panel>

              <v-expansion-panel title="C. Data Layanan Teknis & Trouble Ticket">
                <v-expansion-panel-text>
                  <ul class="ps-4 text-body-2 text-medium-emphasis">
                    <li class="mb-2">Informasi paket langganan internet, alokasi bandwidth, dan status koneksi ONT/modem.</li>
                    <li class="mb-2">Riwayat tiket perbaikan jaringan (Trouble Ticket) dan pesan komunikasi aduan dengan customer care/helpdesk.</li>
                    <li>Foto dokumentasi teknis atau tangkapan layar (screenshot speedtest/lampu indikator modem) yang Anda unggah secara sukarela untuk mempermudah investigasi kendala oleh tim teknisi.</li>
                  </ul>
                </v-expansion-panel-text>
              </v-expansion-panel>
            </v-expansion-panels>
          </v-card-text>
        </v-card>

        <!-- Section 3 -->
        <v-card class="mb-5 rounded-xl border elevation-1">
          <v-card-text class="pa-6 pa-md-8">
            <h2 class="text-h6 font-weight-bold text-primary mb-3 d-flex align-center">
              <v-icon color="primary" class="me-2">mdi-cog-sync</v-icon>
              3. Tujuan Penggunaan Informasi
            </h2>
            <p class="text-body-2 text-high-emphasis leading-relaxed mb-3">
              Kami memproses data pengguna secara sah dan proporsional untuk keperluan:
            </p>
            <v-row dense>
              <v-col cols="12" sm="6">
                <v-card variant="flat" color="blue-lighten-5" class="pa-3 rounded-lg fill-height border">
                  <div class="text-subtitle-2 font-weight-bold text-primary mb-1">
                    <v-icon size="18" class="me-1">mdi-account-check</v-icon>
                    Autentikasi & Akses
                  </div>
                  <div class="text-caption text-medium-emphasis">
                    Memastikan keamanan akun pelanggan serta mencegah akses tidak sah ke data langganan.
                  </div>
                </v-card>
              </v-col>

              <v-col cols="12" sm="6">
                <v-card variant="flat" color="green-lighten-5" class="pa-3 rounded-lg fill-height border">
                  <div class="text-subtitle-2 font-weight-bold text-success mb-1">
                    <v-icon size="18" class="me-1">mdi-receipt-text-check</v-icon>
                    Tagihan & Pembayaran
                  </div>
                  <div class="text-caption text-medium-emphasis">
                    Mencetak faktur tagihan bulanan, memproses pembayaran via payment gateway, dan menerbitkan kuitansi.
                  </div>
                </v-card>
              </v-col>

              <v-col cols="12" sm="6">
                <v-card variant="flat" color="orange-lighten-5" class="pa-3 rounded-lg fill-height border">
                  <div class="text-subtitle-2 font-weight-bold text-warning mb-1">
                    <v-icon size="18" class="me-1">mdi-wrench-clock</v-icon>
                    Resolusi Gangguan (Trouble Ticket)
                  </div>
                  <div class="text-caption text-medium-emphasis">
                    Menugaskan teknisi, melacak downtime jaringan, dan mengirimkan laporan penanganan gangguan.
                  </div>
                </v-card>
              </v-col>

              <v-col cols="12" sm="6">
                <v-card variant="flat" color="purple-lighten-5" class="pa-3 rounded-lg fill-height border">
                  <div class="text-subtitle-2 font-weight-bold text-purple mb-1">
                    <v-icon size="18" class="me-1">mdi-bell-ring</v-icon>
                    Notifikasi Resmi
                  </div>
                  <div class="text-caption text-medium-emphasis">
                    Memberikan pengingat jadwal pemeliharaan jaringan (maintenance), pengingat jatuh tempo, dan rilis aplikasi.
                  </div>
                </v-card>
              </v-col>
            </v-row>
          </v-card-text>
        </v-card>

        <!-- Section 4 -->
        <v-card class="mb-5 rounded-xl border elevation-1">
          <v-card-text class="pa-6 pa-md-8">
            <h2 class="text-h6 font-weight-bold text-primary mb-3 d-flex align-center">
              <v-icon color="primary" class="me-2">mdi-handshake</v-icon>
              4. Pengungkapan kepada Pihak Ketiga
            </h2>
            <v-alert color="primary" variant="tonal" density="compact" class="mb-4 rounded-lg">
              <strong>Jaminan Privasi:</strong> PT Artacomindo Jejaring Nusa <strong>TIDAK AKAN PERNAH</strong> menjual, menyewakan, atau menyebarluaskan data pribadi Anda kepada pihak luar untuk keperluan iklan pihak ketiga atau komersialisasi data.
            </v-alert>
            <p class="text-body-2 text-high-emphasis leading-relaxed mb-3">
              Data hanya dibagikan kepada mitra penyedia infrastruktur terpercaya yang dibutuhkan untuk menjalankan layanan Portal Ajnusa:
            </p>
            <ul class="ps-4 text-body-2 text-medium-emphasis">
              <li class="mb-2">
                <strong>Mitra Payment Gateway (Xendit):</strong> Untuk penyediaan saluran pembayaran resmi seperti Virtual Account, QRIS, gerai ritel, dan e-wallet. Data yang ditransmisikan hanya berupa nomor referensi faktur dan nominal tagihan.
              </li>
              <li class="mb-2">
                <strong>Layanan WhatsApp API (Watzapp / Qontak):</strong> Untuk pengiriman notifikasi faktur tagihan dan notifikasi status trouble ticket langsung ke WhatsApp pengguna.
              </li>
              <li class="mb-2">
                <strong>Google Firebase Cloud Messaging (FCM):</strong> Untuk menyalurkan notifikasi pesan push ke perangkat Android/iOS Anda.
              </li>
              <li>
                <strong>Kewajiban Hukum:</strong> Apabila diwajibkan oleh proses hukum resmi di Republik Indonesia atau putusan pengadilan yang sah.
              </li>
            </ul>
          </v-card-text>
        </v-card>

        <!-- Section 5 -->
        <v-card class="mb-5 rounded-xl border elevation-1">
          <v-card-text class="pa-6 pa-md-8">
            <h2 class="text-h6 font-weight-bold text-primary mb-3 d-flex align-center">
              <v-icon color="primary" class="me-2">mdi-cellphone-cog</v-icon>
              5. Izin Perangkat (Permissions) yang Digunakan
            </h2>
            <p class="text-body-2 text-high-emphasis leading-relaxed mb-3">
              Aplikasi Portal Ajnusa meminta izin perangkat berikut hanya bila diperlukan untuk fitur spesifik:
            </p>
            <div class="d-flex flex-column gap-2">
              <div class="pa-3 bg-grey-lighten-4 rounded-lg border d-flex align-center mb-2">
                <v-icon color="primary" class="me-3">mdi-bell-outline</v-icon>
                <div>
                  <div class="text-subtitle-2 font-weight-bold">Izin Notifikasi (POST_NOTIFICATIONS)</div>
                  <div class="text-caption text-medium-emphasis">Digunakan untuk menerima notifikasi tagihan jatuh tempo dan pembaruan tiket perbaikan.</div>
                </div>
              </div>

              <div class="pa-3 bg-grey-lighten-4 rounded-lg border d-flex align-center mb-2">
                <v-icon color="primary" class="me-3">mdi-camera-outline</v-icon>
                <div>
                  <div class="text-subtitle-2 font-weight-bold">Izin Kamera & Penyimpanan (CAMERA & STORAGE)</div>
                  <div class="text-caption text-medium-emphasis">Digunakan hanya saat Anda mengunggah foto bukti fisik kendala modem, kabel serat optik putus, atau bukti bayar.</div>
                </div>
              </div>

              <div class="pa-3 bg-grey-lighten-4 rounded-lg border d-flex align-center">
                <v-icon color="primary" class="me-3">mdi-web</v-icon>
                <div>
                  <div class="text-subtitle-2 font-weight-bold">Akses Jaringan Internet (INTERNET)</div>
                  <div class="text-caption text-medium-emphasis">Diperlukan untuk sinkronisasi data dengan server backend cloud Portal Ajnusa.</div>
                </div>
              </div>
            </div>
          </v-card-text>
        </v-card>

        <!-- Section 6: Data Deletion & User Rights -->
        <v-card class="mb-5 rounded-xl border elevation-1">
          <v-card-text class="pa-6 pa-md-8">
            <h2 class="text-h6 font-weight-bold text-primary mb-3 d-flex align-center">
              <v-icon color="primary" class="me-2">mdi-account-remove</v-icon>
              6. Hak Pengguna & Prosedur Penghapusan Akun (Data Deletion)
            </h2>
            <p class="text-body-2 text-high-emphasis leading-relaxed mb-3">
              Sesuai dengan Kebijakan Data Pengguna Google Play dan Undang-Undang Perlindungan Data Pribadi (UU PDP Republik Indonesia), setiap pengguna memiliki hak untuk:
            </p>
            <ul class="ps-4 text-body-2 text-medium-emphasis mb-4">
              <li class="mb-1">Meminta akses dan salinan data pribadi yang tersimpan di sistem Kami.</li>
              <li class="mb-1">Memperbaiki atau memperbarui data pribadi yang tidak tepat atau kadaluarsa.</li>
              <li>Meminta <strong>Penghapusan Akun dan Seluruh Data Pribadi Terkait</strong> secara permanen.</li>
            </ul>

            <v-card variant="outlined" color="primary" class="pa-4 rounded-lg bg-blue-lighten-5">
              <div class="text-subtitle-2 font-weight-bold text-primary mb-2">
                <v-icon size="18" class="me-1">mdi-email-fast-outline</v-icon>
                Cara Mengajukan Penghapusan Akun / Data:
              </div>
              <p class="text-caption text-high-emphasis mb-2">
                Anda dapat mengajukan permohonan penghapusan akun kapan saja dengan mengirimkan email ke:
              </p>
              <div class="text-body-2 font-weight-bold text-primary mb-2">
                <a href="mailto:support@ajnusa.com" class="text-decoration-none">support@ajnusa.com</a> atau <a href="mailto:cs@jelantik.com" class="text-decoration-none">cs@jelantik.com</a>
              </div>
              <div class="text-caption text-medium-emphasis">
                Cantumkan: (1) Subjek: "Permohonan Penghapusan Akun Portal Ajnusa", (2) Nama Lengkap, dan (3) Customer ID / Nomor Telepon terdaftar. Tim Data Protection Kami akan memproses verifikasi dan penghapusan data secara tuntas dalam waktu maksimal <strong>14 (empat belas) hari kerja</strong>.
              </div>
            </v-card>
          </v-card-text>
        </v-card>

        <!-- Section 7: Security & Encryption -->
        <v-card class="mb-5 rounded-xl border elevation-1">
          <v-card-text class="pa-6 pa-md-8">
            <h2 class="text-h6 font-weight-bold text-primary mb-3 d-flex align-center">
              <v-icon color="primary" class="me-2">mdi-shield-lock</v-icon>
              7. Keamanan Data & Penyimpanan
            </h2>
            <p class="text-body-2 text-high-emphasis leading-relaxed mb-3">
              Kami memprioritaskan keamanan informasi pelanggan dengan menerapkan langkah-langkah proteksi terstandarisasi:
            </p>
            <ul class="ps-4 text-body-2 text-medium-emphasis">
              <li class="mb-2"><strong>Enkripsi Jaringan:</strong> Seluruh lalu lintas data dienkripsi menggunakan protokol HTTPS / Transport Layer Security (TLS 1.2 / 1.3).</li>
              <li class="mb-2"><strong>Proteksi Basis Data:</strong> Penyimpanan kredensial kata sandi diamankan dengan algoritma hashing mutakhir (bcrypt/argon2).</li>
              <li><strong>Hak Akses Bertingkat (RBAC):</strong> Staf teknis dan operasional hanya dapat mengakses data pelanggan sesuai dengan peran kerja dan kebutuhan penanganan tiket aktif.</li>
            </ul>
          </v-card-text>
        </v-card>

        <!-- Section 8: Children Privacy -->
        <v-card class="mb-5 rounded-xl border elevation-1">
          <v-card-text class="pa-6 pa-md-8">
            <h2 class="text-h6 font-weight-bold text-primary mb-3 d-flex align-center">
              <v-icon color="primary" class="me-2">mdi-account-child</v-icon>
              8. Privasi Anak-Anak
            </h2>
            <p class="text-body-2 text-high-emphasis leading-relaxed">
              Layanan dan aplikasi Portal Ajnusa ditujukan khusus bagi pengguna berusia 18 tahun ke atas yang memiliki kapasitas hukum untuk mengikatkan diri dalam kontrak berlangganan internet. Kami tidak secara sengaja mengumpulkan data dari anak-anak di bawah usia 13 tahun. Jika Kami mendapati adanya data anak di bawah umur yang tersimpan tanpa persetujuan sah wali, data tersebut akan segera dihapus dari server.
            </p>
          </v-card-text>
        </v-card>

        <!-- Section 9: Contact Information -->
        <v-card class="mb-6 rounded-xl border elevation-2">
          <v-card-text class="pa-6 pa-md-8">
            <h2 class="text-h6 font-weight-bold text-primary mb-3 d-flex align-center">
              <v-icon color="primary" class="me-2">mdi-card-account-mail</v-icon>
              9. Informasi Kontak & Layanan Pelanggan
            </h2>
            <p class="text-body-2 text-high-emphasis leading-relaxed mb-4">
              Apabila Anda memiliki pertanyaan, masukan, atau permohonan terkait Kebijakan Privasi ini, silakan hubungi tim dukungan resmi Kami:
            </p>

            <v-card variant="flat" color="grey-lighten-4" class="pa-4 rounded-xl border">
              <v-row dense>
                <v-col cols="12" sm="6">
                  <div class="text-caption text-medium-emphasis">Nama Pengembang Resmi:</div>
                  <div class="text-body-2 font-weight-bold text-high-emphasis mb-2">PT Artacomindo Jejaring Nusa</div>

                  <div class="text-caption text-medium-emphasis">Brand Layanan:</div>
                  <div class="text-body-2 font-weight-bold text-high-emphasis mb-2">Jelantik (FTTH Internet & Network)</div>

                  <div class="text-caption text-medium-emphasis">Nama Aplikasi Google Play:</div>
                  <div class="text-body-2 font-weight-bold text-high-emphasis">Portal Ajnusa</div>
                </v-col>

                <v-col cols="12" sm="6">
                  <div class="text-caption text-medium-emphasis">Email Dukungan Teknis & Privasi:</div>
                  <div class="text-body-2 font-weight-bold mb-2">
                    <a href="mailto:support@ajnusa.com" class="text-primary text-decoration-none">support@ajnusa.com</a> / 
                    <a href="mailto:cs@jelantik.com" class="text-primary text-decoration-none">cs@jelantik.com</a>
                  </div>

                  <div class="text-caption text-medium-emphasis">Situs Web Resmi:</div>
                  <div class="text-body-2 font-weight-bold mb-2">
                    <a href="https://www.ajnusa.com" target="_blank" rel="noopener" class="text-primary text-decoration-none">www.ajnusa.com</a> &bull; 
                    <a href="https://www.jelantik.com" target="_blank" rel="noopener" class="text-primary text-decoration-none">www.jelantik.com</a>
                  </div>

                  <div class="text-caption text-medium-emphasis">Portal Web FTTH:</div>
                  <div class="text-body-2 font-weight-bold">
                    <a href="https://jpo.jelantik.com" target="_blank" rel="noopener" class="text-primary text-decoration-none">https://jpo.jelantik.com</a>
                  </div>
                </v-col>
              </v-row>
            </v-card>
          </v-card-text>
        </v-card>

        <!-- Back to Top / Action Buttons -->
        <div class="d-flex justify-center gap-3 mb-6">
          <v-btn
            color="primary"
            variant="flat"
            prepend-icon="mdi-arrow-up"
            size="large"
            class="text-none font-weight-bold rounded-lg"
            @click="scrollToTop"
          >
            Kembali ke Atas
          </v-btn>
          <v-btn
            color="grey-darken-1"
            variant="outlined"
            prepend-icon="mdi-printer"
            size="large"
            class="text-none font-weight-bold rounded-lg"
            @click="printPolicy"
          >
            Cetak Dokumen
          </v-btn>
        </div>
      </v-container>
    </v-main>

    <!-- Footer -->
    <v-footer class="bg-white border-t py-6 text-center text-caption text-medium-emphasis">
      <v-container>
        <div>&copy; 2026 PT Artacomindo Jejaring Nusa. Seluruh hak cipta dilindungi undang-undang.</div>
        <div class="mt-2">
          <a href="https://jpo.jelantik.com" class="text-decoration-none text-primary me-3">Portal FTTH Jelantik</a>
          <a href="/privacy-policy" class="text-decoration-none text-primary me-3">Kebijakan Privasi</a>
          <a href="https://www.ajnusa.com" target="_blank" rel="noopener" class="text-decoration-none text-primary">PT Artacomindo Jejaring Nusa</a>
        </div>
      </v-container>
    </v-footer>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';

const router = useRouter();

function goHome() {
  router.push('/login');
}

function handleLogoError(event: Event) {
  const target = event.target as HTMLImageElement;
  target.src = '/portal_logo.ico';
}

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: 'smooth' });
}

function printPolicy() {
  window.print();
}
</script>

<style scoped>
.privacy-policy-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: #f8fafc;
}

.header-hero {
  background: linear-gradient(135deg, #1565C0 0%, #0D47A1 100%);
  position: relative;
}

.max-w-4xl {
  max-width: 900px !important;
}

@media print {
  .v-app-bar,
  .v-footer,
  .v-btn,
  .v-chip {
    display: none !important;
  }
  .privacy-policy-page {
    background: white !important;
  }
}
</style>
