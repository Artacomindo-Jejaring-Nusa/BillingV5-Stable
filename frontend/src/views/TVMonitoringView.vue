<template>
  <div class="tv-monitoring-root" :class="{ 'is-fullscreen': isFullscreen }">
    <!-- Fireworks Overlay Canvas (Petasan Meledak) -->
    <canvas ref="fireworksCanvasRef" class="fireworks-canvas"></canvas>

    <!-- Top Bar / Header for NOC Wallboard -->
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

        <!-- Petasan / Celebrate Button -->
        <button
          class="celebrate-btn"
          @click="triggerFireworksBurst"
          title="Nyalakan Efek Petasan Kembang Api"
        >
          <v-icon size="16" color="#f59e0b" class="mr-1">mdi-party-popper</v-icon>
          <span>Petasan</span>
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
          <span>Menampilkan data khusus wilayah: <strong>{{ selectedLocation }}</strong></span>
        </div>
        <button class="reset-filter-btn" @click="resetToAllLocations">
          <v-icon size="14" class="mr-1">mdi-close-circle-outline</v-icon>
          Kembali ke Semua Lokasi
        </button>
      </div>

      <!-- ROW 1: 3 METRIC CARDS (TOTAL KESELURUHAN, AKTIF, SUSPENDED) WITH COUNTING ANIMATION -->
      <section class="metrics-grid-3">
        <!-- CARD 1: TOTAL KESELURUHAN -->
        <div class="metric-card card-total" @click="triggerFireworksBurst" title="Klik untuk memicu petasan">
          <div class="metric-card-inner">
            <div class="metric-content">
              <div class="metric-header">
                <span class="metric-tag">TOTAL KESELURUHAN</span>
                <div class="icon-circle bg-sky-light">
                  <v-icon size="22" color="#0284c7">mdi-account-group</v-icon>
                </div>
              </div>
              <div class="metric-body">
                <div class="metric-big-number text-primary">
                  {{ formatNumber(displayTotalUsers) }}
                </div>
                <div class="metric-sub-badge">
                  <span class="badge-dot blue"></span>
                  <span>Pelanggan FTTH Berjalan (Aktif & Terisolir)</span>
                </div>
              </div>
            </div>
            <div class="metric-anim">
              <img src="/anim-total.gif" alt="Total animation" class="metric-gif" />
            </div>
          </div>
        </div>

        <!-- CARD 2: AKTIF -->
        <div class="metric-card card-active" @click="triggerFireworksBurst" title="Klik untuk memicu petasan">
          <div class="metric-card-inner">
            <div class="metric-content">
              <div class="metric-header">
                <span class="metric-tag text-success">AKTIF</span>
                <div class="icon-circle bg-green-light">
                  <v-icon size="22" color="#059669">mdi-account-check</v-icon>
                </div>
              </div>
              <div class="metric-body">
                <div class="metric-big-number text-success">
                  {{ formatNumber(displayActiveUsers) }}
                </div>
                <div class="metric-sub-badge">
                  <span class="badge-dot green"></span>
                  <span><strong>{{ activePercentage }}%</strong> Pelanggan Online & Aktif</span>
                </div>
              </div>
            </div>
            <div class="metric-anim">
              <img src="/anim-aktif.gif" alt="Aktif animation" class="metric-gif" />
            </div>
          </div>
        </div>

        <!-- CARD 3: SUSPENDED -->
        <div class="metric-card card-suspended" @click="triggerFireworksBurst" title="Klik untuk memicu petasan">
          <div class="metric-card-inner">
            <div class="metric-content">
              <div class="metric-header">
                <span class="metric-tag text-danger">SUSPENDED</span>
                <div class="icon-circle bg-red-light">
                  <v-icon size="22" color="#dc2626">mdi-account-alert</v-icon>
                </div>
              </div>
              <div class="metric-body">
                <div class="metric-big-number text-danger">
                  {{ formatNumber(displaySuspendedUsers) }}
                </div>
                <div class="metric-sub-badge">
                  <span class="badge-dot red"></span>
                  <span><strong>{{ suspendedPercentage }}%</strong> Modem Terisolir (Tunggakan)</span>
                </div>
              </div>
            </div>
            <div class="metric-anim">
              <img src="/anim-suspended.gif" alt="Suspended animation" class="metric-gif" />
            </div>
          </div>
        </div>
      </section>

      <!-- ROW 2: FULL-WIDTH BAR CHART (DIAGRAM BATANG STATISTIK LOKASI) -->
      <section class="chart-section">
        <div class="chart-card">
          <div class="chart-card-header">
            <div class="header-left">
              <div class="chart-icon-box">
                <v-icon size="20" color="#0284c7">mdi-chart-bar</v-icon>
              </div>
              <div>
                <h2 class="chart-title">Statistik Berdasarkan Lokasi</h2>
                <p class="chart-subtitle">
                  Urutan Dari Terkecil (Kiri) ke Terbesar (Kanan) • Klik Batang untuk Filter Wilayah
                </p>
              </div>
            </div>

            <!-- Chart Legend -->
            <div class="chart-legend">
              <div class="legend-item">
                <span class="legend-color-box bg-legend-active"></span>
                <span class="legend-label">Aktif</span>
              </div>
              <div class="legend-item">
                <span class="legend-color-box bg-legend-suspended"></span>
                <span class="legend-label">Suspended</span>
              </div>
            </div>
          </div>

          <div class="chart-container">
            <div ref="barChartRef" class="echarts-dom"></div>
          </div>
        </div>
      </section>
    </main>

    <!-- TV Footer Information -->
    <footer class="tv-footer">
      <div class="footer-left">
        <span :class="isReconnecting ? 'status-dot-warning' : 'status-dot-active'"></span>
        <span v-if="!isReconnecting">Sistem Monitoring Aktif</span>
        <span v-else class="status-reconnecting-text">Menghubungkan Ulang ke Server...</span>
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
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue';
import { useRouter } from 'vue-router';
import * as echarts from 'echarts';
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
const isFullscreen = ref(false);
const isReconnecting = ref(false);

// Animated Counters
const displayTotalUsers = ref(0);
const displayActiveUsers = ref(0);
const displaySuspendedUsers = ref(0);

// Chart DOM Reference
const barChartRef = ref<HTMLDivElement | null>(null);
let myChart: echarts.ECharts | null = null;

// Fireworks Canvas Reference
const fireworksCanvasRef = ref<HTMLCanvasElement | null>(null);
let fireworksAnimationId: number | null = null;

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

// Sorted locations from smallest (left) to largest (right)
const sortedLocations = computed(() => {
  const list = [...(data.value?.locations || [])];
  // Urutkan dari jumlah total user terkecil ke terbesar
  return list.sort((a, b) => (a.total_users || 0) - (b.total_users || 0));
});

// Percentages
const activePercentage = computed(() => {
  const total = data.value?.total_users || 0;
  if (total === 0) return '0.0';
  const active = data.value?.active_users || 0;
  return ((active / total) * 100).toFixed(1);
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

// Helper Formatter
const formatNumber = (num: number) => {
  return new Intl.NumberFormat('id-ID').format(num || 0);
};

// Counting Animation Logic (Cubic Out Easing)
const animateCount = (start: number, end: number, duration: number, callback: (val: number) => void) => {
  const startTime = performance.now();
  const step = (now: number) => {
    const elapsed = now - startTime;
    const progress = Math.min(elapsed / duration, 1);
    // Easing Out Expo
    const ease = progress === 1 ? 1 : 1 - Math.pow(2, -10 * progress);
    const current = Math.round(start + (end - start) * ease);
    callback(current);
    if (progress < 1) {
      requestAnimationFrame(step);
    }
  };
  requestAnimationFrame(step);
};

const runCountersAnimation = (newData: TVMonitoringData) => {
  animateCount(displayTotalUsers.value, newData.total_users || 0, 1500, (v) => {
    displayTotalUsers.value = v;
  });
  animateCount(displayActiveUsers.value, newData.active_users || 0, 1500, (v) => {
    displayActiveUsers.value = v;
  });
  animateCount(displaySuspendedUsers.value, newData.suspended_users || 0, 1500, (v) => {
    displaySuspendedUsers.value = v;
  });
};

// ==========================================
// FIREWORKS (PETASAN MELEDAK) ENGINE
// ==========================================
interface FireworkParticle {
  x: number;
  y: number;
  vx: number;
  vy: number;
  color: string;
  alpha: number;
  decay: number;
  size: number;
  flicker: boolean;
}

interface Rocket {
  x: number;
  y: number;
  targetY: number;
  vy: number;
  color: string;
  exploded: boolean;
}

let particles: FireworkParticle[] = [];
let rockets: Rocket[] = [];

const fireworkColors = [
  '#0284c7', // Sky blue
  '#10b981', // Emerald
  '#f59e0b', // Gold / Amber
  '#ef4444', // Red
  '#a855f7', // Purple
  '#38bdf8', // Cyan
  '#ffffff', // White spark
  '#ec4899', // Pink
];

const createExplosion = (x: number, y: number, color: string) => {
  const count = 70 + Math.floor(Math.random() * 30);
  for (let i = 0; i < count; i++) {
    const angle = Math.random() * Math.PI * 2;
    const speed = 2 + Math.random() * 6;
    particles.push({
      x,
      y,
      vx: Math.cos(angle) * speed,
      vy: Math.sin(angle) * speed,
      color: Math.random() > 0.3 ? color : fireworkColors[Math.floor(Math.random() * fireworkColors.length)],
      alpha: 1,
      decay: 0.015 + Math.random() * 0.015,
      size: 2.5 + Math.random() * 2,
      flicker: Math.random() > 0.5,
    });
  }
};

const triggerFireworksBurst = () => {
  const canvas = fireworksCanvasRef.value;
  if (!canvas) return;

  const w = canvas.width;
  const h = canvas.height;

  // Launch 5 rockets from bottom towards cards and chart
  const launchPoints = [
    { x: w * 0.2, targetY: h * 0.25 },
    { x: w * 0.35, targetY: h * 0.2 },
    { x: w * 0.5, targetY: h * 0.18 },
    { x: w * 0.65, targetY: h * 0.22 },
    { x: w * 0.8, targetY: h * 0.25 },
  ];

  launchPoints.forEach((pt, index) => {
    setTimeout(() => {
      rockets.push({
        x: pt.x + (Math.random() * 60 - 30),
        y: h,
        targetY: pt.targetY + (Math.random() * 50 - 25),
        vy: -11 - Math.random() * 3,
        color: fireworkColors[index % fireworkColors.length],
        exploded: false,
      });
    }, index * 180);
  });
};

const animateFireworks = () => {
  const canvas = fireworksCanvasRef.value;
  if (!canvas) return;
  const ctx = canvas.getContext('2d');
  if (!ctx) return;

  ctx.clearRect(0, 0, canvas.width, canvas.height);

  // Update & Draw Rockets
  for (let i = rockets.length - 1; i >= 0; i--) {
    const r = rockets[i];
    r.y += r.vy;
    r.vy += 0.05; // slight deceleration

    // Draw Rocket trail
    ctx.save();
    ctx.fillStyle = r.color;
    ctx.beginPath();
    ctx.arc(r.x, r.y, 3, 0, Math.PI * 2);
    ctx.fill();

    // Sparkle tail
    ctx.strokeStyle = 'rgba(255, 255, 255, 0.6)';
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.moveTo(r.x, r.y);
    ctx.lineTo(r.x, r.y - r.vy * 2);
    ctx.stroke();
    ctx.restore();

    if (r.y <= r.targetY || r.vy >= -1) {
      createExplosion(r.x, r.y, r.color);
      rockets.splice(i, 1);
    }
  }

  // Update & Draw Particles
  for (let i = particles.length - 1; i >= 0; i--) {
    const p = particles[i];
    p.x += p.vx;
    p.y += p.vy;
    p.vy += 0.08; // gravity
    p.vx *= 0.98; // air drag
    p.vy *= 0.98;
    p.alpha -= p.decay;

    if (p.alpha <= 0) {
      particles.splice(i, 1);
      continue;
    }

    ctx.save();
    ctx.globalAlpha = p.flicker && Math.random() > 0.5 ? p.alpha * 0.4 : p.alpha;
    ctx.fillStyle = p.color;
    ctx.beginPath();
    ctx.arc(p.x, p.y, p.size, 0, Math.PI * 2);
    ctx.fill();
    ctx.restore();
  }

  fireworksAnimationId = requestAnimationFrame(animateFireworks);
};

const resizeFireworksCanvas = () => {
  if (fireworksCanvasRef.value) {
    fireworksCanvasRef.value.width = window.innerWidth;
    fireworksCanvasRef.value.height = window.innerHeight;
  }
};

// ==========================================
// ECHARTS BAR CHART (DIAGRAM BATANG)
// ==========================================
const renderBarChart = () => {
  if (!barChartRef.value) return;

  if (!myChart) {
    myChart = echarts.init(barChartRef.value);
    myChart.on('click', (params: any) => {
      if (params.name) {
        selectLocationRow(params.name);
      }
    });
  }

  const locs = sortedLocations.value;
  const categories = locs.map((l) => l.lokasi);
  const activeSeries = locs.map((l) => l.active_users);
  const suspendedSeries = locs.map((l) => l.suspended_users);
  const totalSeries = locs.map((l) => l.total_users);

  const option: echarts.EChartsOption = {
    animationDuration: 1400,
    animationEasing: 'cubicOut',
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
      backgroundColor: 'rgba(15, 23, 42, 0.95)',
      borderColor: '#334155',
      borderWidth: 1,
      textStyle: {
        color: '#f8fafc',
        fontSize: 13,
      },
      formatter: (params: any) => {
        if (!Array.isArray(params) || params.length === 0) return '';
        const index = params[0].dataIndex;
        const loc = locs[index];
        if (!loc) return '';

        return `
          <div style="font-weight: 800; font-size: 14px; margin-bottom: 6px; color: #38bdf8;">
            📍 ${loc.lokasi}
          </div>
          <div style="display: flex; justify-content: space-between; gap: 20px; font-size: 12px; margin-bottom: 3px;">
            <span style="color: #94a3b8;">Total User:</span>
            <strong style="color: #ffffff;">${formatNumber(loc.total_users)}</strong>
          </div>
          <div style="display: flex; justify-content: space-between; gap: 20px; font-size: 12px; margin-bottom: 3px;">
            <span style="color: #10b981;">● Aktif:</span>
            <strong style="color: #10b981;">${formatNumber(loc.active_users)}</strong>
          </div>
          <div style="display: flex; justify-content: space-between; gap: 20px; font-size: 12px; margin-bottom: 6px;">
            <span style="color: #ef4444;">● Suspended:</span>
            <strong style="color: #ef4444;">${formatNumber(loc.suspended_users)}</strong>
          </div>
          <div style="border-top: 1px solid #334155; padding-top: 4px; font-size: 11px; color: #cbd5e1;">
            Jakinet: ${loc.jakinet_users} • Jelantik: ${loc.jelantik_users} • Nagrak: ${loc.nagrak_users}
          </div>
        `;
      },
    },
    grid: {
      left: '2%',
      right: '2%',
      bottom: '12%',
      top: '18%',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: categories,
      axisLine: {
        lineStyle: {
          color: '#cbd5e1',
        },
      },
      axisTick: {
        alignWithLabel: true,
      },
      axisLabel: {
        color: '#334155',
        fontWeight: 600,
        fontSize: 12,
        interval: 0,
        rotate: categories.length > 7 ? 22 : 0,
      },
    },
    yAxis: {
      type: 'value',
      name: 'Jumlah Pelanggan',
      nameTextStyle: {
        color: '#64748b',
        fontSize: 12,
        padding: [0, 0, 8, 0],
      },
      splitLine: {
        lineStyle: {
          color: '#f1f5f9',
          type: 'dashed',
        },
      },
      axisLabel: {
        color: '#64748b',
        fontSize: 12,
      },
    },
    series: [
      {
        name: 'Aktif',
        type: 'bar',
        stack: 'total',
        barMaxWidth: 56,
        data: activeSeries,
        label: {
          show: true,
          position: 'inside',
          color: '#ffffff',
          fontWeight: 'bold',
          fontSize: 12,
          formatter: (params: any) => (params.value >= 12 ? formatNumber(params.value) : ''),
        },
        itemStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: '#0284c7' },
            { offset: 1, color: '#38bdf8' },
          ]),
          borderRadius: [0, 0, 4, 4],
        },
      },
      {
        name: 'Suspended',
        type: 'bar',
        stack: 'total',
        barMaxWidth: 56,
        data: suspendedSeries,
        label: {
          show: true,
          position: 'inside',
          color: '#ffffff',
          fontWeight: 'bold',
          fontSize: 11,
          formatter: (params: any) => (params.value >= 12 ? formatNumber(params.value) : ''),
        },
        itemStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: '#dc2626' },
            { offset: 1, color: '#f87171' },
          ]),
          borderRadius: [4, 4, 0, 0],
        },
      },
      // Invisible series solely to render the total, active, and suspended directly above the bar without hovering
      {
        name: 'Total Label',
        type: 'bar',
        barMaxWidth: 56,
        data: totalSeries,
        itemStyle: {
          color: 'transparent',
        },
        tooltip: {
          show: false,
        },
        label: {
          show: true,
          position: 'top',
          distance: 6,
          formatter: (params: any) => {
            const loc = locs[params.dataIndex];
            if (!loc) return `${formatNumber(params.value)}`;
            if (loc.suspended_users > 0) {
              return `{total|${formatNumber(loc.total_users)}}\n{aktif|${formatNumber(loc.active_users)}} {susp|${formatNumber(loc.suspended_users)}}`;
            }
            return `{total|${formatNumber(loc.total_users)}}\n{aktif|${formatNumber(loc.active_users)}}`;
          },
          rich: {
            total: {
              color: '#0f172a',
              fontWeight: 'bold',
              fontSize: 14,
              align: 'center',
              lineHeight: 18,
            },
            aktif: {
              color: '#0284c7',
              fontWeight: 'bold',
              fontSize: 11,
              align: 'center',
              backgroundColor: '#e0f2fe',
              padding: [2, 4],
              borderRadius: 3,
            },
            susp: {
              color: '#dc2626',
              fontWeight: 'bold',
              fontSize: 11,
              align: 'center',
              backgroundColor: '#fee2e2',
              padding: [2, 4],
              borderRadius: 3,
            },
          },
        },
        barGap: '-100%',
      },
    ],
  };

  myChart.setOption(option);
};

// Data Fetching
const fetchData = async (showLoading = false) => {
  if (showLoading) loading.value = true;
  try {
    const locParam = selectedLocation.value === 'Semua Lokasi' ? '' : selectedLocation.value;
    const url = `/dashboard/tv-monitoring${locParam ? `?lokasi=${encodeURIComponent(locParam)}` : ''}`;
    const response = await apiClient.get<TVMonitoringData>(url, {
      headers: { 'X-Skip-Network-Interceptor': 'true' },
      timeout: 15000,
    });
    if (response?.data) {
      data.value = response.data;
      isReconnecting.value = false;
      runCountersAnimation(response.data);
      await nextTick();
      renderBarChart();
      // Trigger Petasan Meledak celebration burst hanya saat awal buka / refresh manual
      if (showLoading) {
        triggerFireworksBurst();
      }
    }
  } catch (err) {
    console.warn('[TV Monitoring] Jaringan tertunda/timeout, data monitor tetap dipertahankan:', err);
    isReconnecting.value = true;
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
      setTimeout(() => {
        resizeFireworksCanvas();
        myChart?.resize();
      }, 300);
    }).catch((err) => {
      console.warn('Error attempting fullscreen:', err);
    });
  } else {
    if (document.exitFullscreen) {
      document.exitFullscreen().then(() => {
        isFullscreen.value = false;
        setTimeout(() => {
          resizeFireworksCanvas();
          myChart?.resize();
        }, 300);
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

const handleWindowResize = () => {
  resizeFireworksCanvas();
  myChart?.resize();
};

// Lifecycle
onMounted(() => {
  fetchData(true);
  resizeFireworksCanvas();
  animateFireworks();

  window.addEventListener('resize', handleWindowResize);

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
    setTimeout(() => {
      resizeFireworksCanvas();
      myChart?.resize();
    }, 200);
  });
});

onUnmounted(() => {
  if (clockTimer) clearInterval(clockTimer);
  if (refreshTimer) clearInterval(refreshTimer);
  if (cycleTimer) clearInterval(cycleTimer);
  if (fireworksAnimationId) cancelAnimationFrame(fireworksAnimationId);
  window.removeEventListener('resize', handleWindowResize);
  if (myChart) {
    myChart.dispose();
    myChart = null;
  }
});
</script>

<style scoped>
/* Base: Clean, Modern, Crisp Light Enterprise Theme */
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
  position: relative;
}

/* Fullscreen Transparent Fireworks Canvas */
.fireworks-canvas {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  pointer-events: none;
  z-index: 9999;
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
  gap: 14px;
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

/* Celebrate Petasan Button */
.celebrate-btn {
  display: flex;
  align-items: center;
  background-color: #fef3c7;
  border: 1px solid #fde68a;
  color: #b45309;
  padding: 5px 12px;
  border-radius: 6px;
  font-size: 0.78rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s ease;
}

.celebrate-btn:hover {
  background-color: #fde68a;
  transform: scale(1.03);
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

/* ROW 1: 3 METRICS GRID (TOTAL, AKTIF, SUSPENDED) */
.metrics-grid-3 {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.metric-card {
  background-color: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 18px 20px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  cursor: pointer;
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.metric-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.07);
}

.metric-card-inner {
  display: flex;
  align-items: center;
  gap: 12px;
  height: 100%;
}

.metric-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.metric-anim {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.metric-gif {
  width: 90px;
  height: 90px;
  object-fit: contain;
  filter: drop-shadow(0 4px 8px rgba(0, 0, 0, 0.1));
}

.metric-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.metric-tag {
  font-size: 0.8rem;
  font-weight: 800;
  letter-spacing: 0.06em;
  color: #475569;
}

.icon-circle {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.bg-sky-light { background-color: #e0f2fe; }
.bg-green-light { background-color: #d1fae5; }
.bg-red-light { background-color: #fee2e2; }

.metric-big-number {
  font-size: 2.7rem;
  font-weight: 900;
  line-height: 1.1;
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.02em;
  margin-bottom: 6px;
  transition: all 0.2s ease;
}

.metric-sub-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.78rem;
  color: #64748b;
  font-weight: 500;
}

.badge-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
}

.badge-dot.blue { background-color: #0284c7; }
.badge-dot.green { background-color: #10b981; }
.badge-dot.red { background-color: #ef4444; }

/* Colors */
.text-primary { color: #0284c7 !important; }
.text-success { color: #059669 !important; }
.text-danger { color: #dc2626 !important; }

/* ROW 2: FULL WIDTH DIAGRAM BATANG */
.chart-section {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.chart-card {
  background-color: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  display: flex;
  flex-direction: column;
  flex: 1;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.chart-card-header {
  padding: 14px 20px;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #f8fafc;
}

.chart-card-header .header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.chart-icon-box {
  width: 34px;
  height: 34px;
  background-color: #e0f2fe;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.chart-title {
  font-size: 1.05rem;
  font-weight: 800;
  color: #0f172a;
  margin: 0;
  line-height: 1.2;
}

.chart-subtitle {
  font-size: 0.74rem;
  color: #64748b;
  margin: 2px 0 0 0;
}

.chart-legend {
  display: flex;
  align-items: center;
  gap: 16px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.8rem;
  font-weight: 600;
  color: #334155;
}

.legend-color-box {
  width: 14px;
  height: 14px;
  border-radius: 3px;
}

.bg-legend-active {
  background: linear-gradient(180deg, #0284c7, #38bdf8);
}

.bg-legend-suspended {
  background: linear-gradient(180deg, #dc2626, #f87171);
}

.chart-container {
  flex: 1;
  min-height: 440px;
  width: 100%;
  position: relative;
  padding: 10px 14px;
}

.echarts-dom {
  width: 100%;
  height: 100%;
  min-height: 440px;
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

.status-dot-warning {
  width: 7px;
  height: 7px;
  background-color: #f59e0b;
  border-radius: 50%;
  box-shadow: 0 0 4px #f59e0b;
  animation: pulse-dot 1.5s infinite;
}

.status-reconnecting-text {
  color: #d97706;
  font-weight: 600;
}

@keyframes pulse-dot {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.4; transform: scale(0.85); }
}

.footer-separator {
  color: #cbd5e1;
}

/* Fullscreen adjustments */
.is-fullscreen {
  width: 100vw;
  height: 100vh;
}

@media (max-width: 900px) {
  .metrics-grid-3 {
    grid-template-columns: 1fr;
  }
}
</style>
