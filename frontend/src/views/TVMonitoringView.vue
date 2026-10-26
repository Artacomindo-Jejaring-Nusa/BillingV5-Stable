<template>
  <div class="tv-monitoring-root" :class="{ 'is-fullscreen': isFullscreen }">
    <!-- Top Bar / Header for NOC Wallboard -->
    <header class="tv-header">
      <div class="tv-header-left">
        <div class="brand-badge">
          <v-icon color="#0284c7" size="24" class="mr-2">mdi-access-point-network</v-icon>
          <div class="brand-titles">
            <h1 class="tv-title">NOC & BILLING MONITORING</h1>
            <p class="tv-subtitle">PT Artacomindo Jejaring Nusa • Single Wallboard Monitor</p>
          </div>
        </div>
      </div>

      <!-- Center Controls: Location Filter & Auto Cycle -->
      <div class="tv-header-center">
        <div class="location-control">
          <v-icon size="18" color="#94a3b8" class="mr-1">mdi-map-marker-radius</v-icon>
          <span class="location-label">Lokasi:</span>
          <v-select
            v-model="selectedLocation"
            :items="availableLocations"
            density="compact"
            variant="outlined"
            hide-details
            class="location-select"
            @update:model-value="onLocationChange"
          ></v-select>
        </div>

        <button
          class="cycle-toggle-btn"
          :class="{ active: autoCycleEnabled }"
          @click="toggleAutoCycle"
          title="Otomatis rotasi tampilan lokasi setiap 15 detik"
        >
          <v-icon size="16" class="mr-1">mdi-autorenew</v-icon>
          <span>Auto Rotate</span>
          <span v-if="autoCycleEnabled" class="cycle-indicator">{{ cycleCountdown }}s</span>
        </button>
      </div>

      <!-- Right Controls: Live Clock, Refresh, Fullscreen -->
      <div class="tv-header-right">
        <div class="clock-display">
          <span class="clock-date">{{ currentFormattedDate }}</span>
          <span class="clock-time">{{ currentFormattedTime }}</span>
        </div>

        <div class="header-actions">
          <button
            class="action-btn"
            :class="{ spinning: loading }"
            @click="fetchData(true)"
            title="Muat ulang data sekarang"
          >
            <v-icon size="18">mdi-refresh</v-icon>
            <span class="refresh-countdown">{{ refreshCountdown }}s</span>
          </button>

          <button
            class="action-btn"
            @click="toggleFullscreen"
            title="Mode Layar Penuh (Fullscreen)"
          >
            <v-icon size="20">{{ isFullscreen ? 'mdi-fullscreen-exit' : 'mdi-fullscreen' }}</v-icon>
          </button>

          <button
            class="action-btn close-btn"
            @click="navigateBack"
            title="Keluar ke Dashboard"
          >
            <v-icon size="20">mdi-close</v-icon>
          </button>
        </div>
      </div>
    </header>

    <!-- Main TV Body -->
    <main class="tv-main-content">
      <!-- Loading Bar -->
      <v-progress-linear
        v-if="loading"
        indeterminate
        color="#0284c7"
        height="2"
        class="tv-loader"
      ></v-progress-linear>

      <!-- Active Filter Banner (if filtering by specific location) -->
      <div v-if="selectedLocation && selectedLocation !== 'Semua Lokasi'" class="active-location-banner">
        <v-icon size="16" color="#38bdf8" class="mr-1">mdi-filter-check</v-icon>
        <span>Menampilkan data khusus lokasi: <strong>{{ selectedLocation }}</strong></span>
        <button class="reset-filter-btn" @click="resetToAllLocations">
          Reset ke Semua Lokasi
        </button>
      </div>

      <!-- ROW 1: 4 Key Metric Cards with Big Legible Numbers -->
      <section class="metrics-grid">
        <!-- CARD 1: TOTAL USERS -->
        <div class="metric-card card-users">
          <div class="metric-header">
            <span class="metric-tag">TOTAL PELANGGAN</span>
            <v-icon size="22" color="#38bdf8">mdi-account-group</v-icon>
          </div>
          <div class="metric-body">
            <div class="metric-big-number">{{ formatNumber(data?.total_users || 0) }}</div>
            <div class="metric-stats-row">
              <div class="stat-pill pill-active">
                <span class="pill-dot"></span>
                <span class="pill-label">Aktif:</span>
                <strong>{{ formatNumber(data?.active_users || 0) }}</strong>
              </div>
              <div class="stat-pill pill-suspended">
                <span class="pill-dot red"></span>
                <span class="pill-label">Suspended:</span>
                <strong>{{ formatNumber(data?.suspended_users || 0) }}</strong>
              </div>
              <div class="stat-pill pill-stopped">
                <span class="pill-label">Berhenti:</span>
                <strong>{{ formatNumber(data?.stopped_users || 0) }}</strong>
              </div>
            </div>
          </div>
          <div class="card-footer-info">
            {{ selectedLocation === 'Semua Lokasi' ? 'Keseluruhan Pelanggan Terdaftar' : `Pelanggan Wilayah ${selectedLocation}` }}
          </div>
        </div>

        <!-- CARD 2: BRAND BREAKDOWN (JAKINET, JELANTIK, NAGRAK) -->
        <div class="metric-card card-brands">
          <div class="metric-header">
            <span class="metric-tag">DISTRIBUSI BRAND</span>
            <v-icon size="22" color="#a855f7">mdi-domain</v-icon>
          </div>
          <div class="metric-body brands-body">
            <div
              v-for="brand in brandMetrics"
              :key="brand.name"
              class="brand-sub-box"
              :class="brand.class"
            >
              <div class="brand-sub-header">
                <span class="brand-name">{{ brand.name }}</span>
                <span class="brand-share">{{ getBrandPercentage(brand.total) }}%</span>
              </div>
              <div class="brand-sub-val">{{ formatNumber(brand.total) }}</div>
              <div class="brand-sub-footer">
                <span class="brand-active">Aktif: {{ formatNumber(brand.active) }}</span>
                <span v-if="brand.suspended > 0" class="brand-susp"> • Susp: {{ brand.suspended }}</span>
              </div>
            </div>
          </div>
          <div class="card-footer-info">
            Pemisahan Akun & Infrastruktur FTTH
          </div>
        </div>

        <!-- CARD 3: INVOICES (TOTAL, PAID, PENDING) -->
        <div class="metric-card card-invoices">
          <div class="metric-header">
            <span class="metric-tag">STATUS INVOICE TAGIHAN</span>
            <v-icon size="22" color="#10b981">mdi-receipt-text-outline</v-icon>
          </div>
          <div class="metric-body">
            <div class="invoice-primary-row">
              <div class="invoice-col">
                <span class="inv-col-label">Total Tagihan</span>
                <span class="inv-col-number">{{ formatNumber(data?.invoices?.total_count || 0) }}</span>
                <span class="inv-col-sub">{{ formatRupiahShort(data?.invoices?.total_amount || 0) }}</span>
              </div>
              <div class="invoice-divider"></div>
              <div class="invoice-col col-paid">
                <span class="inv-col-label text-success">Lunas (Paid)</span>
                <span class="inv-col-number text-success">{{ formatNumber(data?.invoices?.paid_count || 0) }}</span>
                <span class="inv-col-sub text-success">{{ formatRupiahShort(data?.invoices?.paid_amount || 0) }}</span>
              </div>
              <div class="invoice-divider"></div>
              <div class="invoice-col col-pending">
                <span class="inv-col-label text-warning">Pending (Unpaid)</span>
                <span class="inv-col-number text-warning">{{ formatNumber(data?.invoices?.unpaid_count || 0) }}</span>
                <span class="inv-col-sub text-warning">{{ formatRupiahShort(data?.invoices?.unpaid_amount || 0) }}</span>
              </div>
            </div>

            <!-- Progress bar lunas vs pending -->
            <div class="invoice-progress-container">
              <div class="progress-labels">
                <span>Rasio Lunas: <strong>{{ paidPercentage }}%</strong></span>
                <span v-if="(data?.invoices?.expired_count || 0) > 0">Expired: {{ data?.invoices?.expired_count }}</span>
              </div>
              <div class="progress-track">
                <div class="progress-fill" :style="{ width: `${paidPercentage}%` }"></div>
              </div>
            </div>
          </div>
          <div class="card-footer-info">
            Monitoring Pembayaran Pelanggan
          </div>
        </div>

        <!-- CARD 4: USER SUSPENDED (TERISOLIR) -->
        <div class="metric-card card-suspended" :class="{ 'has-alerts': (data?.suspended_users || 0) > 0 }">
          <div class="metric-header">
            <span class="metric-tag">USER TERISOLIR (SUSPENDED)</span>
            <v-icon size="22" color="#f43f5e">mdi-account-alert-outline</v-icon>
          </div>
          <div class="metric-body suspended-main-body">
            <div class="suspended-metric-container">
              <div class="metric-big-number text-danger">
                {{ formatNumber(data?.suspended_users || 0) }}
              </div>
              <div class="suspended-status-badge">
                <span v-if="(data?.suspended_users || 0) > 0" class="badge-alert">
                  <v-icon size="14" color="white" class="mr-1">mdi-alert-circle</v-icon>
                  Modem Terisolir
                </span>
                <span v-else class="badge-clean">
                  <v-icon size="14" color="white" class="mr-1">mdi-check-circle</v-icon>
                  Normal (0 Suspended)
                </span>
              </div>
            </div>

            <div class="suspended-meta">
              <span>Persentase Gangguan / Tunggakan:</span>
              <strong>{{ suspendedPercentage }}% dari total pelanggan</strong>
            </div>
          </div>
          <div class="card-footer-info">
            Pelanggan Belum Bayar Melewati Batas Jatuh Tempo
          </div>
        </div>
      </section>

      <!-- ROW 2: DETAILED DATA (LEFT: LOCATIONS TABLE, RIGHT: SUSPENDED FEED) -->
      <section class="details-grid">
        <!-- LEFT: BREAKDOWN PER LOKASI -->
        <div class="detail-card locations-card">
          <div class="detail-card-header">
            <div class="header-left">
              <v-icon size="20" color="#38bdf8" class="mr-2">mdi-map-marker-multiple</v-icon>
              <div>
                <h2 class="detail-title">Statistik Berdasarkan Lokasi</h2>
                <p class="detail-desc">Klik baris lokasi untuk memfilter monitoring tampilan ke lokasi tersebut</p>
              </div>
            </div>

            <div class="table-search-box">
              <v-icon size="16" color="#94a3b8">mdi-magnify</v-icon>
              <input
                v-model="locationSearch"
                type="text"
                placeholder="Cari lokasi..."
                class="search-input"
              />
            </div>
          </div>

          <div class="table-responsive">
            <table class="tv-table">
              <thead>
                <tr>
                  <th class="text-left">LOKASI / WILAYAH</th>
                  <th class="text-center">TOTAL USER</th>
                  <th class="text-center">AKTIF</th>
                  <th class="text-center">SUSPEND</th>
                  <th class="text-center">JAKINET</th>
                  <th class="text-center">JELANTIK</th>
                  <th class="text-center">NAGRAK</th>
                  <th class="text-center">INVOICE PAID</th>
                  <th class="text-center">UNPAID</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="loc in filteredLocations"
                  :key="loc.lokasi"
                  :class="{ selected: selectedLocation === loc.lokasi }"
                  @click="selectLocationRow(loc.lokasi)"
                >
                  <td class="text-left font-weight-medium">
                    <div class="loc-cell">
                      <v-icon size="16" color="#0284c7" class="mr-1">mdi-map-marker</v-icon>
                      <span>{{ loc.lokasi }}</span>
                    </div>
                  </td>
                  <td class="text-center font-weight-bold text-white">{{ loc.total_users }}</td>
                  <td class="text-center text-success font-weight-medium">{{ loc.active_users }}</td>
                  <td class="text-center">
                    <span :class="loc.suspended_users > 0 ? 'cell-badge-red' : 'cell-badge-muted'">
                      {{ loc.suspended_users }}
                    </span>
                  </td>
                  <td class="text-center text-sky">{{ loc.jakinet_users }}</td>
                  <td class="text-center text-violet">{{ loc.jelantik_users }}</td>
                  <td class="text-center text-amber">{{ loc.nagrak_users }}</td>
                  <td class="text-center text-success">{{ loc.paid_invoices }}</td>
                  <td class="text-center">
                    <span :class="loc.unpaid_invoices > 0 ? 'cell-badge-amber' : 'cell-badge-muted'">
                      {{ loc.unpaid_invoices }}
                    </span>
                  </td>
                </tr>
                <tr v-if="filteredLocations.length === 0">
                  <td colspan="9" class="text-center py-6 text-muted">
                    Tidak ada lokasi yang cocok dengan pencarian "{{ locationSearch }}"
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- RIGHT: SUSPENDED USERS FEED -->
        <div class="detail-card suspended-feed-card">
          <div class="detail-card-header">
            <div class="header-left">
              <v-icon size="20" color="#f43f5e" class="mr-2">mdi-account-alert</v-icon>
              <div>
                <h2 class="detail-title">Daftar Pelanggan Terisolir</h2>
                <p class="detail-desc">Total {{ data?.suspended_list?.length || 0 }} akun terisolir terdeteksi</p>
              </div>
            </div>
            <v-chip size="x-small" color="#f43f5e" variant="flat" class="font-weight-bold">
              LIVE NOC
            </v-chip>
          </div>

          <div class="suspended-list-container">
            <div v-if="data?.suspended_list && data.suspended_list.length > 0" class="suspended-items">
              <div
                v-for="user in data.suspended_list"
                :key="user.id"
                class="suspended-item"
              >
                <div class="item-left">
                  <div class="user-id-badge">{{ user.id_pelanggan || `PLG-${user.id}` }}</div>
                  <div class="user-info">
                    <span class="user-name">{{ user.nama }}</span>
                    <span class="user-loc">
                      <v-icon size="12" class="mr-1">mdi-map-marker</v-icon>
                      {{ user.alamat }}
                    </span>
                  </div>
                </div>
                <div class="item-right">
                  <span class="user-brand-tag">{{ user.brand }}</span>
                  <span class="user-plan">{{ user.paket }}</span>
                </div>
              </div>
            </div>
            <div v-else class="empty-suspended-state">
              <v-icon size="48" color="#10b981" class="mb-2">mdi-check-circle-outline</v-icon>
              <p class="empty-title">Semua Koneksi Aktif Normal</p>
              <p class="empty-desc">Tidak ada pelanggan berstatus suspended pada filter lokasi ini</p>
            </div>
          </div>
        </div>
      </section>
    </main>

    <!-- TV Footer Information -->
    <footer class="tv-footer">
      <div class="footer-left">
        <span class="status-indicator"></span>
        <span>Status Jaringan: <strong>SISTEM OPERASIONAL</strong></span>
        <span class="divider">•</span>
        <span>Update Terakhir: <strong>{{ lastUpdatedFormatted }}</strong></span>
      </div>
      <div class="footer-right">
        <span>Tekan <strong>F11</strong> atau ikon pojok kanan untuk Fullscreen TV</span>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import apiClient from '@/services/api';

interface TVBrandStats {
  brand: string;
  total_users: number;
  active_users: number;
  suspended_users: number;
}

interface TVLocationStats {
  lokasi: string;
  total_users: number;
  active_users: number;
  suspended_users: number;
  jakinet_users: number;
  jelantik_users: number;
  nagrak_users: number;
  total_invoices: number;
  paid_invoices: number;
  unpaid_invoices: number;
}

interface TVInvoiceStats {
  total_count: number;
  paid_count: number;
  unpaid_count: number;
  expired_count: number;
  total_amount: number;
  paid_amount: number;
  unpaid_amount: number;
}

interface TVSuspendedUser {
  id: number;
  id_pelanggan: string;
  nama: string;
  alamat: string;
  brand: string;
  paket: string;
  no_telp: string;
}

interface TVMonitoringData {
  total_users: number;
  active_users: number;
  suspended_users: number;
  stopped_users: number;
  brands: TVBrandStats[];
  invoices: TVInvoiceStats;
  locations: TVLocationStats[];
  suspended_list: TVSuspendedUser[];
  active_location: string;
  updated_at: string;
}

const router = useRouter();

// State
const data = ref<TVMonitoringData | null>(null);
const loading = ref(false);
const selectedLocation = ref<string>('Semua Lokasi');
const locationSearch = ref('');
const isFullscreen = ref(false);

// Auto-Refresh (30s)
const refreshIntervalSeconds = 30;
const refreshCountdown = ref(refreshIntervalSeconds);
let refreshTimer: any = null;

// Auto-Cycle Locations (15s)
const autoCycleEnabled = ref(false);
const cycleIntervalSeconds = 15;
const cycleCountdown = ref(cycleIntervalSeconds);
let cycleTimer: any = null;
let currentCycleIndex = 0;

// Clock
const currentTime = ref(new Date());
let clockTimer: any = null;

// Available Locations for dropdown
const availableLocations = computed(() => {
  const list = ['Semua Lokasi'];
  if (data.value?.locations) {
    data.value.locations.forEach((loc) => {
      if (loc.lokasi && !list.includes(loc.lokasi)) {
        list.push(loc.lokasi);
      }
    });
  }
  return list;
});

// Brand Metrics organized for Jakinet, Jelantik, and Jelantik Nagrak
const brandMetrics = computed(() => {
  const brands = data.value?.brands || [];
  const findBrand = (keyword: string) => {
    return brands.find((b) => b.brand.toLowerCase().includes(keyword.toLowerCase()));
  };

  const jakinet = findBrand('jakinet');
  const nagrak = findBrand('nagrak');
  const jelantik = brands.find(
    (b) => b.brand.toLowerCase().includes('jelantik') && !b.brand.toLowerCase().includes('nagrak')
  );

  return [
    {
      name: 'Jakinet',
      total: jakinet?.total_users || 0,
      active: jakinet?.active_users || 0,
      suspended: jakinet?.suspended_users || 0,
      class: 'brand-jakinet',
    },
    {
      name: 'Jelantik',
      total: jelantik?.total_users || 0,
      active: jelantik?.active_users || 0,
      suspended: jelantik?.suspended_users || 0,
      class: 'brand-jelantik',
    },
    {
      name: 'Jelantik Nagrak',
      total: nagrak?.total_users || 0,
      active: nagrak?.active_users || 0,
      suspended: nagrak?.suspended_users || 0,
      class: 'brand-nagrak',
    },
  ];
});

// Filtered locations based on search box
const filteredLocations = computed(() => {
  const list = data.value?.locations || [];
  if (!locationSearch.value.trim()) return list;
  const q = locationSearch.value.toLowerCase().trim();
  return list.filter((l) => l.lokasi.toLowerCase().includes(q));
});

// Percentages
const paidPercentage = computed(() => {
  const total = data.value?.invoices?.total_count || 0;
  if (total === 0) return 0;
  const paid = data.value?.invoices?.paid_count || 0;
  return Math.round((paid / total) * 100);
});

const suspendedPercentage = computed(() => {
  const total = data.value?.total_users || 0;
  if (total === 0) return 0;
  const susp = data.value?.suspended_users || 0;
  return ((susp / total) * 100).toFixed(1);
});

// Time and Date formatters
const currentFormattedDate = computed(() => {
  return currentTime.value.toLocaleDateString('id-ID', {
    weekday: 'long',
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  });
});

const currentFormattedTime = computed(() => {
  return currentTime.value.toLocaleTimeString('id-ID', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  }) + ' WIB';
});

const lastUpdatedFormatted = computed(() => {
  if (!data.value?.updated_at) return '-';
  const d = new Date(data.value.updated_at);
  return d.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' }) + ' WIB';
});

// Helper Formatters
const formatNumber = (num: number) => {
  return new Intl.NumberFormat('id-ID').format(num || 0);
};

const formatRupiahShort = (amount: number) => {
  if (!amount) return 'Rp 0';
  if (amount >= 1_000_000_000) {
    return `Rp ${(amount / 1_000_000_000).toFixed(2)} M`;
  }
  if (amount >= 1_000_000) {
    return `Rp ${(amount / 1_000_000).toFixed(1)} Jt`;
  }
  return `Rp ${new Intl.NumberFormat('id-ID').format(amount)}`;
};

const getBrandPercentage = (val: number) => {
  const total = data.value?.total_users || 0;
  if (total === 0) return 0;
  return Math.round((val / total) * 100);
};

// Data Fetching
const fetchData = async (showLoading = false) => {
  if (showLoading) loading.value = true;
  try {
    const locParam = selectedLocation.value === 'Semua Lokasi' ? '' : selectedLocation.value;
    const url = `/dashboard/tv-monitoring${locParam ? `?lokasi=${encodeURIComponent(locParam)}` : ''}`;
    const response = await apiClient.get<TVMonitoringData>(url);
    if (response?.data) {
      data.value = response.data;
    }
  } catch (err) {
    console.error('Failed to load TV monitoring data:', err);
  } finally {
    loading.value = false;
    refreshCountdown.value = refreshIntervalSeconds;
  }
};

const onLocationChange = (newLoc: string) => {
  selectedLocation.value = newLoc;
  fetchData(true);
};

const selectLocationRow = (lokasi: string) => {
  if (selectedLocation.value === lokasi) {
    selectedLocation.value = 'Semua Lokasi';
  } else {
    selectedLocation.value = lokasi;
  }
  fetchData(true);
};

const resetToAllLocations = () => {
  selectedLocation.value = 'Semua Lokasi';
  fetchData(true);
};

// Fullscreen Toggle
const toggleFullscreen = () => {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen().then(() => {
      isFullscreen.value = true;
    }).catch(err => {
      console.warn('Error attempting fullscreen:', err);
    });
  } else {
    if (document.exitFullscreen) {
      document.exitFullscreen().then(() => {
        isFullscreen.value = false;
      });
    }
  }
};

// Auto-Cycle Toggle
const toggleAutoCycle = () => {
  autoCycleEnabled.value = !autoCycleEnabled.value;
  cycleCountdown.value = cycleIntervalSeconds;
};

// Back Navigation
const navigateBack = () => {
  if (document.fullscreenElement) {
    document.exitFullscreen();
  }
  router.push('/dashboard');
};

// Lifecycle
onMounted(() => {
  fetchData(true);

  // Clock Timer
  clockTimer = setInterval(() => {
    currentTime.value = new Date();
  }, 1000);

  // Refresh Timer
  refreshTimer = setInterval(() => {
    refreshCountdown.value--;
    if (refreshCountdown.value <= 0) {
      fetchData(false);
    }
  }, 1000);

  // Auto-Cycle Timer
  cycleTimer = setInterval(() => {
    if (!autoCycleEnabled.value) return;
    cycleCountdown.value--;
    if (cycleCountdown.value <= 0) {
      cycleCountdown.value = cycleIntervalSeconds;
      const locs = availableLocations.value;
      if (locs.length > 1) {
        currentCycleIndex = (currentCycleIndex + 1) % locs.length;
        selectedLocation.value = locs[currentCycleIndex];
        fetchData(false);
      }
    }
  }, 1000);

  // Handle Fullscreen change events
  document.addEventListener('fullscreenchange', () => {
    isFullscreen.value = !!document.fullscreenElement;
  });
});

onUnmounted(() => {
  if (clockTimer) clearInterval(clockTimer);
  if (refreshTimer) clearInterval(refreshTimer);
  if (cycleTimer) clearInterval(cycleTimer);
});
</script>

<style scoped>
/* Base Theme: Professional Deep NOC Dark Slate (No generic blue-purple gradient) */
.tv-monitoring-root {
  min-height: 100vh;
  background-color: #080c14;
  color: #f8fafc;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  display: flex;
  flex-direction: column;
  padding: 0;
  margin: 0;
  overflow-x: hidden;
  box-sizing: border-box;
}

/* Header */
.tv-header {
  height: 68px;
  background-color: #0f172a;
  border-bottom: 1px solid #1e293b;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  z-index: 20;
}

.tv-header-left .brand-badge {
  display: flex;
  align-items: center;
}

.brand-titles .tv-title {
  font-size: 1.15rem;
  font-weight: 800;
  letter-spacing: 0.05em;
  color: #f8fafc;
  margin: 0;
  line-height: 1.2;
}

.brand-titles .tv-subtitle {
  font-size: 0.72rem;
  color: #94a3b8;
  margin: 0;
  letter-spacing: 0.02em;
}

.tv-header-center {
  display: flex;
  align-items: center;
  gap: 16px;
}

.location-control {
  display: flex;
  align-items: center;
  background-color: #1e293b;
  padding: 4px 10px;
  border-radius: 6px;
  border: 1px solid #334155;
}

.location-label {
  font-size: 0.8rem;
  color: #94a3b8;
  font-weight: 600;
  margin-right: 8px;
}

.location-select {
  width: 190px;
  font-size: 0.82rem;
}

.location-select :deep(.v-field__input) {
  padding: 2px 8px;
  min-height: 30px;
  color: #f8fafc;
  font-size: 0.82rem;
}

.location-select :deep(.v-field__outline) {
  display: none;
}

.cycle-toggle-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  background-color: #1e293b;
  border: 1px solid #334155;
  color: #94a3b8;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.cycle-toggle-btn.active {
  background-color: rgba(2, 132, 199, 0.2);
  border-color: #0284c7;
  color: #38bdf8;
}

.cycle-indicator {
  background-color: #0284c7;
  color: white;
  padding: 1px 6px;
  border-radius: 10px;
  font-size: 0.7rem;
  margin-left: 4px;
}

.tv-header-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

.clock-display {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  border-right: 1px solid #334155;
  padding-right: 18px;
}

.clock-date {
  font-size: 0.72rem;
  color: #94a3b8;
  font-weight: 500;
}

.clock-time {
  font-size: 1.1rem;
  font-weight: 700;
  color: #38bdf8;
  font-family: monospace;
  letter-spacing: 0.05em;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.action-btn {
  background-color: #1e293b;
  border: 1px solid #334155;
  color: #f8fafc;
  width: 36px;
  height: 36px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  position: relative;
  transition: background-color 0.2s;
}

.action-btn:hover {
  background-color: #334155;
}

.refresh-countdown {
  position: absolute;
  bottom: -6px;
  right: -4px;
  font-size: 0.6rem;
  background-color: #0284c7;
  color: white;
  padding: 0 4px;
  border-radius: 4px;
  font-weight: bold;
}

.action-btn.close-btn:hover {
  background-color: #dc2626;
  border-color: #ef4444;
}

/* Main Content Area */
.tv-main-content {
  flex: 1;
  padding: 18px 24px;
  display: flex;
  flex-direction: column;
  gap: 18px;
  position: relative;
}

.tv-loader {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  z-index: 10;
}

.active-location-banner {
  background-color: rgba(2, 132, 199, 0.15);
  border: 1px solid rgba(56, 189, 248, 0.4);
  color: #e0f2fe;
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 0.82rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.reset-filter-btn {
  background: none;
  border: 1px solid #38bdf8;
  color: #38bdf8;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 0.72rem;
  cursor: pointer;
}

.reset-filter-btn:hover {
  background-color: #38bdf8;
  color: #0f172a;
}

/* Metrics Row (4 Cards) */
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 18px;
}

.metric-card {
  background-color: #0f172a;
  border: 1px solid #1e293b;
  border-radius: 10px;
  padding: 16px 18px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.2);
}

.metric-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.metric-tag {
  font-size: 0.75rem;
  font-weight: 700;
  color: #94a3b8;
  letter-spacing: 0.06em;
}

.metric-big-number {
  font-size: 2.75rem;
  font-weight: 800;
  line-height: 1.1;
  color: #f8fafc;
  letter-spacing: -0.02em;
}

.card-footer-info {
  margin-top: 14px;
  padding-top: 8px;
  border-top: 1px solid #1e293b;
  font-size: 0.7rem;
  color: #64748b;
  font-weight: 500;
}

/* Card 1: Users Details */
.metric-stats-row {
  display: flex;
  gap: 8px;
  margin-top: 10px;
  flex-wrap: wrap;
}

.stat-pill {
  display: flex;
  align-items: center;
  gap: 5px;
  background-color: #1e293b;
  padding: 3px 8px;
  border-radius: 5px;
  font-size: 0.72rem;
}

.pill-dot {
  width: 6px;
  height: 6px;
  background-color: #10b981;
  border-radius: 50%;
}

.pill-dot.red {
  background-color: #f43f5e;
}

.pill-label {
  color: #94a3b8;
}

/* Card 2: Brands Sub-Boxes */
.brands-body {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

.brand-sub-box {
  background-color: #1e293b;
  border: 1px solid #334155;
  border-radius: 6px;
  padding: 8px 10px;
}

.brand-sub-box.brand-jakinet {
  border-top: 3px solid #0284c7;
}

.brand-sub-box.brand-jelantik {
  border-top: 3px solid #8b5cf6;
}

.brand-sub-box.brand-nagrak {
  border-top: 3px solid #f59e0b;
}

.brand-sub-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.brand-name {
  font-size: 0.68rem;
  font-weight: 700;
  color: #cbd5e1;
}

.brand-share {
  font-size: 0.65rem;
  color: #94a3b8;
}

.brand-sub-val {
  font-size: 1.25rem;
  font-weight: 800;
  color: #f8fafc;
  margin: 4px 0;
}

.brand-sub-footer {
  font-size: 0.65rem;
  color: #94a3b8;
}

.brand-susp {
  color: #f43f5e;
  font-weight: 600;
}

/* Card 3: Invoices */
.invoice-primary-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #1e293b;
  padding: 10px 12px;
  border-radius: 6px;
  margin-bottom: 10px;
}

.invoice-col {
  display: flex;
  flex-direction: column;
}

.inv-col-label {
  font-size: 0.68rem;
  color: #94a3b8;
  font-weight: 600;
}

.inv-col-number {
  font-size: 1.35rem;
  font-weight: 800;
  color: #f8fafc;
}

.inv-col-sub {
  font-size: 0.68rem;
  color: #64748b;
  font-weight: 600;
}

.invoice-divider {
  width: 1px;
  height: 36px;
  background-color: #334155;
}

.invoice-progress-container {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.progress-labels {
  display: flex;
  justify-content: space-between;
  font-size: 0.72rem;
  color: #94a3b8;
}

.progress-track {
  width: 100%;
  height: 6px;
  background-color: #334155;
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background-color: #10b981;
  transition: width 0.4s ease;
}

/* Card 4: Suspended Alert */
.card-suspended.has-alerts {
  border-color: rgba(244, 63, 94, 0.4);
}

.suspended-metric-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.text-danger {
  color: #f43f5e;
}

.suspended-status-badge {
  display: flex;
}

.badge-alert {
  display: flex;
  align-items: center;
  background-color: #e11d48;
  color: white;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 700;
}

.badge-clean {
  display: flex;
  align-items: center;
  background-color: #059669;
  color: white;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 700;
}

.suspended-meta {
  font-size: 0.72rem;
  color: #94a3b8;
  margin-top: 8px;
}

/* ROW 2: DETAILS GRID */
.details-grid {
  display: grid;
  grid-template-columns: 62% 38%;
  gap: 18px;
  flex: 1;
}

.detail-card {
  background-color: #0f172a;
  border: 1px solid #1e293b;
  border-radius: 10px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.2);
}

.detail-card-header {
  padding: 14px 18px;
  border-bottom: 1px solid #1e293b;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #111e36;
}

.detail-card-header .header-left {
  display: flex;
  align-items: center;
}

.detail-title {
  font-size: 0.95rem;
  font-weight: 700;
  color: #f8fafc;
  margin: 0;
  line-height: 1.2;
}

.detail-desc {
  font-size: 0.72rem;
  color: #94a3b8;
  margin: 0;
}

/* Locations Table */
.table-search-box {
  display: flex;
  align-items: center;
  background-color: #1e293b;
  border: 1px solid #334155;
  border-radius: 6px;
  padding: 4px 10px;
  width: 170px;
}

.search-input {
  background: none;
  border: none;
  color: #f8fafc;
  font-size: 0.75rem;
  margin-left: 6px;
  outline: none;
  width: 100%;
}

.table-responsive {
  flex: 1;
  overflow-y: auto;
  max-height: 480px;
}

.tv-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.78rem;
}

.tv-table th {
  background-color: #1e293b;
  color: #94a3b8;
  font-weight: 700;
  font-size: 0.7rem;
  letter-spacing: 0.05em;
  padding: 10px 12px;
  border-bottom: 1px solid #334155;
  position: sticky;
  top: 0;
  z-index: 5;
}

.tv-table td {
  padding: 9px 12px;
  border-bottom: 1px solid #1e293b;
  color: #cbd5e1;
}

.tv-table tr {
  cursor: pointer;
  transition: background-color 0.15s ease;
}

.tv-table tr:hover {
  background-color: #1e293b;
}

.tv-table tr.selected {
  background-color: rgba(2, 132, 199, 0.2);
  border-left: 3px solid #0284c7;
}

.loc-cell {
  display: flex;
  align-items: center;
}

.cell-badge-red {
  background-color: rgba(244, 63, 94, 0.2);
  color: #f43f5e;
  padding: 2px 7px;
  border-radius: 4px;
  font-weight: 700;
}

.cell-badge-amber {
  background-color: rgba(245, 158, 11, 0.2);
  color: #f59e0b;
  padding: 2px 7px;
  border-radius: 4px;
  font-weight: 700;
}

.cell-badge-muted {
  color: #64748b;
}

/* Suspended Users Feed */
.suspended-list-container {
  flex: 1;
  overflow-y: auto;
  max-height: 480px;
  padding: 10px 14px;
}

.suspended-items {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.suspended-item {
  background-color: #1e293b;
  border: 1px solid #334155;
  border-left: 3px solid #f43f5e;
  border-radius: 6px;
  padding: 8px 12px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.suspended-item .item-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.user-id-badge {
  background-color: #0f172a;
  border: 1px solid #475569;
  color: #38bdf8;
  font-family: monospace;
  font-size: 0.72rem;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
}

.user-info {
  display: flex;
  flex-direction: column;
}

.user-name {
  font-size: 0.82rem;
  font-weight: 700;
  color: #f8fafc;
}

.user-loc {
  font-size: 0.7rem;
  color: #94a3b8;
  display: flex;
  align-items: center;
}

.item-right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
}

.user-brand-tag {
  font-size: 0.65rem;
  font-weight: 700;
  background-color: #334155;
  color: #e2e8f0;
  padding: 1px 6px;
  border-radius: 3px;
}

.user-plan {
  font-size: 0.68rem;
  color: #94a3b8;
}

.empty-suspended-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 280px;
  text-align: center;
}

.empty-title {
  font-size: 0.95rem;
  font-weight: 700;
  color: #10b981;
  margin: 0;
}

.empty-desc {
  font-size: 0.75rem;
  color: #94a3b8;
  margin: 4px 0 0 0;
}

/* Footer */
.tv-footer {
  height: 38px;
  background-color: #0f172a;
  border-top: 1px solid #1e293b;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  font-size: 0.72rem;
  color: #94a3b8;
}

.footer-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-indicator {
  width: 8px;
  height: 8px;
  background-color: #10b981;
  border-radius: 50%;
  box-shadow: 0 0 6px rgba(16, 185, 129, 0.8);
}

.divider {
  color: #475569;
}

/* Color Helpers */
.text-success { color: #10b981 !important; }
.text-warning { color: #f59e0b !important; }
.text-sky { color: #38bdf8 !important; }
.text-violet { color: #a855f7 !important; }
.text-amber { color: #f59e0b !important; }
.text-muted { color: #64748b !important; }

/* Fullscreen adjustments */
.is-fullscreen {
  width: 100vw;
  height: 100vh;
}

/* Responsive adjustments */
@media (max-width: 1200px) {
  .metrics-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .details-grid {
    grid-template-columns: 1fr;
  }
}
</style>
