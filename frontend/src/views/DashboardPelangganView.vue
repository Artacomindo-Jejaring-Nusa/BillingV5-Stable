<template>
  <v-container fluid class="pa-4 pa-md-6">
    <!-- Header Section (Standard Clean SaaS Style) -->
    <div class="header-card mb-4 mb-md-6">
      <div class="d-flex flex-column flex-md-row align-start align-md-center justify-space-between gap-4">
        <div class="d-flex align-center header-info">
          <div class="header-avatar-wrapper">
            <v-avatar class="header-avatar" color="transparent" size="50">
              <v-icon color="primary" size="28">mdi-view-dashboard-outline</v-icon>
            </v-avatar>
          </div>
          <div class="ml-4">
            <h1 class="header-title">Dashboard Monitoring Pelanggan JakiNet</h1>
            <p class="header-subtitle">
              Pantau performa, pertumbuhan, dan aktivitas finansial pelanggan JakiNet secara real-time
            </p>
          </div>
        </div>

        <div class="d-flex align-center gap-2 flex-wrap header-actions">
          <div class="realtime-status-badge">
            <span class="status-dot"></span>
            <span class="status-text">Sistem Real-time</span>
          </div>
          <v-btn
            variant="outlined"
            class="refresh-btn text-none font-weight-medium"
            prepend-icon="mdi-refresh"
            :loading="loading"
            @click="fetchDashboardData(revenueTimeframe)"
          >
            Segarkan Data
          </v-btn>
        </div>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading">
      <!-- Stats Cards Skeleton -->
      <v-row class="mb-4 mb-md-6">
        <v-col cols="12" sm="6" md="3" v-for="i in 4" :key="i">
          <v-card class="skeleton-card pa-4 rounded-xl border" elevation="0">
            <v-skeleton-loader type="article" />
          </v-card>
        </v-col>
      </v-row>

      <!-- Revenue Card Skeleton -->
      <v-row class="mb-4 mb-md-6">
        <v-col cols="12">
          <v-card class="skeleton-card pa-6 rounded-xl border" elevation="0">
            <v-skeleton-loader type="heading, paragraph" />
          </v-card>
        </v-col>
      </v-row>

      <!-- Charts Skeleton -->
      <v-row class="mb-4 mb-md-6">
        <v-col cols="12" md="6" v-for="i in 2" :key="i">
          <v-card class="skeleton-card pa-6 rounded-xl border" height="400" elevation="0">
            <v-skeleton-loader type="heading" class="mb-4" />
            <div style="height: 280px;">
              <SkeletonLoader type="chart" />
            </div>
          </v-card>
        </v-col>
      </v-row>
    </div>

    <!-- Dashboard Content -->
    <div v-else>
      <!-- Stats Cards Row -->
      <v-row class="mb-4 mb-md-6">
        <!-- 1. Pelanggan Aktif -->
        <v-col cols="12" sm="6" md="3">
          <div class="metric-card">
            <div class="d-flex align-center justify-space-between mb-2">
              <span class="metric-label">PELANGGAN AKTIF</span>
              <div class="metric-icon-box box-blue">
                <v-icon size="20" color="primary">mdi-account-check</v-icon>
              </div>
            </div>
            <div class="metric-value">{{ stats.pelanggan_aktif.toLocaleString('id-ID') }}</div>
            <div class="metric-meta mt-2 text-caption text-medium-emphasis">
              Total seluruh pelanggan berstatus aktif
            </div>
          </div>
        </v-col>

        <!-- 2. Pelanggan Baru -->
        <v-col cols="12" sm="6" md="3">
          <div class="metric-card">
            <div class="d-flex align-center justify-space-between mb-2">
              <span class="metric-label">PELANGGAN BARU</span>
              <div class="metric-icon-box box-green">
                <v-icon size="20" color="success">mdi-account-plus</v-icon>
              </div>
            </div>
            <div class="metric-value text-success">+{{ stats.pelanggan_baru_bulan_ini.toLocaleString('id-ID') }}</div>
            <div class="d-flex align-center mt-2">
              <span class="meta-pill pill-success">
                <v-icon size="12" class="mr-1">mdi-trending-up</v-icon>
                Bulan ini
              </span>
            </div>
          </div>
        </v-col>

        <!-- 3. Pelanggan Berhenti -->
        <v-col cols="12" sm="6" md="3">
          <div class="metric-card">
            <div class="d-flex align-center justify-space-between mb-2">
              <span class="metric-label">BERHENTI</span>
              <div class="metric-icon-box box-rose">
                <v-icon size="20" color="error">mdi-account-minus</v-icon>
              </div>
            </div>
            <div class="metric-value text-error">{{ stats.pelanggan_berhenti_bulan_ini.toLocaleString('id-ID') }}</div>
            <div class="d-flex align-center mt-2">
              <span class="meta-pill pill-rose">
                <v-icon size="12" class="mr-1">mdi-trending-down</v-icon>
                Bulan ini
              </span>
            </div>
          </div>
        </v-col>

        <!-- 4. JakiNet Aktif -->
        <v-col cols="12" sm="6" md="3">
          <div class="metric-card">
            <div class="d-flex align-center justify-space-between mb-2">
              <span class="metric-label">JAKINET AKTIF</span>
              <div class="metric-icon-box box-cyan">
                <v-icon size="20" color="info">mdi-wifi</v-icon>
              </div>
            </div>
            <div class="metric-value text-info">{{ stats.pelanggan_jakinet_aktif.toLocaleString('id-ID') }}</div>
            <div class="metric-meta mt-2 text-caption text-medium-emphasis">
              <strong>{{ jakiNetPercentage }}%</strong> dari total pelanggan
            </div>
          </div>
        </v-col>
      </v-row>

      <!-- Revenue Hero Card -->
      <v-row class="mb-4 mb-md-6">
        <v-col cols="12">
          <div class="revenue-hero-card pa-5 pa-md-6">
            <div class="d-flex flex-column flex-sm-row align-start align-sm-center justify-space-between gap-3 mb-3">
              <div>
                <span class="revenue-tag">FINANCIAL PERFORMANCE</span>
                <h3 class="revenue-title mb-1">Pendapatan JakiNet</h3>
                <p class="revenue-desc mb-0">
                  Estimasi total tagihan dan penerimaan pelanggan JakiNet bulan berjalan
                </p>
              </div>
              <div class="revenue-icon-box">
                <v-icon size="26" color="success">mdi-cash-multiple</v-icon>
              </div>
            </div>

            <div class="revenue-amount-row d-flex flex-column flex-sm-row align-start align-sm-baseline gap-2 gap-sm-3 mb-4">
              <span class="revenue-amount-value">{{ formatCurrency(stats.pendapatan_jakinet_bulan_ini) }}</span>
              <span class="revenue-amount-sub">/ bulan ini</span>
            </div>

            <div class="revenue-pills-row d-flex align-center gap-2 flex-wrap">
              <div class="revenue-meta-pill">
                <v-icon size="14" class="mr-1 text-success">mdi-check-circle-outline</v-icon>
                <span>Penetrasi JakiNet: <strong>{{ jakiNetPercentage }}%</strong></span>
              </div>
              <div class="revenue-meta-pill">
                <v-icon size="14" class="mr-1 text-primary">mdi-account-cash-outline</v-icon>
                <span>ARPU: <strong>{{ formatCurrency(averageRevenuePerUser) }}</strong> / user</span>
              </div>
            </div>
          </div>
        </v-col>
      </v-row>

      <!-- Charts Section Row 1 -->
      <v-row class="mb-4 mb-md-6">
        <!-- Customer Distribution Donut Chart -->
        <v-col cols="12" md="6">
          <div class="chart-panel-card">
            <div class="chart-panel-header pa-4 pa-sm-5 d-flex align-center justify-space-between border-b">
              <div class="d-flex align-center gap-2">
                <div class="chart-header-icon-box">
                  <v-icon size="18" color="primary">mdi-chart-pie</v-icon>
                </div>
                <div>
                  <h3 class="chart-panel-title mb-0">Distribusi Pelanggan JakiNet</h3>
                  <p class="chart-panel-subtitle mb-0">Komposisi aktif, baru, dan berhenti</p>
                </div>
              </div>
            </div>
            <div class="chart-panel-body pa-4 pa-sm-5">
              <div v-if="stats.pelanggan_aktif === 0 && stats.pelanggan_baru_bulan_ini === 0 && stats.pelanggan_berhenti_bulan_ini === 0" class="empty-chart-state">
                <div class="empty-chart-icon-box mb-2">
                  <v-icon size="32" color="slate-400">mdi-chart-pie</v-icon>
                </div>
                <h4 class="empty-chart-title mb-1">Tidak Ada Data Pelanggan</h4>
                <p class="empty-chart-desc mb-0">Belum ada data pelanggan JakiNet yang tersedia.</p>
              </div>
              <div v-else class="chart-canvas-wrapper" style="height: 280px; position: relative;">
                <canvas ref="pieChartCanvas" style="max-height: 280px;"></canvas>
              </div>
            </div>
          </div>
        </v-col>

        <!-- Customer Growth Trend Line Chart -->
        <v-col cols="12" md="6">
          <div class="chart-panel-card">
            <div class="chart-panel-header pa-4 pa-sm-5 d-flex align-center justify-space-between border-b">
              <div class="d-flex align-center gap-2">
                <div class="chart-header-icon-box">
                  <v-icon size="18" color="info">mdi-chart-line</v-icon>
                </div>
                <div>
                  <h3 class="chart-panel-title mb-0">Tren Pertumbuhan Kumulatif</h3>
                  <p class="chart-panel-subtitle mb-0">Grafik akumulasi pelanggan JakiNet</p>
                </div>
              </div>
            </div>
            <div class="chart-panel-body pa-4 pa-sm-5">
              <div v-if="!growthChartData || !growthChartData.labels || growthChartData.labels.length === 0" class="empty-chart-state">
                <div class="empty-chart-icon-box mb-2">
                  <v-icon size="32" color="slate-400">mdi-chart-line</v-icon>
                </div>
                <h4 class="empty-chart-title mb-1">Tidak Ada Data Pertumbuhan</h4>
                <p class="empty-chart-desc mb-0">Belum ada data riwayat pertumbuhan pelanggan JakiNet.</p>
              </div>
              <div v-else class="chart-canvas-wrapper" style="height: 280px; position: relative;">
                <canvas ref="lineChartCanvas" style="max-height: 280px;"></canvas>
              </div>
            </div>
          </div>
        </v-col>
      </v-row>

      <!-- Charts Section Row 2: Monthly Comparison & Quick Metrics -->
      <v-row class="mb-4 mb-md-6">
        <!-- Customer Status Overview (Bar Chart) -->
        <v-col cols="12" md="8">
          <div class="chart-panel-card h-100">
            <div class="chart-panel-header pa-4 pa-sm-5 d-flex align-center justify-space-between border-b">
              <div class="d-flex align-center gap-2">
                <div class="chart-header-icon-box">
                  <v-icon size="18" color="primary">mdi-chart-bar</v-icon>
                </div>
                <div>
                  <h3 class="chart-panel-title mb-0">Pelanggan Baru vs Berhenti</h3>
                  <p class="chart-panel-subtitle mb-0">Perbandingan mutasi pelanggan bulan berjalan</p>
                </div>
              </div>
              <span class="meta-pill pill-subtle">Bulan Ini</span>
            </div>
            <div class="chart-panel-body pa-4 pa-sm-5">
              <div class="chart-canvas-wrapper" style="height: 250px; position: relative;">
                <canvas ref="barChartCanvas" style="max-height: 250px;"></canvas>
              </div>
            </div>
          </div>
        </v-col>

        <!-- Quick Stats Panel -->
        <v-col cols="12" md="4">
          <div class="chart-panel-card h-100">
            <div class="chart-panel-header pa-4 pa-sm-5 d-flex align-center justify-space-between border-b">
              <div class="d-flex align-center gap-2">
                <div class="chart-header-icon-box">
                  <v-icon size="18" color="warning">mdi-speedometer</v-icon>
                </div>
                <div>
                  <h3 class="chart-panel-title mb-0">Metrik Performa</h3>
                  <p class="chart-panel-subtitle mb-0">KPI retensi dan pertumbuhan</p>
                </div>
              </div>
            </div>
            <div class="chart-panel-body pa-4 pa-sm-5">
              <div class="quick-stats-list">
                <!-- Growth Rate -->
                <div class="quick-stat-item mb-4">
                  <div class="d-flex align-center justify-space-between mb-2">
                    <span class="stat-item-label">Growth Rate Bulan Ini</span>
                    <span
                      class="stat-item-badge"
                      :class="growthRate >= 0 ? 'badge-positive' : 'badge-negative'"
                    >
                      {{ growthRate >= 0 ? '+' : '' }}{{ isNaN(growthRate) ? '0.0' : growthRate.toFixed(1) }}%
                    </span>
                  </div>
                  <v-progress-linear 
                    :model-value="isNaN(growthRate) ? 0 : Math.min(100, Math.abs(growthRate))" 
                    :color="growthRate >= 0 ? 'success' : 'error'" 
                    height="6" 
                    rounded
                    bg-color="slate-200"
                  ></v-progress-linear>
                </div>

                <!-- JakiNet Penetration -->
                <div class="quick-stat-item mb-4">
                  <div class="d-flex align-center justify-space-between mb-2">
                    <span class="stat-item-label">Penetrasi JakiNet</span>
                    <span class="stat-item-val">{{ jakiNetPenetration.toFixed(1) }}%</span>
                  </div>
                  <v-progress-linear 
                    :model-value="jakiNetPenetration" 
                    color="info" 
                    height="6" 
                    rounded
                    bg-color="slate-200"
                  ></v-progress-linear>
                </div>

                <!-- Customer Retention -->
                <div class="quick-stat-item mb-4">
                  <div class="d-flex align-center justify-space-between mb-2">
                    <span class="stat-item-label">Customer Retention</span>
                    <span class="stat-item-val text-success">{{ customerRetention.toFixed(1) }}%</span>
                  </div>
                  <v-progress-linear 
                    :model-value="customerRetention" 
                    color="success" 
                    height="6" 
                    rounded
                    bg-color="slate-200"
                  ></v-progress-linear>
                </div>

                <!-- Average Revenue Per User -->
                <div class="arpu-box pa-3 rounded-lg border mt-2">
                  <div class="d-flex align-center justify-space-between">
                    <div>
                      <span class="text-caption text-medium-emphasis d-block">ARPU JakiNet Bulan Ini</span>
                      <span class="arpu-val text-h6 font-weight-bold">{{ formatCurrency(averageRevenuePerUser) }}</span>
                    </div>
                    <v-icon size="24" color="primary" class="opacity-70">mdi-cash-check</v-icon>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </v-col>
      </v-row>

      <!-- Monthly Revenue Trend Row -->
      <v-row>
        <v-col cols="12">
          <div class="chart-panel-card">
            <div class="chart-panel-header pa-4 pa-sm-5 d-flex flex-column flex-sm-row align-start align-sm-center justify-space-between gap-3 border-b">
              <div class="d-flex align-center gap-2">
                <div class="chart-header-icon-box">
                  <v-icon size="18" color="success">mdi-chart-areaspline</v-icon>
                </div>
                <div>
                  <h3 class="chart-panel-title mb-0">Tren Pendapatan Bulanan JakiNet</h3>
                  <p class="chart-panel-subtitle mb-0">Riwayat penerimaan tagihan berdasar periode pilihan</p>
                </div>
              </div>
              
              <!-- Clean Segmented Timeframe Toggle -->
              <div class="timeframe-toggle-wrapper">
                <v-btn-toggle
                  v-model="revenueTimeframe"
                  color="primary"
                  density="compact"
                  mandatory
                  variant="outlined"
                  class="segmented-toggle"
                >
                  <v-btn value="3m" size="small" class="text-none">3 Bulan</v-btn>
                  <v-btn value="6m" size="small" class="text-none">6 Bulan</v-btn>
                  <v-btn value="1y" size="small" class="text-none">1 Tahun</v-btn>
                </v-btn-toggle>
              </div>
            </div>
            <div class="chart-panel-body pa-4 pa-sm-5">
              <div v-if="hasRevenueData" class="chart-canvas-wrapper" style="height: 300px; position: relative;">
                <canvas ref="revenueChartCanvas" style="max-height: 300px;"></canvas>
              </div>
              <div v-else class="empty-chart-state py-8">
                <div class="empty-chart-icon-box mb-2">
                  <v-icon size="36" color="slate-400">mdi-chart-bar-off</v-icon>
                </div>
                <h4 class="empty-chart-title mb-1">Tidak Ada Data Pendapatan</h4>
                <p class="empty-chart-desc mb-0">Tidak ada transaksi tagihan yang tercatat pada rentang waktu yang dipilih.</p>
              </div>
            </div>
          </div>
        </v-col>
      </v-row>
    </div>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue';
import apiClient from '@/services/api';
import Chart from 'chart.js/auto';
import type { ChartConfiguration } from 'chart.js';
import SkeletonLoader from '@/components/SkeletonLoader.vue';

const loading = ref(true);
const revenueTimeframe = ref('6m');

// State data dari API
const stats = ref({
  pelanggan_aktif: 0,
  pelanggan_baru_bulan_ini: 0,
  pelanggan_berhenti_bulan_ini: 0,
  pelanggan_jakinet_aktif: 0,
  pendapatan_jakinet_bulan_ini: 0,
});
const growthChartData = ref({ labels: [] as string[], data: [] as number[] });
const revenueChartData = ref({ labels: [] as string[], data: [] as number[] });

// Chart canvas refs
const pieChartCanvas = ref<HTMLCanvasElement>();
const lineChartCanvas = ref<HTMLCanvasElement>();
const barChartCanvas = ref<HTMLCanvasElement>();
const revenueChartCanvas = ref<HTMLCanvasElement>();

let pieChart: Chart | null = null;
let lineChart: Chart | null = null;
let barChart: Chart | null = null;
let revenueChart: Chart | null = null;

// Safe Computed Properties (Zero-division Protection)
const jakiNetPercentage = computed(() => {
  if (!stats.value.pelanggan_aktif || stats.value.pelanggan_aktif <= 0) return '0.0';
  return ((stats.value.pelanggan_jakinet_aktif / stats.value.pelanggan_aktif) * 100).toFixed(1);
});

const growthRate = computed(() => {
  const base = stats.value.pelanggan_aktif - stats.value.pelanggan_baru_bulan_ini;
  if (base <= 0) return stats.value.pelanggan_baru_bulan_ini > 0 ? 100 : 0;
  const rate = ((stats.value.pelanggan_baru_bulan_ini - stats.value.pelanggan_berhenti_bulan_ini) / base) * 100;
  return isNaN(rate) ? 0 : rate;
});

const jakiNetPenetration = computed(() => {
  if (!stats.value.pelanggan_aktif || stats.value.pelanggan_aktif <= 0) return 0;
  const val = (stats.value.pelanggan_jakinet_aktif / stats.value.pelanggan_aktif) * 100;
  return isNaN(val) ? 0 : val;
});

const customerRetention = computed(() => {
  const startOfMonth = stats.value.pelanggan_aktif - stats.value.pelanggan_baru_bulan_ini + stats.value.pelanggan_berhenti_bulan_ini;
  if (startOfMonth <= 0) return 100;
  const retained = startOfMonth - stats.value.pelanggan_berhenti_bulan_ini;
  const val = (retained / startOfMonth) * 100;
  return isNaN(val) ? 0 : Math.max(0, Math.min(100, val));
});

const averageRevenuePerUser = computed(() => {
  if (!stats.value.pelanggan_jakinet_aktif || stats.value.pelanggan_jakinet_aktif <= 0) return 0;
  const val = stats.value.pendapatan_jakinet_bulan_ini / stats.value.pelanggan_jakinet_aktif;
  return isNaN(val) ? 0 : val;
});

const hasRevenueData = computed(() => {
  if (!revenueChartData.value ||
      !revenueChartData.value.data ||
      !Array.isArray(revenueChartData.value.data) ||
      revenueChartData.value.data.length === 0) {
    return false;
  }
  return revenueChartData.value.data.some(value => value > 0);
});

async function fetchDashboardData(timespan: string = '6m') {
  loading.value = true;
  try {
    const response = await apiClient.get('/dashboard-pelanggan', {
      params: { timespan }
    });
    const data = response.data;

    stats.value = data.main_stats || stats.value;
    growthChartData.value = data.growth_chart || { labels: [], data: [] };
    revenueChartData.value = data.revenue_chart || { labels: [], data: [] };

    initializeCharts();
  } catch (error) {
    console.error("Gagal mengambil data dashboard pelanggan:", error);
  } finally {
    loading.value = false;
  }
}

function formatCurrency(value: number) {
  if (isNaN(value)) return "Rp 0";
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0
  }).format(value);
}

function createPieChart() {
  if (!pieChartCanvas.value) return;
  const ctx = pieChartCanvas.value.getContext('2d');
  if (!ctx) return;

  if (pieChart) pieChart.destroy();

  const activeExisting = Math.max(0, stats.value.pelanggan_jakinet_aktif - stats.value.pelanggan_baru_bulan_ini);
  const total = activeExisting + stats.value.pelanggan_baru_bulan_ini + stats.value.pelanggan_berhenti_bulan_ini;

  const config: ChartConfiguration<'doughnut'> = {
    type: 'doughnut',
    data: {
      labels: ['Pelanggan Aktif (Lama)', 'Pelanggan Baru (Bulan Ini)', 'Pelanggan Berhenti (Bulan Ini)'],
      datasets: [{
        data: [
          activeExisting,
          stats.value.pelanggan_baru_bulan_ini,
          stats.value.pelanggan_berhenti_bulan_ini
        ],
        backgroundColor: [
          '#0f172a', // Slate-900
          '#059669', // Emerald-600
          '#e11d48'  // Rose-600
        ],
        borderColor: '#ffffff',
        borderWidth: 2,
        hoverOffset: 4
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      cutout: '70%',
      plugins: {
        legend: { 
          position: 'bottom', 
          labels: { 
            padding: 16, 
            usePointStyle: true, 
            font: { size: 12, weight: 500 },
            color: '#475569'
          } 
        },
        tooltip: {
          backgroundColor: '#0f172a',
          titleColor: '#ffffff',
          bodyColor: '#e2e8f0',
          padding: 10,
          cornerRadius: 6,
          displayColors: true,
          callbacks: {
            label: (context) => {
              const label = context.label || '';
              const value = Number(context.parsed) || 0;
              const pct = total > 0 ? ((value / total) * 100).toFixed(1) : '0';
              return `${label}: ${value} (${pct}%)`;
            }
          }
        }
      }
    }
  };
  pieChart = new Chart(ctx, config as any);
}

function createLineChart() {
  if (!lineChartCanvas.value) return;
  const ctx = lineChartCanvas.value.getContext('2d');
  if (!ctx) return;

  if (lineChart) lineChart.destroy();

  if (!growthChartData.value || !growthChartData.value.labels || !growthChartData.value.data) {
    return;
  }

  const gradient = ctx.createLinearGradient(0, 0, 0, 260);
  gradient.addColorStop(0, 'rgba(5, 150, 105, 0.16)');
  gradient.addColorStop(1, 'rgba(5, 150, 105, 0.0)');

  const config: ChartConfiguration<'line'> = {
    type: 'line',
    data: {
      labels: growthChartData.value.labels,
      datasets: [
        {
          label: 'Total Pelanggan JakiNet',
          data: growthChartData.value.data,
          borderColor: '#059669',
          backgroundColor: gradient,
          pointBackgroundColor: '#059669',
          pointBorderColor: '#ffffff',
          pointHoverBackgroundColor: '#ffffff',
          pointHoverBorderColor: '#059669',
          pointRadius: 4,
          pointHoverRadius: 6,
          borderWidth: 2.5,
          tension: 0.35,
          fill: true,
        }
      ]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      scales: { 
        y: { 
          beginAtZero: false,
          grid: {
            color: 'rgba(226, 232, 240, 0.6)'
          },
          ticks: {
            color: '#64748b'
          }
        }, 
        x: { 
          grid: { 
            display: false 
          },
          ticks: {
            color: '#64748b'
          }
        } 
      },
      plugins: { 
        legend: { 
          position: 'top', 
          labels: { 
            usePointStyle: true, 
            padding: 16,
            font: { size: 12, weight: 500 },
            color: '#475569'
          } 
        },
        tooltip: {
          backgroundColor: '#0f172a',
          titleColor: '#ffffff',
          bodyColor: '#e2e8f0',
          padding: 10,
          cornerRadius: 6,
          displayColors: false,
          callbacks: {
            label: (context) => `${context.dataset.label}: ${context.parsed.y} pelanggan`
          }
        }
      }
    }
  };
  lineChart = new Chart(ctx, config);
}

function createBarChart() {
  if (!barChartCanvas.value) return;
  const ctx = barChartCanvas.value.getContext('2d');
  if (!ctx) return;

  if (barChart) barChart.destroy();

  const config: ChartConfiguration<'bar'> = {
    type: 'bar',
    data: {
      labels: ['Mutasi Bulan Ini'],
      datasets: [
        {
          label: 'Pelanggan Baru',
          data: [stats.value.pelanggan_baru_bulan_ini],
          backgroundColor: '#059669',
          borderColor: '#047857',
          borderWidth: 1,
          borderRadius: 6,
          borderSkipped: false,
        },
        {
          label: 'Pelanggan Berhenti',
          data: [stats.value.pelanggan_berhenti_bulan_ini],
          backgroundColor: '#e11d48',
          borderColor: '#be123c',
          borderWidth: 1,
          borderRadius: 6,
          borderSkipped: false,
        }
      ]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      scales: { 
        y: { 
          beginAtZero: true,
          grid: {
            color: 'rgba(226, 232, 240, 0.6)'
          },
          ticks: {
            color: '#64748b'
          }
        }, 
        x: { 
          grid: { 
            display: false 
          },
          ticks: {
            color: '#64748b'
          }
        } 
      },
      plugins: { 
        legend: { 
          position: 'top',
          labels: { 
            usePointStyle: true, 
            padding: 16,
            font: { size: 12, weight: 500 },
            color: '#475569'
          }
        },
        tooltip: {
          backgroundColor: '#0f172a',
          titleColor: '#ffffff',
          bodyColor: '#e2e8f0',
          padding: 10,
          cornerRadius: 6,
          displayColors: true
        }
      }
    }
  };
  barChart = new Chart(ctx, config);
}

function createRevenueChart() {
  if (!revenueChartCanvas.value) return;
  const ctx = revenueChartCanvas.value.getContext('2d');
  if (!ctx) return;

  if (revenueChart) revenueChart.destroy();

  if (!revenueChartData.value || !revenueChartData.value.labels || !revenueChartData.value.data) {
    return;
  }

  const gradient = ctx.createLinearGradient(0, 0, 0, 280);
  gradient.addColorStop(0, 'rgba(5, 150, 105, 0.18)');
  gradient.addColorStop(1, 'rgba(5, 150, 105, 0.0)');

  const config: ChartConfiguration<'line'> = {
    type: 'line',
    data: {
      labels: revenueChartData.value.labels,
      datasets: [{
        label: 'Pendapatan (IDR)',
        data: revenueChartData.value.data,
        borderColor: '#059669',
        backgroundColor: gradient,
        pointBackgroundColor: '#059669',
        pointBorderColor: '#ffffff',
        pointHoverBackgroundColor: '#ffffff',
        pointHoverBorderColor: '#059669',
        pointRadius: 4,
        pointHoverRadius: 6,
        borderWidth: 2.5,
        tension: 0.35,
        fill: true,
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        y: {
          beginAtZero: true,
          ticks: {
            callback: (value) => new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0, notation: 'compact' }).format(value as number),
            color: '#64748b'
          },
          grid: {
            color: 'rgba(226, 232, 240, 0.6)'
          }
        },
        x: { 
          grid: { 
            display: false 
          },
          ticks: {
            color: '#64748b'
          }
        }
      },
      plugins: {
        legend: { 
          display: false
        },
        tooltip: {
          backgroundColor: '#0f172a',
          titleColor: '#ffffff',
          bodyColor: '#e2e8f0',
          padding: 10,
          cornerRadius: 6,
          displayColors: false,
          callbacks: {
            label: (context) => `Pendapatan: ${formatCurrency(Number(context.parsed.y) || 0)}`
          }
        }
      }
    }
  };
  revenueChart = new Chart(ctx, config);
}

function initializeCharts() {
  setTimeout(() => {
    createPieChart();
    createBarChart();

    if (growthChartData.value && growthChartData.value.labels && growthChartData.value.data) {
      createLineChart();
    }

    if (hasRevenueData.value) {
      createRevenueChart();
    }
  }, 100);
}

watch(revenueTimeframe, (newTimeframe) => {
  fetchDashboardData(newTimeframe);
});

watch([stats, growthChartData], () => {
  initializeCharts();
}, { deep: true });

onMounted(() => {
  fetchDashboardData(revenueTimeframe.value);
});
</script>

<style scoped>
.gap-1 { gap: 0.25rem; }
.gap-2 { gap: 0.5rem; }
.gap-3 { gap: 0.75rem; }
.gap-4 { gap: 1rem; }

/* ==========================================================================
   Header Card (Standard Clean SaaS Style - NO PURPLE GRADIENT)
   ========================================================================== */
.header-card {
  background: #ffffff;
  border-radius: 8px;
  padding: 24px;
  color: #0f172a;
  border: 1px solid #e2e8f0;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

@media (min-width: 768px) {
  .header-card {
    padding: 28px 32px;
    border-radius: 12px;
  }
}

.header-avatar-wrapper {
  width: 50px;
  height: 50px;
  border-radius: 12px;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.header-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0;
  line-height: 1.2;
  letter-spacing: -0.01em;
}

.header-subtitle {
  font-size: 0.875rem;
  color: #64748b;
  margin: 4px 0 0 0;
}

.realtime-status-badge {
  display: inline-flex;
  align-items: center;
  padding: 6px 12px;
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: 8px;
  font-size: 0.8125rem;
  font-weight: 600;
  color: #166534;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #22c55e;
  margin-right: 8px;
  display: inline-block;
  animation: pulse-dot 2s infinite;
}

@keyframes pulse-dot {
  0% { transform: scale(0.95); opacity: 0.8; }
  50% { transform: scale(1.15); opacity: 1; }
  100% { transform: scale(0.95); opacity: 0.8; }
}

.refresh-btn {
  background: #ffffff !important;
  color: #334155 !important;
  border: 1px solid #cbd5e1 !important;
  border-radius: 8px !important;
  height: 38px !important;
}

.refresh-btn:hover {
  background: #f8fafc !important;
  border-color: #94a3b8 !important;
  color: #0f172a !important;
}

/* ==========================================================================
   Metric Stats Cards
   ========================================================================== */
.metric-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.metric-card:hover {
  border-color: #cbd5e1;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.04);
}

.metric-label {
  font-size: 0.75rem;
  font-weight: 700;
  color: #64748b;
  letter-spacing: 0.05em;
}

.metric-value {
  font-size: 1.875rem;
  font-weight: 700;
  color: #0f172a;
  line-height: 1.15;
  font-variant-numeric: tabular-nums;
}

.metric-icon-box {
  width: 38px;
  height: 38px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.box-blue {
  background: #eff6ff;
  border: 1px solid #dbeafe;
}

.box-green {
  background: #ecfdf5;
  border: 1px solid #a7f3d0;
}

.box-rose {
  background: #fff1f2;
  border: 1px solid #fecdd3;
}

.box-cyan {
  background: #f0f9ff;
  border: 1px solid #bae6fd;
}

.meta-pill {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 600;
}

.pill-success {
  background: #ecfdf5;
  color: #047857;
  border: 1px solid #a7f3d0;
}

.pill-rose {
  background: #fff1f2;
  color: #be123c;
  border: 1px solid #fecdd3;
}

.pill-subtle {
  background: #f1f5f9;
  color: #475569;
  border: 1px solid #e2e8f0;
}

/* ==========================================================================
   Revenue Hero Card
   ========================================================================== */
.revenue-hero-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.revenue-tag {
  font-size: 0.6875rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  color: #059669;
  display: block;
  margin-bottom: 2px;
}

.revenue-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: #0f172a;
}

.revenue-desc {
  font-size: 0.8125rem;
  color: #64748b;
}

.revenue-icon-box {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: #ecfdf5;
  border: 1px solid #a7f3d0;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.revenue-amount-value {
  font-size: 2.25rem;
  font-weight: 700;
  color: #047857;
  line-height: 1.1;
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.02em;
}

.revenue-amount-sub {
  font-size: 0.875rem;
  color: #64748b;
  font-weight: 500;
}

.revenue-meta-pill {
  display: inline-flex;
  align-items: center;
  padding: 4px 12px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 0.8125rem;
  color: #334155;
}

/* ==========================================================================
   Chart Panel Cards
   ========================================================================== */
.chart-panel-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  overflow: hidden;
}

.chart-panel-header {
  background: #ffffff;
}

.chart-header-icon-box {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.chart-panel-title {
  font-size: 0.9375rem;
  font-weight: 700;
  color: #0f172a;
}

.chart-panel-subtitle {
  font-size: 0.75rem;
  color: #64748b;
}

/* Empty Chart State */
.empty-chart-state {
  height: 260px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 24px;
}

.empty-chart-icon-box {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-chart-title {
  font-size: 1rem;
  font-weight: 700;
  color: #0f172a;
}

.empty-chart-desc {
  font-size: 0.8125rem;
  color: #64748b;
  max-width: 320px;
}

/* Quick Stats */
.stat-item-label {
  font-size: 0.8125rem;
  color: #64748b;
  font-weight: 500;
}

.stat-item-val {
  font-size: 0.875rem;
  font-weight: 700;
  color: #0f172a;
}

.stat-item-badge {
  font-size: 0.75rem;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 6px;
}

.badge-positive {
  background: #ecfdf5;
  color: #047857;
  border: 1px solid #a7f3d0;
}

.badge-negative {
  background: #fff1f2;
  color: #be123c;
  border: 1px solid #fecdd3;
}

.arpu-box {
  background: #f8fafc;
  border-color: #e2e8f0 !important;
}

.arpu-val {
  color: #0f172a;
  font-variant-numeric: tabular-nums;
}

/* Segmented Toggle */
.segmented-toggle {
  border-radius: 8px !important;
  border-color: #cbd5e1 !important;
  background: #ffffff !important;
}

/* ==========================================================================
   Dark Theme Adjustments
   ========================================================================== */
.v-theme--dark .header-card,
.v-theme--dark .metric-card,
.v-theme--dark .revenue-hero-card,
.v-theme--dark .chart-panel-card,
.v-theme--dark .chart-panel-header {
  background: #1e293b !important;
  border-color: #334155 !important;
  color: #f8fafc !important;
}

.v-theme--dark .header-title,
.v-theme--dark .revenue-title,
.v-theme--dark .chart-panel-title,
.v-theme--dark .metric-value,
.v-theme--dark .stat-item-val,
.v-theme--dark .arpu-val,
.v-theme--dark .empty-chart-title {
  color: #f8fafc !important;
}

.v-theme--dark .header-avatar-wrapper,
.v-theme--dark .chart-header-icon-box,
.v-theme--dark .empty-chart-icon-box,
.v-theme--dark .arpu-box {
  background: #0f172a !important;
  border-color: #334155 !important;
  color: #f8fafc !important;
}

.v-theme--dark .refresh-btn,
.v-theme--dark .segmented-toggle {
  background: #1e293b !important;
  border-color: #334155 !important;
  color: #f8fafc !important;
}

.v-theme--dark .revenue-meta-pill,
.v-theme--dark .pill-subtle {
  background: #0f172a !important;
  border-color: #334155 !important;
  color: #cbd5e1 !important;
}

.v-theme--dark .realtime-status-badge {
  background: rgba(34, 197, 94, 0.15) !important;
  border-color: rgba(34, 197, 94, 0.3) !important;
  color: #4ade80 !important;
}

.v-theme--dark .box-blue {
  background: rgba(59, 130, 246, 0.15) !important;
  border-color: rgba(59, 130, 246, 0.3) !important;
}

.v-theme--dark .box-green,
.v-theme--dark .revenue-icon-box {
  background: rgba(34, 197, 94, 0.15) !important;
  border-color: rgba(34, 197, 94, 0.3) !important;
}

.v-theme--dark .box-rose {
  background: rgba(244, 63, 94, 0.15) !important;
  border-color: rgba(244, 63, 94, 0.3) !important;
}

.v-theme--dark .box-cyan {
  background: rgba(6, 182, 212, 0.15) !important;
  border-color: rgba(6, 182, 212, 0.3) !important;
}

@media (max-width: 600px) {
  .revenue-amount-value {
    font-size: 1.75rem;
  }
}
</style>
