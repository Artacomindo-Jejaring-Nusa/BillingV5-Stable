<template>
  <div class="trouble-ticket-report-container">
    <!-- ===== HEADER SECTION ===== -->
    <div class="header-card mb-4 mb-md-6">
      <div class="d-flex flex-column flex-md-row align-start align-md-center justify-space-between gap-4">
        <div class="d-flex align-center header-info">
          <div class="header-avatar-box">
            <v-icon color="white" size="24">mdi-chart-box-outline</v-icon>
          </div>
          <div class="ml-4">
            <div class="d-flex align-center gap-2">
              <h1 class="header-title mb-0">Trouble Ticket Reports</h1>
              <v-chip size="x-small" color="primary" variant="tonal" class="font-weight-bold ml-1">
                Analytics
              </v-chip>
            </div>
            <p class="header-subtitle mb-0 mt-0.5">
              Analisis komprehensif, tracking downtime, dan performa penanganan gangguan
            </p>
          </div>
        </div>
        <div class="header-actions">
          <v-btn
            prepend-icon="mdi-refresh"
            @click="refreshAllData"
            :loading="loading"
            class="action-btn-refresh text-none font-weight-medium"
            variant="outlined"
            rounded="pill"
            height="38"
          >
            Segarkan Data
          </v-btn>
          <v-btn
            prepend-icon="mdi-file-export-outline"
            @click="exportReport"
            :disabled="loading || exporting"
            :loading="exporting"
            class="export-btn-emerald text-none font-weight-bold"
            variant="flat"
            rounded="pill"
            elevation="0"
            height="38"
          >
            Export Laporan
          </v-btn>
        </div>
      </div>
    </div>

    <!-- ===== DATE RANGE FILTER ===== -->
    <div class="filter-card mb-4 mb-md-6 pa-4 pa-sm-5">
      <div class="d-flex align-center justify-space-between mb-3">
        <div class="d-flex align-center">
          <div class="filter-icon-box mr-3">
            <v-icon color="primary" size="20">mdi-calendar-range</v-icon>
          </div>
          <div>
            <h3 class="filter-title mb-0">Periode Laporan</h3>
            <p class="filter-subtitle mb-0">Pilih rentang tanggal untuk memfilter analitik tiket</p>
          </div>
        </div>
        <v-chip
          v-if="dateRange.from || dateRange.to"
          color="primary"
          variant="tonal"
          size="small"
          prepend-icon="mdi-filter-check"
          class="font-weight-medium"
        >
          Filter Aktif
        </v-chip>
      </div>

      <div class="date-range-filters mt-3">
        <v-row dense class="align-center">
          <v-col cols="12" sm="5" md="4" lg="3">
            <v-text-field
              v-model="dateRange.from"
              type="date"
              label="Tanggal Mulai"
              prepend-inner-icon="mdi-calendar-start"
              variant="outlined"
              density="compact"
              hide-details
              @update:modelValue="applyDateFilter"
              class="date-input rounded-lg bg-white"
            />
          </v-col>
          <v-col cols="12" sm="5" md="4" lg="3">
            <v-text-field
              v-model="dateRange.to"
              type="date"
              label="Tanggal Akhir"
              prepend-inner-icon="mdi-calendar-end"
              variant="outlined"
              density="compact"
              hide-details
              @update:modelValue="applyDateFilter"
              class="date-input rounded-lg bg-white"
            />
          </v-col>
          <v-col cols="12" sm="2" md="2" lg="2">
            <v-btn
              variant="tonal"
              prepend-icon="mdi-filter-off-outline"
              @click="clearDateFilter"
              :disabled="!dateRange.from && !dateRange.to"
              block
              color="error"
              class="btn-clear text-none font-weight-bold"
              height="40"
              rounded="lg"
            >
              Reset Filter
            </v-btn>
          </v-col>
        </v-row>
      </div>
    </div>

    <!-- ===== KEY METRICS OVERVIEW ===== -->
    <div class="metrics-overview mb-4 mb-md-6 mt-6">
      <div class="section-heading mb-4">
        <div class="d-flex align-center">
          <div class="section-icon-badge mr-3">
            <v-icon size="20" color="primary">mdi-view-dashboard-outline</v-icon>
          </div>
          <div>
            <h2 class="section-title mb-0">Key Performance Metrics</h2>
            <p class="section-subtitle mb-0">Ringkasan performa tiket dan kecepatan penanganan gangguan</p>
          </div>
        </div>
      </div>

      <v-row>
        <!-- Total Tickets Card -->
        <v-col cols="12" sm="6" md="3">
          <div class="metric-card pa-4 pa-sm-5">
            <div class="metric-accent-line line-blue"></div>
            <div class="d-flex align-center justify-space-between mb-3 mt-1">
              <div class="d-flex align-center">
                <span class="metric-dot dot-blue mr-2"></span>
                <span class="metric-label">TOTAL TICKETS</span>
              </div>
              <div class="metric-icon-box box-blue">
                <v-icon size="20" color="primary">mdi-ticket-confirmation-outline</v-icon>
              </div>
            </div>
            <div class="metric-value mb-2">{{ statistics?.total_tickets || 0 }}</div>
            <div class="d-flex align-center justify-space-between flex-wrap gap-2 mt-auto pt-2 border-t">
              <span class="meta-pill pill-subtle">
                <v-icon size="12" class="mr-1">mdi-history</v-icon>
                All time
              </span>
              <span class="text-caption text-medium-emphasis">Total dibuat</span>
            </div>
          </div>
        </v-col>

        <!-- Resolution Rate Card -->
        <v-col cols="12" sm="6" md="3">
          <div class="metric-card pa-4 pa-sm-5">
            <div class="metric-accent-line line-green"></div>
            <div class="d-flex align-center justify-space-between mb-3 mt-1">
              <div class="d-flex align-center">
                <span class="metric-dot dot-green mr-2"></span>
                <span class="metric-label">RESOLUTION RATE</span>
              </div>
              <div class="metric-icon-box box-green">
                <v-icon size="20" color="success">mdi-check-circle-outline</v-icon>
              </div>
            </div>
            <div class="metric-value text-success mb-2">{{ calculateResolutionRate() }}%</div>
            <div class="d-flex align-center justify-space-between flex-wrap gap-2 mt-auto pt-2 border-t">
              <span class="meta-pill pill-success">
                <v-icon size="12" class="mr-1">mdi-check</v-icon>
                Performa baik
              </span>
              <span class="text-caption text-medium-emphasis">Tiket terselesaikan</span>
            </div>
          </div>
        </v-col>

        <!-- Avg Resolution Time Card -->
        <v-col cols="12" sm="6" md="3">
          <div class="metric-card pa-4 pa-sm-5">
            <div class="metric-accent-line line-amber"></div>
            <div class="d-flex align-center justify-space-between mb-3 mt-1">
              <div class="d-flex align-center">
                <span class="metric-dot dot-amber mr-2"></span>
                <span class="metric-label">AVG RESOLUTION TIME</span>
              </div>
              <div class="metric-icon-box box-amber">
                <v-icon size="20" color="warning">mdi-clock-fast</v-icon>
              </div>
            </div>
            <div class="metric-value text-warning mb-2">{{ formatHours(statistics?.avg_resolution_time_hours || null) }}</div>
            <div class="d-flex align-center justify-space-between flex-wrap gap-2 mt-auto pt-2 border-t">
              <span class="meta-pill pill-amber">
                <v-icon size="12" class="mr-1">mdi-clock-outline</v-icon>
                {{ getResolutionTimeStatus() }}
              </span>
              <span class="text-caption text-medium-emphasis">Rata-rata waktu</span>
            </div>
          </div>
        </v-col>

        <!-- Open Tickets Card -->
        <v-col cols="12" sm="6" md="3">
          <div class="metric-card pa-4 pa-sm-5">
            <div class="metric-accent-line line-rose"></div>
            <div class="d-flex align-center justify-space-between mb-3 mt-1">
              <div class="d-flex align-center">
                <span class="metric-dot dot-rose mr-2"></span>
                <span class="metric-label">OPEN TICKETS</span>
              </div>
              <div class="metric-icon-box box-rose">
                <v-icon size="20" color="error">mdi-alert-circle-outline</v-icon>
              </div>
            </div>
            <div class="metric-value text-error mb-2">{{ statistics?.open_tickets || 0 }}</div>
            <div class="d-flex align-center justify-space-between flex-wrap gap-2 mt-auto pt-2 border-t">
              <span class="meta-pill pill-rose">
                <v-icon size="12" class="mr-1">mdi-alert</v-icon>
                Perlu atensi
              </span>
              <span class="text-caption text-medium-emphasis">Sedang ditangani</span>
            </div>
          </div>
        </v-col>
      </v-row>
    </div>

    <!-- ===== CHARTS SECTION ===== -->
    <div class="charts-section mb-4 mb-md-6">
      <v-row>
        <!-- Monthly Trends Chart -->
        <v-col cols="12" lg="8">
          <div class="chart-panel-card h-100">
            <div class="chart-panel-header pa-4 pa-sm-5 d-flex flex-column flex-sm-row align-start align-sm-center justify-space-between gap-3 border-b">
              <div class="d-flex align-center gap-2">
                <div class="chart-header-icon-box">
                  <v-icon color="primary" size="18">mdi-chart-line</v-icon>
                </div>
                <div>
                  <h3 class="chart-panel-title mb-0">Monthly Trends</h3>
                  <p class="chart-panel-subtitle mb-0">Tren volume dan resolusi tiket bulanan</p>
                </div>
              </div>
              <v-btn-toggle
                v-model="trendPeriod"
                mandatory
                variant="outlined"
                density="compact"
                class="segmented-toggle"
              >
                <v-btn value="6" size="small" class="text-none">6 Bulan</v-btn>
                <v-btn value="12" size="small" class="text-none">1 Tahun</v-btn>
                <v-btn value="24" size="small" class="text-none">2 Tahun</v-btn>
              </v-btn-toggle>
            </div>
            <div class="chart-panel-body pa-4 pa-sm-6">
              <div v-if="loadingMonthly" class="loading-state py-8 text-center">
                <v-progress-circular 
                  indeterminate 
                  color="primary" 
                  size="48"
                  width="4"
                ></v-progress-circular>
                <p class="loading-text mt-3 text-caption text-medium-emphasis">Memuat tren bulanan...</p>
              </div>
              <div v-else-if="monthlyTrends.length > 0" class="trends-chart">
                <Line
                  :data="trendsChartData"
                  :options="trendsChartOptions"
                  height="120"
                />
              </div>
              <div v-else class="empty-chart-state py-8">
                <div class="empty-chart-icon-box mb-2">
                  <v-icon size="32" color="slate-400">mdi-chart-line</v-icon>
                </div>
                <h4 class="empty-chart-title mb-1">Tidak Ada Data Tren</h4>
                <p class="empty-chart-desc mb-0">Data grafik akan muncul ketika tiket mulai tercatat dalam rentang waktu ini.</p>
              </div>
            </div>
          </div>
        </v-col>

        <!-- Category Distribution -->
        <v-col cols="12" lg="4">
          <div class="chart-panel-card h-100">
            <div class="chart-panel-header pa-4 pa-sm-5 d-flex align-center justify-space-between border-b">
              <div class="d-flex align-center gap-2">
                <div class="chart-header-icon-box">
                  <v-icon color="primary" size="18">mdi-chart-pie</v-icon>
                </div>
                <div>
                  <h3 class="chart-panel-title mb-0">Category Distribution</h3>
                  <p class="chart-panel-subtitle mb-0">Persentase tiket berdasarkan kategori</p>
                </div>
              </div>
            </div>
            <div class="chart-panel-body pa-4 pa-sm-6">
              <div v-if="loadingCategory" class="loading-state py-8 text-center">
                <v-progress-circular 
                  indeterminate 
                  color="primary" 
                  size="48"
                  width="4"
                ></v-progress-circular>
                <p class="loading-text mt-3 text-caption text-medium-emphasis">Memuat kategori...</p>
              </div>
              <div v-else-if="categoryPerformance.length > 0" class="category-chart">
                <Pie
                  :data="categoryChartData"
                  :options="categoryChartOptions"
                  height="200"
                />
              </div>
              <div v-else class="empty-chart-state py-8">
                <div class="empty-chart-icon-box mb-2">
                  <v-icon size="32" color="slate-400">mdi-chart-pie</v-icon>
                </div>
                <h4 class="empty-chart-title mb-1">Tidak Ada Data Kategori</h4>
                <p class="empty-chart-desc mb-0">Data akan muncul setelah tiket memiliki kategori masalah.</p>
              </div>
            </div>
          </div>
        </v-col>
      </v-row>
    </div>

    <!-- ===== DETAILED PERFORMANCE TABLES SECTION ===== -->
    <div class="tables-section mb-6">
      <v-row>
        <!-- Category Performance Table -->
        <v-col cols="12" lg="6">
          <div class="chart-panel-card h-100">
            <div class="chart-panel-header pa-4 pa-sm-5 d-flex align-center justify-space-between border-b">
              <div class="d-flex align-center gap-2">
                <div class="chart-header-icon-box">
                  <v-icon color="primary" size="18">mdi-tag-outline</v-icon>
                </div>
                <div>
                  <h3 class="chart-panel-title mb-0">Category Performance</h3>
                  <p class="chart-panel-subtitle mb-0">Metrik performa per kategori gangguan</p>
                </div>
              </div>
              <v-chip size="small" variant="tonal" color="primary">
                {{ categoryPerformance.length }} Kategori
              </v-chip>
            </div>
            <div class="pa-0">
              <div v-if="loadingCategory" class="loading-state py-8 text-center">
                <v-progress-circular 
                  indeterminate 
                  color="primary" 
                  size="40"
                  width="3"
                ></v-progress-circular>
                <p class="loading-text mt-3 text-caption text-medium-emphasis">Memuat performa kategori...</p>
              </div>
              <v-data-table
                v-else
                :headers="categoryHeaders"
                :items="categoryPerformance"
                :loading="loadingCategory"
                class="modern-data-table"
                density="comfortable"
                hide-default-footer
                :items-per-page="-1"
              >
                <template v-slot:item.category="{ item }">
                  <v-chip
                    size="small"
                    variant="tonal"
                    color="primary"
                    prepend-icon="mdi-tag-outline"
                    class="font-weight-medium"
                  >
                    {{ item.category_display }}
                  </v-chip>
                </template>
                <template v-slot:item.resolution_rate="{ item }">
                  <div class="d-flex align-center gap-2" style="max-width: 160px;">
                    <v-progress-linear
                      :model-value="item.resolution_rate_percent"
                      :color="getResolutionRateColor(item.resolution_rate_percent)"
                      height="8"
                      rounded
                      class="flex-grow-1"
                    ></v-progress-linear>
                    <span class="text-caption font-weight-bold" style="min-width: 36px; text-align: right;">
                      {{ Math.ceil(item.resolution_rate_percent) }}%
                    </span>
                  </div>
                </template>
                <template v-slot:item.avg_resolution="{ item }">
                  <span class="text-body-2 font-weight-medium text-slate-700">
                    {{ formatHours(item.avg_resolution_hours) }}
                  </span>
                </template>
              </v-data-table>
            </div>
          </div>
        </v-col>

        <!-- User Performance Table -->
        <v-col cols="12" lg="6">
          <div class="chart-panel-card h-100">
            <div class="chart-panel-header pa-4 pa-sm-5 d-flex align-center justify-space-between border-b">
              <div class="d-flex align-center gap-2">
                <div class="chart-header-icon-box">
                  <v-icon color="primary" size="18">mdi-account-group-outline</v-icon>
                </div>
                <div>
                  <h3 class="chart-panel-title mb-0">Technician Performance</h3>
                  <p class="chart-panel-subtitle mb-0">Metrik produktivitas masing-masing teknisi</p>
                </div>
              </div>
              <v-chip size="small" variant="tonal" color="success">
                {{ userPerformance.length }} Teknisi
              </v-chip>
            </div>
            <div class="pa-0">
              <div v-if="loadingUser" class="loading-state py-8 text-center">
                <v-progress-circular 
                  indeterminate 
                  color="primary" 
                  size="40"
                  width="3"
                ></v-progress-circular>
                <p class="loading-text mt-3 text-caption text-medium-emphasis">Memuat performa teknisi...</p>
              </div>
              <v-data-table
                v-else
                :headers="userHeaders"
                :items="userPerformance"
                :loading="loadingUser"
                class="modern-data-table"
                density="comfortable"
                hide-default-footer
                :items-per-page="-1"
              >
                <template v-slot:item.user_name="{ item }">
                  <div class="d-flex align-center gap-2">
                    <v-avatar size="28" color="slate-100" class="border">
                      <span class="text-caption font-weight-bold text-slate-700">
                        {{ item.user_name.charAt(0).toUpperCase() }}
                      </span>
                    </v-avatar>
                    <span class="font-weight-medium text-body-2 text-slate-800">{{ item.user_name }}</span>
                  </div>
                </template>
                <template v-slot:item.resolution_rate="{ item }">
                  <div class="d-flex align-center gap-2" style="max-width: 160px;">
                    <v-progress-linear
                      :model-value="item.resolution_rate_percent"
                      :color="getResolutionRateColor(item.resolution_rate_percent)"
                      height="8"
                      rounded
                      class="flex-grow-1"
                    ></v-progress-linear>
                    <span class="text-caption font-weight-bold" style="min-width: 36px; text-align: right;">
                      {{ Math.ceil(item.resolution_rate_percent) }}%
                    </span>
                  </div>
                </template>
                <template v-slot:item.avg_resolution="{ item }">
                  <span class="text-body-2 font-weight-medium text-slate-700">
                    {{ formatHours(item.avg_resolution_hours) }}
                  </span>
                </template>
              </v-data-table>
            </div>
          </div>
        </v-col>
      </v-row>
    </div>

    <!-- ===== DOWNTIME ANALYSIS SECTION ===== -->
    <div class="downtime-section mb-6">
      <div class="chart-panel-card">
        <div class="chart-panel-header pa-4 pa-sm-5 d-flex align-center justify-space-between border-b">
          <div class="d-flex align-center gap-2">
            <div class="chart-header-icon-box">
              <v-icon color="primary" size="18">mdi-clock-alert-outline</v-icon>
            </div>
            <div>
              <h3 class="chart-panel-title mb-0">Downtime Analysis</h3>
              <p class="chart-panel-subtitle mb-0">Pelacakan durasi downtime dan dampaknya pada pelanggan</p>
            </div>
          </div>
          <v-btn
            variant="outlined"
            prepend-icon="mdi-refresh"
            @click="loadDowntimeAnalysis"
            :loading="loadingDowntime"
            size="small"
            color="primary"
          >
            Refresh Data
          </v-btn>
        </div>

        <div class="pa-4 pa-sm-6">
          <div v-if="loadingDowntime" class="loading-state py-10 text-center">
            <v-progress-circular 
              indeterminate 
              color="primary" 
              size="44"
              width="3"
            ></v-progress-circular>
            <p class="loading-text mt-3 text-caption text-medium-emphasis">Menganalisis data downtime...</p>
          </div>
          <div v-else-if="downtimeAnalysis && downtimeAnalysis.overall_statistics" class="downtime-content">
            <!-- Overall Statistics Cards -->
            <div class="mb-6">
              <h4 class="text-subtitle-2 font-weight-bold text-slate-800 mb-3 d-flex align-center gap-2">
                <v-icon size="16" color="primary">mdi-chart-box-outline</v-icon>
                Statistik Akumulasi Downtime
              </h4>
              <v-row>
                <!-- Total Downtime -->
                <v-col cols="12" sm="6" lg="3">
                  <div class="metric-card pa-4 pa-sm-5">
                    <div class="metric-accent-line line-blue"></div>
                    <div class="d-flex align-center justify-space-between mb-3 mt-1">
                      <div class="d-flex align-center">
                        <span class="metric-dot dot-blue mr-2"></span>
                        <span class="metric-label">TOTAL DOWNTIME</span>
                      </div>
                      <div class="metric-icon-box box-blue">
                        <v-icon size="20" color="primary">mdi-clock-outline</v-icon>
                      </div>
                    </div>
                    <div class="metric-value mb-2">{{ formatHours(downtimeAnalysis.overall_statistics.total_downtime_hours) }}</div>
                    <div class="d-flex align-center justify-space-between flex-wrap gap-2 mt-auto pt-2 border-t">
                      <span class="meta-pill pill-subtle">
                        <v-icon size="12" class="mr-1">mdi-clock-time-four-outline</v-icon>
                        Semua Tiket
                      </span>
                      <span class="text-caption text-medium-emphasis">Akumulasi durasi</span>
                    </div>
                  </div>
                </v-col>

                <!-- Average Downtime -->
                <v-col cols="12" sm="6" lg="3">
                  <div class="metric-card pa-4 pa-sm-5">
                    <div class="metric-accent-line line-green"></div>
                    <div class="d-flex align-center justify-space-between mb-3 mt-1">
                      <div class="d-flex align-center">
                        <span class="metric-dot dot-green mr-2"></span>
                        <span class="metric-label">AVG DOWNTIME</span>
                      </div>
                      <div class="metric-icon-box box-green">
                        <v-icon size="20" color="success">mdi-chart-timeline-variant</v-icon>
                      </div>
                    </div>
                    <div class="metric-value text-success mb-2">{{ formatHours(downtimeAnalysis.overall_statistics.avg_downtime_hours) }}</div>
                    <div class="d-flex align-center justify-space-between flex-wrap gap-2 mt-auto pt-2 border-t">
                      <span class="meta-pill pill-success">
                        <v-icon size="12" class="mr-1">mdi-timer-outline</v-icon>
                        Rata-rata
                      </span>
                      <span class="text-caption text-medium-emphasis">Per tiket gangguan</span>
                    </div>
                  </div>
                </v-col>

                <!-- Maximum Downtime -->
                <v-col cols="12" sm="6" lg="3">
                  <div class="metric-card pa-4 pa-sm-5">
                    <div class="metric-accent-line line-amber"></div>
                    <div class="d-flex align-center justify-space-between mb-3 mt-1">
                      <div class="d-flex align-center">
                        <span class="metric-dot dot-amber mr-2"></span>
                        <span class="metric-label">MAX DOWNTIME</span>
                      </div>
                      <div class="metric-icon-box box-amber">
                        <v-icon size="20" color="warning">mdi-clock-alert-outline</v-icon>
                      </div>
                    </div>
                    <div class="metric-value text-warning mb-2">{{ formatHours(downtimeAnalysis.overall_statistics.max_downtime_hours) }}</div>
                    <div class="d-flex align-center justify-space-between flex-wrap gap-2 mt-auto pt-2 border-t">
                      <span class="meta-pill pill-amber">
                        <v-icon size="12" class="mr-1">mdi-alert-outline</v-icon>
                        Tertinggi
                      </span>
                      <span class="text-caption text-medium-emphasis">Durasi terlama</span>
                    </div>
                  </div>
                </v-col>

                <!-- Affected Tickets -->
                <v-col cols="12" sm="6" lg="3">
                  <div class="metric-card pa-4 pa-sm-5">
                    <div class="metric-accent-line line-rose"></div>
                    <div class="d-flex align-center justify-space-between mb-3 mt-1">
                      <div class="d-flex align-center">
                        <span class="metric-dot dot-rose mr-2"></span>
                        <span class="metric-label">TIKET TERDAMPAK</span>
                      </div>
                      <div class="metric-icon-box box-rose">
                        <v-icon size="20" color="error">mdi-ticket-alert-outline</v-icon>
                      </div>
                    </div>
                    <div class="metric-value text-error mb-2">{{ downtimeAnalysis.overall_statistics.tickets_with_downtime }}</div>
                    <div class="d-flex align-center justify-space-between flex-wrap gap-2 mt-auto pt-2 border-t">
                      <span class="meta-pill pill-rose">
                        <v-icon size="12" class="mr-1">mdi-alert-circle-outline</v-icon>
                        Terdampak
                      </span>
                      <span class="text-caption text-medium-emphasis">Dengan downtime</span>
                    </div>
                  </div>
                </v-col>
              </v-row>
            </div>

            <!-- Top Customers with Downtime -->
            <div class="top-customers" v-if="downtimeAnalysis.top_customers && downtimeAnalysis.top_customers.length > 0">
              <h4 class="text-subtitle-2 font-weight-bold text-slate-800 mb-3 d-flex align-center gap-2">
                <v-icon size="16" color="warning">mdi-account-alert-outline</v-icon>
                Pelanggan dengan Downtime Tertinggi
              </h4>
              <v-data-table
                :headers="customerDowntimeHeaders"
                :items="downtimeAnalysis.top_customers.slice(0, 5)"
                density="comfortable"
                hide-default-footer
                class="modern-data-table"
              >
                <template v-slot:item.customer_name="{ item }">
                  <div class="d-flex align-center gap-2">
                    <v-avatar size="28" color="slate-100" class="border">
                      <span class="text-caption font-weight-bold text-slate-700">
                        {{ item.customer_name.charAt(0).toUpperCase() }}
                      </span>
                    </v-avatar>
                    <span class="font-weight-medium text-body-2 text-slate-800">{{ item.customer_name }}</span>
                  </div>
                </template>
                <template v-slot:item.total_downtime="{ item }">
                  <v-chip
                    size="small"
                    :color="getDowntimeSeverityColor(item.total_downtime_hours)"
                    variant="tonal"
                    prepend-icon="mdi-clock-alert-outline"
                    class="font-weight-medium"
                  >
                    {{ formatHours(item.total_downtime_hours) }}
                  </v-chip>
                </template>
                <template v-slot:item.avg_downtime_hours="{ item }">
                  <span class="text-body-2 font-weight-medium text-slate-700">
                    {{ formatHours(item.avg_downtime_hours) }}
                  </span>
                </template>
              </v-data-table>
            </div>
          </div>
          <div v-else class="empty-state py-10 text-center">
            <div class="empty-chart-icon-box mb-2 mx-auto">
              <v-icon size="36" color="slate-400">mdi-clock-check-outline</v-icon>
            </div>
            <h4 class="text-subtitle-1 font-weight-bold text-slate-800 mb-1">Tidak Ada Data Downtime</h4>
            <p class="text-caption text-slate-500 mb-0">Semua layanan stabil atau belum ada tiket yang mencatat waktu downtime pada periode ini.</p>
          </div>
        </div>
      </div>
    </div>

    <!-- ===== TICKET DETAILS SECTION ===== -->
    <div class="ticket-details-section mb-6">
      <div class="chart-panel-card">
        <div class="chart-panel-header pa-4 pa-sm-5 d-flex align-center justify-space-between border-b flex-wrap gap-3">
          <div class="d-flex align-center gap-2">
            <div class="chart-header-icon-box">
              <v-icon color="primary" size="18">mdi-ticket-account</v-icon>
            </div>
            <div>
              <h3 class="chart-panel-title mb-0">Detail Tiket & Informasi Pelanggan</h3>
              <p class="chart-panel-subtitle mb-0">Data lengkap riwayat penanganan gangguan dan data teknis pelanggan</p>
            </div>
          </div>
          <div class="d-flex align-center gap-2">
            <v-chip size="small" variant="tonal" color="primary">
              {{ ticketDetails.length }} Total Tiket
            </v-chip>
            <v-btn
              variant="outlined"
              prepend-icon="mdi-refresh"
              @click="loadTicketDetails"
              :loading="loadingTicketDetails"
              size="small"
              color="primary"
            >
              Refresh
            </v-btn>
          </div>
        </div>

        <div class="pa-4 pa-sm-6">
          <div v-if="loadingTicketDetails" class="loading-state py-10 text-center">
            <v-progress-circular 
              indeterminate 
              color="primary" 
              size="44"
              width="3"
            ></v-progress-circular>
            <p class="loading-text mt-3 text-caption text-medium-emphasis">Memuat detail tiket...</p>
          </div>
          <div v-else-if="ticketDetails.length > 0" class="ticket-details-content">
            <!-- Filter Bar -->
            <div class="ticket-filter-bar mb-4 pa-3 pa-sm-4 rounded-xl border">
              <v-row dense align="center">
                <v-col cols="12" md="4">
                  <v-text-field
                    v-model="ticketSearch"
                    placeholder="Cari no tiket, pelanggan, atau teknisi..."
                    prepend-inner-icon="mdi-magnify"
                    variant="outlined"
                    density="compact"
                    hide-details
                    clearable
                    class="bg-white filter-input-control"
                  ></v-text-field>
                </v-col>
                <v-col cols="12" sm="6" md="3">
                  <v-select
                    v-model="ticketStatusFilter"
                    :items="ticketStatusOptions"
                    placeholder="Filter Status"
                    prepend-inner-icon="mdi-filter-outline"
                    variant="outlined"
                    density="compact"
                    hide-details
                    clearable
                    class="bg-white filter-input-control"
                  ></v-select>
                </v-col>
                <v-col cols="12" sm="6" md="3">
                  <v-select
                    v-model="ticketCategoryFilter"
                    :items="ticketCategoryOptions"
                    placeholder="Filter Kategori"
                    prepend-inner-icon="mdi-tag-outline"
                    variant="outlined"
                    density="compact"
                    hide-details
                    clearable
                    class="bg-white filter-input-control"
                  ></v-select>
                </v-col>
                <v-col cols="12" md="2">
                  <v-btn
                    variant="tonal"
                    prepend-icon="mdi-filter-off-outline"
                    @click="clearTicketFilters"
                    size="small"
                    block
                    color="slate-600"
                    height="40"
                    rounded="lg"
                    class="font-weight-bold text-none"
                  >
                    Reset
                  </v-btn>
                </v-col>
              </v-row>
            </div>

            <!-- Modern Ticket Details Table -->
            <v-data-table
              :headers="ticketDetailsHeaders"
              :items="filteredTicketDetails"
              :loading="loadingTicketDetails"
              density="comfortable"
              :items-per-page="ticketDetailsPerPage"
              v-model:page="ticketDetailsPage"
              class="modern-data-table"
            >
              <!-- Ticket Number -->
              <template v-slot:item.ticket_number="{ item }">
                <span class="ticket-number-badge">
                  <v-icon size="14" class="mr-1 text-slate-500">mdi-ticket-outline</v-icon>
                  {{ item.ticket_number }}
                </span>
              </template>

              <!-- Customer Information -->
              <template v-slot:item.customer_info="{ item }">
                <div class="py-1">
                  <div class="font-weight-bold text-slate-900 text-body-2 mb-0">
                    {{ item.customer_name || 'N/A' }}
                  </div>
                  <div class="text-caption text-slate-500 d-flex align-center gap-1 my-0.5">
                    <v-icon size="12" color="slate-400">mdi-phone-outline</v-icon>
                    <span>{{ item.customer_phone || '-' }}</span>
                  </div>
                  <div v-if="item.customer_address" class="text-caption text-slate-400 text-truncate" style="max-width: 220px;" :title="item.customer_address">
                    <v-icon size="12" color="slate-400" class="mr-0.5">mdi-map-marker-outline</v-icon>
                    {{ item.customer_address }}
                  </div>
                </div>
              </template>

              <!-- Technical Details -->
              <template v-slot:item.technical_info="{ item }">
                <div class="d-flex flex-column gap-1 py-1">
                  <div class="d-flex align-center gap-1">
                    <span class="text-caption text-slate-400 font-weight-medium" style="min-width: 30px;">IP:</span>
                    <span class="text-caption font-mono font-weight-semibold text-slate-700">{{ item.ip_pelanggan || '-' }}</span>
                  </div>
                  <div class="d-flex align-center gap-1">
                    <span class="text-caption text-slate-400 font-weight-medium" style="min-width: 30px;">ID:</span>
                    <span class="text-caption font-weight-medium text-slate-700">{{ item.id_pelanggan || '-' }}</span>
                  </div>
                  <div class="d-flex align-center gap-1">
                    <span class="text-caption text-slate-400 font-weight-medium" style="min-width: 30px;">ONU:</span>
                    <v-chip
                      size="x-small"
                      variant="tonal"
                      :color="item.onu_power && item.onu_power !== 'N/A' && parseFloat(item.onu_power) < -27 ? 'error' : (item.onu_power && item.onu_power !== 'N/A' ? 'success' : 'slate')"
                      class="font-weight-medium font-mono"
                    >
                      {{ item.onu_power || '-' }} dBm
                    </v-chip>
                  </div>
                </div>
              </template>

              <!-- Problem Description -->
              <template v-slot:item.problem="{ item }">
                <div class="py-1" style="max-width: 260px;">
                  <div class="font-weight-bold text-slate-800 text-body-2 text-truncate" :title="item.title">{{ item.title }}</div>
                  <div class="text-caption text-slate-500 text-truncate mt-0.5" :title="item.description">{{ item.description || '-' }}</div>
                </div>
              </template>

              <!-- Status & Category -->
              <template v-slot:item.status_category="{ item }">
                <div class="d-flex flex-column gap-1.5 py-1 align-start">
                  <v-chip
                    :color="getStatusColor(item.status)"
                    size="small"
                    variant="tonal"
                    class="font-weight-bold text-uppercase"
                    style="font-size: 0.6875rem; letter-spacing: 0.03em;"
                  >
                    {{ formatStatus(item.status) }}
                  </v-chip>
                  <v-chip
                    color="slate"
                    size="x-small"
                    variant="outlined"
                    class="text-uppercase"
                    style="font-size: 0.625rem; font-weight: 600; letter-spacing: 0.03em;"
                  >
                    {{ formatCategory(item.category) }}
                  </v-chip>
                </div>
              </template>

              <!-- Downtime -->
              <template v-slot:item.downtime="{ item }">
                <div class="py-1">
                  <v-chip
                    :color="getDowntimeColor(item.total_downtime_minutes)"
                    size="small"
                    variant="tonal"
                    prepend-icon="mdi-clock-outline"
                    class="font-weight-bold"
                  >
                    {{ item.downtime_hours }} jam
                  </v-chip>
                  <div class="text-caption text-slate-400 mt-1 font-weight-medium">
                    {{ item.total_downtime_minutes }} menit
                  </div>
                </div>
              </template>

              <!-- Assignment -->
              <template v-slot:item.assignment="{ item }">
                <div class="py-1">
                  <div class="d-flex align-center gap-1.5 text-body-2 font-weight-semibold text-slate-800">
                    <v-icon size="15" color="success">mdi-account-check-outline</v-icon>
                    <span>{{ item.assigned_to || 'Belum ditugaskan' }}</span>
                  </div>
                  <span class="text-caption text-slate-400 d-inline-block mt-0.5 font-weight-medium">
                    {{ item.latest_actions ? item.latest_actions.length : 0 }} riwayat tindakan
                  </span>
                </div>
              </template>
            </v-data-table>
          </div>
          <div v-else class="empty-state py-10 text-center">
            <div class="empty-chart-icon-box mb-2 mx-auto">
              <v-icon size="36" color="slate-400">mdi-ticket-outline</v-icon>
            </div>
            <h4 class="text-subtitle-1 font-weight-bold text-slate-800 mb-1">Tidak Ada Tiket</h4>
            <p class="text-caption text-slate-500 mb-3">Klik tombol di bawah untuk memuat informasi tiket pada periode terpilih.</p>
            <v-btn
              variant="outlined"
              prepend-icon="mdi-refresh"
              @click="loadTicketDetails"
              color="primary"
              size="small"
            >
              Muat Data Tiket
            </v-btn>
          </div>
        </div>
      </div>
    </div>

    <!-- ===== EXPORT DIALOG ===== -->
    <v-dialog v-model="showExportDialog" max-width="520" persistent>
      <div class="export-dialog-card bg-white rounded-xl border elevation-4 overflow-hidden">
        <div class="export-dialog-header pa-4 pa-sm-5 border-b d-flex align-center justify-space-between">
          <div class="d-flex align-center gap-2">
            <div class="chart-header-icon-box">
              <v-icon color="primary" size="20">mdi-file-export-outline</v-icon>
            </div>
            <div>
              <h3 class="text-subtitle-1 font-weight-bold text-slate-900 mb-0">Export Laporan</h3>
              <p class="text-caption text-slate-500 mb-0">Unduh data analitik trouble ticket</p>
            </div>
          </div>
          <v-btn
            icon="mdi-close"
            variant="text"
            density="compact"
            color="slate-400"
            @click="closeExportDialog"
            :disabled="exporting"
          ></v-btn>
        </div>

        <div class="pa-4 pa-sm-5">
          <label class="text-caption font-weight-bold text-slate-700 mb-2 d-block">Filter Brand (Opsional)</label>
          <v-select
            v-model="exportBrandFilter"
            :items="brandOptions"
            placeholder="Semua Brand"
            prepend-inner-icon="mdi-domain"
            variant="outlined"
            density="compact"
            class="mb-4"
            hide-details
            clearable
          ></v-select>

          <label class="text-caption font-weight-bold text-slate-700 mb-2 d-block">Pilih Format File</label>
          <div class="d-flex flex-column gap-2 mb-3">
            <div
              v-for="format in exportFormats"
              :key="format.value"
              @click="!exporting && selectExportFormat(format.value)"
              class="export-option-box pa-3 rounded-lg border d-flex align-center justify-space-between cursor-pointer transition-all"
              :class="{ 'opacity-50 pointer-events-none': exporting }"
            >
              <div class="d-flex align-center gap-3">
                <div class="export-format-badge rounded-md pa-2 d-flex align-center justify-center" :class="`${format.color}-light`">
                  <v-icon size="20" :color="format.color === 'blue' ? 'primary' : format.color === 'green' ? 'success' : format.color === 'red' ? 'error' : 'warning'">
                    {{ format.icon }}
                  </v-icon>
                </div>
                <div>
                  <div class="text-body-2 font-weight-bold text-slate-800">{{ format.title }}</div>
                  <div class="text-caption text-slate-500">{{ format.description }}</div>
                </div>
              </div>
              <v-icon size="18" color="slate-400">mdi-chevron-right</v-icon>
            </div>
          </div>

          <v-alert
            v-if="exporting"
            type="info"
            variant="tonal"
            density="compact"
            class="mt-3"
          >
            <div class="d-flex align-center gap-2">
              <v-progress-circular indeterminate size="16" width="2" color="info"></v-progress-circular>
              <span class="text-caption">Sedang menyiapkan dan mengunduh berkas laporan...</span>
            </div>
          </v-alert>
        </div>

        <div class="pa-4 border-t d-flex justify-end bg-slate-50">
          <v-btn
            variant="text"
            @click="closeExportDialog"
            :disabled="exporting"
            size="small"
            color="slate-600"
          >
            Batal
          </v-btn>
        </div>
      </div>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import apiClient from '@/services/api'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { Line, Pie } from 'vue-chartjs'

// Register Chart.js components
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

// Types
interface Statistics {
  total_tickets: number
  open_tickets: number
  in_progress_tickets: number
  resolved_tickets: number
  closed_tickets: number
  high_priority_tickets: number
  critical_priority_tickets: number
  avg_resolution_time_hours: number | null
  tickets_this_month: number
  unresolved_over_24h: number
}

interface ExcelSheet {
  name: string
  data: string[][]
}

interface CategoryPerformance {
  category: string
  category_display: string
  total_tickets: number
  resolved_tickets: number
  resolution_rate_percent: number
  avg_resolution_hours: number
  avg_downtime_minutes: number
  total_downtime_minutes: number
}

interface UserPerformance {
  user_id: number
  user_name: string
  total_assigned: number
  resolved_tickets: number
  resolution_rate_percent: number
  avg_resolution_hours: number
  first_ticket_date: string | null
  last_ticket_date: string | null
}

interface DowntimeAnalysis {
  overall_statistics: {
    total_downtime_minutes: number
    total_downtime_hours: number
    total_downtime_days: number
    avg_downtime_minutes: number
    avg_downtime_hours: number
    max_downtime_minutes: number
    max_downtime_hours: number
    tickets_with_downtime: number
  }
  by_category: Array<{
    category: string
    category_display: string
    total_downtime_minutes: number
    avg_downtime_minutes: number
    ticket_count: number
    total_downtime_hours: number
    avg_downtime_hours: number
  }>
  top_customers: Array<{
    customer_id: number
    customer_name: string
    total_downtime_minutes: number
    ticket_count: number
    avg_downtime_minutes: number
    total_downtime_hours: number
    avg_downtime_hours: number
  }>
}

interface MonthlyTrend {
  month: string
  month_name: string
  statistics: {
    total: number
    resolved: number
    avg_resolution_hours: number
    by_category: Record<string, number>
    by_priority: Record<string, number>
  }
}

// State
const loading = ref(false)
const loadingMonthly = ref(false)
const loadingCategory = ref(false)
const loadingUser = ref(false)
const loadingDowntime = ref(false)
const loadingTicketDetails = ref(false)
const exporting = ref(false)
const showExportDialog = ref(false)

// Ticket details data for display
const ticketDetails = ref<any[]>([])
const ticketDetailsPage = ref(1)
const ticketDetailsPerPage = ref(10)

// Filter states
const ticketSearch = ref('')
const ticketStatusFilter = ref('')
const ticketCategoryFilter = ref('')

// Filter options
const ticketStatusOptions = [
  { title: 'Open', value: 'open' },
  { title: 'In Progress', value: 'in_progress' },
  { title: 'Pending Customer', value: 'pending_customer' },
  { title: 'Pending Vendor', value: 'pending_vendor' },
  { title: 'Resolved', value: 'resolved' },
  { title: 'Closed', value: 'closed' },
  { title: 'Cancelled', value: 'cancelled' }
]

const ticketCategoryOptions = [
  { title: 'All Categories', value: '' },
  { title: 'No Connection', value: 'no_connection' },
  { title: 'Slow Connection', value: 'slow_connection' },
  { title: 'Intermittent', value: 'intermittent' },
  { title: 'Hardware Issue', value: 'hardware_issue' },
  { title: 'Cable Issue', value: 'cable_issue' },
  { title: 'ONU Issue', value: 'onu_issue' },
  { title: 'OLT Issue', value: 'olt_issue' },
  { title: 'Mikrotik Issue', value: 'mikrotik_issue' },
  { title: 'Other', value: 'other' }
]

const brandOptions = [
  { title: 'All Brands', value: '' },
  { title: 'Jakinet', value: 'JAKINET' },
  { title: 'Jelantik', value: 'JELANTIK' },
  { title: 'Jelantik Nagrak', value: 'JELANTIK NAGRAK' }
]

const exportBrandFilter = ref('')

// Computed filtered tickets
const filteredTicketDetails = computed(() => {
  return ticketDetails.value.filter(ticket => {
    const matchesSearch = !ticketSearch.value ||
      ticket.ticket_number.toLowerCase().includes(ticketSearch.value.toLowerCase()) ||
      ticket.customer_name.toLowerCase().includes(ticketSearch.value.toLowerCase()) ||
      ticket.title.toLowerCase().includes(ticketSearch.value.toLowerCase()) ||
      ticket.description.toLowerCase().includes(ticketSearch.value.toLowerCase())

    const matchesStatus = !ticketStatusFilter.value || ticket.status === ticketStatusFilter.value
    const matchesCategory = !ticketCategoryFilter.value || ticket.category === ticketCategoryFilter.value

    return matchesSearch && matchesStatus && matchesCategory
  })
})

const statistics = ref<Statistics | null>(null)
const monthlyTrends = ref<MonthlyTrend[]>([])
const categoryPerformance = ref<CategoryPerformance[]>([])
const userPerformance = ref<UserPerformance[]>([])
const downtimeAnalysis = ref<DowntimeAnalysis | null>(null)

const trendPeriod = ref('12')
const dateRange = reactive({
  from: '',
  to: ''
})


// Chart data properties
const trendsChartData = ref({
  labels: [] as string[],
  datasets: [
    {
      label: 'Tickets Created',
      data: [] as number[],
      borderColor: 'rgb(59, 130, 246)',
      backgroundColor: 'rgba(59, 130, 246, 0.1)',
      tension: 0.4,
      fill: true
    },
    {
      label: 'Tickets Resolved',
      data: [] as number[],
      borderColor: 'rgb(34, 197, 94)',
      backgroundColor: 'rgba(34, 197, 94, 0.1)',
      tension: 0.4,
      fill: true
    }
  ]
})

const categoryChartData = ref({
  labels: [] as string[],
  datasets: [
    {
      label: 'Tickets by Category',
      data: [] as number[],
      backgroundColor: [
        '#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6',
        '#EC4899', '#14B8A6', '#F97316', '#06B6D4', '#84CC16'
      ],
      borderWidth: 2,
      borderColor: '#ffffff'
    }
  ]
})

// Chart options
const trendsChartOptions = ref({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'top' as const,
      labels: {
        usePointStyle: true,
        padding: 20
      }
    },
    tooltip: {
      mode: 'index' as const,
      intersect: false,
      backgroundColor: 'rgba(0, 0, 0, 0.8)',
      titleColor: '#ffffff',
      bodyColor: '#ffffff',
      borderColor: 'rgba(255, 255, 255, 0.1)',
      borderWidth: 1
    }
  },
  scales: {
    x: {
      display: true,
      grid: {
        display: false
      }
    },
    y: {
      display: true,
      beginAtZero: true,
      ticks: {
        stepSize: 1
      }
    }
  }
})

const categoryChartOptions = ref({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'right' as const,
      labels: {
        usePointStyle: true,
        padding: 15,
        generateLabels: function(chart: any) {
          const data = chart.data
          if (data.labels.length && data.datasets.length) {
            const dataset = data.datasets[0]
            const total = dataset.data.reduce((a: number, b: number) => a + b, 0)
            return data.labels.map((label: string, i: number) => {
              const value = dataset.data[i]
              const percentage = ((value / total) * 100).toFixed(1)
              return {
                text: `${label} (${percentage}%)`,
                fillStyle: dataset.backgroundColor[i],
                hidden: false,
                index: i
              }
            })
          }
          return []
        }
      }
    },
    tooltip: {
      backgroundColor: 'rgba(0, 0, 0, 0.8)',
      titleColor: '#ffffff',
      bodyColor: '#ffffff',
      borderColor: 'rgba(255, 255, 255, 0.1)',
      borderWidth: 1,
      callbacks: {
        label: function(context: any) {
          const total = context.dataset.data.reduce((a: number, b: number) => a + b, 0)
          const percentage = ((context.parsed / total) * 100).toFixed(1)
          return `${context.label}: ${context.parsed} (${percentage}%)`
        }
      }
    }
  }
})

// Table headers
const categoryHeaders = [
  { title: 'Category', key: 'category', sortable: false },
  { title: 'Total', key: 'total_tickets', sortable: false },
  { title: 'Resolved', key: 'resolved_tickets', sortable: false },
  { title: 'Resolution Rate', key: 'resolution_rate', sortable: false },
  { title: 'Avg Resolution', key: 'avg_resolution', sortable: false }
]

const userHeaders = [
  { title: 'Technician', key: 'user_name', sortable: false },
  { title: 'Assigned', key: 'total_assigned', sortable: false },
  { title: 'Resolved', key: 'resolved_tickets', sortable: false },
  { title: 'Resolution Rate', key: 'resolution_rate', sortable: false },
  { title: 'Avg Resolution', key: 'avg_resolution', sortable: false }
]

const customerDowntimeHeaders = [
  { title: 'Customer', key: 'customer_name', sortable: false },
  { title: 'Tickets', key: 'ticket_count', sortable: false },
  { title: 'Total Downtime', key: 'total_downtime', sortable: false },
  { title: 'Avg Downtime', key: 'avg_downtime_hours', sortable: false }
]

// Ticket Details Table Headers
const ticketDetailsHeaders = [
  { title: 'Ticket #', key: 'ticket_number', sortable: false, width: '120px' },
  { title: 'Customer Information', key: 'customer_info', sortable: false, width: '200px' },
  { title: 'Technical Details', key: 'technical_info', sortable: false, width: '150px' },
  { title: 'Problem', key: 'problem', sortable: false, width: '250px' },
  { title: 'Status & Category', key: 'status_category', sortable: false, width: '140px' },
  { title: 'Downtime', key: 'downtime', sortable: false, width: '100px' },
  { title: 'Assignment', key: 'assignment', sortable: false, width: '150px' }
]

// Helper methods
const calculateResolutionRate = () => {
  if (!statistics.value) return 0
  const total = statistics.value.total_tickets
  const resolved = statistics.value.resolved_tickets + statistics.value.closed_tickets
  return total > 0 ? Math.round((resolved / total) * 100) : 0
}

const getResolutionTimeClass = () => {
  const hours = statistics.value?.avg_resolution_time_hours || 0
  if (hours < 24) return 'positive'
  if (hours < 72) return 'warning'
  return 'negative'
}

const getResolutionTimeStatus = () => {
  const hours = statistics.value?.avg_resolution_time_hours || 0
  if (hours < 24) return 'Excellent'
  if (hours < 72) return 'Good'
  return 'Needs Improvement'
}

const formatHours = (hours: number | null) => {
  if (!hours) return '0h'
  if (hours < 1) return `${Math.round(hours * 60)}m`
  if (hours < 24) return `${Math.round(hours)}h`
  return `${Math.round(hours / 24)}d ${Math.round(hours % 24)}h`
}

const getResolutionRateColor = (rate: number) => {
  if (rate >= 90) return 'success'
  if (rate >= 70) return 'warning'
  return 'error'
}

const getDowntimeSeverityColor = (hours: number) => {
  if (hours < 1) return 'success'
  if (hours < 6) return 'warning'
  return 'error'
}

const formatDate = (dateString: string | null | undefined) => {
  if (!dateString) return 'N/A'
  return new Date(dateString).toLocaleDateString('id-ID')
}

const showNotification = (message: string, type: 'success' | 'error' | 'info' | 'warning' = 'info') => {
  const colors = {
    success: '#4caf50',
    error: '#f44336',
    info: '#2196f3',
    warning: '#ff9800'
  }

  const notification = document.createElement('div')
  notification.style.cssText = `
    position: fixed;
    top: 20px;
    right: 20px;
    background: ${colors[type]};
    color: white;
    padding: 16px 24px;
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0,0,0,0.15);
    z-index: 9999;
    font-family: Arial, sans-serif;
    font-size: 14px;
    max-width: 300px;
    animation: slideIn 0.3s ease;
  `
  notification.textContent = message

  document.body.appendChild(notification)

  setTimeout(() => {
    if (notification.parentNode) {
      notification.style.animation = 'slideOut 0.3s ease'
      setTimeout(() => {
        notification.parentNode?.removeChild(notification)
      }, 300)
    }
  }, 5000)
}

// Chart rendering is now handled by vue-chartjs components

// Methods
const loadStatistics = async () => {
  try {
    const response = await apiClient.get('/trouble-tickets/statistics/dashboard')
    statistics.value = response.data
  } catch (error) {
    console.error('Failed to load statistics:', error)
  }
}

const loadMonthlyTrends = async () => {
  loadingMonthly.value = true
  try {
    const params: any = { months: parseInt(trendPeriod.value) }
    if (dateRange.from) params.date_from = dateRange.from
    if (dateRange.to) params.date_to = dateRange.to

    const response = await apiClient.get('/trouble-tickets/reports/monthly-trends', { params })
    monthlyTrends.value = response.data.trends || []

    // Update chart data
    if (monthlyTrends.value.length > 0) {
      const sortedTrends = [...monthlyTrends.value].sort((a, b) =>
        new Date(a.month).getTime() - new Date(b.month).getTime()
      )

      trendsChartData.value.labels = sortedTrends.map(item => {
        const date = new Date(item.month)
        return date.toLocaleDateString('id-ID', { month: 'short', year: 'numeric' })
      })
      trendsChartData.value.datasets[0].data = sortedTrends.map(item => item.statistics.total)
      trendsChartData.value.datasets[1].data = sortedTrends.map(item => item.statistics.resolved)
    }
  } catch (error) {
    console.error('Failed to load monthly trends:', error)
    monthlyTrends.value = []
  } finally {
    loadingMonthly.value = false
  }
}

const loadCategoryPerformance = async () => {
  loadingCategory.value = true
  try {
    const params: any = {}
    if (dateRange.from) params.date_from = dateRange.from
    if (dateRange.to) params.date_to = dateRange.to

    const response = await apiClient.get('/trouble-tickets/reports/category-performance', { params })
    categoryPerformance.value = response.data.category_performance || []

    // Update chart data
    if (categoryPerformance.value.length > 0) {
      categoryChartData.value.labels = categoryPerformance.value.map(item => item.category)
      categoryChartData.value.datasets[0].data = categoryPerformance.value.map(item => item.total_tickets)
    }
  } catch (error: any) {
    console.error('Failed to load category performance:', error)
    showNotification('Failed to load category performance data. Please try again later.', 'error')
    categoryPerformance.value = []
  } finally {
    loadingCategory.value = false
  }
}

const loadUserPerformance = async () => {
  loadingUser.value = true
  try {
    const params: any = {}
    if (dateRange.from) params.date_from = dateRange.from
    if (dateRange.to) params.date_to = dateRange.to

    const response = await apiClient.get('/trouble-tickets/reports/user-performance', { params })
    userPerformance.value = response.data.user_performance || []
  } catch (error: any) {
    console.error('Failed to load user performance:', error)
    showNotification('Failed to load user performance data. Please try again later.', 'error')
    userPerformance.value = []
  } finally {
    loadingUser.value = false
  }
}

const loadDowntimeAnalysis = async () => {
  loadingDowntime.value = true
  try {
    const params: any = {}
    if (dateRange.from) params.date_from = dateRange.from
    if (dateRange.to) params.date_to = dateRange.to

    const response = await apiClient.get('/trouble-tickets/reports/downtime-analysis', { params })
    
    // Safety check: Ensure the response contains the expected structure
    if (response.data && response.data.overall_statistics) {
      downtimeAnalysis.value = response.data
    } else {
      downtimeAnalysis.value = null
    }
  } catch (error: any) {
    console.error('Failed to load downtime analysis:', error)
    showNotification('Failed to load downtime analysis data. Please try again later.', 'error')
    downtimeAnalysis.value = null
  } finally {
    loadingDowntime.value = false
  }
}

const refreshAllData = async () => {
  loading.value = true
  try {
    await Promise.all([
      loadStatistics(),
      loadMonthlyTrends(),
      loadCategoryPerformance(),
      loadUserPerformance(),
      loadDowntimeAnalysis(),
      loadTicketDetails()
    ])
  } finally {
    loading.value = false
  }
}

const loadTicketDetails = async () => {
  loadingTicketDetails.value = true
  try {
    const params: any = {
      skip: (ticketDetailsPage.value - 1) * ticketDetailsPerPage.value,
      limit: ticketDetailsPerPage.value,
      include_relations: true,
      ...(dateRange.from && { date_from: dateRange.from }),
      ...(dateRange.to && { date_to: dateRange.to })
    }

    const response = await apiClient.get('/trouble-tickets', { params })
    const tickets = response.data.data

    const detailedTickets = await Promise.all(
      tickets.map(async (ticket: any) => {
        try {
          const actionsResponse = await apiClient.get(`/trouble-tickets/${ticket.id}/actions`)
          const actions = actionsResponse.data.data || actionsResponse.data || []

          return {
            ticket_number: ticket.ticket_number,
            title: ticket.title,
            description: ticket.description,
            status: ticket.status,
            priority: ticket.priority,
            category: ticket.category,
            customer_name: ticket.pelanggan?.nama || 'N/A',
            customer_address: ticket.pelanggan?.alamat || 'N/A',
            customer_phone: ticket.pelanggan?.no_telp || 'N/A',
            customer_email: ticket.pelanggan?.email || 'N/A',
            id_pelanggan: ticket.data_teknis?.id_pelanggan || 'N/A',
            ip_pelanggan: ticket.data_teknis?.ip_pelanggan || 'N/A',
            customer_brand: ticket.pelanggan?.harga_layanan?.brand || 'N/A',
            onu_power: ticket.data_teknis?.onu_power || 'N/A',
            onu_sn: ticket.data_teknis?.onu_sn || 'N/A',
            distance_olt: ticket.data_teknis?.distance_olt || 'N/A',
            created_at: ticket.created_at,
            resolved_at: ticket.resolved_at,
            total_downtime_minutes: ticket.total_downtime_minutes,
            downtime_hours: ticket.total_downtime_minutes ? (ticket.total_downtime_minutes / 60).toFixed(2) : '0',
            assigned_to: ticket.assigned_user?.name || 'Unassigned',
            resolution_notes: ticket.resolution_notes || '',
            latest_actions: actions.slice(0, 3).map((action: any) => ({
              action_description: action.action_description || '',
              summary_problem: action.summary_problem || '',
              summary_action: action.summary_action || '',
              action_date: action.created_at,
              taken_by: action.taken_user?.name || 'System'
            })),
            id: ticket.id
          }
        } catch (error) {
          console.error(`Failed to fetch details for ticket ${ticket.id}:`, error)
          return {
            ticket_number: ticket.ticket_number,
            title: ticket.title,
            description: ticket.description,
            status: ticket.status,
            priority: ticket.priority,
            category: ticket.category,
            customer_name: 'Error loading details',
            error: 'Failed to load ticket details',
            id: ticket.id
          }
        }
      })
    )

    ticketDetails.value = detailedTickets
  } catch (error) {
    console.error('Failed to load ticket details:', error)
    showNotification('Failed to load ticket details', 'error')
  } finally {
    loadingTicketDetails.value = false
  }
}

const clearTicketFilters = () => {
  ticketSearch.value = ''
  ticketStatusFilter.value = ''
  ticketCategoryFilter.value = ''
}

const formatStatus = (status: string) => {
  return status.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase())
}

const formatPriority = (priority: string) => {
  return priority.charAt(0).toUpperCase() + priority.slice(1)
}

const formatCategory = (category: string) => {
  return category.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase())
}

const getStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    'open': 'info',
    'in_progress': 'warning',
    'pending_customer': 'orange',
    'pending_vendor': 'amber',
    'resolved': 'success',
    'closed': 'blue-grey',
    'cancelled': 'error'
  }
  return colors[status] || 'blue-grey'
}

const getPriorityColor = (priority: string) => {
  const colors: Record<string, string> = {
    'low': 'success',
    'medium': 'warning',
    'high': 'error',
    'critical': 'deep-purple'
  }
  return colors[priority] || 'blue-grey'
}

const getDowntimeColor = (minutes: number | null | undefined) => {
  if (!minutes || minutes <= 0) return 'success'
  if (minutes <= 60) return 'warning'
  if (minutes <= 360) return 'error'
  return 'deep-purple'
}

const applyDateFilter = () => {
  refreshAllData()
}

const clearDateFilter = () => {
  dateRange.from = ''
  dateRange.to = ''
  refreshAllData()
}

// Export formats configuration
const exportFormats = ref([
  {
    value: 'excel',
    title: 'Excel (XLS)',
    description: 'Multi-sheet workbook with proper Excel formatting & complete ticket data',
    icon: 'mdi-file-excel',
    color: 'success'
  },
  {
    value: 'csv',
    title: 'CSV',
    description: 'Complete ticket details with customer info, actions, and history',
    icon: 'mdi-file-delimited',
    color: 'info'
  },
  {
    value: 'pdf',
    title: 'PDF Report',
    description: 'Summary report with key ticket details (20 latest tickets)',
    icon: 'mdi-file-pdf',
    color: 'error'
  },
  {
    value: 'json',
    title: 'JSON Data',
    description: 'Complete raw data including all ticket details and history',
    icon: 'mdi-code-json',
    color: 'warning'
  }
])

const exportReport = () => {
  showExportDialog.value = true
}

const selectExportFormat = async (format: string) => {
  exporting.value = true

  try {
    if (format === 'excel' || format === 'csv') {
      showNotification('Fetching detailed ticket data... This may take a moment.', 'info')
    }

    let ticketDetailsData: any[] = []
    try {
      ticketDetailsData = await fetchTicketDetailsForExport()
    } catch (error) {
      console.error('Failed to fetch ticket details for export:', error)
      showNotification('Warning: Using basic data. Detailed ticket data not available.', 'warning')
    }

    const reportData = {
      metadata: {
        title: exportBrandFilter.value ? `Trouble Ticket Analysis Report - ${exportBrandFilter.value}` : 'Trouble Ticket Analysis Report',
        generated_at: new Date().toISOString(),
        period: {
          from: dateRange.from || 'All time',
          to: dateRange.to || 'Present',
          brand: exportBrandFilter.value || 'All Brands'
        },
        summary: {
          total_tickets: statistics.value?.total_tickets || 0,
          resolution_rate: calculateResolutionRate(),
          avg_resolution_time: formatHours(statistics.value?.avg_resolution_time_hours || null)
        }
      },
      statistics: statistics.value,
      monthly_trends: monthlyTrends.value,
      category_performance: categoryPerformance.value,
      user_performance: userPerformance.value,
      downtime_analysis: downtimeAnalysis.value,
      ticket_details: ticketDetailsData
    }

    switch (format) {
      case 'excel':
        await exportToExcel(reportData)
        break
      case 'csv':
        await exportToCSV(reportData)
        break
      case 'pdf':
        await exportToPDF(reportData)
        break
      case 'json':
        await exportToJSON(reportData)
        break
    }

    closeExportDialog()

  } catch (error) {
    console.error('Export failed:', error)
    showNotification('Export failed. Please try again.', 'error')
  } finally {
    exporting.value = false
  }
}

const closeExportDialog = () => {
  showExportDialog.value = false
}

const fetchTicketDetailsForExport = async () => {
  const detailedTickets: any[] = []

  try {
    const params: any = {
      skip: 0,
      limit: 100,
      include_relations: true,
      ...(dateRange.from && { date_from: dateRange.from }),
      ...(dateRange.to && { date_to: dateRange.to }),
      ...(exportBrandFilter.value && { brand: exportBrandFilter.value })
    }

    const response = await apiClient.get('/trouble-tickets', { params })
    const tickets = response.data.data

    const batchSize = 10

    for (let i = 0; i < tickets.length; i += batchSize) {
      const batch = tickets.slice(i, i + batchSize)

      const batchResults = await Promise.allSettled(
        batch.map(async (ticket: any) => {
          try {
            const actionsResponse = await apiClient.get(`/trouble-tickets/${ticket.id}/actions`)
            const actions = actionsResponse.data.data || actionsResponse.data || []

            const historyResponse = await apiClient.get(`/trouble-tickets/${ticket.id}/history`)
            const history = historyResponse.data.data || historyResponse.data || []

            return {
              ticket_number: ticket.ticket_number,
              title: ticket.title,
              description: ticket.description,
              status: ticket.status,
              priority: ticket.priority,
              category: ticket.category,
              customer_id: ticket.pelanggan?.id || '',
              customer_name: ticket.pelanggan?.nama || 'N/A',
              customer_address: ticket.pelanggan?.alamat || 'N/A',
              customer_phone: ticket.pelanggan?.no_telp || 'N/A',
              customer_email: ticket.pelanggan?.email || 'N/A',
              id_pelanggan: ticket.data_teknis?.id_pelanggan || 'N/A',
              ip_pelanggan: ticket.data_teknis?.ip_pelanggan || 'N/A',
              customer_brand: ticket.pelanggan?.harga_layanan?.brand || 'N/A',
              onu_power: ticket.data_teknis?.onu_power || 'N/A',
              onu_sn: ticket.data_teknis?.onu_sn || 'N/A',
              distance_olt: ticket.data_teknis?.distance_olt || 'N/A',
              created_at: ticket.created_at,
              updated_at: ticket.updated_at,
              resolved_at: ticket.resolved_at,
              downtime_start: ticket.downtime_start,
              downtime_end: ticket.downtime_end,
              total_downtime_minutes: ticket.total_downtime_minutes,
              total_pending_minutes: ticket.total_pending_minutes,
              pending_start: ticket.pending_start,
              downtime_hours: ticket.total_downtime_minutes ? (ticket.total_downtime_minutes / 60).toFixed(2) : '0',
              assigned_to: ticket.assigned_user?.name || 'Unassigned',
              assigned_user_id: ticket.assigned_to || '',
              resolution_notes: ticket.resolution_notes || '',
              latest_actions: actions.slice(0, 3).map((action: any) => ({
                action_description: action.action_description || '',
                summary_problem: action.summary_problem || '',
                summary_action: action.summary_action || '',
                action_date: action.created_at,
                taken_by: action.taken_user?.name || 'System'
              })),
              status_history: history.slice(0, 5).map((h: any) => ({
                old_status: h.old_status || 'Created',
                new_status: h.new_status,
                changed_date: h.created_at,
                changed_by: h.changed_user?.name || 'System',
                notes: h.notes || ''
              })),
              evidence: ticket.evidence || '',
              customer_notified: ticket.customer_notified ? 'Yes' : 'No',
              last_customer_contact: ticket.last_customer_contact || 'N/A'
            }
          } catch (error) {
            console.error(`Failed to fetch details for ticket ${ticket.id}:`, error)
            return {
              ticket_number: ticket.ticket_number,
              title: ticket.title,
              description: ticket.description,
              status: ticket.status,
              priority: ticket.priority,
              category: ticket.category,
              customer_name: 'Error loading details',
              error: 'Failed to load ticket details'
            }
          }
        })
      )

      batchResults.forEach((result) => {
        if (result.status === 'fulfilled') {
          detailedTickets.push(result.value)
        }
      })

      if (i + batchSize < tickets.length) {
        await new Promise(resolve => setTimeout(resolve, 100))
      }
    }

    return detailedTickets
  } catch (error) {
    console.error('Failed to fetch ticket details for export:', error)
    showNotification('Failed to load detailed ticket data', 'error')
    return []
  }
}

const exportToExcel = async (data: any) => {
  const workbook = createExcelWorkbook(data)
  const timestamp = new Date().toISOString().split('T')[0]
  const brandSuffix = exportBrandFilter.value ? `-${exportBrandFilter.value.toLowerCase().replace(/\s+/g, '-')}` : ''
  const filename = `trouble-ticket-report${brandSuffix}-${timestamp}.xlsx`
  downloadExcelFile(workbook, filename)
  showNotification('Excel report exported successfully!', 'success')
}

const exportToCSV = async (data: any) => {
  let csvContent = generateCSVContent(data)
  const timestamp = new Date().toISOString().split('T')[0]
  const brandSuffix = exportBrandFilter.value ? `-${exportBrandFilter.value.toLowerCase().replace(/\s+/g, '-')}` : ''
  const filename = `trouble-ticket-report${brandSuffix}-${timestamp}.csv`
  downloadCSVFile(csvContent, filename)
  showNotification('CSV report exported successfully!', 'success')
}

const exportToPDF = async (data: any) => {
  const pdfContent = generatePDFContent(data)
  const timestamp = new Date().toISOString().split('T')[0]
  const brandSuffix = exportBrandFilter.value ? `-${exportBrandFilter.value.toLowerCase().replace(/\s+/g, '-')}` : ''
  const filename = `trouble-ticket-report${brandSuffix}-${timestamp}.pdf`
  downloadPDFFile(pdfContent, filename)
  showNotification('PDF report exported successfully!', 'success')
}

const exportToJSON = async (data: any) => {
  const timestamp = new Date().toISOString().split('T')[0]
  const brandSuffix = exportBrandFilter.value ? `-${exportBrandFilter.value.toLowerCase().replace(/\s+/g, '-')}` : ''
  const filename = `trouble-ticket-report${brandSuffix}-${timestamp}.json`
  downloadJSONFile(data, filename)
  showNotification('JSON report exported successfully!', 'success')
}

// Excel Helper Functions
const formatReportDateTime = (dateStr: string | null | undefined) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  if (isNaN(date.getTime())) return dateStr
  const pad = (n: number) => n.toString().padStart(2, '0')
  const day = pad(date.getDate())
  const month = pad(date.getMonth() + 1)
  const year = date.getFullYear()
  const hours = pad(date.getHours())
  const minutes = pad(date.getMinutes())
  const seconds = pad(date.getSeconds())
  return `${day}/${month}/${year} ${hours}:${minutes}:${seconds}`
}

const formatDetailedDuration = (ms: number) => {
  if (isNaN(ms) || ms <= 0) return '0 jam, 0 menit, 0 detik'
  const totalSeconds = Math.floor(ms / 1000)
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60
  return `${hours} jam, ${minutes} menit, ${seconds} detik`
}

const getDowntimeDuration = (t: any) => {
  if (!t.downtime_start) {
    const minutes = t.total_downtime_minutes || 0
    const hours = Math.floor(minutes / 60)
    const mins = minutes % 60
    return `${hours} jam, ${mins} menit, 0 detik`
  }
  const start = new Date(t.downtime_start)
  const end = t.downtime_end ? new Date(t.downtime_end) : new Date()
  let durationMs = end.getTime() - start.getTime()
  let pendingMs = (t.total_pending_minutes || 0) * 60 * 1000
  if (t.pending_start && !t.downtime_end) {
    const pStart = new Date(t.pending_start)
    pendingMs += new Date().getTime() - pStart.getTime()
  }
  let netMs = durationMs - pendingMs
  if (netMs < 0) netMs = 0
  return formatDetailedDuration(netMs)
}

const getPendingDuration = (t: any) => {
  let pendingMs = (t.total_pending_minutes || 0) * 60 * 1000
  if (t.pending_start && !t.downtime_end) {
    const pStart = new Date(t.pending_start)
    pendingMs += new Date().getTime() - pStart.getTime()
  }
  return formatDetailedDuration(pendingMs)
}

// Excel Helper Functions
const createExcelWorkbook = (data: any): ExcelSheet[] => {
  const sheets: ExcelSheet[] = []

  sheets.push({
    name: 'Summary',
    data: [
      ['Trouble Ticket Analysis Report', '', '', ''],
      ['Generated:', new Date().toLocaleString(), '', ''],
      ['Period:', `${data.metadata.period.from} - ${data.metadata.period.to}`, '', ''],
      ['Brand:', data.metadata.period.brand || 'All Brands', '', ''],
      ['', '', '', ''],
      ['Key Metrics', '', '', ''],
      ['Total Tickets:', data.statistics?.total_tickets || 0, '', ''],
      ['Resolution Rate:', `${calculateResolutionRate()}%`, '', ''],
      ['Avg Resolution Time:', formatHours(data.statistics?.avg_resolution_time_hours || null), '', ''],
      ['Open Tickets:', data.statistics?.open_tickets || 0, '', ''],
      ['Resolved Tickets:', data.statistics?.resolved_tickets || 0, '', ''],
    ]
  })

  if (data.category_performance?.length > 0) {
    const categorySheet: string[][] = [
      ['Category', 'Total Tickets', 'Resolved', 'Resolution Rate (%)', 'Avg Resolution Time']
    ]

    data.category_performance.forEach((cat: any) => {
      categorySheet.push([
        cat.category_display,
        cat.total_tickets,
        cat.resolved_tickets,
        cat.resolution_rate_percent,
        formatHours(cat.avg_resolution_hours)
      ])
    })

    sheets.push({ name: 'Category Performance', data: categorySheet })
  }

  if (data.user_performance?.length > 0) {
    const userSheet: string[][] = [
      ['Technician', 'Assigned Tickets', 'Resolved', 'Resolution Rate (%)', 'Avg Resolution Time']
    ]

    data.user_performance.forEach((user: any) => {
      userSheet.push([
        user.user_name,
        user.total_assigned,
        user.resolved_tickets,
        user.resolution_rate_percent,
        formatHours(user.avg_resolution_hours)
      ])
    })

    sheets.push({ name: 'Technician Performance', data: userSheet })
  }

  if (data.monthly_trends?.length > 0) {
    const trendsSheet: string[][] = [
      ['Month', 'Total Tickets', 'Resolved', 'Avg Resolution Time (hours)']
    ]

    data.monthly_trends.forEach((trend: any) => {
      trendsSheet.push([
        trend.month_name,
        trend.statistics.total,
        trend.statistics.resolved,
        trend.statistics.avg_resolution_hours
      ])
    })

    sheets.push({ name: 'Monthly Trends', data: trendsSheet })
  }

  if (data.downtime_analysis) {
    const downtimeSheet: string[][] = [
      ['Downtime Analysis', '', '', ''],
      ['Total Downtime:', formatHours(data.downtime_analysis.overall_statistics.total_downtime_hours), '', ''],
      ['Average Downtime:', formatHours(data.downtime_analysis.overall_statistics.avg_downtime_hours), '', ''],
      ['Maximum Downtime:', formatHours(data.downtime_analysis.overall_statistics.max_downtime_hours), '', ''],
      ['Affected Tickets:', data.downtime_analysis.overall_statistics.tickets_with_downtime, '', ''],
      ['', '', '', ''],
      ['Top Customers by Downtime', '', '', ''],
      ['Customer', 'Ticket Count', 'Total Downtime', 'Avg Downtime']
    ]

    data.downtime_analysis.top_customers.slice(0, 10).forEach((customer: any) => {
      downtimeSheet.push([
        customer.customer_name,
        customer.ticket_count,
        formatHours(customer.total_downtime_hours),
        formatHours(customer.avg_downtime_hours)
      ])
    })

    sheets.push({ name: 'Downtime Analysis', data: downtimeSheet })
  }

  if (data.ticket_details?.length > 0) {
    const ticketDetailsSheet: string[][] = []

    ticketDetailsSheet.push(['Ticket Details with Customer Information'])
    ticketDetailsSheet.push([''])

    ticketDetailsSheet.push([
      'Ticket #',
      'Brand',
      'Customer Name',
      'Customer Phone',
      'Customer Address',
      'IP Pelanggan',
      'ID Pelanggan',
      'ONU Power (dBm)',
      'Problem Description',
      'Category',
      'Status',
      'Start Clock',
      'Pending Clock',
      'Total Downtime',
      'Assigned To',
      'Total Actions'
    ])

    data.ticket_details.forEach((ticket: any) => {
      const rowData = [
        ticket.ticket_number || '',
        ticket.customer_brand || 'N/A',
        ticket.customer_name || '',
        ticket.customer_phone || '',
        ticket.customer_address || '',
        ticket.ip_pelanggan || '',
        ticket.id_pelanggan || '',
        ticket.onu_power || '',
        ticket.description || '',
        ticket.category || '',
        ticket.status || '',
        formatReportDateTime(ticket.downtime_start || ticket.created_at),
        getPendingDuration(ticket),
        getDowntimeDuration(ticket),
        ticket.assigned_to || '',
        (ticket.latest_actions ? ticket.latest_actions.length : 0).toString()
      ]

      while (rowData.length < 16) {
        rowData.push('')
      }

      ticketDetailsSheet.push(rowData)
    })

    sheets.push({ name: 'Ticket Details', data: ticketDetailsSheet })

    const detailedProblemsSheet: string[][] = []

    detailedProblemsSheet.push(['Detailed Problem Information & Actions'])
    detailedProblemsSheet.push([''])

    detailedProblemsSheet.push([
      'Ticket #',
      'Problem Title',
      'Problem Description',
      'Latest Action Description',
      'Problem Summary',
      'Action Summary',
      'Action Date',
      'Taken By',
      'Resolution Notes',
      'Status Changes'
    ])

    data.ticket_details.forEach((ticket: any) => {
      const latestAction = ticket.latest_actions[0]
      const statusChanges = ticket.status_history.length > 0
        ? ticket.status_history.map((h: any) => `${h.old_status} → ${h.new_status}`).join(', ')
        : 'No status changes'

      const rowData = [
        ticket.ticket_number || '',
        ticket.title || '',
        ticket.description || '',
        latestAction?.action_description || 'No actions yet',
        latestAction?.summary_problem || 'No problem summary',
        latestAction?.summary_action || 'No action summary',
        latestAction ? formatDate(latestAction.action_date) : 'N/A',
        latestAction?.taken_by || 'N/A',
        ticket.resolution_notes || 'No resolution notes',
        statusChanges
      ]

      while (rowData.length < 10) {
        rowData.push('')
      }

      detailedProblemsSheet.push(rowData)
    })

    sheets.push({ name: 'Problem Details', data: detailedProblemsSheet })
  }

  return sheets
}

const downloadExcelFile = (sheets: any[], filename: string) => {
  createTrueExcelFile(sheets, filename)
}

const createTrueExcelFile = (sheets: any[], filename: string) => {
  try {
    const excelXML = createExcelXML(sheets)

    const blob = new Blob([excelXML], {
      type: 'application/vnd.ms-excel'
    })

    const link = document.createElement('a')
    if (link.download !== undefined) {
      const url = URL.createObjectURL(blob)
      link.setAttribute('href', url)
      const xlsFilename = filename.replace('.xlsx', '.xls')
      link.setAttribute('download', xlsFilename)
      link.style.visibility = 'hidden'
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)

      showNotification(`Excel file downloaded successfully as "${xlsFilename}"!`, 'success')
    }
  } catch (error) {
    console.error('Error creating Excel file:', error)
    showNotification('Failed to create Excel file. Downloading CSV format instead.', 'error')
    createExcelCSVWorkbook(sheets, filename)
  }
}

const createExcelXML = (sheets: any[]) => {
  let xmlContent = `<?xml version="1.0"?>
<?mso-application progid="Excel.Sheet"?>
<Workbook xmlns="urn:schemas-microsoft-com:office:spreadsheet"
 xmlns:o="urn:schemas-microsoft-com:office:office"
 xmlns:x="urn:schemas-microsoft-com:office:excel"
 xmlns:ss="urn:schemas-microsoft-com:office:spreadsheet"
 xmlns:html="http://www.w3.org/TR/REC-html40">
  <DocumentProperties xmlns="urn:schemas-microsoft-com:office:office">
    <Title>Trouble Ticket Analysis Report</Title>
    <Subject>Trouble Ticket Analysis</Subject>
    <Author>Artacom FTTH Billing System</Author>
    <Created>${new Date().toISOString()}</Created>
    <Company>Artacom</Company>
  </DocumentProperties>
  <ExcelWorkbook xmlns="urn:schemas-microsoft-com:office:excel">
    <WindowHeight>9000</WindowHeight>
    <WindowWidth>13860</WindowWidth>
    <WindowTopX>240</WindowTopX>
    <WindowTopY>75</WindowTopY>
    <ProtectStructure>False</ProtectStructure>
    <ProtectWindows>False</ProtectWindows>
  </ExcelWorkbook>
  <Styles>
    <Style ss:ID="Default" ss:Name="Normal">
      <Alignment ss:Vertical="Bottom"/>
      <Borders/>
      <Font ss:FontName="Arial" ss:Size="11" ss:Color="#000000"/>
      <Interior/>
      <NumberFormat/>
      <Protection/>
    </Style>
    <Style ss:ID="Header">
      <Font ss:FontName="Arial" ss:Size="12" ss:Bold="1" ss:Color="#FFFFFF"/>
      <Interior ss:Color="#4CAF50" ss:Pattern="Solid"/>
      <Borders>
        <Border ss:Position="Bottom" ss:LineStyle="Continuous" ss:Weight="1"/>
        <Border ss:Position="Left" ss:LineStyle="Continuous" ss:Weight="1"/>
        <Border ss:Position="Right" ss:LineStyle="Continuous" ss:Weight="1"/>
        <Border ss:Position="Top" ss:LineStyle="Continuous" ss:Weight="1"/>
      </Borders>
      <Alignment ss:Horizontal="Center" ss:Vertical="Center"/>
    </Style>
    <Style ss:ID="TicketNumber">
      <Font ss:FontName="Arial" ss:Size="11" ss:Bold="1"/>
      <Interior ss:Color="#E8F5E8" ss:Pattern="Solid"/>
      <Borders>
        <Border ss:Position="Bottom" ss:LineStyle="Continuous" ss:Weight="1"/>
        <Border ss:Position="Left" ss:LineStyle="Continuous" ss:Weight="1"/>
        <Border ss:Position="Right" ss:LineStyle="Continuous" ss:Weight="1"/>
        <Border ss:Position="Top" ss:LineStyle="Continuous" ss:Weight="1"/>
      </Borders>
      <Alignment ss:Horizontal="Center" ss:Vertical="Center"/>
    </Style>
    <Style ss:ID="CustomerName">
      <Font ss:FontName="Arial" ss:Size="11" ss:Bold="1"/>
      <Borders>
        <Border ss:Position="Bottom" ss:LineStyle="Continuous" ss:Weight="1"/>
        <Border ss:Position="Left" ss:LineStyle="Continuous" ss:Weight="1"/>
        <Border ss:Position="Right" ss:LineStyle="Continuous" ss:Weight="1"/>
        <Border ss:Position="Top" ss:LineStyle="Continuous" ss:Weight="1"/>
      </Borders>
    </Style>
    <Style ss:ID="Number">
      <NumberFormat ss:Format="#,##0.00"/>
      <Font ss:FontName="Arial" ss:Size="11"/>
      <Borders>
        <Border ss:Position="Bottom" ss:LineStyle="Continuous" ss:Weight="1"/>
        <Border ss:Position="Left" ss:LineStyle="Continuous" ss:Weight="1"/>
        <Border ss:Position="Right" ss:LineStyle="Continuous" ss:Weight="1"/>
        <Border ss:Position="Top" ss:LineStyle="Continuous" ss:Weight="1"/>
      </Borders>
      <Alignment ss:Horizontal="Right" ss:Vertical="Center"/>
    </Style>
    <Style ss:ID="Centered">
      <Alignment ss:Horizontal="Center" ss:Vertical="Center"/>
      <Font ss:FontName="Arial" ss:Size="11" ss:Bold="1"/>
      <Borders>
        <Border ss:Position="Bottom" ss:LineStyle="Continuous" ss:Weight="1"/>
        <Border ss:Position="Left" ss:LineStyle="Continuous" ss:Weight="1"/>
        <Border ss:Position="Right" ss:LineStyle="Continuous" ss:Weight="1"/>
        <Border ss:Position="Top" ss:LineStyle="Continuous" ss:Weight="1"/>
      </Borders>
    </Style>
    <Style ss:ID="DefaultCell">
      <Font ss:FontName="Arial" ss:Size="11"/>
      <Borders>
        <Border ss:Position="Bottom" ss:LineStyle="Continuous" ss:Weight="1"/>
        <Border ss:Position="Left" ss:LineStyle="Continuous" ss:Weight="1"/>
        <Border ss:Position="Right" ss:LineStyle="Continuous" ss:Weight="1"/>
        <Border ss:Position="Top" ss:LineStyle="Continuous" ss:Weight="1"/>
      </Borders>
    </Style>
  </Styles>
`

  sheets.forEach(sheet => {
    xmlContent += createWorksheetXML(sheet)
  })

  xmlContent += '</Workbook>'
  return xmlContent
}

const createWorksheetXML = (sheet: ExcelSheet) => {
  let worksheetXML = `<Worksheet ss:Name="${sheet.name.replace(/[^\w\s-]/g, '_')}">`
  worksheetXML += '<Table>'

  let rowIndex = 0
  sheet.data.forEach((row: any[]) => {
    if (row.length === 0) {
      return
    }

    const isHeaderRow = row.some(cell => {
      const cellStr = String(cell).toLowerCase()
      return cellStr.includes('ticket') && cellStr.includes('#') ||
             cellStr.includes('customer name') ||
             cellStr.includes('problem title') ||
             cellStr.includes('metric') ||
             cellStr.includes('category')
    })

    if (row.length === 1 && (row[0].includes('Details') || row[0].includes('Analysis') || row[0].includes('Summary'))) {
      return
    }

    worksheetXML += '<Row>'
    row.forEach((cell, colIndex) => {
      let cellStyle = 'Default'
      let cellValue = cell === null || cell === undefined ? '' : String(cell)

      if (isHeaderRow) {
        cellStyle = 'Header'
      } else if (sheet.name === 'Ticket Details' || sheet.name === 'Ticket_Details') {
        if (colIndex === 0) cellStyle = 'TicketNumber'
        else if (colIndex === 2) cellStyle = 'CustomerName'
        else if (colIndex === 9) cellStyle = 'Centered'
        else if (colIndex === 10) cellStyle = 'Centered'
        else if (colIndex === 15) cellStyle = 'Centered'
        else cellStyle = 'DefaultCell'
      } else if (sheet.name === 'Problem_Details') {
        if (colIndex === 0) cellStyle = 'TicketNumber'
        else if (colIndex === 6) cellStyle = 'DefaultCell'
        else cellStyle = 'DefaultCell'
      } else {
        cellStyle = 'DefaultCell'
      }

      const isNumeric = !isHeaderRow && !isNaN(parseFloat(String(cellValue))) && String(cellValue) !== '' && !isNaN(Number(cellValue))

      let processedCellValue: string | number = cellValue
      if (isNumeric) {
        processedCellValue = parseFloat(String(cellValue))
      }

      const cellType = isNumeric ? ('Number' as const) : ('String' as const)

      worksheetXML += `<Cell ss:StyleID="${cellStyle}"><Data ss:Type="${cellType}">${escapeXML(processedCellValue)}</Data></Cell>`
    })
    worksheetXML += '</Row>'
    rowIndex++
  })

  worksheetXML += '</Table></Worksheet>'
  return worksheetXML
}

const escapeXML = (text: any): string => {
  return String(text)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

const createExcelCSVWorkbook = (sheets: ExcelSheet[], filename: string) => {
  let csvContent = ''

  sheets.forEach((sheet, sheetIndex) => {
    if (sheetIndex > 0) {
      csvContent += '\n\n'
    }

    csvContent += `# Sheet: ${sheet.name}\n`

    sheet.data.forEach((row: string[], rowIndex: number) => {
      if (row.length === 0 || (row.length === 1 && row[0].includes('Details'))) {
        return
      }

      if (rowIndex === 0 || (rowIndex === 1 && !row[0].includes('#'))) {
        return
      }

      if (rowIndex === 2 || (row.some(cell => String(cell).includes('#')))) {
        const headerRow = row.map(cell => `"${String(cell)}"`).join(',')
        csvContent += headerRow + '\n'
      } else {
        const dataRow = row.map(cell => {
          if (cell === null || cell === undefined) return '""'
          return `"${String(cell).replace(/"/g, '""')}"`
        }).join(',')
        csvContent += dataRow + '\n'
      }
    })
  })

  const blob = new Blob([csvContent], {
    type: 'text/csv;charset=utf-8;'
  })

  const link = document.createElement('a')
  if (link.download !== undefined) {
    const url = URL.createObjectURL(blob)
    link.setAttribute('href', url)

    const csvFilename = filename.replace('.xlsx', '.csv')
    link.setAttribute('download', csvFilename)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)

    showNotification(`Excel-compatible CSV file downloaded as "${csvFilename}". Open with Excel for proper formatting.`, 'success')
  }
}

// CSV Helper Functions
const generateCSVContent = (data: any): string => {
  let csvContent = `Trouble Ticket Analysis Report\n`
  csvContent += `Generated: ${new Date().toLocaleString()}\n`
  csvContent += `Period: ${data.metadata.period.from} - ${data.metadata.period.to}\n\n`

  csvContent += `SUMMARY\n`
  csvContent += `Metric,Value\n`
  csvContent += `Total Tickets,${data.statistics?.total_tickets || 0}\n`
  csvContent += `Resolution Rate,${calculateResolutionRate()}%\n`
  csvContent += `Avg Resolution Time,${formatHours(data.statistics?.avg_resolution_time_hours || null)}\n`
  csvContent += `Open Tickets,${data.statistics?.open_tickets || 0}\n`
  csvContent += `Resolved Tickets,${data.statistics?.resolved_tickets || 0}\n`
  csvContent += `Brand,${data.metadata.period.brand || 'All Brands'}\n\n`

  if (data.category_performance?.length > 0) {
    csvContent += `CATEGORY PERFORMANCE\n`
    csvContent += `Category,Total Tickets,Resolved,Resolution Rate (%),Avg Resolution Time\n`

    data.category_performance.forEach((cat: any) => {
      csvContent += `"${cat.category_display}",${cat.total_tickets},${cat.resolved_tickets},${cat.resolution_rate_percent},"${formatHours(cat.avg_resolution_hours)}"\n`
    })
    csvContent += `\n`
  }

  if (data.user_performance?.length > 0) {
    csvContent += `TECHNICIAN PERFORMANCE\n`
    csvContent += `Technician,Assigned Tickets,Resolved,Resolution Rate (%),Avg Resolution Time\n`

    data.user_performance.forEach((user: any) => {
      csvContent += `"${user.user_name}",${user.total_assigned},${user.resolved_tickets},${user.resolution_rate_percent},"${formatHours(user.avg_resolution_hours)}"\n`
    })
    csvContent += `\n`
  }

  if (data.monthly_trends?.length > 0) {
    csvContent += `MONTHLY TRENDS\n`
    csvContent += `Month,Total Tickets,Resolved,Avg Resolution Time (hours)\n`

    data.monthly_trends.forEach((trend: any) => {
      csvContent += `"${trend.month_name}",${trend.statistics.total},${trend.statistics.resolved},${trend.statistics.avg_resolution_hours}\n`
    })
    csvContent += `\n`
  }

  if (data.downtime_analysis) {
    csvContent += `DOWNTIME ANALYSIS\n`
    csvContent += `Metric,Value\n`
    csvContent += `Total Downtime,${formatHours(data.downtime_analysis.overall_statistics.total_downtime_hours)}\n`
    csvContent += `Average Downtime,${formatHours(data.downtime_analysis.overall_statistics.avg_downtime_hours)}\n`
    csvContent += `Maximum Downtime,${formatHours(data.downtime_analysis.overall_statistics.max_downtime_hours)}\n`
    csvContent += `Affected Tickets,${data.downtime_analysis.overall_statistics.tickets_with_downtime}\n\n`

    if (data.downtime_analysis.top_customers?.length > 0) {
      csvContent += `TOP CUSTOMERS BY DOWNTIME\n`
      csvContent += `Customer,Ticket Count,Total Downtime,Avg Downtime\n`

      data.downtime_analysis.top_customers.slice(0, 10).forEach((customer: any) => {
        csvContent += `"${customer.customer_name}",${customer.ticket_count},"${formatHours(customer.total_downtime_hours)}","${formatHours(customer.avg_downtime_hours)}"\n`
      })
    }
  }

  if (data.ticket_details?.length > 0) {
    csvContent += `\nTROUBLE TICKET DETAILS\n`
    csvContent += `Ticket Number,Customer Name,Customer Address,Phone,Email,Brand,ID Pelanggan,IP Pelanggan,ONU Power,Status,Category,Start Clock,Pending Clock,Total Downtime,Assigned To,Problem Description,Resolution Notes,Total Actions\n`

    data.ticket_details.forEach((ticket: any) => {
      csvContent += `"${ticket.ticket_number}","${ticket.customer_name}","${ticket.customer_address}","${ticket.customer_phone}","${ticket.customer_email}","${ticket.customer_brand}","${ticket.id_pelanggan}","${ticket.ip_pelanggan}","${ticket.onu_power}","${ticket.status}","${ticket.category}","${formatReportDateTime(ticket.downtime_start || ticket.created_at)}","${getPendingDuration(ticket)}","${getDowntimeDuration(ticket)}","${ticket.assigned_to}","${ticket.description.substring(0, 200).replace(/"/g, '""')}","${ticket.resolution_notes.replace(/"/g, '""')}","${ticket.latest_actions.length}"\n`
    })

    csvContent += `\nTICKET ACTION DETAILS\n`
    csvContent += `Ticket Number,Action Date,Action Description,Problem Summary,Action Summary,Taken By\n`

    data.ticket_details.forEach((ticket: any) => {
      if (ticket.latest_actions.length > 0) {
        ticket.latest_actions.forEach((action: any) => {
          csvContent += `"${ticket.ticket_number}","${formatDate(action.action_date)}","${action.action_description.replace(/"/g, '""')}","${action.summary_problem.replace(/"/g, '""')}","${action.summary_action.replace(/"/g, '""')}","${action.taken_by}"\n`
        })
      } else {
        csvContent += `"${ticket.ticket_number}","No actions recorded","-","-","-","-"\n`
      }
    })

    csvContent += `\nTICKET STATUS HISTORY\n`
    csvContent += `Ticket Number,Changed Date,Old Status,New Status,Changed By,Notes\n`

    data.ticket_details.forEach((ticket: any) => {
      if (ticket.status_history.length > 0) {
        ticket.status_history.forEach((history: any) => {
          csvContent += `"${ticket.ticket_number}","${formatDate(history.changed_date)}","${history.old_status}","${history.new_status}","${history.changed_by}","${history.notes.replace(/"/g, '""')}"\n`
        })
      } else {
        csvContent += `"${ticket.ticket_number}","No status changes recorded","-","-","-","-"\n`
      }
    })
  }

  return csvContent
}

const downloadCSVFile = (content: string, filename: string) => {
  const blob = new Blob([content], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')

  if (link.download !== undefined) {
    const url = URL.createObjectURL(blob)
    link.setAttribute('href', url)
    link.setAttribute('download', filename)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }
}

// PDF Helper Functions
const generatePDFContent = (data: any): string => {
  let htmlContent = `
    <!DOCTYPE html>
    <html>
    <head>
      <title>Trouble Ticket Report</title>
      <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { text-align: center; margin-bottom: 30px; }
        .section { margin-bottom: 25px; }
        .section h2 { color: #1976d2; border-bottom: 2px solid #1976d2; padding-bottom: 5px; }
        table { width: 100%; border-collapse: collapse; margin-bottom: 15px; font-size: 11px; }
        th, td { border: 1px solid #ddd; padding: 6px 8px; text-align: left; }
        th { background-color: #f5f5f5; font-weight: bold; }
        .summary-metric { display: inline-block; margin: 10px; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
        .metric-value { font-size: 24px; font-weight: bold; color: #1976d2; }
        .metric-label { font-size: 12px; color: #666; }
      </style>
    </head>
    <body>
      <div class="header">
        <h1>Trouble Ticket Analysis Report</h1>
        <p>Generated: ${new Date().toLocaleString()}</p>
        <p>Period: ${data.metadata.period.from} - ${data.metadata.period.to}</p>
        <p>Brand: <strong>${data.metadata.period.brand || 'All Brands'}</strong></p>
      </div>

      <div class="section">
        <h2>Summary</h2>
        <div class="summary-metric">
          <div class="metric-value">${data.statistics?.total_tickets || 0}</div>
          <div class="metric-label">Total Tickets</div>
        </div>
        <div class="summary-metric">
          <div class="metric-value">${calculateResolutionRate()}%</div>
          <div class="metric-label">Resolution Rate</div>
        </div>
        <div class="summary-metric">
          <div class="metric-value">${formatHours(data.statistics?.avg_resolution_time_hours || null)}</div>
          <div class="metric-label">Avg Resolution Time</div>
        </div>
      </div>
  `

  if (data.category_performance?.length > 0) {
    htmlContent += `
      <div class="section">
        <h2>Category Performance</h2>
        <table>
          <thead>
            <tr>
              <th>Category</th>
              <th>Total</th>
              <th>Resolved</th>
              <th>Resolution Rate</th>
              <th>Avg Resolution Time</th>
            </tr>
          </thead>
          <tbody>
    `

    data.category_performance.forEach((cat: any) => {
      htmlContent += `
        <tr>
          <td>${cat.category_display}</td>
          <td>${cat.total_tickets}</td>
          <td>${cat.resolved_tickets}</td>
          <td>${cat.resolution_rate_percent}%</td>
          <td>${formatHours(cat.avg_resolution_hours)}</td>
        </tr>
      `
    })

    htmlContent += `
          </tbody>
        </table>
      </div>
    `
  }

  if (data.user_performance?.length > 0) {
    htmlContent += `
      <div class="section">
        <h2>Technician Performance</h2>
        <table>
          <thead>
            <tr>
              <th>Technician</th>
              <th>Assigned</th>
              <th>Resolved</th>
              <th>Resolution Rate</th>
              <th>Avg Resolution Time</th>
            </tr>
          </thead>
          <tbody>
    `

    data.user_performance.forEach((user: any) => {
      htmlContent += `
        <tr>
          <td>${user.user_name}</td>
          <td>${user.total_assigned}</td>
          <td>${user.resolved_tickets}</td>
          <td>${user.resolution_rate_percent}%</td>
          <td>${formatHours(user.avg_resolution_hours)}</td>
        </tr>
      `
    })

    htmlContent += `
          </tbody>
        </table>
      </div>
    `
  }

  if (data.ticket_details?.length > 0) {
    htmlContent += `
      <div class="section">
        <h2>Trouble Ticket Details (Latest 20)</h2>
        <table>
          <thead>
            <tr>
              <th>Ticket #</th>
              <th>Brand</th>
              <th>Customer</th>
              <th>Problem</th>
              <th>Status</th>
              <th>Start Clock</th>
              <th>Pending Clock</th>
              <th>Total Downtime</th>
              <th>Assigned To</th>
            </tr>
          </thead>
          <tbody>
    `

    data.ticket_details.slice(0, 20).forEach((ticket: any) => {
      htmlContent += `
        <tr>
          <td>${ticket.ticket_number}</td>
          <td>${ticket.customer_brand}</td>
          <td>${ticket.customer_name}</td>
          <td><strong>${ticket.title || 'Lain-lain'}</strong>: ${ticket.description.substring(0, 80)}${ticket.description.length > 80 ? '...' : ''}</td>
          <td>${ticket.status}</td>
          <td>${formatReportDateTime(ticket.downtime_start || ticket.created_at)}</td>
          <td>${getPendingDuration(ticket)}</td>
          <td>${getDowntimeDuration(ticket)}</td>
          <td>${ticket.assigned_to}</td>
        </tr>
      `
    })

    htmlContent += `
          </tbody>
        </table>
        ${data.ticket_details.length > 20 ? `<p><em>*Showing 20 of ${data.ticket_details.length} total tickets. Use Excel or CSV export for complete data.</em></p>` : ''}
      </div>
    `
  }

  htmlContent += `
    </body>
    </html>
  `

  return htmlContent
}

const downloadPDFFile = (htmlContent: string, filename: string) => {
  const printWindow = window.open('', '_blank', 'width=1000,height=800')
  if (printWindow) {
    printWindow.document.write(htmlContent)
    printWindow.document.close()
    printWindow.focus()

    printWindow.onload = () => {
      setTimeout(() => {
        printWindow.print()

        showNotification('Print dialog opened. Choose "Save as PDF" as the destination.', 'info')

        printWindow.onafterprint = () => {
          printWindow.close()
        }
      }, 1000)
    }
  } else {
    showNotification('Please allow popups for this site to use PDF export.', 'warning')

    const blob = new Blob([htmlContent], { type: 'text/html;charset=utf-8;' })
    const link = document.createElement('a')

    if (link.download !== undefined) {
      const url = URL.createObjectURL(blob)
      link.setAttribute('href', url)
      link.setAttribute('download', filename.replace('.pdf', '.html'))
      link.style.visibility = 'hidden'
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)

      showNotification('HTML file downloaded. Open and print to PDF.', 'info')
    }
  }
}

// JSON Helper Functions
const downloadJSONFile = (data: any, filename: string) => {
  const jsonString = JSON.stringify(data, null, 2)
  const blob = new Blob([jsonString], { type: 'application/json;charset=utf-8;' })
  const link = document.createElement('a')

  if (link.download !== undefined) {
    const url = URL.createObjectURL(blob)
    link.setAttribute('href', url)
    link.setAttribute('download', filename)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }
}

// Watch for changes
watch(trendPeriod, () => {
  loadMonthlyTrends()
})

// Lifecycle
onMounted(() => {
  refreshAllData()
})
</script>

<style scoped>
.trouble-ticket-report-container {
  width: 100%;
  max-width: 100%;
  padding: 20px 24px;
  margin: 0;
  box-sizing: border-box;
}

/* Header & Panels */
.header-card,
.filter-card,
.chart-panel-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.04), 0 1px 2px 0 rgba(0, 0, 0, 0.02);
  transition: all 0.2s ease;
}

.header-card {
  padding: 18px 24px;
}

.header-avatar-box {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 4px 10px rgba(15, 23, 42, 0.15);
}

.header-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: #0f172a;
  letter-spacing: -0.01em;
}

.header-subtitle,
.chart-panel-subtitle {
  font-size: 0.8125rem;
  color: #64748b;
}

/* Action Buttons Header */
.header-actions {
  display: flex !important;
  align-items: center !important;
  gap: 12px !important;
}

.header-actions .v-btn + .v-btn {
  margin-left: 12px !important;
}

.action-btn-refresh {
  border-color: #cbd5e1 !important;
  color: #334155 !important;
}

.action-btn-refresh:hover {
  background-color: #f8fafc !important;
  border-color: #94a3b8 !important;
}

.export-btn-emerald {
  background-color: #10b981 !important;
  color: #ffffff !important;
  font-weight: 600 !important;
}

.export-btn-emerald:hover {
  background-color: #059669 !important;
}

/* Filter Card Elements */
.filter-title {
  font-size: 0.9375rem;
  font-weight: 700;
  color: #0f172a;
}

.filter-subtitle {
  font-size: 0.8125rem;
  color: #64748b;
  margin-top: 2px;
}

.filter-icon-box {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: #eff6ff;
  border: 1px solid #dbeafe;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

/* Section Header */
.section-icon-badge {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: #eff6ff;
  border: 1px solid #dbeafe;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.section-title {
  font-size: 1.125rem;
  font-weight: 700;
  color: #0f172a;
  line-height: 1.25;
}

.section-subtitle {
  font-size: 0.8125rem;
  color: #64748b;
  margin-top: 2px;
}

.chart-header-icon-box {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: #f1f5f9;
  display: flex;
  align-items: center;
  justify-content: center;
}

.chart-panel-title {
  font-size: 0.9375rem;
  font-weight: 700;
  color: #0f172a;
}

/* Metric Cards */
.metric-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  position: relative;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.metric-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.06);
  border-color: #cbd5e1;
}

.metric-accent-line {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
}

.line-blue { background-color: #0284c7; }
.line-green { background-color: #10b981; }
.line-amber { background-color: #f59e0b; }
.line-rose { background-color: #f43f5e; }

.metric-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  display: inline-block;
}

.dot-blue { background-color: #0284c7; }
.dot-green { background-color: #10b981; }
.dot-amber { background-color: #f59e0b; }
.dot-rose { background-color: #f43f5e; }

.metric-label {
  font-size: 0.6875rem;
  font-weight: 700;
  color: #64748b;
  letter-spacing: 0.05em;
}

.metric-value {
  font-size: 1.875rem;
  font-weight: 800;
  color: #0f172a;
  line-height: 1.15;
  letter-spacing: -0.02em;
}

.metric-icon-box {
  width: 36px;
  height: 36px;
  border-radius: 8px;
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

.box-amber {
  background: #fffbeb;
  border: 1px solid #fde68a;
}

.box-rose {
  background: #fff1f2;
  border: 1px solid #fecdd3;
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

.pill-amber {
  background: #fffbeb;
  color: #b45309;
  border: 1px solid #fde68a;
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

/* Ticket Number Badge */
.ticket-number-badge {
  display: inline-flex;
  align-items: center;
  padding: 3px 8px;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-family: monospace;
  font-size: 0.8125rem;
  font-weight: 600;
  color: #334155;
  letter-spacing: -0.01em;
}

/* Modern Data Table Tweaks */
.modern-data-table :deep(th) {
  font-size: 0.75rem !important;
  font-weight: 700 !important;
  text-transform: uppercase !important;
  letter-spacing: 0.05em !important;
  color: #475569 !important;
  background-color: #f8fafc !important;
  border-bottom: 1px solid #e2e8f0 !important;
  padding: 12px 16px !important;
}

.modern-data-table :deep(td) {
  border-bottom: 1px solid #f1f5f9 !important;
  padding: 10px 16px !important;
}

.modern-data-table :deep(tr:hover td) {
  background-color: #f8fafc !important;
}

/* Filter Bar */
.ticket-filter-bar {
  background: #f8fafc;
  border-color: #e2e8f0 !important;
}

/* Export Option Box */
.export-option-box {
  border-color: rgba(0, 0, 0, 0.08);
  transition: all 0.15s ease;
}

.export-option-box:hover {
  background-color: #f8fafc;
  border-color: rgba(0, 0, 0, 0.16);
}

.export-format-badge.blue-light { background: #eff6ff; }
.export-format-badge.green-light { background: #f0fdf4; }
.export-format-badge.red-light { background: #fef2f2; }
.export-format-badge.orange-light { background: #fffbeb; }

/* Empty Chart States */
.empty-chart-state {
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 220px;
}

.empty-chart-icon-box {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: #f1f5f9;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-chart-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: #334155;
}

.empty-chart-desc {
  font-size: 0.75rem;
  color: #64748b;
  max-width: 320px;
}

/* Segmented Buttons */
.segmented-toggle {
  border: 1px solid #e2e8f0 !important;
  border-radius: 8px !important;
  background-color: #f8fafc !important;
  padding: 2px !important;
}

.segmented-toggle .v-btn {
  border: none !important;
  border-radius: 6px !important;
  color: #64748b !important;
  font-weight: 500 !important;
  font-size: 0.75rem !important;
  transition: all 0.2s ease !important;
}

.segmented-toggle .v-btn--active {
  background-color: #ffffff !important;
  color: #0f172a !important;
  font-weight: 700 !important;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08) !important;
}

.segmented-btn {
  font-size: 0.75rem !important;
  font-weight: 600 !important;
  min-width: 48px !important;
  text-transform: none !important;
}

/* Dark Mode Overrides */
:deep(.v-theme--dark) .header-card,
:deep(.v-theme--dark) .filter-card,
:deep(.v-theme--dark) .metric-card,
:deep(.v-theme--dark) .chart-panel-card,
:deep(.v-theme--dark) .export-dialog-card {
  background: #1e293b !important;
  border-color: rgba(255, 255, 255, 0.08) !important;
}

:deep(.v-theme--dark) .header-title,
:deep(.v-theme--dark) .section-title,
:deep(.v-theme--dark) .filter-title,
:deep(.v-theme--dark) .chart-panel-title,
:deep(.v-theme--dark) .metric-value {
  color: #f8fafc !important;
}

:deep(.v-theme--dark) .header-subtitle,
:deep(.v-theme--dark) .section-subtitle,
:deep(.v-theme--dark) .filter-subtitle,
:deep(.v-theme--dark) .chart-panel-subtitle,
:deep(.v-theme--dark) .metric-label {
  color: #94a3b8 !important;
}

:deep(.v-theme--dark) .header-avatar-box {
  background: #0f172a !important;
}

:deep(.v-theme--dark) .filter-icon-box,
:deep(.v-theme--dark) .section-icon-badge,
:deep(.v-theme--dark) .chart-header-icon-box,
:deep(.v-theme--dark) .empty-chart-icon-box {
  background: #334155 !important;
  border-color: #475569 !important;
}

:deep(.v-theme--dark) .action-btn-refresh {
  border-color: #475569 !important;
  color: #cbd5e1 !important;
}

:deep(.v-theme--dark) .modern-data-table :deep(th) {
  background-color: #0f172a !important;
  color: #cbd5e1 !important;
  border-bottom-color: rgba(255, 255, 255, 0.08) !important;
}

:deep(.v-theme--dark) .modern-data-table :deep(td) {
  border-bottom-color: rgba(255, 255, 255, 0.05) !important;
}

:deep(.v-theme--dark) .ticket-filter-bar {
  background: #0f172a !important;
  border-color: rgba(255, 255, 255, 0.08) !important;
}

:deep(.v-theme--dark) .ticket-number-badge {
  background: #1e293b !important;
  border-color: #334155 !important;
  color: #94a3b8 !important;
}

:deep(.v-theme--dark) .export-option-box {
  border-color: rgba(255, 255, 255, 0.08) !important;
}

:deep(.v-theme--dark) .export-option-box:hover {
  background-color: #334155 !important;
}

@media (max-width: 600px) {
  .trouble-ticket-report-container {
    padding: 12px;
  }
}

@media print {
  .trouble-ticket-report-container {
    padding: 0;
  }
  .header-card .v-btn,
  .filter-card,
  .ticket-filter-bar {
    display: none !important;
  }
}
</style>
