<template>
  <div class="tv-monitoring-root" :class="{ 'is-fullscreen': isFullscreen }">
    <!-- Top Bar / Header for NOC Wallboard (Clean Light Theme) -->
    <header class="tv-header">
      <div class="tv-header-left">
        <div class="brand-badge">
          <div class="brand-icon-box">
            <v-icon color="#0284c7" size="22">mdi-access-point-network</v-icon>
          </div>
          <div class="brand-titles">
            <h1 class="tv-title">NOC & BILLING WALLBOARD</h1>
            <p class="tv-subtitle">PT Artacomindo Jejaring Nusa</p>
          </div>
        </div>
      </div>

      <!-- Center Controls: Location Filter & Auto Cycle -->
      <div class="tv-header-center">
        <div class="location-control">
          <v-icon size="17" color="#64748b" class="mr-1">mdi-map-marker-radius</v-icon>
          <span class="location-label">Lokasi:</span>
          <select
            v-model="selectedLocation"
            class="custom-location-select"
            @change="onLocationSelectChange"
          >
            <option v-for="loc in availableLocations" :key="loc" :value="loc">
              {{ loc }}
            </option>
          </select>
        </div>

        <button
          class="cycle-toggle-btn"
          :class="{ active: autoCycleEnabled }"
          @click="toggleAutoCycle"
          title="Otomatis rotasi tampilan lokasi setiap 15 detik"
        >
          <v-icon size="15" class="mr-1">mdi-autorenew</v-icon>
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
        height="3"
        class="tv-loader"
      ></v-progress-linear>

      <!-- Active Filter Banner (if filtering by specific location) -->
      <div v-if="selectedLocation && selectedLocation !== 'Semua Lokasi'" class="active-location-banner">
        <div class="banner-left">
          <v-icon size="16" color="#0284c7" class="mr-2">mdi-filter-check</v-icon>
          <span>Menampilkan data wilayah: <strong>{{ selectedLocation }}</strong></span>
        </div>
        <button class="reset-filter-btn" @click="resetToAllLocations">
          <v-icon size="14" class="mr-1">mdi-close-circle-outline</v-icon>
          Kembali ke Semua Lokasi
        </button>
      </div>

      <!-- ROW 1: 4 Key Metric Cards (No filler text, No currency nominals) -->
      <section class="metrics-grid">
        <!-- CARD 1: TOTAL USERS -->
        <div class="metric-card card-users">
          <div class="metric-header">
            <span class="metric-tag">TOTAL PELANGGAN</span>
            <v-icon size="20" color="#0284c7">mdi-account-group</v-icon>
          </div>
          <div class="metric-body">
            <div class="metric-big-number">{{ formatNumber(data?.total_users || 0) }}</div>
            <div class="metric-stats-row">
              <div class="stat-pill pill-active">
                <span class="pill-dot green"></span>
                <span class="pill-label">Aktif:</span>
                <strong>{{ formatNumber(data?.active_users || 0) }}</strong>
              </div>
              <div class="stat-pill pill-suspended">
                <span class="pill-dot red"></span>
                <span class="pill-label">Suspended:</span>
                <strong>{{ formatNumber(data?.suspended_users || 0) }}</strong>
              </div>
              <div class="stat-pill pill-stopped">
                <span class="pill-dot gray"></span>
                <span class="pill-label">Berhenti:</span>
                <strong>{{ formatNumber(data?.stopped_users || 0) }}</strong>
              </div>
            </div>
          </div>
        </div>

        <!-- CARD 2: BRAND BREAKDOWN (JAKINET, JELANTIK, NAGRAK) -->
        <div class="metric-card card-brands">
          <div class="metric-header">
            <span class="metric-tag">DISTRIBUSI BRAND</span>
            <v-icon size="20" color="#7c3aed">mdi-domain</v-icon>
          </div>
          <div class="metric-body brands-body">
            <!-- Mini Segmented Bar -->
            <div class="brand-bar-track">
              <div
                class="brand-bar-segment seg-jakinet"
                :style="{ width: `${getBrandPercentage(brandMetrics[0].total)}%` }"
                title="Jakinet"
              ></div>
              <div
                class="brand-bar-segment seg-jelantik"
                :style="{ width: `${getBrandPercentage(brandMetrics[1].total)}%` }"
                title="Jelantik"
              ></div>
              <div
                class="brand-bar-segment seg-nagrak"
                :style="{ width: `${getBrandPercentage(brandMetrics[2].total)}%` }"
                title="Jelantik Nagrak"
              ></div>
            </div>

            <div class="brand-columns-row">
              <div
                v-for="brand in brandMetrics"
                :key="brand.name"
                class="brand-col"
              >
                <div class="brand-col-header">
                  <span class="brand-col-name">{{ brand.name }}</span>
                  <span class="brand-col-share">{{ getBrandPercentage(brand.total) }}%</span>
                </div>
                <div class="brand-col-val">{{ formatNumber(brand.total) }}</div>
                <div class="brand-col-sub">
                  <span>Aktif: <strong>{{ formatNumber(brand.active) }}</strong></span>
                  <span v-if="brand.suspended > 0" class="brand-col-susp"> • Susp: {{ brand.suspended }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- CARD 3: INVOICES (TOTAL, PAID, PENDING - QUANTITY ONLY, NO RUPIAH) -->
        <div class="metric-card card-invoices">
          <div class="metric-header">
            <span class="metric-tag">STATUS INVOICE TAGIHAN</span>
            <v-icon size="20" color="#059669">mdi-receipt-text-outline</v-icon>
          </div>
          <div class="metric-body">
            <div class="invoice-primary-row">
              <div class="invoice-col">
                <span class="inv-col-label">Total Tagihan</span>
                <span class="inv-col-number">{{ formatNumber(data?.invoices?.total_count || 0) }}</span>
              </div>
              <div class="invoice-divider"></div>
              <div class="invoice-col">
                <span class="inv-col-label text-success">Lunas (Paid)</span>
                <span class="inv-col-number text-success">{{ formatNumber(data?.invoices?.paid_count || 0) }}</span>
              </div>
              <div class="invoice-divider"></div>
              <div class="invoice-col">
                <span class="inv-col-label" :class="(data?.invoices?.unpaid_count || 0) > 0 ? 'text-warning' : 'text-muted'">
                  Pending (Unpaid)
                </span>
                <span class="inv-col-number" :class="(data?.invoices?.unpaid_count || 0) > 0 ? 'text-warning' : 'text-muted'">
                  {{ formatNumber(data?.invoices?.unpaid_count || 0) }}
                </span>
              </div>
            </div>

            <!-- Progress bar lunas vs pending -->
            <div class="invoice-progress-container">
              <div class="progress-labels">
                <span>Pelunasan: <strong>{{ paidPercentage }}%</strong></span>
                <span v-if="(data?.invoices?.expired_count || 0) > 0" class="text-danger">
                  Expired: {{ data?.invoices?.expired_count }}
                </span>
              </div>
              <div class="progress-track">
                <div class="progress-fill" :style="{ width: `${paidPercentage}%` }"></div>
              </div>
            </div>
          </div>
        </div>

        <!-- CARD 4: USER SUSPENDED (TERISOLIR) -->
        <div class="metric-card card-suspended" :class="{ 'has-alerts': (data?.suspended_users || 0) > 0 }">
          <div class="metric-header">
            <span class="metric-tag text-danger">PELANGGAN TERISOLIR</span>
            <v-icon size="20" color="#dc2626">mdi-account-alert-outline</v-icon>
          </div>
          <div class="metric-body">
            <div class="suspended-metric-container">
              <div class="metric-big-number text-danger">
                {{ formatNumber(data?.suspended_users || 0) }}
              </div>
              <div class="suspended-meta-tag">
                <span v-if="(data?.suspended_users || 0) > 0" class="status-chip-alert">
                  {{ suspendedPercentage }}% Pelanggan
                </span>
                <span v-else class="status-chip-clean">
                  100% Normal
                </span>
              </div>
            </div>

            <div class="suspended-note">
              <span v-if="(data?.suspended_users || 0) > 0">
                Menunggu pelunasan melewati jatuh tempo
              </span>
              <span v-else class="text-success">
                Tidak ada koneksi modem yang diisolir
              </span>
            </div>
          </div>
        </div>
      </section>

      <!-- ROW 2: DETAILED DATA (LEFT: LOCATIONS TABLE, RIGHT: SUSPENDED FEED) -->
      <section class="details-grid">
        <!-- LEFT: BREAKDOWN PER LOKASI -->
        <div class="detail-card locations-card">
          <div class="detail-card-header">
            <div class="header-left">
              <v-icon size="18" color="#0284c7" class="mr-2">mdi-map-marker-multiple</v-icon>
              <h2 class="detail-title">Statistik Berdasarkan Lokasi</h2>
            </div>

            <div class="table-search-box">
              <v-icon size="16" color="#64748b">mdi-magnify</v-icon>
              <input
                v-model="locationSearch"
                type="text"
                placeholder="Cari lokasi..."
                class="search-input"
              />
              <button v-if="locationSearch" class="clear-search-btn" @click="locationSearch = ''">
                <v-icon size="14">mdi-close</v-icon>
              </button>
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
                      <v-icon size="15" color="#0284c7" class="mr-1">mdi-map-marker</v-icon>
                      <span class="loc-name">{{ loc.lokasi }}</span>
                    </div>
                  </td>
                  <td class="text-center font-weight-bold text-slate">{{ loc.total_users }}</td>
                  <td class="text-center text-success font-weight-semibold">{{ loc.active_users }}</td>
                  <td class="text-center">
                    <span :class="loc.suspended_users > 0 ? 'badge-danger-pill' : 'text-muted'">
                      {{ loc.suspended_users }}
                    </span>
                  </td>
                  <td class="text-center text-sky">{{ loc.jakinet_users }}</td>
                  <td class="text-center text-violet">{{ loc.jelantik_users }}</td>
                  <td class="text-center text-amber">{{ loc.nagrak_users }}</td>
                  <td class="text-center text-success">{{ loc.paid_invoices }}</td>
                  <td class="text-center">
                    <span :class="loc.unpaid_invoices > 0 ? 'badge-amber-pill' : 'text-muted'">
                      {{ loc.unpaid_invoices }}
                    </span>
                  </td>
                </tr>
                <tr v-if="filteredLocations.length === 0">
                  <td colspan="9" class="text-center py-8 text-muted">
                    Tidak ada lokasi yang cocok dengan kata kunci "{{ locationSearch }}"
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
              <v-icon size="18" color="#dc2626" class="mr-2">mdi-account-alert</v-icon>
              <h2 class="detail-title">Daftar Pelanggan Terisolir ({{ data?.suspended_list?.length || 0 }})</h2>
            </div>
            <span class="live-pill">
              <span class="live-pulse"></span>
              LIVE
            </span>
          </div>

          <div class="suspended-list-container">
            <div v-if="data?.suspended_list && data.suspended_list.length > 0" class="suspended-items">
              <div
                v-for="user in data.suspended_list"
                :key="user.id"
                class="suspended-item"
              >
                <div class="item-left">
                  <span class="user-id-code">{{ user.id_pelanggan || `PLG-${user.id}` }}</span>
                  <div class="user-details">
                    <span class="user-fullname">{{ user.nama }}</span>
                    <span class="user-address">
                      <v-icon size="12" class="mr-1" color="#94a3b8">mdi-map-marker</v-icon>
                      {{ user.alamat }}
                    </span>
                  </div>
                </div>
                <div class="item-right">
                  <span class="brand-tag-clean">{{ user.brand }}</span>
                  <span class="plan-tag-clean">{{ user.paket }}</span>
                </div>
              </div>
            </div>
            <div v-else class="empty-suspended-state">
              <v-icon size="42" color="#059669" class="mb-2">mdi-check-circle-outline</v-icon>
              <p class="empty-title">Semua Koneksi Normal</p>
              <p class="empty-desc">Tidak ada pelanggan berstatus terisolir pada filter ini</p>
            </div>
          </div>
        </div>
      </section>
    </main>

    <!-- TV Footer Information -->
    <footer class="tv-footer">
      <div class="footer-left">
        <span class="status-dot-active"></span>
        <span>Sistem Monitoring Aktif</span>
        <span class="footer-separator">•</span>
        <span>Pembaruan Data: <strong>{{ lastUpdatedFormatted }}</strong></span>
      </div>
      <div class="footer-right">
        <span>Tekan <strong>F11</strong> untuk Fullscreen Layar TV</span>
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
    },
    {
      name: 'Jelantik',
      total: jelantik?.total_users || 0,
      active: jelantik?.active_users || 0,
      suspended: jelantik?.suspended_users || 0,
    },
    {
      name: 'Jelantik Nagrak',
      total: nagrak?.total_users || 0,
      active: nagrak?.active_users || 0,
      suspended: nagrak?.suspended_users || 0,
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
  if (total === 0) return '0.0';
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

const onLocationSelectChange = () => {
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
    }).catch((err) => {
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
/* Base: Clean, Modern, Crisp Light Enterprise Theme (No forced dark mode) */
.tv-monitoring-root {
  min-height: 100vh;
  background-color: #f1f5f9;
  color: #1e293b;
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
  height: 64px;
  background-color: #ffffff;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  z-index: 20;
}

.tv-header-left .brand-badge {
  display: flex;
  align-items: center;
  gap: 12px;
}

.brand-icon-box {
  width: 38px;
  height: 38px;
  background-color: #e0f2fe;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.brand-titles .tv-title {
  font-size: 1.05rem;
  font-weight: 800;
  letter-spacing: 0.03em;
  color: #0f172a;
  margin: 0;
  line-height: 1.2;
}

.brand-titles .tv-subtitle {
  font-size: 0.72rem;
  color: #64748b;
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
  background-color: #f8fafc;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  padding: 4px 10px;
}

.location-label {
  font-size: 0.78rem;
  font-weight: 600;
  color: #475569;
  margin-right: 8px;
}

.custom-location-select {
  background: transparent;
  border: none;
  font-size: 0.8rem;
  font-weight: 600;
  color: #0f172a;
  outline: none;
  cursor: pointer;
  min-width: 150px;
}

.cycle-toggle-btn {
  display: flex;
  align-items: center;
  background-color: #f8fafc;
  border: 1px solid #cbd5e1;
  color: #475569;
  padding: 5px 12px;
  border-radius: 6px;
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.cycle-toggle-btn:hover {
  background-color: #f1f5f9;
}

.cycle-toggle-btn.active {
  background-color: #0284c7;
  border-color: #0284c7;
  color: #ffffff;
}

.cycle-indicator {
  background-color: rgba(255, 255, 255, 0.25);
  padding: 1px 6px;
  border-radius: 4px;
  margin-left: 6px;
  font-size: 0.7rem;
}

/* Header Right */
.tv-header-right {
  display: flex;
  align-items: center;
  gap: 18px;
}

.clock-display {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  line-height: 1.2;
}

.clock-date {
  font-size: 0.7rem;
  font-weight: 500;
  color: #64748b;
}

.clock-time {
  font-size: 1rem;
  font-weight: 800;
  color: #0f172a;
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.04em;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  color: #475569;
  width: 36px;
  height: 36px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
  position: relative;
}

.action-btn:hover {
  background-color: #e2e8f0;
  color: #0f172a;
}

.refresh-countdown {
  position: absolute;
  bottom: -4px;
  right: -4px;
  background-color: #0284c7;
  color: #ffffff;
  font-size: 0.58rem;
  font-weight: 700;
  padding: 1px 4px;
  border-radius: 4px;
}

.action-btn.spinning v-icon {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.close-btn:hover {
  background-color: #fee2e2;
  color: #dc2626;
  border-color: #fca5a5;
}

/* Main TV Content */
.tv-main-content {
  flex: 1;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  position: relative;
}

.tv-loader {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
}

/* Active location banner */
.active-location-banner {
  background-color: #e0f2fe;
  border: 1px solid #bae6fd;
  color: #0369a1;
  padding: 8px 16px;
  border-radius: 6px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.82rem;
}

.reset-filter-btn {
  display: flex;
  align-items: center;
  background-color: #0284c7;
  color: #ffffff;
  border: none;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
}

.reset-filter-btn:hover {
  background-color: #0369a1;
}

/* ROW 1: METRICS GRID */
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
}

.metric-card {
  background-color: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.metric-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.metric-tag {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  color: #64748b;
}

.metric-big-number {
  font-size: 2.1rem;
  font-weight: 800;
  line-height: 1.1;
  color: #0f172a;
  font-variant-numeric: tabular-nums;
  margin-bottom: 8px;
}

/* Card 1: Users Breakdown */
.metric-stats-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.stat-pill {
  display: flex;
  align-items: center;
  gap: 5px;
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 5px;
  padding: 3px 8px;
  font-size: 0.74rem;
  color: #475569;
}

.pill-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.pill-dot.green { background-color: #10b981; }
.pill-dot.red { background-color: #ef4444; }
.pill-dot.gray { background-color: #94a3b8; }

/* Card 2: Brands Breakdown */
.brand-bar-track {
  width: 100%;
  height: 6px;
  background-color: #e2e8f0;
  border-radius: 3px;
  display: flex;
  overflow: hidden;
  margin-bottom: 10px;
}

.brand-bar-segment {
  height: 100%;
  transition: width 0.3s ease;
}

.seg-jakinet { background-color: #0284c7; }
.seg-jelantik { background-color: #7c3aed; }
.seg-nagrak { background-color: #d97706; }

.brand-columns-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

.brand-col {
  display: flex;
  flex-direction: column;
}

.brand-col-header {
  display: flex;
  justify-content: space-between;
  font-size: 0.7rem;
  color: #64748b;
  font-weight: 600;
}

.brand-col-name {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.brand-col-share {
  font-weight: 700;
  color: #0f172a;
}

.brand-col-val {
  font-size: 1.15rem;
  font-weight: 800;
  color: #0f172a;
  font-variant-numeric: tabular-nums;
  line-height: 1.2;
}

.brand-col-sub {
  font-size: 0.68rem;
  color: #64748b;
}

.brand-col-susp {
  color: #dc2626;
  font-weight: 600;
}

/* Card 3: Invoices (Quantities only) */
.invoice-primary-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.invoice-col {
  display: flex;
  flex-direction: column;
}

.inv-col-label {
  font-size: 0.7rem;
  font-weight: 600;
  color: #64748b;
}

.inv-col-number {
  font-size: 1.35rem;
  font-weight: 800;
  color: #0f172a;
  font-variant-numeric: tabular-nums;
  line-height: 1.2;
}

.invoice-divider {
  width: 1px;
  height: 32px;
  background-color: #e2e8f0;
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
  color: #64748b;
}

.progress-track {
  width: 100%;
  height: 6px;
  background-color: #e2e8f0;
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background-color: #059669;
  transition: width 0.4s ease;
}

/* Card 4: Suspended */
.card-suspended.has-alerts {
  border-color: #fca5a5;
  background-color: #fffafb;
}

.suspended-metric-container {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.status-chip-alert {
  background-color: #fee2e2;
  color: #dc2626;
  font-size: 0.74rem;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: 4px;
}

.status-chip-clean {
  background-color: #ecfdf5;
  color: #059669;
  font-size: 0.74rem;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: 4px;
}

.suspended-note {
  font-size: 0.74rem;
  color: #64748b;
  margin-top: 4px;
}

/* ROW 2: DETAILS GRID */
.details-grid {
  display: grid;
  grid-template-columns: 63% 37%;
  gap: 14px;
  flex: 1;
}

.detail-card {
  background-color: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.detail-card-header {
  padding: 12px 16px;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #f8fafc;
}

.detail-card-header .header-left {
  display: flex;
  align-items: center;
}

.detail-title {
  font-size: 0.92rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0;
}

.live-pill {
  display: flex;
  align-items: center;
  gap: 5px;
  background-color: #fee2e2;
  color: #dc2626;
  font-size: 0.68rem;
  font-weight: 800;
  padding: 2px 7px;
  border-radius: 4px;
  letter-spacing: 0.05em;
}

.live-pulse {
  width: 6px;
  height: 6px;
  background-color: #dc2626;
  border-radius: 50%;
  box-shadow: 0 0 4px #dc2626;
}

/* Table */
.table-search-box {
  display: flex;
  align-items: center;
  background-color: #ffffff;
  border: 1px solid #cbd5e1;
  border-radius: 5px;
  padding: 3px 8px;
  width: 180px;
}

.search-input {
  background: none;
  border: none;
  color: #0f172a;
  font-size: 0.75rem;
  margin-left: 6px;
  outline: none;
  width: 100%;
}

.clear-search-btn {
  background: none;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  padding: 0;
}

.table-responsive {
  flex: 1;
  overflow-y: auto;
  max-height: 490px;
}

.tv-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8rem;
  background-color: #ffffff;
}

.tv-table th {
  background-color: #f8fafc;
  color: #475569;
  font-weight: 700;
  font-size: 0.7rem;
  letter-spacing: 0.04em;
  padding: 10px 12px;
  border-bottom: 1px solid #e2e8f0;
  position: sticky;
  top: 0;
  z-index: 5;
}

.tv-table td {
  padding: 9px 12px;
  border-bottom: 1px solid #f1f5f9;
  color: #334155;
  font-variant-numeric: tabular-nums;
}

.tv-table tbody tr {
  background-color: #ffffff;
  cursor: pointer;
  transition: background-color 0.1s ease;
}

.tv-table tbody tr:hover {
  background-color: #f8fafc;
}

.tv-table tbody tr.selected {
  background-color: #e0f2fe;
}

.loc-cell {
  display: flex;
  align-items: center;
}

.loc-name {
  color: #0f172a;
  font-weight: 600;
}

.badge-danger-pill {
  background-color: #fee2e2;
  color: #dc2626;
  font-size: 0.72rem;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
}

.badge-amber-pill {
  background-color: #fef3c7;
  color: #d97706;
  font-size: 0.72rem;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
}

/* Suspended Feed */
.suspended-list-container {
  flex: 1;
  overflow-y: auto;
  max-height: 490px;
  padding: 10px 12px;
}

.suspended-items {
  display: flex;
  flex-direction: column;
  gap: 7px;
}

.suspended-item {
  background-color: #ffffff;
  border: 1px solid #e2e8f0;
  border-left: 3px solid #dc2626;
  border-radius: 6px;
  padding: 8px 10px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  transition: all 0.15s ease;
}

.suspended-item:hover {
  background-color: #fffafb;
  border-color: #fca5a5;
  border-left-color: #dc2626;
}

.suspended-item .item-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.user-id-code {
  background-color: #f1f5f9;
  border: 1px solid #cbd5e1;
  color: #0284c7;
  font-family: monospace;
  font-size: 0.72rem;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
}

.user-details {
  display: flex;
  flex-direction: column;
}

.user-fullname {
  font-size: 0.8rem;
  font-weight: 700;
  color: #0f172a;
}

.user-address {
  font-size: 0.7rem;
  color: #64748b;
  display: flex;
  align-items: center;
}

.item-right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
}

.brand-tag-clean {
  font-size: 0.65rem;
  font-weight: 700;
  background-color: #f1f5f9;
  color: #334155;
  padding: 1px 5px;
  border-radius: 3px;
  border: 1px solid #e2e8f0;
}

.plan-tag-clean {
  font-size: 0.68rem;
  color: #64748b;
  font-weight: 500;
}

.empty-suspended-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 260px;
  text-align: center;
}

.empty-title {
  font-size: 0.92rem;
  font-weight: 700;
  color: #059669;
  margin: 0;
}

.empty-desc {
  font-size: 0.74rem;
  color: #64748b;
  margin: 4px 0 0 0;
}

/* Footer */
.tv-footer {
  height: 36px;
  background-color: #ffffff;
  border-top: 1px solid #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  font-size: 0.72rem;
  color: #64748b;
}

.footer-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-dot-active {
  width: 7px;
  height: 7px;
  background-color: #10b981;
  border-radius: 50%;
  box-shadow: 0 0 4px #10b981;
}

.footer-separator {
  color: #cbd5e1;
}

/* Color classes */
.text-slate { color: #0f172a !important; }
.text-success { color: #059669 !important; }
.text-warning { color: #d97706 !important; }
.text-danger { color: #dc2626 !important; }
.text-muted { color: #94a3b8 !important; }
.text-sky { color: #0284c7 !important; }
.text-violet { color: #7c3aed !important; }
.text-amber { color: #d97706 !important; }

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
