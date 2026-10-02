<template>
  <div class="pp-root">
    <!-- Navbar -->
    <nav class="pp-nav">
      <div class="pp-nav-inner">
        <div class="pp-nav-brand" @click="goHome">
          <img
            src="/portal_logo.svg"
            alt="Logo"
            class="pp-nav-logo"
            @error="handleLogoError"
          />
          <div class="pp-nav-brand-text">
            <span class="pp-nav-title">Portal Ajnusa</span>
            <span class="pp-nav-subtitle">PT Artacomindo Jejaring Nusa</span>
          </div>
        </div>

        <div class="pp-nav-actions">
          <span class="pp-badge-public">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/><polyline points="9 12 12 15 16 10"/></svg>
            Dokumen Publik
          </span>
          <router-link to="/login" class="pp-btn-login">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"/><polyline points="10 17 15 12 10 7"/><line x1="15" y1="12" x2="3" y2="12"/></svg>
            Login Portal
          </router-link>
        </div>
      </div>
    </nav>

    <!-- Hero -->
    <header class="pp-hero">
      <div class="pp-hero-bg"></div>
      <div class="pp-hero-content">
        <div class="pp-hero-label">KEBIJAKAN PRIVASI</div>
        <h1 class="pp-hero-title">Kebijakan Privasi<br/>Portal Ajnusa</h1>
        <p class="pp-hero-desc">
          PT Artacomindo Jejaring Nusa — Sistem Manajemen FTTH & Billing Jelantik
        </p>
        <div class="pp-hero-meta">
          <span class="pp-hero-meta-item">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg>
            Terakhir diperbarui: 2 Oktober 2026
          </span>
          <span class="pp-hero-meta-item">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
            Akses Publik — Bebas Login
          </span>
        </div>
      </div>
    </header>

    <!-- Table of Contents -->
    <div class="pp-container">
      <aside class="pp-toc">
        <div class="pp-toc-label">DAFTAR ISI</div>
        <ol class="pp-toc-list">
          <li><a href="#s1" @click.prevent="scrollToSection('s1')">Identitas & Ruang Lingkup</a></li>
          <li><a href="#s2" @click.prevent="scrollToSection('s2')">Data yang Dikumpulkan</a></li>
          <li><a href="#s3" @click.prevent="scrollToSection('s3')">Tujuan Penggunaan</a></li>
          <li><a href="#s4" @click.prevent="scrollToSection('s4')">Pihak Ketiga</a></li>
          <li><a href="#s5" @click.prevent="scrollToSection('s5')">Izin Perangkat</a></li>
          <li><a href="#s6" @click.prevent="scrollToSection('s6')">Keamanan Data</a></li>
          <li><a href="#s7" @click.prevent="scrollToSection('s7')">Hak Pengguna & Penghapusan</a></li>
          <li><a href="#s8" @click.prevent="scrollToSection('s8')">Privasi Anak</a></li>
          <li><a href="#s9" @click.prevent="scrollToSection('s9')">Perubahan Kebijakan</a></li>
          <li><a href="#s10" @click.prevent="scrollToSection('s10')">Hubungi Kami</a></li>
        </ol>
      </aside>

      <!-- Sections -->
      <main class="pp-main">

        <!-- Section 1 -->
        <section id="s1" class="pp-section">
          <div class="pp-section-number">01</div>
          <h2 class="pp-section-title">Identitas Pengembang & Ruang Lingkup</h2>
          <p>
            Kebijakan Privasi ini mengatur bagaimana <strong>PT Artacomindo Jejaring Nusa</strong>
            ("Kami", "Perusahaan", atau "Jelantik") mengumpulkan, menggunakan, memproses,
            menyimpan, dan melindungi informasi data pribadi pengguna aplikasi
            <strong>Portal Ajnusa</strong> serta portal web resmi di
            <strong>https://jpo.jelantik.com</strong>.
          </p>
          <p>
            <strong>Portal Ajnusa</strong> merupakan aplikasi manajemen layanan internet
            serat optik (FTTH — Fiber To The Home) dan penagihan (Billing) yang dirancang
            untuk pelanggan dan staf teknis PT Artacomindo Jejaring Nusa. Aplikasi ini
            memudahkan pelanggan untuk memantau status langganan, melihat tagihan/invoice
            bulanan, melakukan konfirmasi pembayaran, serta melaporkan dan memantau status
            kendala teknis (Trouble Ticket).
          </p>
          <div class="pp-callout pp-callout-info">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
            <div>
              <strong>Akses Terbuka:</strong> Kebijakan privasi ini dapat diakses secara publik
              tanpa memerlukan kredensial login atau autentikasi akun.
            </div>
          </div>
        </section>

        <!-- Section 2 -->
        <section id="s2" class="pp-section">
          <div class="pp-section-number">02</div>
          <h2 class="pp-section-title">Data Pribadi yang Kami Kumpulkan</h2>
          <p>
            Untuk menyediakan fungsionalitas aplikasi dan layanan konektivitas internet
            secara optimal, Kami mengumpulkan kategori informasi berikut:
          </p>

          <div class="pp-data-group">
            <h3 class="pp-data-group-title">
              <span class="pp-data-group-icon pp-icon-blue">A</span>
              Data Akun & Identitas Pelanggan
            </h3>
            <ul class="pp-list">
              <li><strong>Nomor Pelanggan (Customer ID)</strong> — Identitas unik langganan FTTH Anda.</li>
              <li><strong>Nama Lengkap</strong> — Untuk validasi data pelanggan dan pencetakan faktur tagihan resmi.</li>
              <li><strong>Nomor Telepon & WhatsApp</strong> — Untuk autentikasi, penyampaian notifikasi faktur, dan koordinasi teknis penanganan gangguan.</li>
              <li><strong>Alamat Email</strong> — Untuk pengiriman bukti pembayaran elektronik dan login akun.</li>
              <li><strong>Alamat Pemasangan FTTH</strong> — Lokasi fisik modem untuk penugasan tim teknisi dan penanganan kendala fisik kabel.</li>
            </ul>
          </div>

          <div class="pp-data-group">
            <h3 class="pp-data-group-title">
              <span class="pp-data-group-icon pp-icon-teal">B</span>
              Data Perangkat & Diagnostik
            </h3>
            <ul class="pp-list">
              <li><strong>FCM Token (Firebase Cloud Messaging)</strong> — Token unik perangkat dari Google Firebase untuk mengirimkan push notification saat ada tagihan baru atau pembaruan status tiket gangguan.</li>
              <li><strong>Informasi Perangkat</strong> — Model perangkat, versi sistem operasi (Android/iOS), dan ID perangkat unik untuk mendeteksi ancaman keamanan.</li>
              <li><strong>Alamat IP dan Log Aktivitas</strong> — Untuk pencatatan audit keamanan dan verifikasi sesi.</li>
            </ul>
          </div>

          <div class="pp-data-group">
            <h3 class="pp-data-group-title">
              <span class="pp-data-group-icon pp-icon-amber">C</span>
              Data Layanan Teknis & Trouble Ticket
            </h3>
            <ul class="pp-list">
              <li>Informasi paket langganan internet, alokasi bandwidth, dan status koneksi ONT/modem.</li>
              <li>Riwayat tiket perbaikan jaringan (Trouble Ticket) dan pesan komunikasi aduan dengan customer care/helpdesk.</li>
              <li>Foto dokumentasi teknis atau tangkapan layar (screenshot speedtest/lampu indikator modem) yang Anda unggah secara sukarela untuk mempermudah investigasi kendala oleh tim teknisi.</li>
            </ul>
          </div>
        </section>

        <!-- Section 3 -->
        <section id="s3" class="pp-section">
          <div class="pp-section-number">03</div>
          <h2 class="pp-section-title">Tujuan Penggunaan Informasi</h2>
          <p>Kami memproses data pengguna secara sah dan proporsional untuk keperluan:</p>

          <div class="pp-purpose-grid">
            <div class="pp-purpose-card">
              <div class="pp-purpose-icon" style="background: #EEF2FF; color: #4F46E5;">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
              </div>
              <h4>Autentikasi & Akses</h4>
              <p>Memastikan keamanan akun pelanggan dan mencegah akses tidak sah ke data langganan.</p>
            </div>
            <div class="pp-purpose-card">
              <div class="pp-purpose-icon" style="background: #ECFDF5; color: #059669;">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="1" y="4" width="22" height="16" rx="2" ry="2"/><line x1="1" y1="10" x2="23" y2="10"/></svg>
              </div>
              <h4>Tagihan & Pembayaran</h4>
              <p>Mencetak faktur bulanan, memproses pembayaran via payment gateway, dan menerbitkan kuitansi.</p>
            </div>
            <div class="pp-purpose-card">
              <div class="pp-purpose-icon" style="background: #FFF7ED; color: #EA580C;">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/></svg>
              </div>
              <h4>Resolusi Gangguan</h4>
              <p>Menugaskan teknisi, melacak downtime jaringan, dan mengirimkan laporan penanganan.</p>
            </div>
            <div class="pp-purpose-card">
              <div class="pp-purpose-icon" style="background: #F5F3FF; color: #7C3AED;">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/></svg>
              </div>
              <h4>Notifikasi Resmi</h4>
              <p>Pengingat jadwal pemeliharaan jaringan, jatuh tempo tagihan, dan status tiket perbaikan.</p>
            </div>
          </div>
        </section>

        <!-- Section 4 -->
        <section id="s4" class="pp-section">
          <div class="pp-section-number">04</div>
          <h2 class="pp-section-title">Pengungkapan kepada Pihak Ketiga</h2>

          <div class="pp-callout pp-callout-success">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
            <div>
              <strong>Jaminan Privasi:</strong> Kami <strong>TIDAK AKAN PERNAH</strong> menjual,
              menyewakan, atau menyebarluaskan data pribadi Anda kepada pihak luar untuk
              iklan pihak ketiga atau komersialisasi data.
            </div>
          </div>

          <p>
            Data hanya dibagikan kepada mitra penyedia infrastruktur terpercaya yang
            dibutuhkan untuk menjalankan layanan Portal Ajnusa:
          </p>

          <div class="pp-partner-list">
            <div class="pp-partner-item">
              <div class="pp-partner-label">Payment Gateway</div>
              <div class="pp-partner-name">Xendit</div>
              <p>Penyediaan saluran pembayaran resmi (Virtual Account, QRIS, e-wallet). Data yang ditransmisikan hanya berupa nomor referensi faktur dan nominal tagihan.</p>
            </div>
            <div class="pp-partner-item">
              <div class="pp-partner-label">WhatsApp API</div>
              <div class="pp-partner-name">Watzapp / Qontak</div>
              <p>Pengiriman notifikasi faktur tagihan dan status trouble ticket langsung ke WhatsApp pengguna.</p>
            </div>
            <div class="pp-partner-item">
              <div class="pp-partner-label">Push Notification</div>
              <div class="pp-partner-name">Google Firebase (FCM)</div>
              <p>Menyalurkan notifikasi pesan push ke perangkat Android/iOS Anda.</p>
            </div>
            <div class="pp-partner-item">
              <div class="pp-partner-label">Kepatuhan Hukum</div>
              <div class="pp-partner-name">Republik Indonesia</div>
              <p>Apabila diwajibkan oleh proses hukum resmi atau putusan pengadilan yang sah.</p>
            </div>
          </div>
        </section>

        <!-- Section 5 -->
        <section id="s5" class="pp-section">
          <div class="pp-section-number">05</div>
          <h2 class="pp-section-title">Izin Perangkat (Permissions)</h2>
          <p>Aplikasi Portal Ajnusa meminta izin perangkat berikut hanya bila diperlukan:</p>

          <div class="pp-perm-list">
            <div class="pp-perm-item">
              <div class="pp-perm-icon">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/></svg>
              </div>
              <div>
                <strong>Notifikasi (POST_NOTIFICATIONS)</strong>
                <p>Menerima pemberitahuan tagihan jatuh tempo dan pembaruan tiket perbaikan.</p>
              </div>
            </div>
            <div class="pp-perm-item">
              <div class="pp-perm-icon">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z"/><circle cx="12" cy="13" r="4"/></svg>
              </div>
              <div>
                <strong>Kamera & Penyimpanan (CAMERA & STORAGE)</strong>
                <p>Mengunggah foto bukti kendala modem, kabel serat optik putus, atau bukti bayar manual.</p>
              </div>
            </div>
            <div class="pp-perm-item">
              <div class="pp-perm-icon">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
              </div>
              <div>
                <strong>Akses Internet (INTERNET)</strong>
                <p>Sinkronisasi data dengan server backend cloud Portal Ajnusa.</p>
              </div>
            </div>
          </div>

          <p class="pp-note">
            Anda dapat mengontrol, mengubah, atau mencabut izin ini kapan saja melalui
            pengaturan perangkat Anda (Settings → Apps → Portal Ajnusa → Permissions).
          </p>
        </section>

        <!-- Section 6 -->
        <section id="s6" class="pp-section">
          <div class="pp-section-number">06</div>
          <h2 class="pp-section-title">Keamanan Data & Penyimpanan</h2>
          <p>
            Kami memprioritaskan keamanan informasi pelanggan dengan menerapkan
            langkah-langkah proteksi terstandarisasi:
          </p>
          <ul class="pp-list">
            <li><strong>Enkripsi Jaringan:</strong> Seluruh lalu lintas data dienkripsi menggunakan protokol HTTPS / Transport Layer Security (TLS 1.2 / 1.3).</li>
            <li><strong>Proteksi Kata Sandi:</strong> Kredensial kata sandi diamankan dengan algoritma hashing (bcrypt/argon2). Kami tidak menyimpan kata sandi dalam bentuk teks biasa.</li>
            <li><strong>Hak Akses Bertingkat (RBAC):</strong> Staf teknis dan operasional hanya dapat mengakses data pelanggan sesuai dengan peran kerja dan kebutuhan penanganan tiket aktif.</li>
            <li><strong>Retensi Data:</strong> Data transaksi penagihan disimpan sesuai dengan ketentuan akuntansi dan peraturan perpajakan yang berlaku di Indonesia.</li>
          </ul>
        </section>

        <!-- Section 7 -->
        <section id="s7" class="pp-section">
          <div class="pp-section-number">07</div>
          <h2 class="pp-section-title">Hak Pengguna & Penghapusan Akun</h2>
          <p>
            Sesuai dengan Kebijakan Data Pengguna Google Play dan Undang-Undang
            Perlindungan Data Pribadi (UU PDP Republik Indonesia), setiap pengguna
            memiliki hak untuk:
          </p>
          <ul class="pp-list">
            <li>Meminta akses dan salinan data pribadi yang tersimpan di sistem Kami.</li>
            <li>Memperbaiki atau memperbarui data pribadi yang tidak tepat atau kadaluarsa.</li>
            <li>Meminta <strong>Penghapusan Akun dan Seluruh Data Pribadi Terkait</strong> secara permanen.</li>
          </ul>

          <div class="pp-deletion-box">
            <h4>
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>
              Cara Mengajukan Penghapusan Akun
            </h4>
            <p>
              Kirim email ke <a href="mailto:support@ajnusa.com">support@ajnusa.com</a>
              atau <a href="mailto:cs@jelantik.com">cs@jelantik.com</a> dengan mencantumkan:
            </p>
            <ol>
              <li>Subjek: <em>"Permohonan Penghapusan Akun Portal Ajnusa"</em></li>
              <li>Nama Lengkap dan Customer ID / Nomor Telepon terdaftar</li>
              <li>Alasan permohonan (opsional)</li>
            </ol>
            <p class="pp-deletion-timeline">
              Tim Kami akan memverifikasi dan memproses penghapusan dalam
              <strong>14 hari kerja</strong>.
            </p>
          </div>
        </section>

        <!-- Section 8 -->
        <section id="s8" class="pp-section">
          <div class="pp-section-number">08</div>
          <h2 class="pp-section-title">Privasi Anak-Anak</h2>
          <p>
            Layanan Portal Ajnusa ditujukan bagi pengguna berusia 18 tahun ke atas yang
            memiliki kapasitas hukum untuk mengikatkan diri dalam kontrak berlangganan
            internet. Kami tidak secara sengaja mengumpulkan data dari anak-anak di bawah
            usia 13 tahun. Jika Kami mendapati adanya data anak di bawah umur yang
            tersimpan tanpa persetujuan orang tua/wali, data tersebut akan segera dihapus
            dari server.
          </p>
        </section>

        <!-- Section 9 -->
        <section id="s9" class="pp-section">
          <div class="pp-section-number">09</div>
          <h2 class="pp-section-title">Perubahan Kebijakan Privasi</h2>
          <p>
            Kami dapat memperbarui Kebijakan Privasi ini dari waktu ke waktu untuk
            menyesuaikan dengan perubahan layanan atau regulasi hukum. Setiap pembaruan
            akan dipublikasikan di halaman ini dengan tanggal pembaruan terbaru. Kami
            menyarankan Anda untuk memeriksa halaman ini secara berkala.
          </p>
        </section>

        <!-- Section 10 -->
        <section id="s10" class="pp-section">
          <div class="pp-section-number">10</div>
          <h2 class="pp-section-title">Hubungi Kami</h2>
          <p>
            Jika Anda memiliki pertanyaan, masukan, atau permohonan terkait Kebijakan
            Privasi ini, silakan hubungi tim dukungan resmi Kami:
          </p>

          <div class="pp-contact-grid">
            <div class="pp-contact-item">
              <div class="pp-contact-label">Pengembang</div>
              <div class="pp-contact-value">PT Artacomindo Jejaring Nusa</div>
            </div>
            <div class="pp-contact-item">
              <div class="pp-contact-label">Brand Layanan</div>
              <div class="pp-contact-value">Jelantik (FTTH Internet & Network)</div>
            </div>
            <div class="pp-contact-item">
              <div class="pp-contact-label">Nama Aplikasi</div>
              <div class="pp-contact-value">Portal Ajnusa</div>
            </div>
            <div class="pp-contact-item">
              <div class="pp-contact-label">Email Privasi & Dukungan</div>
              <div class="pp-contact-value">
                <a href="mailto:support@ajnusa.com">support@ajnusa.com</a>&nbsp;/
                <a href="mailto:cs@jelantik.com">cs@jelantik.com</a>
              </div>
            </div>
            <div class="pp-contact-item">
              <div class="pp-contact-label">Situs Resmi</div>
              <div class="pp-contact-value">
                <a href="https://www.ajnusa.com" target="_blank" rel="noopener">www.ajnusa.com</a>&nbsp;•
                <a href="https://www.jelantik.com" target="_blank" rel="noopener">www.jelantik.com</a>
              </div>
            </div>
            <div class="pp-contact-item">
              <div class="pp-contact-label">Portal FTTH</div>
              <div class="pp-contact-value">
                <a href="https://jpo.jelantik.com" target="_blank" rel="noopener">https://jpo.jelantik.com</a>
              </div>
            </div>
          </div>
        </section>

        <!-- Bottom Actions -->
        <div class="pp-bottom-actions">
          <button class="pp-btn-top" @click="scrollToTop">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="18 15 12 9 6 15"/></svg>
            Kembali ke Atas
          </button>
          <button class="pp-btn-print" @click="printPolicy">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 6 2 18 2 18 9"/><path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2"/><rect x="6" y="14" width="12" height="8"/></svg>
            Cetak / Simpan PDF
          </button>
        </div>
      </main>
    </div>

    <!-- Footer -->
    <footer class="pp-footer">
      <p>&copy; 2026 PT Artacomindo Jejaring Nusa. Seluruh hak cipta dilindungi undang-undang.</p>
      <div class="pp-footer-links">
        <a href="https://jpo.jelantik.com">Portal FTTH Jelantik</a>
        <span>•</span>
        <a href="/privacy-policy">Kebijakan Privasi</a>
        <span>•</span>
        <a href="https://www.ajnusa.com" target="_blank" rel="noopener">PT Artacomindo Jejaring Nusa</a>
      </div>
    </footer>
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

function scrollToSection(id: string) {
  const el = document.getElementById(id);
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }
}

function printPolicy() {
  window.print();
}
</script>

<style scoped>
/* ===== Foundation ===== */
.pp-root {
  min-height: 100vh;
  background: #F7F8FA;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  color: #1a1a2e;
  line-height: 1.7;
  -webkit-font-smoothing: antialiased;
}

/* ===== Navbar ===== */
.pp-nav {
  position: sticky;
  top: 0;
  z-index: 100;
  background: #fff;
  border-bottom: 1px solid #e5e7eb;
  height: 60px;
}
.pp-nav-inner {
  max-width: 1100px;
  margin: 0 auto;
  padding: 0 24px;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.pp-nav-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}
.pp-nav-logo {
  width: 32px;
  height: 32px;
  object-fit: contain;
}
.pp-nav-brand-text {
  display: flex;
  flex-direction: column;
  line-height: 1.2;
}
.pp-nav-title {
  font-weight: 700;
  font-size: 0.95rem;
  color: #1565C0;
}
.pp-nav-subtitle {
  font-size: 0.72rem;
  color: #94a3b8;
  font-weight: 500;
}
.pp-nav-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}
.pp-badge-public {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 0.73rem;
  font-weight: 600;
  color: #059669;
  background: #ECFDF5;
  padding: 4px 10px;
  border-radius: 20px;
  border: 1px solid #A7F3D0;
}
.pp-btn-login {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 0.82rem;
  font-weight: 600;
  color: #1565C0;
  border: 1.5px solid #1565C0;
  padding: 6px 14px;
  border-radius: 8px;
  text-decoration: none;
  transition: all 0.15s ease;
}
.pp-btn-login:hover {
  background: #1565C0;
  color: #fff;
}

/* ===== Hero ===== */
.pp-hero {
  position: relative;
  background: linear-gradient(145deg, #0D47A1 0%, #1565C0 40%, #1976D2 100%);
  padding: 56px 24px 48px;
  overflow: hidden;
}
.pp-hero-bg {
  position: absolute;
  inset: 0;
  opacity: 0.07;
  background-image: url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23ffffff' fill-opacity='1'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E");
}
.pp-hero-content {
  position: relative;
  max-width: 720px;
  margin: 0 auto;
  text-align: center;
  color: #fff;
}
.pp-hero-label {
  display: inline-block;
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  background: rgba(255, 255, 255, 0.18);
  padding: 5px 14px;
  border-radius: 20px;
  margin-bottom: 20px;
  border: 1px solid rgba(255, 255, 255, 0.25);
}
.pp-hero-title {
  font-size: 2.1rem;
  font-weight: 800;
  line-height: 1.25;
  margin-bottom: 12px;
  letter-spacing: -0.02em;
}
.pp-hero-desc {
  font-size: 1rem;
  opacity: 0.85;
  margin-bottom: 20px;
  font-weight: 400;
}
.pp-hero-meta {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: 20px;
  font-size: 0.8rem;
  opacity: 0.75;
}
.pp-hero-meta-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

/* ===== Container ===== */
.pp-container {
  max-width: 1100px;
  margin: 0 auto;
  padding: 40px 24px 60px;
  display: grid;
  grid-template-columns: 220px 1fr;
  gap: 48px;
  align-items: start;
}

/* ===== TOC ===== */
.pp-toc {
  position: sticky;
  top: 80px;
}
.pp-toc-label {
  font-size: 0.65rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  color: #94a3b8;
  margin-bottom: 12px;
  padding-left: 4px;
}
.pp-toc-list {
  list-style: none;
  padding: 0;
  margin: 0;
  counter-reset: toc;
}
.pp-toc-list li {
  counter-increment: toc;
}
.pp-toc-list li a {
  display: block;
  padding: 6px 12px;
  font-size: 0.82rem;
  color: #64748b;
  text-decoration: none;
  border-left: 2px solid transparent;
  transition: all 0.15s ease;
  font-weight: 500;
}
.pp-toc-list li a:hover {
  color: #1565C0;
  border-left-color: #1565C0;
  background: #EEF4FF;
}

/* ===== Section ===== */
.pp-section {
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 32px;
  margin-bottom: 20px;
  position: relative;
}
.pp-section-number {
  position: absolute;
  top: 28px;
  right: 28px;
  font-size: 2.4rem;
  font-weight: 800;
  color: #f0f1f3;
  line-height: 1;
  pointer-events: none;
  user-select: none;
}
.pp-section-title {
  font-size: 1.2rem;
  font-weight: 700;
  color: #0D47A1;
  margin-bottom: 16px;
  padding-bottom: 10px;
  border-bottom: 2px solid #E3F2FD;
}
.pp-section p {
  color: #374151;
  margin-bottom: 12px;
  font-size: 0.93rem;
}
.pp-section strong {
  color: #1a1a2e;
}

/* ===== Callout ===== */
.pp-callout {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  padding: 14px 18px;
  border-radius: 8px;
  margin: 16px 0;
  font-size: 0.9rem;
  line-height: 1.6;
}
.pp-callout svg {
  flex-shrink: 0;
  margin-top: 2px;
}
.pp-callout-info {
  background: #EFF6FF;
  border: 1px solid #BFDBFE;
  color: #1e40af;
}
.pp-callout-info svg { color: #3B82F6; }
.pp-callout-success {
  background: #F0FDF4;
  border: 1px solid #BBF7D0;
  color: #166534;
}
.pp-callout-success svg { color: #22C55E; }

/* ===== Data Groups ===== */
.pp-data-group {
  margin-top: 20px;
  margin-bottom: 8px;
}
.pp-data-group-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 0.95rem;
  font-weight: 700;
  color: #1e293b;
  margin-bottom: 10px;
}
.pp-data-group-icon {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  font-weight: 800;
  color: #fff;
  flex-shrink: 0;
}
.pp-icon-blue { background: #3B82F6; }
.pp-icon-teal { background: #14B8A6; }
.pp-icon-amber { background: #F59E0B; }

/* ===== List ===== */
.pp-list {
  padding-left: 20px;
  margin: 8px 0 12px;
}
.pp-list li {
  font-size: 0.9rem;
  color: #4b5563;
  margin-bottom: 8px;
  line-height: 1.6;
}
.pp-list li strong {
  color: #1e293b;
}

/* ===== Purpose Grid ===== */
.pp-purpose-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-top: 16px;
}
.pp-purpose-card {
  padding: 18px;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  background: #fafbfc;
}
.pp-purpose-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 10px;
}
.pp-purpose-card h4 {
  font-size: 0.88rem;
  font-weight: 700;
  color: #1e293b;
  margin-bottom: 4px;
}
.pp-purpose-card p {
  font-size: 0.82rem;
  color: #6b7280;
  margin-bottom: 0;
  line-height: 1.5;
}

/* ===== Partner ===== */
.pp-partner-list {
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  gap: 1px;
  background: #e5e7eb;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  overflow: hidden;
}
.pp-partner-item {
  padding: 16px 20px;
  background: #fff;
}
.pp-partner-label {
  font-size: 0.7rem;
  font-weight: 700;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  margin-bottom: 2px;
}
.pp-partner-name {
  font-size: 0.95rem;
  font-weight: 700;
  color: #1e293b;
  margin-bottom: 4px;
}
.pp-partner-item p {
  font-size: 0.85rem;
  color: #6b7280;
  margin: 0;
}

/* ===== Permission ===== */
.pp-perm-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin: 16px 0;
}
.pp-perm-item {
  display: flex;
  gap: 14px;
  align-items: flex-start;
  padding: 14px 18px;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  background: #fafbfc;
}
.pp-perm-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: #EEF2FF;
  color: #4F46E5;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.pp-perm-item strong {
  font-size: 0.88rem;
  color: #1e293b;
  display: block;
  margin-bottom: 2px;
}
.pp-perm-item p {
  font-size: 0.83rem;
  color: #6b7280;
  margin: 0;
}
.pp-note {
  font-size: 0.82rem;
  color: #94a3b8;
  font-style: italic;
}

/* ===== Deletion Box ===== */
.pp-deletion-box {
  margin-top: 20px;
  border: 1.5px solid #93C5FD;
  border-radius: 10px;
  padding: 20px 24px;
  background: #EFF6FF;
}
.pp-deletion-box h4 {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.95rem;
  font-weight: 700;
  color: #1e40af;
  margin-bottom: 10px;
}
.pp-deletion-box p {
  font-size: 0.88rem;
  color: #374151;
}
.pp-deletion-box a {
  color: #1565C0;
  font-weight: 600;
  text-decoration: none;
}
.pp-deletion-box a:hover {
  text-decoration: underline;
}
.pp-deletion-box ol {
  padding-left: 20px;
  margin: 8px 0 12px;
}
.pp-deletion-box ol li {
  font-size: 0.85rem;
  color: #374151;
  margin-bottom: 4px;
}
.pp-deletion-timeline {
  margin-top: 8px;
  padding-top: 10px;
  border-top: 1px dashed #93C5FD;
  font-size: 0.85rem;
  color: #1e40af;
}

/* ===== Contact Grid ===== */
.pp-contact-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1px;
  background: #e5e7eb;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  overflow: hidden;
  margin-top: 16px;
}
.pp-contact-item {
  padding: 14px 18px;
  background: #fff;
}
.pp-contact-label {
  font-size: 0.7rem;
  font-weight: 700;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 2px;
}
.pp-contact-value {
  font-size: 0.9rem;
  font-weight: 600;
  color: #1e293b;
}
.pp-contact-value a {
  color: #1565C0;
  text-decoration: none;
}
.pp-contact-value a:hover {
  text-decoration: underline;
}

/* ===== Bottom Actions ===== */
.pp-bottom-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
  margin-top: 8px;
  padding: 24px 0;
}
.pp-btn-top,
.pp-btn-print {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  border: none;
  transition: all 0.15s ease;
}
.pp-btn-top {
  background: #1565C0;
  color: #fff;
}
.pp-btn-top:hover {
  background: #0D47A1;
}
.pp-btn-print {
  background: #fff;
  color: #4b5563;
  border: 1.5px solid #d1d5db;
}
.pp-btn-print:hover {
  background: #f3f4f6;
  border-color: #9ca3af;
}

/* ===== Footer ===== */
.pp-footer {
  text-align: center;
  padding: 28px 24px;
  background: #fff;
  border-top: 1px solid #e5e7eb;
  font-size: 0.82rem;
  color: #94a3b8;
}
.pp-footer-links {
  margin-top: 8px;
  display: flex;
  justify-content: center;
  gap: 10px;
  flex-wrap: wrap;
}
.pp-footer-links a {
  color: #1565C0;
  text-decoration: none;
  font-weight: 500;
}
.pp-footer-links a:hover {
  text-decoration: underline;
}
.pp-footer-links span {
  color: #d1d5db;
}

/* ===== Responsive ===== */
@media (max-width: 860px) {
  .pp-container {
    grid-template-columns: 1fr;
    gap: 0;
    padding-top: 24px;
  }
  .pp-toc {
    position: static;
    margin-bottom: 24px;
    padding: 16px 20px;
    background: #fff;
    border: 1px solid #e5e7eb;
    border-radius: 10px;
  }
  .pp-toc-list {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }
  .pp-toc-list li a {
    border-left: none;
    padding: 4px 10px;
    border-radius: 6px;
    font-size: 0.78rem;
  }
  .pp-toc-list li a:hover {
    border-left: none;
  }
  .pp-hero-title {
    font-size: 1.6rem;
  }
  .pp-section {
    padding: 24px 20px;
  }
  .pp-purpose-grid {
    grid-template-columns: 1fr;
  }
  .pp-contact-grid {
    grid-template-columns: 1fr;
  }
  .pp-badge-public {
    display: none;
  }
}

@media (max-width: 480px) {
  .pp-hero {
    padding: 36px 16px 32px;
  }
  .pp-hero-title {
    font-size: 1.35rem;
  }
  .pp-hero-meta {
    flex-direction: column;
    align-items: center;
    gap: 6px;
  }
  .pp-bottom-actions {
    flex-direction: column;
    align-items: stretch;
  }
  .pp-btn-top,
  .pp-btn-print {
    justify-content: center;
  }
  .pp-nav-brand-text {
    display: none;
  }
}

/* ===== Print ===== */
@media print {
  .pp-nav,
  .pp-toc,
  .pp-bottom-actions,
  .pp-footer,
  .pp-btn-login,
  .pp-badge-public {
    display: none !important;
  }
  .pp-root {
    background: #fff !important;
  }
  .pp-container {
    display: block !important;
    padding: 0 !important;
  }
  .pp-hero {
    background: #1565C0 !important;
    -webkit-print-color-adjust: exact;
    print-color-adjust: exact;
  }
  .pp-section {
    break-inside: avoid;
    border: 1px solid #ddd;
    margin-bottom: 12px;
  }
}
</style>
