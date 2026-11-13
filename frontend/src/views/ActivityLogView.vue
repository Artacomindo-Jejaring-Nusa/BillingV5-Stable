<template>
  <v-container fluid class="pa-4 pa-md-6">
    <div class="header-card mb-4 mb-md-6">
      <div class="d-flex align-center header-info">
        <div class="header-avatar-wrapper">
          <v-avatar class="header-avatar" size="48">
            <v-icon size="26">mdi-history</v-icon>
          </v-avatar>
        </div>
        <div class="ml-4">
          <h1 class="header-title">Log Aktivitas</h1>
          <p class="header-subtitle">Melacak semua perubahan data penting dalam sistem</p>
        </div>
      </div>
    </div>

    <!-- Tabs Section -->
    <v-card class="log-tabs-card mb-4 mb-md-6" elevation="0" rounded="xl">
      <v-tabs
        v-model="activeTab"
        color="#0e7c8c"
        slider-color="#0e7c8c"
        density="comfortable"
        grow
        class="custom-log-tabs"
      >
        <v-tab value="user" class="text-none font-weight-bold">
          <v-icon start size="18">mdi-account-clock-outline</v-icon>
          Log Aktivitas Pengguna
          <v-chip size="x-small" color="primary" variant="tonal" class="ms-2 font-weight-medium">Audit User</v-chip>
        </v-tab>
        <v-tab value="system" class="text-none font-weight-bold">
          <v-icon start size="18">mdi-server-network</v-icon>
          Log Sistem & Otomatisasi
          <v-chip size="x-small" color="teal" variant="tonal" class="ms-2 font-weight-medium">Cron & Xendit</v-chip>
        </v-tab>
      </v-tabs>
    </v-card>

    <!-- TAB 1: User Activity Logs -->
    <template v-if="activeTab === 'user'">
      <!-- Filters Section -->
      <v-card class="filter-card mb-4 mb-md-6" elevation="0" rounded="xl">
        <div class="filter-card-header pa-4 pa-md-5 d-flex align-center flex-wrap gap-3">
          <div class="d-flex align-center">
            <div class="filter-icon-box me-3">
              <v-icon color="#0e7c8c" size="20">mdi-account-clock-outline</v-icon>
            </div>
            <div>
              <h3 class="filter-card-title">Filter Pencarian Aktivitas User</h3>
              <p class="filter-card-subtitle">Temukan riwayat audit dan aktivitas pengguna sistem</p>
            </div>
          </div>
          <v-spacer></v-spacer>
          <v-btn
            v-if="hasActiveFilters"
            variant="tonal"
            color="error"
            prepend-icon="mdi-filter-off-outline"
            size="small"
            @click="clearFilters"
            class="clear-filters-btn"
          >
            Reset Filter
          </v-btn>
        </div>

        <v-divider class="filter-divider"></v-divider>

        <v-card-text class="filter-content pa-4 pa-md-5">
          <v-row dense>
            <!-- Search Keyword -->
            <v-col cols="12" md="3">
              <v-text-field
                v-model="filters.search"
                label="Cari aktivitas..."
                prepend-inner-icon="mdi-magnify"
                variant="outlined"
                density="compact"
                hide-details
                clearable
                placeholder="Kata kunci..."
                class="search-field"
                @update:model-value="debouncedSearch"
              ></v-text-field>
            </v-col>

            <!-- User Filter -->
            <v-col cols="12" md="3">
              <v-select
                v-model="filters.user_id"
                :items="userOptions"
                label="Pengguna"
                prepend-inner-icon="mdi-account"
                variant="outlined"
                density="compact"
                hide-details
                clearable
                placeholder="Semua pengguna"
                item-title="title"
                item-value="value"
                class="user-filter"
                @update:model-value="applyFilters"
              ></v-select>
            </v-col>

            <!-- Action Filter -->
            <v-col cols="12" md="2">
              <v-select
                v-model="filters.action"
                :items="actionOptions"
                label="Aksi"
                prepend-inner-icon="mdi-cog"
                variant="outlined"
                density="compact"
                hide-details
                clearable
                placeholder="Semua aksi"
                class="action-filter"
                @update:model-value="applyFilters"
              ></v-select>
            </v-col>

            <!-- Date Range Filter -->
            <v-col cols="12" sm="6" md="2">
              <v-text-field
                v-model="filters.date_from"
                type="date"
                label="Dari tanggal"
                prepend-inner-icon="mdi-calendar-start"
                variant="outlined"
                density="compact"
                hide-details
                clearable
                class="date-from-filter"
                @update:model-value="applyFilters"
              ></v-text-field>
            </v-col>
            <v-col cols="12" sm="6" md="2">
              <v-text-field
                v-model="filters.date_to"
                type="date"
                label="Sampai tanggal"
                prepend-inner-icon="mdi-calendar-end"
                variant="outlined"
                density="compact"
                hide-details
                clearable
                class="date-to-filter"
                @update:model-value="applyFilters"
              ></v-text-field>
            </v-col>
          </v-row>

          <!-- Active Filters Display -->
          <div v-if="hasActiveFilters" class="active-filters mt-4">
            <div class="d-flex align-center flex-wrap gap-2">
              <span class="text-caption font-weight-bold text-medium-emphasis me-2">Filter aktif:</span>
              <v-chip
                v-if="filters.search"
                color="primary"
                variant="tonal"
                size="small"
                closable
                @click:close="filters.search = ''; applyFilters()"
              >
                <v-icon start size="14">mdi-magnify</v-icon>
                {{ filters.search }}
              </v-chip>
              <v-chip
                v-if="filters.user_id"
                color="primary"
                variant="tonal"
                size="small"
                closable
                @click:close="filters.user_id = null; applyFilters()"
              >
                <v-icon start size="14">mdi-account</v-icon>
                {{ getUserDisplayName(filters.user_id) }}
              </v-chip>
              <v-chip
                v-if="filters.action"
                :color="getActionColor(filters.action)"
                variant="tonal"
                size="small"
                closable
                @click:close="filters.action = ''; applyFilters()"
              >
                <v-icon start size="14">{{ getActionIcon(filters.action) }}</v-icon>
                {{ filters.action }}
              </v-chip>
              <v-chip
                v-if="filters.date_from"
                color="teal"
                variant="tonal"
                size="small"
                closable
                @click:close="filters.date_from = ''; applyFilters()"
              >
                <v-icon start size="14">mdi-calendar-start</v-icon>
                {{ formatDateFilter(filters.date_from) }}
              </v-chip>
              <v-chip
                v-if="filters.date_to"
                color="teal"
                variant="tonal"
                size="small"
                closable
                @click:close="filters.date_to = ''; applyFilters()"
              >
                <v-icon start size="14">mdi-calendar-end</v-icon>
                {{ formatDateFilter(filters.date_to) }}
              </v-chip>
            </div>
          </div>
        </v-card-text>
      </v-card>

      <!-- Activity Logs Card -->
      <v-card class="activity-card" elevation="0" rounded="xl">
        <!-- Card Header -->
        <div class="table-header">
          <v-card-title class="d-flex align-center pa-4 pa-md-5">
            <div class="d-flex align-center flex-grow-1">
              <v-icon color="#0e7c8c" size="22" class="me-3">mdi-format-list-bulleted</v-icon>
              <div>
                <h2 class="text-subtitle-1 font-weight-bold mb-0">Riwayat Aktivitas Pengguna</h2>
                <p class="text-caption text-medium-emphasis mb-0">
                  {{ totalLogs }} total aktivitas tercatat
                </p>
              </div>
            </div>

            <!-- Refresh Button -->
            <v-btn
              icon="mdi-refresh"
              variant="tonal"
              size="small"
              color="primary"
              :loading="loading"
              @click="refreshData"
              class="refresh-btn rounded-lg"
              title="Segarkan Data"
            ></v-btn>
          </v-card-title>
        </div>

        <v-divider class="card-divider"></v-divider>

        <!-- Data Table -->
        <div class="table-container">
          <v-data-table-server
            v-model:items-per-page="itemsPerPage"
            :headers="headers"
            :items="logs"
            :items-length="totalLogs"
            :loading="loading"
            @update:options="fetchLogs"
            class="activity-table"
            :no-data-text="'Belum ada aktivitas yang tercatat'"
            loading-text="Memuat data log..."
            :items-per-page-options="[10, 25, 50, 100]"
          >
            <!-- User Column -->
            <template v-slot:item.user="{ item }">
              <div class="user-cell d-flex align-center py-2">
                <div class="user-avatar-wrapper me-3">
                  <v-avatar size="36" class="user-avatar">
                    <div class="user-avatar-bg">
                      <span class="user-initial">{{ item.user.name.charAt(0).toUpperCase() }}</span>
                    </div>
                  </v-avatar>
                </div>
                <div class="user-info">
                  <div class="user-name font-weight-bold text-body-2">{{ item.user.name }}</div>
                  <div class="user-email text-caption text-medium-emphasis">{{ item.user.email }}</div>
                </div>
              </div>
            </template>

            <!-- Action Column -->
            <template v-slot:item.action="{ item }">
              <v-chip 
                :color="getActionColor(item.action)" 
                variant="flat" 
                size="small" 
                class="action-chip font-weight-bold"
                label
              >
                <v-icon start size="16">{{ getActionIcon(item.action) }}</v-icon>
                {{ item.action }}
              </v-chip>
            </template>

            <!-- Timestamp Column -->
            <template v-slot:item.timestamp="{ item }">
              <div class="timestamp-cell">
                <div class="timestamp-primary text-body-2 font-weight-medium">
                  {{ formatDate(item.timestamp) }}
                </div>
                <div class="timestamp-secondary text-caption text-medium-emphasis">
                  {{ formatTime(item.timestamp) }}
                </div>
              </div>
            </template>

            <!-- Details Column -->
            <template v-slot:item.details="{ item }">
              <div class="details-cell text-center">
                <v-btn
                  icon="mdi-eye-outline"
                  variant="tonal"
                  size="x-small"
                  color="primary"
                  class="action-icon-btn rounded-lg"
                  @click="showDetails(item.details || '', item)"
                  title="Lihat Detail Aktivitas"
                ></v-btn>
              </div>
            </template>

            <!-- Loading State -->
            <template v-slot:loading>
              <SkeletonLoader type="table" :rows="8" />
            </template>

            <!-- No Data State -->
            <template v-slot:no-data>
              <div class="no-data-container d-flex flex-column align-center justify-center pa-8">
                <v-icon size="64" color="disabled" class="mb-4">mdi-history-off</v-icon>
                <p class="text-h6 font-weight-medium text-primary mb-2">Belum ada aktivitas</p>
                <p class="text-body-2 text-info mb-0">Data aktivitas akan muncul setelah ada perubahan sistem</p>
              </div>
            </template>
          </v-data-table-server>
        </div>
      </v-card>
    </template>

    <!-- TAB 2: System & Automation Logs -->
    <template v-else-if="activeTab === 'system'">
      <!-- System Filters Section -->
      <v-card class="filter-card mb-4 mb-md-6" elevation="0" rounded="xl">
        <div class="filter-card-header pa-4 pa-md-5 d-flex align-center flex-wrap gap-3">
          <div class="d-flex align-center">
            <div class="filter-icon-box me-3">
              <v-icon color="#0e7c8c" size="22">mdi-server-network</v-icon>
            </div>
            <div>
              <h3 class="filter-card-title">Filter Log Sistem & Otomatisasi</h3>
              <p class="filter-card-subtitle">Pantau proses cron auto-billing, suspensi, & callback Xendit</p>
            </div>
          </div>
          <v-spacer></v-spacer>
          <div class="d-flex align-center gap-3">
            <div class="live-refresh-pill d-flex align-center px-3 py-1.5 rounded-pill">
              <div class="live-pulse-dot" :class="{ 'live-pulse-dot--active': autoRefresh }"></div>
              <span class="live-label text-caption font-weight-medium ms-2 me-2">Live Update (15s)</span>
              <v-switch
                v-model="autoRefresh"
                color="#0e7c8c"
                density="compact"
                hide-details
                class="live-switch ma-0 pa-0"
              ></v-switch>
            </div>
            <v-btn
              v-if="hasActiveSystemFilters"
              variant="tonal"
              color="error"
              prepend-icon="mdi-refresh"
              size="small"
              rounded="lg"
              @click="clearSystemFilters"
              class="clear-filters-btn"
            >
              Reset Filter
            </v-btn>
          </div>
        </div>

        <v-divider class="filter-divider"></v-divider>

        <v-card-text class="filter-content pa-4 pa-md-5">
          <div class="d-flex flex-column gap-4">
            <!-- Row 1: Tingkat Log Segmented Filter Toolbar & Date Quick Presets -->
            <div class="filter-toolbar-row d-flex align-center justify-space-between flex-wrap gap-3 pb-3">
              <div class="d-flex align-center flex-wrap gap-2">
                <span class="filter-section-label">Tingkat Log:</span>
                <div class="segmented-control">
                  <button
                    type="button"
                    class="segmented-btn"
                    :class="{ active: systemFilters.level === '' }"
                    @click="setSystemLevel('')"
                  >
                    Semua
                  </button>
                  <button
                    type="button"
                    class="segmented-btn btn-info"
                    :class="{ active: systemFilters.level === 'INFO' }"
                    @click="setSystemLevel('INFO')"
                  >
                    <span class="status-indicator-dot dot-info"></span>
                    INFO
                  </button>
                  <button
                    type="button"
                    class="segmented-btn btn-warning"
                    :class="{ active: systemFilters.level === 'WARN' }"
                    @click="setSystemLevel('WARN')"
                  >
                    <span class="status-indicator-dot dot-warning"></span>
                    WARN
                  </button>
                  <button
                    type="button"
                    class="segmented-btn btn-error"
                    :class="{ active: systemFilters.level === 'ERROR' }"
                    @click="setSystemLevel('ERROR')"
                  >
                    <span class="status-indicator-dot dot-error"></span>
                    ERROR
                  </button>
                </div>
              </div>

              <!-- Quick Date Range Shortcuts -->
              <div class="d-flex align-center flex-wrap gap-1">
                <span class="filter-section-label text-caption me-1">Rentang Cepat:</span>
                <v-btn
                  size="x-small"
                  variant="tonal"
                  color="primary"
                  class="text-none font-weight-medium rounded-pill"
                  @click="setDateRangeQuick('today')"
                >
                  Hari Ini
                </v-btn>
                <v-btn
                  size="x-small"
                  variant="tonal"
                  color="primary"
                  class="text-none font-weight-medium rounded-pill"
                  @click="setDateRangeQuick('7days')"
                >
                  7 Hari
                </v-btn>
                <v-btn
                  size="x-small"
                  variant="tonal"
                  color="primary"
                  class="text-none font-weight-medium rounded-pill"
                  @click="setDateRangeQuick('month')"
                >
                  Bulan Ini
                </v-btn>
              </div>
            </div>

            <!-- Row 2: Search and Date Range Inputs -->
            <v-row dense>
              <!-- Search Keyword -->
              <v-col cols="12" md="6">
                <v-text-field
                  v-model="systemFilters.search"
                  label="Cari modul, aksi, atau pesan sistem..."
                  prepend-inner-icon="mdi-magnify"
                  variant="outlined"
                  density="compact"
                  hide-details
                  clearable
                  placeholder="Contoh: SchedulerManager, XENDIT, SUSPEND, CRON..."
                  class="filter-input-field"
                  @update:model-value="debouncedSystemSearch"
                ></v-text-field>
              </v-col>

              <!-- Date Range Filter -->
              <v-col cols="12" sm="6" md="3">
                <v-text-field
                  v-model="systemFilters.date_from"
                  type="date"
                  label="Dari tanggal"
                  prepend-inner-icon="mdi-calendar-start"
                  variant="outlined"
                  density="compact"
                  hide-details
                  clearable
                  class="filter-input-field"
                  @update:model-value="applySystemFilters"
                ></v-text-field>
              </v-col>
              <v-col cols="12" sm="6" md="3">
                <v-text-field
                  v-model="systemFilters.date_to"
                  type="date"
                  label="Sampai tanggal"
                  prepend-inner-icon="mdi-calendar-end"
                  variant="outlined"
                  density="compact"
                  hide-details
                  clearable
                  class="filter-input-field"
                  @update:model-value="applySystemFilters"
                ></v-text-field>
              </v-col>
            </v-row>

            <!-- Active System Filters Display -->
            <div v-if="hasActiveSystemFilters" class="active-filters mt-2">
              <div class="d-flex align-center flex-wrap gap-2">
                <span class="text-caption font-weight-bold text-medium-emphasis me-2">Filter aktif:</span>
                <v-chip
                  v-if="systemFilters.search"
                  color="primary"
                  variant="tonal"
                  size="small"
                  closable
                  @click:close="systemFilters.search = ''; applySystemFilters()"
                >
                  <v-icon start size="14">mdi-magnify</v-icon>
                  {{ systemFilters.search }}
                </v-chip>
                <v-chip
                  v-if="systemFilters.level"
                  :color="getSystemLevelColor(systemFilters.level)"
                  variant="tonal"
                  size="small"
                  closable
                  @click:close="setSystemLevel('')"
                >
                  <v-icon start size="14">{{ getSystemLevelIcon(systemFilters.level) }}</v-icon>
                  {{ systemFilters.level }}
                </v-chip>
                <v-chip
                  v-if="systemFilters.date_from"
                  color="teal"
                  variant="tonal"
                  size="small"
                  closable
                  @click:close="systemFilters.date_from = ''; applySystemFilters()"
                >
                  <v-icon start size="14">mdi-calendar-start</v-icon>
                  {{ formatDateFilter(systemFilters.date_from) }}
                </v-chip>
                <v-chip
                  v-if="systemFilters.date_to"
                  color="teal"
                  variant="tonal"
                  size="small"
                  closable
                  @click:close="systemFilters.date_to = ''; applySystemFilters()"
                >
                  <v-icon start size="14">mdi-calendar-end</v-icon>
                  {{ formatDateFilter(systemFilters.date_to) }}
                </v-chip>
              </div>
            </div>
          </div>
        </v-card-text>
      </v-card>

      <!-- System Logs Table Card -->
      <v-card class="activity-card" elevation="0" rounded="xl">
        <div class="table-header">
          <v-card-title class="d-flex align-center pa-4 pa-md-5">
            <div class="d-flex align-center flex-grow-1">
              <v-icon color="#0e7c8c" size="22" class="me-3">mdi-server-network</v-icon>
              <div>
                <h2 class="text-subtitle-1 font-weight-bold mb-0">Riwayat Log Sistem & Otomatisasi</h2>
                <p class="text-caption text-medium-emphasis mb-0">
                  {{ totalSystemLogs }} total catatan sistem
                </p>
              </div>
            </div>

            <v-btn
              icon="mdi-refresh"
              variant="tonal"
              size="small"
              color="primary"
              :loading="systemLoading"
              @click="refreshSystemData"
              class="refresh-btn rounded-lg"
              title="Segarkan Data"
            ></v-btn>
          </v-card-title>
        </div>

        <v-divider class="card-divider"></v-divider>

        <div class="table-container">
          <v-data-table-server
            v-model:items-per-page="systemItemsPerPage"
            :headers="systemHeaders"
            :items="systemLogs"
            :items-length="totalSystemLogs"
            :loading="systemLoading"
            @update:options="onSystemTableOptionsChange"
            class="activity-table"
            :no-data-text="'Belum ada log sistem yang tercatat'"
            loading-text="Memuat data log sistem..."
            :items-per-page-options="[10, 25, 50, 100]"
          >
            <!-- Waktu Column -->
            <template v-slot:item.created_at="{ item }">
              <div class="timestamp-cell">
                <div class="timestamp-primary text-body-2 font-weight-medium">
                  {{ formatDate(item.created_at) }}
                </div>
                <div class="timestamp-secondary text-caption text-medium-emphasis">
                  {{ formatTime(item.created_at) }}
                </div>
              </div>
            </template>

            <!-- Level Column -->
            <template v-slot:item.level="{ item }">
              <v-chip
                :color="getSystemLevelColor(item.level)"
                variant="flat"
                size="small"
                class="font-weight-bold rounded-pill"
                label
              >
                <v-icon start size="14">{{ getSystemLevelIcon(item.level) }}</v-icon>
                {{ item.level }}
              </v-chip>
            </template>

            <!-- Module Column -->
            <template v-slot:item.module="{ item }">
              <div class="d-flex align-center">
                <v-chip
                  :color="getModuleColor(item.module)"
                  variant="tonal"
                  size="small"
                  class="font-weight-medium rounded-pill"
                >
                  <v-icon start size="14">{{ getModuleIcon(item.module) }}</v-icon>
                  {{ item.module }}
                </v-chip>
              </div>
            </template>

            <!-- Message Column -->
            <template v-slot:item.message="{ item }">
              <div class="py-2">
                <div class="font-weight-bold text-caption text-primary mb-1">
                  {{ item.action }}
                </div>
                <div class="text-body-2 font-weight-medium" style="word-break: break-word; line-height: 1.4;">
                  {{ item.message }}
                </div>
              </div>
            </template>

            <!-- Details Column -->
            <template v-slot:item.details="{ item }">
              <div class="details-cell text-center">
                <v-btn
                  icon="mdi-eye-outline"
                  variant="tonal"
                  size="x-small"
                  color="primary"
                  class="action-icon-btn rounded-lg"
                  @click="showSystemDetails(item)"
                  title="Lihat Detail Log Sistem"
                ></v-btn>
              </div>
            </template>

            <!-- Loading State -->
            <template v-slot:loading>
              <SkeletonLoader type="table" :rows="8" />
            </template>

            <!-- No Data State -->
            <template v-slot:no-data>
              <div class="no-data-container d-flex flex-column align-center justify-center pa-8">
                <v-icon size="64" color="disabled" class="mb-4">mdi-server-off</v-icon>
                <p class="text-h6 font-weight-medium text-primary mb-2">Belum ada log sistem</p>
                <p class="text-body-2 text-info mb-0">Log otomatisasi sistem akan muncul di sini</p>
              </div>
            </template>
          </v-data-table-server>
        </div>
      </v-card>
    </template>

    <!-- Enhanced Details Dialog (Supports both User and System Logs) -->
    <v-dialog v-model="dialog" max-width="800" class="details-dialog" persistent>
      <v-card class="dialog-card" rounded="xl">
        <!-- Dialog Header with Clean High-Contrast Layout -->
        <div class="dialog-header-enhanced pa-4 pa-md-5">
          <div class="d-flex align-center justify-space-between">
            <div class="d-flex align-center">
              <div class="dialog-icon-wrapper-enhanced me-3">
                <v-icon color="#0e7c8c" size="24">mdi-information-outline</v-icon>
              </div>
              <div>
                <h3 class="dialog-title-enhanced text-h6 font-weight-bold mb-0">
                  {{ selectedSystemLog ? 'Detail Log Sistem' : 'Detail Aktivitas Pengguna' }}
                </h3>
                <p class="dialog-subtitle-enhanced text-caption mb-0">
                  {{ selectedSystemLog ? 'Informasi payload dan trace eksekusi sistem' : 'Informasi lengkap perubahan data sistem' }}
                </p>
              </div>
            </div>
            <v-btn
              icon="mdi-close"
              variant="text"
              size="small"
              class="close-btn-enhanced"
              @click="closeDetailDialog"
            ></v-btn>
          </div>
        </div>

        <!-- Dialog Content with Cards -->
        <div class="dialog-content-enhanced pa-6">
          <!-- User Summary Section -->
          <div v-if="selectedLog" class="summary-section mb-6">
            <div class="summary-header d-flex align-center mb-4">
              <div class="summary-avatar-wrapper me-3">
                <v-avatar size="48" class="summary-avatar">
                  <div class="summary-avatar-bg">
                    <span class="summary-initial">{{ selectedLog.user.name.charAt(0).toUpperCase() }}</span>
                  </div>
                </v-avatar>
              </div>
              <div class="summary-info flex-grow-1">
                <h4 class="summary-user-name text-h6 font-weight-bold mb-1">{{ selectedLog.user.name }}</h4>
                <div class="summary-user-email text-body-2 text-medium-emphasis mb-2">{{ selectedLog.user.email }}</div>
                <div class="d-flex align-center gap-2 flex-wrap">
                  <v-chip
                    :color="getActionColor(selectedLog.action)"
                    variant="elevated"
                    size="small"
                    class="summary-action-chip font-weight-bold"
                    label
                  >
                    <v-icon start size="16">{{ getActionIcon(selectedLog.action) }}</v-icon>
                    {{ selectedLog.action }}
                  </v-chip>
                  <v-chip
                    color="info"
                    variant="tonal"
                    size="small"
                    class="timestamp-chip"
                    label
                  >
                    <v-icon start size="14">mdi-clock-outline</v-icon>
                    {{ formatFullDateTime(selectedLog.timestamp) }}
                  </v-chip>
                </div>
              </div>
            </div>
          </div>

          <!-- System Log Summary Section -->
          <div v-if="selectedSystemLog" class="summary-section mb-6">
            <div class="summary-header d-flex align-center mb-4">
              <div class="summary-avatar-wrapper me-3">
                <v-avatar size="48" :color="getSystemLevelColor(selectedSystemLog.level)">
                  <v-icon color="white" size="24">{{ getSystemLevelIcon(selectedSystemLog.level) }}</v-icon>
                </v-avatar>
              </div>
              <div class="summary-info flex-grow-1">
                <div class="d-flex align-center gap-2 mb-1">
                  <v-chip
                    :color="getSystemLevelColor(selectedSystemLog.level)"
                    variant="flat"
                    size="small"
                    class="font-weight-bold"
                    label
                  >
                    {{ selectedSystemLog.level }}
                  </v-chip>
                  <v-chip
                    :color="getModuleColor(selectedSystemLog.module)"
                    variant="tonal"
                    size="small"
                    class="font-weight-bold"
                  >
                    {{ selectedSystemLog.module }}
                  </v-chip>
                  <v-chip
                    color="info"
                    variant="tonal"
                    size="small"
                    label
                  >
                    <v-icon start size="14">mdi-clock-outline</v-icon>
                    {{ formatFullDateTime(selectedSystemLog.created_at) }}
                  </v-chip>
                </div>
                <h4 class="text-subtitle-1 font-weight-bold text-primary mt-2 mb-1">
                  {{ selectedSystemLog.action }}
                </h4>
                <div class="text-body-2 text-medium-emphasis">
                  {{ selectedSystemLog.message }}
                </div>
              </div>
            </div>
          </div>

          <!-- Details Content -->
          <div v-if="detailsObject" class="details-content">
            <!-- Action Path / Info -->
            <v-card class="detail-card mb-4" elevation="0" variant="outlined">
              <v-card-title class="detail-card-header d-flex align-center pa-4">
                <v-icon color="primary" size="20" class="me-2">mdi-route</v-icon>
                <span class="text-subtitle-1 font-weight-bold">
                  {{ selectedSystemLog ? 'Aksi / Event' : 'Action Path' }}
                </span>
              </v-card-title>
              <v-divider class="detail-divider"></v-divider>
              <v-card-text class="pa-4">
                <code class="action-path">
                  {{ selectedSystemLog ? `${selectedSystemLog.module} : ${selectedSystemLog.action}` : (selectedLog?.action || 'N/A') }}
                </code>
              </v-card-text>
            </v-card>

            <!-- Payload Details -->
            <v-card class="detail-card mb-4" elevation="0" variant="outlined">
              <v-card-title class="detail-card-header d-flex align-center pa-4">
                <v-icon color="success" size="20" class="me-2">mdi-code-json</v-icon>
                <span class="text-subtitle-1 font-weight-bold">Data Payload & Context</span>
                <v-spacer></v-spacer>
                <v-btn
                  icon="mdi-content-copy"
                  variant="text"
                  size="small"
                  color="primary"
                  @click="copyToClipboard(JSON.stringify(detailsObject, null, 2))"
                  class="copy-btn"
                >
                  <v-icon size="18">mdi-content-copy</v-icon>
                </v-btn>
              </v-card-title>
              <v-divider class="detail-divider"></v-divider>
              <v-card-text class="pa-4">
                <div class="details-grid">
                  <template v-for="(value, key, index) in detailsObject" :key="key">
                    <div class="detail-item-enhanced">
                      <div class="detail-label">
                        <v-icon size="16" class="label-icon">mdi-chevron-right</v-icon>
                        <span class="label-text">{{ formatKey(key) }}</span>
                      </div>
                      <div class="detail-value-wrapper">
                        <div class="detail-value-content">
                          {{ formatValueEnhanced(value) }}
                        </div>
                        <v-chip
                          v-if="getValueType(value) !== 'string'"
                          :color="getValueTypeColor(getValueType(value))"
                          variant="tonal"
                          size="x-small"
                          class="value-type-chip ml-2"
                        >
                          {{ getValueType(value) }}
                        </v-chip>
                      </div>
                    </div>
                    <v-divider
                      v-if="index < Object.keys(detailsObject).length - 1"
                      class="detail-divider-enhanced my-3"
                    ></v-divider>
                  </template>
                </div>
              </v-card-text>
            </v-card>
          </div>

          <!-- Empty State -->
          <div v-else class="empty-state text-center py-8">
            <v-icon size="64" color="disabled" class="mb-4">mdi-information-off-outline</v-icon>
            <p class="text-h6 font-weight-medium text-primary mb-2">Tidak Ada Detail</p>
            <p class="text-body-2 text-info mb-0">Catatan ini tidak memiliki informasi detail payload tambahan</p>
          </div>
        </div>

        <!-- Dialog Actions -->
        <div class="dialog-actions-enhanced pa-4 pa-md-6">
          <v-btn
            variant="tonal"
            color="info"
            prepend-icon="mdi-download"
            @click="downloadDetails"
            class="action-btn me-3"
          >
            Export JSON
          </v-btn>
          <v-spacer></v-spacer>
          <v-btn
            variant="elevated"
            color="primary"
            prepend-icon="mdi-close"
            @click="closeDetailDialog"
            class="action-btn"
          >
            Tutup
          </v-btn>
        </div>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import apiClient from '@/services/api';
import { debounce } from 'lodash-es';
import SkeletonLoader from '@/components/SkeletonLoader.vue';

// --- Tab State ---
const activeTab = ref<'user' | 'system'>('user');

// --- User Activity Log Interfaces & State ---
interface User {
  id: number;
  name: string;
  email: string;
}

interface ActivityLog {
  id: number;
  user: User;
  timestamp: string;
  action: string;
  details?: string;
}

interface Filters {
  search: string;
  user_id: number | null;
  action: string;
  date_from: string;
  date_to: string;
}

const logs = ref<ActivityLog[]>([]);
const loading = ref(true);
const totalLogs = ref(0);
const itemsPerPage = ref(10);
const users = ref<User[]>([]);

const dialog = ref(false);
const detailsObject = ref<Record<string, any> | null>(null);
const selectedLog = ref<ActivityLog | null>(null);
const selectedSystemLog = ref<SystemLog | null>(null);

// Filters state
const filters = ref<Filters>({
  search: '',
  user_id: null,
  action: '',
  date_from: '',
  date_to: ''
});

const headers = [
  { title: 'Pengguna', key: 'user', sortable: false, width: '28%' },
  { title: 'Aksi', key: 'action', sortable: false, width: '22%' },
  { title: 'Waktu', key: 'timestamp', sortable: false, width: '25%' },
  { title: 'Detail', key: 'details', sortable: false, align: 'center', width: '100px' },
] as const;

// --- System Log Interfaces & State ---
interface SystemLog {
  id: number;
  level: string;
  module: string;
  action: string;
  message: string;
  details?: string | null;
  created_at: string;
  timestamp?: string;
}

interface SystemFilters {
  search: string;
  level: string;
  date_from: string;
  date_to: string;
}

const systemLogs = ref<SystemLog[]>([]);
const systemLoading = ref(false);
const totalSystemLogs = ref(0);
const systemItemsPerPage = ref(10);
const systemCurrentPage = ref(1);
const autoRefresh = ref(false);
let autoRefreshTimer: any = null;

const systemFilters = ref<SystemFilters>({
  search: '',
  level: '',
  date_from: '',
  date_to: ''
});

const systemHeaders = [
  { title: 'Waktu', key: 'created_at', sortable: false, width: '160px' },
  { title: 'Level', key: 'level', sortable: false, width: '110px' },
  { title: 'Modul', key: 'module', sortable: false, width: '160px' },
  { title: 'Aksi & Pesan', key: 'message', sortable: false },
  { title: 'Detail', key: 'details', sortable: false, align: 'center', width: '90px' },
] as const;

// --- Parsing helper for System Logs ---
function parseSystemLogItem(raw: any): SystemLog {
  const ts = raw.timestamp || raw.created_at || new Date().toISOString();
  let rawMsg = (raw.message || '').trim();
  let mod = (raw.module || '').trim();
  let act = (raw.action || '').trim();
  let displayMsg = rawMsg;

  if (!mod && rawMsg.includes(':')) {
    const colonIdx = rawMsg.indexOf(':');
    const candidate = rawMsg.substring(0, colonIdx).trim();
    if (candidate.length > 0 && candidate.length < 35 && !candidate.includes('\n')) {
      mod = candidate;
      displayMsg = rawMsg.substring(colonIdx + 1).trim();
    }
  }

  if (!mod) {
    mod = 'SYSTEM';
  }

  if (!act) {
    const combined = (rawMsg + ' ' + mod).toLowerCase();
    if (combined.includes('retry_failed_whatsapp') || combined.includes('whatsapp')) {
      act = 'WHATSAPP RETRY';
    } else if (combined.includes('retry_mikrotik_sync') || combined.includes('mikrotik')) {
      act = 'MIKROTIK SYNC';
    } else if (combined.includes('job') || combined.includes('scheduler') || combined.includes('cron')) {
      act = 'CRON JOB';
    } else if (combined.includes('xendit') || combined.includes('invoice') || combined.includes('payment')) {
      act = 'PAYMENT';
    } else if (combined.includes('suspend') || combined.includes('isolir')) {
      act = 'SUSPENSION';
    } else {
      act = 'SYSTEM EVENT';
    }
  }

  return {
    id: raw.id,
    level: raw.level || 'INFO',
    module: mod,
    action: act,
    message: displayMsg || rawMsg,
    details: raw.details || null,
    created_at: ts,
    timestamp: ts
  };
}

// --- Computed Properties ---
const hasActiveFilters = computed(() => {
  return !!(filters.value.search ||
         filters.value.user_id ||
         filters.value.action ||
         filters.value.date_from ||
         filters.value.date_to);
});

const hasActiveSystemFilters = computed(() => {
  return !!(systemFilters.value.search ||
         systemFilters.value.level ||
         systemFilters.value.date_from ||
         systemFilters.value.date_to);
});

const userOptions = computed(() => {
  return [
    { title: 'Semua pengguna', value: null },
    ...users.value.map(user => ({
      title: user.name,
      value: user.id,
      subtitle: user.email
    }))
  ];
});

const actionOptions = computed(() => [
  { title: 'POST (Create)', value: 'POST' },
  { title: 'PATCH (Update)', value: 'PATCH' },
  { title: 'DELETE (Remove)', value: 'DELETE' },
  { title: 'GET (View)', value: 'GET' }
]);

// Debounced search functions
const debouncedSearch = debounce(() => {
  applyFilters();
}, 500);

const debouncedSystemSearch = debounce(() => {
  applySystemFilters();
}, 500);

// Watchers
watch(activeTab, (newTab) => {
  if (newTab === 'system' && systemLogs.value.length === 0) {
    fetchSystemLogs({ page: 1, itemsPerPage: systemItemsPerPage.value });
  }
});

watch(autoRefresh, (enabled) => {
  if (enabled) {
    if (autoRefreshTimer) clearInterval(autoRefreshTimer);
    autoRefreshTimer = setInterval(() => {
      if (activeTab.value === 'system') {
        fetchSystemLogs({ page: systemCurrentPage.value, itemsPerPage: systemItemsPerPage.value, silent: true });
      }
    }, 15000);
  } else if (autoRefreshTimer) {
    clearInterval(autoRefreshTimer);
    autoRefreshTimer = null;
  }
});

onUnmounted(() => {
  if (autoRefreshTimer) {
    clearInterval(autoRefreshTimer);
    autoRefreshTimer = null;
  }
});

// --- User Logs Methods ---
async function fetchUsers() {
  try {
    const response = await apiClient.get('/users', {
      params: {
        skip: 0,
        limit: 1000,
      },
    });
    users.value = Array.isArray(response.data) ? response.data : (response.data.data || response.data.items || []);
  } catch (error) {
    console.error("Gagal mengambil data users:", error);
  }
}

async function fetchLogs({ page, itemsPerPage }: { page: number, itemsPerPage: number }) {
  loading.value = true;
  try {
    const params: any = {
      skip: (page - 1) * itemsPerPage,
      limit: itemsPerPage,
    };

    if (filters.value.search) params.search = filters.value.search;
    if (filters.value.user_id) params.user_id = filters.value.user_id;
    if (filters.value.action) params.action = filters.value.action;
    if (filters.value.date_from) params.date_from = filters.value.date_from;
    if (filters.value.date_to) params.date_to = filters.value.date_to;

    const response = await apiClient.get('/activity-logs', { params });
    logs.value = response.data.items || [];
    totalLogs.value = response.data.total || 0;
  } catch (error) {
    console.error("Gagal mengambil data log aktivitas:", error);
  } finally {
    loading.value = false;
  }
}

function refreshData() {
  fetchLogs({ page: 1, itemsPerPage: itemsPerPage.value });
}

function applyFilters() {
  fetchLogs({ page: 1, itemsPerPage: itemsPerPage.value });
}

function clearFilters() {
  filters.value = {
    search: '',
    user_id: null,
    action: '',
    date_from: '',
    date_to: ''
  };
  applyFilters();
}

// --- System Logs Methods ---
async function fetchSystemLogs(options?: { page?: number, itemsPerPage?: number, silent?: boolean }) {
  const page = options?.page || systemCurrentPage.value;
  const pageSize = options?.itemsPerPage || systemItemsPerPage.value;
  systemCurrentPage.value = page;
  systemItemsPerPage.value = pageSize;

  if (!options?.silent) {
    systemLoading.value = true;
  }

  try {
    const params: any = {
      page: page,
      page_size: pageSize,
    };

    if (systemFilters.value.search) params.search = systemFilters.value.search;
    if (systemFilters.value.level) params.level = systemFilters.value.level;
    if (systemFilters.value.date_from) params.date_from = systemFilters.value.date_from;
    if (systemFilters.value.date_to) params.date_to = systemFilters.value.date_to;

    const response = await apiClient.get('/system-logs', { params });
    const rawList = response.data?.data || (Array.isArray(response.data) ? response.data : []);
    systemLogs.value = rawList.map(parseSystemLogItem);
    totalSystemLogs.value = response.data?.total || systemLogs.value.length;
  } catch (error) {
    console.error("Gagal mengambil system logs:", error);
  } finally {
    if (!options?.silent) {
      systemLoading.value = false;
    }
  }
}

function onSystemTableOptionsChange(options: { page: number, itemsPerPage: number }) {
  fetchSystemLogs(options);
}

function refreshSystemData() {
  fetchSystemLogs({ page: 1, itemsPerPage: systemItemsPerPage.value });
}

function applySystemFilters() {
  fetchSystemLogs({ page: 1, itemsPerPage: systemItemsPerPage.value });
}

function setSystemLevel(level: string) {
  systemFilters.value.level = level;
  applySystemFilters();
}

function clearSystemFilters() {
  systemFilters.value = {
    search: '',
    level: '',
    date_from: '',
    date_to: ''
  };
  applySystemFilters();
}

function setDateRangeQuick(preset: 'today' | '7days' | 'month') {
  const now = new Date();
  const formatYMD = (d: Date) => {
    const year = d.getFullYear();
    const month = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
  };

  if (preset === 'today') {
    const todayStr = formatYMD(now);
    systemFilters.value.date_from = todayStr;
    systemFilters.value.date_to = todayStr;
  } else if (preset === '7days') {
    const past = new Date();
    past.setDate(past.getDate() - 7);
    systemFilters.value.date_from = formatYMD(past);
    systemFilters.value.date_to = formatYMD(now);
  } else if (preset === 'month') {
    const startOfMonth = new Date(now.getFullYear(), now.getMonth(), 1);
    systemFilters.value.date_from = formatYMD(startOfMonth);
    systemFilters.value.date_to = formatYMD(now);
  }
  applySystemFilters();
}

// --- Style & Format Helpers ---
function getSystemLevelColor(level: string): string {
  switch (level?.toUpperCase()) {
    case 'ERROR': return 'error';
    case 'WARN': return 'warning';
    case 'INFO': return 'info';
    default: return 'grey';
  }
}

function getSystemLevelIcon(level: string): string {
  switch (level?.toUpperCase()) {
    case 'ERROR': return 'mdi-alert-circle';
    case 'WARN': return 'mdi-alert';
    case 'INFO': return 'mdi-information';
    default: return 'mdi-circle-medium';
  }
}

function getModuleColor(module: string): string {
  const m = module?.toUpperCase() || '';
  if (m.includes('SCHEDULER') || m.includes('CRON') || m.includes('BILLING')) return 'teal';
  if (m.includes('SUSPEND') || m.includes('ISOLIR')) return 'deep-orange';
  if (m.includes('XENDIT') || m.includes('PAYMENT')) return 'indigo';
  if (m.includes('WHATSAPP') || m.includes('WA')) return 'success';
  if (m.includes('MIKROTIK') || m.includes('NETWORK')) return 'cyan';
  if (m.includes('SUBSCRIPTION')) return 'blue';
  return 'blue-grey';
}

function getModuleIcon(module: string): string {
  const m = module?.toUpperCase() || '';
  if (m.includes('SCHEDULER') || m.includes('CRON')) return 'mdi-clock-check-outline';
  if (m.includes('SUSPEND')) return 'mdi-account-lock-outline';
  if (m.includes('XENDIT')) return 'mdi-credit-card-sync-outline';
  if (m.includes('WHATSAPP')) return 'mdi-whatsapp';
  if (m.includes('MIKROTIK')) return 'mdi-router-wireless';
  if (m.includes('SUBSCRIPTION')) return 'mdi-account-cog-outline';
  return 'mdi-cog-outline';
}

function getUserDisplayName(userId: number): string {
  const user = users.value.find(u => u.id === userId);
  return user ? user.name : `User ${userId}`;
}

function formatDateFilter(dateString: string): string {
  const d = parseSafeDate(dateString);
  if (!d) return dateString;
  return d.toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric'
  });
}

function getActionColor(action: string): string {
  if (action.startsWith('POST')) return 'success';
  if (action.startsWith('PATCH')) return 'warning';
  if (action.startsWith('DELETE')) return 'error';
  if (action.startsWith('GET')) return 'info';
  return 'grey';
}

function getActionIcon(action: string): string {
  if (action.startsWith('POST')) return 'mdi-plus-circle-outline';
  if (action.startsWith('PATCH')) return 'mdi-pencil-outline';
  if (action.startsWith('DELETE')) return 'mdi-delete-outline';
  if (action.startsWith('GET')) return 'mdi-eye-outline';
  return 'mdi-cog-outline';
}

function showDetails(details: string, log: ActivityLog) {
  selectedSystemLog.value = null;
  selectedLog.value = log;
  if (details && details.trim() !== '') {
    try {
      detailsObject.value = JSON.parse(details);
    } catch {
      detailsObject.value = { "raw_data": details };
    }
  } else {
    detailsObject.value = {
      id: log.id,
      user: log.user ? log.user.name : '-',
      email: log.user ? log.user.email : '-',
      action: log.action,
      timestamp: log.timestamp
    };
  }
  dialog.value = true;
}

function showSystemDetails(log: SystemLog) {
  selectedLog.value = null;
  selectedSystemLog.value = log;
  if (log.details && log.details.trim() !== '') {
    try {
      detailsObject.value = JSON.parse(log.details);
    } catch {
      detailsObject.value = { "raw_data": log.details };
    }
  } else {
    detailsObject.value = {
      id: log.id,
      level: log.level,
      module: log.module,
      action: log.action,
      message: log.message,
      timestamp: log.created_at
    };
  }
  dialog.value = true;
}

function closeDetailDialog() {
  dialog.value = false;
  selectedLog.value = null;
  selectedSystemLog.value = null;
  detailsObject.value = null;
}

function parseSafeDate(timestamp: any): Date | null {
  if (!timestamp) return null;
  let str = String(timestamp).trim();
  if (str.includes(' ') && !str.includes('T')) {
    str = str.replace(' ', 'T');
  }
  const d = new Date(str);
  return isNaN(d.getTime()) ? null : d;
}

function formatFullDateTime(timestamp: string): string {
  const d = parseSafeDate(timestamp);
  if (!d) return '-';
  return d.toLocaleString('id-ID', {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  });
}

function formatValueEnhanced(value: any): string {
  if (value === null || value === undefined) return '-';
  if (typeof value === 'boolean') return value ? '✓ Ya' : '✗ Tidak';
  if (typeof value === 'number') return value.toLocaleString('id-ID');
  if (typeof value === 'object') {
    return JSON.stringify(value, null, 2);
  }
  return String(value);
}

function getValueType(value: any): string {
  if (value === null || value === undefined) return 'null';
  if (typeof value === 'boolean') return 'boolean';
  if (typeof value === 'number') return 'number';
  if (typeof value === 'string') return 'string';
  if (Array.isArray(value)) return 'array';
  return 'object';
}

function getValueTypeColor(type: string): string {
  const colors: Record<string, string> = {
    'string': 'blue',
    'number': 'green',
    'boolean': 'purple',
    'object': 'orange',
    'array': 'teal',
    'null': 'grey'
  };
  return colors[type] || 'grey';
}

async function copyToClipboard(text: string) {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text);
    } else {
      const textArea = document.createElement("textarea");
      textArea.value = text;
      textArea.style.position = "fixed";
      textArea.style.left = "-999999px";
      textArea.style.top = "-999999px";
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();
      document.execCommand('copy');
      textArea.remove();
    }
  } catch (err) {
    console.error('Failed to copy:', err);
  }
}

function downloadDetails() {
  if (!detailsObject.value) return;

  const data = selectedSystemLog.value ? {
    system_log: selectedSystemLog.value,
    details: detailsObject.value
  } : {
    log_info: selectedLog.value,
    details: detailsObject.value
  };

  const id = selectedSystemLog.value?.id || selectedLog.value?.id || 'log';
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `log-${id}-${new Date().toISOString().split('T')[0]}.json`;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

function formatKey(key: string): string {
  return key.replace(/_/g, ' ').replace(/\b\w/g, char => char.toUpperCase());
}

function formatDate(timestamp: string): string {
  const d = parseSafeDate(timestamp);
  if (!d) return '-';
  return d.toLocaleDateString('id-ID', { 
    day: 'numeric',
    month: 'short', 
    year: 'numeric' 
  });
}

function formatTime(timestamp: string): string {
  const d = parseSafeDate(timestamp);
  if (!d) return '-';
  return d.toLocaleTimeString('id-ID', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  });
}

// Initialize component
fetchUsers();
</script>

<style scoped>
/* =====================================================================
   DESIGN TOKENS — consistent with CustomerChatView ISP fiber-optic theme
   ===================================================================== */

/* Header Card */
.header-card {
  background: #ffffff;
  border-radius: 16px;
  padding: 20px 24px;
  border: 1px solid #e2e7eb;
  box-shadow: 0 1px 3px rgba(16, 24, 39, 0.06);
  position: relative;
}

.v-theme--dark .header-card {
  background: #1a1f2e;
  border-color: #2d3548;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}

.header-info {
  width: 100%;
  justify-content: flex-start;
}

.header-avatar-wrapper {
  flex-shrink: 0;
}

.header-avatar {
  background: linear-gradient(155deg, #0e7c8c 0%, #0a5866 100%) !important;
  color: #ffffff !important;
}

.header-title {
  font-size: 1.5rem;
  font-weight: 700;
  line-height: 1.3;
  margin-bottom: 2px;
  color: #101827 !important;
  letter-spacing: -0.01em;
}

.v-theme--dark .header-title {
  color: #f0f2f5 !important;
}

.header-subtitle {
  font-size: 0.85rem;
  line-height: 1.4;
  color: #47566e !important;
  font-weight: 400;
  margin-bottom: 0;
}

.v-theme--dark .header-subtitle {
  color: #8994a6 !important;
}

/* Tabs Card */
.log-tabs-card {
  background: #ffffff;
  border: 1px solid #e2e7eb;
}

.v-theme--dark .log-tabs-card {
  background: #1a1f2e;
  border-color: #2d3548;
}

.custom-log-tabs :deep(.v-tab) {
  letter-spacing: normal;
  transition: all 0.2s ease;
}

.custom-log-tabs :deep(.v-tab--selected) {
  color: #0e7c8c !important;
}

.custom-log-tabs :deep(.v-tab__slider) {
  height: 3px !important;
  background-color: #0e7c8c !important;
  border-radius: 3px 3px 0 0 !important;
}

/* Filter Card */
.filter-card {
  background: #ffffff;
  border: 1px solid #e2e7eb;
  box-shadow: 0 1px 3px rgba(16, 24, 39, 0.06);
  overflow: hidden;
}

.v-theme--dark .filter-card {
  background: #1a1f2e;
  border-color: #2d3548;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}

.filter-card-header {
  background: #f8fafc;
}

.v-theme--dark .filter-card-header {
  background: #141825;
}

.filter-icon-box {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: rgba(14, 124, 140, 0.08);
  border: 1px solid rgba(14, 124, 140, 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

/* Live Refresh Pill */
.live-refresh-pill {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  display: inline-flex;
  align-items: center;
  transition: all 0.2s ease;
}

.v-theme--dark .live-refresh-pill {
  background: #1e293b;
  border-color: #334155;
}

.live-pulse-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #94a3b8;
  transition: all 0.3s ease;
}

.live-pulse-dot--active {
  background-color: #10b981;
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.2);
  animation: pulseGreen 2s infinite;
}

@keyframes pulseGreen {
  0% {
    box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.4);
  }
  70% {
    box-shadow: 0 0 0 6px rgba(16, 185, 129, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(16, 185, 129, 0);
  }
}

.live-label {
  font-size: 0.8rem;
  color: #47566e;
  white-space: nowrap;
}

.v-theme--dark .live-label {
  color: #cbd5e1;
}

/* Filter Toolbar Row (Tingkat Log & Quick Range) */
.filter-toolbar-row {
  border-bottom: 1px solid #e2e8f0;
}

.v-theme--dark .filter-toolbar-row {
  border-bottom-color: #2d3548;
}

.filter-section-label {
  font-size: 0.8rem;
  font-weight: 600;
  color: #47566e;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.v-theme--dark .filter-section-label {
  color: #94a3b8;
}

/* Segmented Control for Log Levels */
.segmented-control {
  display: inline-flex;
  align-items: center;
  background: #f1f5f9;
  padding: 3px;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
  gap: 2px;
}

.v-theme--dark .segmented-control {
  background: #1e293b;
  border-color: #334155;
}

.segmented-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 5px 14px;
  border-radius: 7px;
  font-size: 0.8rem;
  font-weight: 500;
  color: #47566e;
  background: transparent;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease;
  line-height: 1.2;
}

.v-theme--dark .segmented-btn {
  color: #94a3b8;
}

.segmented-btn:hover {
  color: #101827;
  background: rgba(255, 255, 255, 0.5);
}

.v-theme--dark .segmented-btn:hover {
  color: #f8fafc;
  background: rgba(255, 255, 255, 0.05);
}

.segmented-btn.active {
  background: #ffffff;
  color: #0f172a;
  font-weight: 600;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08), 0 1px 2px rgba(0, 0, 0, 0.04);
}

.v-theme--dark .segmented-btn.active {
  background: #0f172a;
  color: #f8fafc;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.4);
}

.status-indicator-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  margin-right: 6px;
  flex-shrink: 0;
}

.dot-info {
  background-color: #0284c7;
}

.dot-warning {
  background-color: #f59e0b;
}

.dot-error {
  background-color: #ef4444;
}

.segmented-btn.btn-info.active {
  color: #0284c7;
}

.segmented-btn.btn-warning.active {
  color: #d97706;
}

.segmented-btn.btn-error.active {
  color: #dc2626;
}

.action-icon-btn {
  width: 32px !important;
  height: 32px !important;
  transition: all 0.2s ease;
}

.action-icon-btn:hover {
  transform: scale(1.08);
}

.filter-card-title {
  font-size: 0.95rem;
  font-weight: 600;
  color: #101827;
  margin-bottom: 0;
  line-height: 1.4;
}

.v-theme--dark .filter-card-title {
  color: #f0f2f5;
}

.filter-card-subtitle {
  font-size: 0.75rem;
  color: #47566e;
  margin-bottom: 0;
  line-height: 1.4;
}

.v-theme--dark .filter-card-subtitle {
  color: #8994a6;
}

.filter-divider {
  border-color: #e2e7eb !important;
}

.v-theme--dark .filter-divider {
  border-color: #2d3548 !important;
}

.filter-content {
  background: transparent;
}

.search-field,
.user-filter,
.action-filter,
.date-from-filter,
.date-to-filter {
  transition: all 0.3s ease;
}

.search-field :deep(.v-field),
.user-filter :deep(.v-field),
.action-filter :deep(.v-field),
.date-from-filter :deep(.v-field),
.date-to-filter :deep(.v-field) {
  border-radius: 8px;
  transition: all 0.3s ease;
}

.search-field :deep(.v-field:hover),
.user-filter :deep(.v-field:hover),
.action-filter :deep(.v-field:hover),
.date-from-filter :deep(.v-field:hover),
.date-to-filter :deep(.v-field:hover) {
  box-shadow: 0 2px 8px rgba(var(--v-theme-primary), 0.1);
}

.search-field :deep(.v-field--focused),
.user-filter :deep(.v-field--focused),
.action-filter :deep(.v-field--focused),
.date-from-filter :deep(.v-field--focused),
.date-to-filter :deep(.v-field--focused) {
  box-shadow: 0 2px 12px rgba(var(--v-theme-primary), 0.2);
}

.clear-filters-btn {
  text-transform: none;
  font-weight: 600;
  border-radius: 8px;
  transition: all 0.3s ease;
}

.clear-filters-btn:hover {
  background-color: rgba(var(--v-theme-error), 0.08);
  color: rgb(var(--v-theme-error));
}

.active-filters {
  background: rgba(var(--v-theme-surface-variant), 0.3);
  border-radius: 8px;
  padding: 12px;
}

.active-filters .v-chip {
  font-weight: 600;
  transition: all 0.3s ease;
}

.active-filters .v-chip:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

/* Activity Card */
.activity-card {
  background: rgb(var(--v-theme-surface));
  border: 1px solid rgba(var(--v-border-color), 0.12);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
  overflow: hidden;
}

.v-theme--dark .activity-card {
  background: rgb(var(--v-theme-surface));
  border: 1px solid rgba(var(--v-border-color), 0.3);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(20px);
}

/* Table Header */
.table-header {
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
}

.table-header .v-card-title {
  background: transparent;
}



.refresh-btn {
  color: rgb(var(--v-theme-primary));
  transition: all 0.3s ease;
}

.refresh-btn:hover {
  background-color: rgba(var(--v-theme-primary), 0.1);
  color: rgb(var(--v-theme-primary));
  transform: rotate(180deg);
}

/* Table Styles */
.table-container {
  background: transparent;
}

.activity-table {
  background: transparent !important;
}

.activity-table :deep(.v-data-table-header) {
  background-color: rgba(var(--v-theme-surface-variant), 0.4) !important;
}

.activity-table :deep(.v-data-table-header th) {
  font-weight: 600 !important;
  color: rgb(var(--v-theme-on-surface)) !important;
  font-size: 0.875rem !important;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 2px solid rgba(var(--v-theme-primary), 0.1) !important;
  padding: 16px !important;
}

.activity-table :deep(tbody tr) {
  transition: all 0.3s ease;
  border-bottom: 1px solid rgba(var(--v-border-color), 0.08) !important;
}

.activity-table :deep(tbody tr:hover) {
  background-color: rgba(var(--v-theme-primary), 0.04) !important;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(var(--v-theme-primary), 0.1);
}

.v-theme--dark .activity-table :deep(tbody tr:hover) {
  background-color: rgba(var(--v-theme-primary), 0.08) !important;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.activity-table :deep(td) {
  border-bottom: 1px solid rgba(var(--v-border-color), 0.08) !important;
  font-size: 0.9rem;
  padding: 16px !important;
}

/* User Cell */
.user-cell {
  min-height: 60px;
}

.user-avatar-wrapper {
  position: relative;
}

.user-avatar {
  box-shadow: 0 4px 12px rgba(var(--v-theme-primary), 0.15);
  transition: all 0.3s ease;
}

.user-avatar-bg {
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, rgb(var(--v-theme-primary)) 0%, rgb(var(--v-theme-secondary)) 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.user-initial {
  color: white;
  font-weight: 700;
  font-size: 0.875rem;
}

.user-name {
  color: rgb(var(--v-theme-on-surface));
  line-height: 1.2;
}

.user-email {
  color: rgb(var(--v-theme-primary));
  line-height: 1.2;
}

/* Action Chip */
.action-chip {
  font-size: 0.75rem !important;
  height: 28px !important;
  border-radius: 8px !important;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
}

.action-chip:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

/* Timestamp Cell */
.timestamp-cell {
  min-width: 120px;
}

.timestamp-primary {
  color: rgb(var(--v-theme-on-surface));
  line-height: 1.2;
}

.timestamp-secondary {
  color: rgb(var(--v-theme-primary));
  line-height: 1.2;
}

/* Details Cell */
.details-btn {
  color: rgb(var(--v-theme-primary));
  transition: all 0.3s ease;
}

.details-btn:hover {
  background-color: rgba(var(--v-theme-primary), 0.1);
  color: rgb(var(--v-theme-primary));
  transform: scale(1.1);
}

/* Loading & No Data States */
.loading-container,
.no-data-container {
  min-height: 200px;
}

/* Enhanced Dialog Styles */
.details-dialog :deep(.v-overlay__content) {
  border-radius: 20px;
  max-height: 90vh;
  overflow-y: auto;
}

.dialog-card {
  background: rgb(var(--v-theme-surface));
  border: none;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
  overflow: hidden;
  transition: all 0.3s ease;
}

.v-theme--dark .dialog-card {
  background: rgb(var(--v-theme-surface));
  border: 1px solid rgba(var(--v-border-color), 0.3);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(30px);
}

/* Enhanced Header */
.dialog-header-enhanced {
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
  color: #0f172a;
  position: relative;
}

.v-theme--dark .dialog-header-enhanced {
  background: #1e293b;
  border-bottom: 1px solid #334155;
  color: #f8fafc;
}

.dialog-icon-wrapper-enhanced {
  width: 44px;
  height: 44px;
  background: rgba(14, 124, 140, 0.1);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dialog-title-enhanced {
  color: #0f172a !important;
  font-size: 1.15rem;
}

.v-theme--dark .dialog-title-enhanced {
  color: #f8fafc !important;
}

.dialog-subtitle-enhanced {
  color: #64748b !important;
}

.v-theme--dark .dialog-subtitle-enhanced {
  color: #94a3b8 !important;
}

.close-btn-enhanced {
  color: #64748b !important;
  transition: all 0.2s ease;
}

.v-theme--dark .close-btn-enhanced {
  color: #cbd5e1 !important;
}

.close-btn-enhanced:hover {
  background: rgba(0, 0, 0, 0.05) !important;
}

/* Summary Section */
.summary-section {
  background: #f8fafc;
  border-radius: 14px;
  padding: 16px 20px;
  border: 1px solid #e2e8f0;
}

.v-theme--dark .summary-section {
  background: #1e293b;
  border-color: #334155;
}

.summary-avatar-wrapper {
  position: relative;
}

.summary-avatar {
  box-shadow: 0 4px 12px rgba(14, 124, 140, 0.2);
  transition: all 0.3s ease;
}

.summary-avatar-bg {
  width: 100%;
  height: 100%;
  background: #0e7c8c;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.summary-initial {
  color: white;
  font-weight: 700;
  font-size: 1.2rem;
}

.summary-user-name {
  color: rgb(var(--v-theme-on-surface));
}

.summary-user-email {
  color: rgb(var(--v-theme-primary));
}

.summary-action-chip {
  text-transform: uppercase;
  letter-spacing: 0.5px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.timestamp-chip {
  font-size: 0.75rem;
}

/* Detail Cards */
.detail-card {
  background: rgb(var(--v-theme-surface));
  border-radius: 12px;
  overflow: hidden;
  transition: all 0.3s ease;
}

.detail-card:hover {
  box-shadow: 0 4px 20px rgba(var(--v-theme-primary), 0.1);
  transform: translateY(-2px);
}

.detail-card-header {
  background: rgba(var(--v-theme-surface-variant), 0.3);
  border-bottom: 1px solid rgba(var(--v-border-color), 0.08);
  color: rgb(var(--v-theme-on-surface));
  font-weight: 600;
}

.detail-divider {
  border-color: rgba(var(--v-border-color), 0.08) !important;
  margin: 0;
}

.action-path {
  background: rgba(var(--v-theme-primary), 0.08);
  padding: 8px 12px;
  border-radius: 6px;
  font-family: 'Courier New', monospace;
  font-size: 0.875rem;
  color: rgb(var(--v-theme-primary));
  border: 1px solid rgba(var(--v-theme-primary), 0.2);
  display: inline-block;
}

/* Details Grid */
.details-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.detail-item-enhanced {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 16px;
  background: rgba(var(--v-theme-surface-variant), 0.2);
  border-radius: 8px;
  border: 1px solid rgba(var(--v-border-color), 0.06);
  transition: all 0.3s ease;
}

.detail-item-enhanced:hover {
  background: rgba(var(--v-theme-primary), 0.04);
  box-shadow: 0 2px 8px rgba(var(--v-theme-primary), 0.1);
  transform: translateY(-1px);
}

.detail-label {
  display: flex;
  align-items: center;
  gap: 8px;
}

.label-icon {
  color: rgb(var(--v-theme-primary));
}

.label-text {
  font-weight: 600;
  color: rgb(var(--v-theme-on-surface));
  font-size: 0.875rem;
  text-transform: capitalize;
}

.detail-value-wrapper {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.detail-value-content {
  flex: 1;
  font-family: 'Courier New', monospace;
  font-size: 0.875rem;
  background: rgba(var(--v-theme-surface), 0.5);
  padding: 8px 12px;
  border-radius: 6px;
  border: 1px solid rgba(var(--v-border-color), 0.1);
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 200px;
  overflow-y: auto;
  line-height: 1.4;
}

.value-type-chip {
  font-weight: 600;
  text-transform: uppercase;
  font-size: 0.625rem;
  letter-spacing: 0.5px;
  flex-shrink: 0;
}

.detail-divider-enhanced {
  border-color: rgba(var(--v-border-color), 0.12) !important;
  margin: 0 !important;
  border-width: 1px;
  border-style: dashed;
}

/* Copy Button */
.copy-btn {
  transition: all 0.3s ease;
}

.copy-btn:hover {
  background-color: rgba(var(--v-theme-primary), 0.08);
  transform: scale(1.05);
}

/* Empty State */
.empty-state {
  padding: 32px;
  background: rgba(var(--v-theme-surface-variant), 0.1);
  border-radius: 12px;
  border: 2px dashed rgba(var(--v-border-color), 0.2);
}

/* Enhanced Dialog Actions */
.dialog-actions-enhanced {
  background: linear-gradient(135deg, rgba(var(--v-theme-surface-variant), 0.4) 0%, rgba(var(--v-theme-surface-variant), 0.2) 100%);
  border-top: 1px solid rgba(var(--v-border-color), 0.12);
}

.action-btn {
  text-transform: none;
  font-weight: 600;
  border-radius: 10px;
  transition: all 0.3s ease;
  min-width: 120px;
}

.action-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

/* Scrollbar Styles */
.detail-value-content::-webkit-scrollbar {
  width: 6px;
}

.detail-value-content::-webkit-scrollbar-track {
  background: rgba(var(--v-theme-surface), 0.3);
  border-radius: 3px;
}

.detail-value-content::-webkit-scrollbar-thumb {
  background: rgba(var(--v-theme-primary), 0.3);
  border-radius: 3px;
}

.detail-value-content::-webkit-scrollbar-thumb:hover {
  background: rgba(var(--v-theme-primary), 0.5);
}

/* Dark mode adjustments */
.v-theme--dark .summary-section {
  background: linear-gradient(135deg, rgba(var(--v-theme-primary), 0.08) 0%, rgba(var(--v-theme-secondary), 0.08) 100%);
  border: 1px solid rgba(var(--v-border-color), 0.15);
}

.v-theme--dark .detail-item-enhanced {
  background: rgba(var(--v-theme-surface-variant), 0.1);
  border: 1px solid rgba(var(--v-border-color), 0.1);
}

.v-theme--dark .detail-value-content {
  background: rgba(var(--v-theme-surface), 0.3);
  border: 1px solid rgba(var(--v-border-color), 0.15);
}

.v-theme--dark .dialog-actions-enhanced {
  background: linear-gradient(135deg, rgba(var(--v-theme-surface-variant), 0.2) 0%, rgba(var(--v-theme-surface-variant), 0.1) 100%);
}

/* Details List */
.details-list {
  max-height: 400px;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: rgba(var(--v-theme-primary), 0.3) transparent;
}

.details-list::-webkit-scrollbar {
  width: 6px;
}

.details-list::-webkit-scrollbar-track {
  background: transparent;
}

.details-list::-webkit-scrollbar-thumb {
  background: rgba(var(--v-theme-primary), 0.3);
  border-radius: 3px;
}

.detail-item {
  transition: all 0.3s ease;
  border-radius: 8px;
  margin: 4px 8px;
}

.detail-item:hover {
  background-color: rgba(var(--v-theme-primary), 0.04);
}

.detail-key {
  font-size: 0.875rem;
}

.detail-value {
  font-size: 0.875rem;
  word-break: break-all;
  white-space: pre-wrap;
  max-height: 120px;
  overflow-y: auto;
}

.dialog-actions {
  background: rgba(var(--v-theme-surface-variant), 0.2);
  border-top: 1px solid rgba(var(--v-border-color), 0.12);
}

.close-dialog-btn {
  text-transform: none;
  font-weight: 600;
  border-radius: 8px;
}

/* Responsive Design */
@media (max-width: 960px) {
  .header-title {
    font-size: 1.5rem;
  }

  .header-subtitle {
    font-size: 0.95rem;
  }
}

@media (max-width: 768px) {
  .activity-log-container {
    padding: 1rem !important;
  }

  .header-card {
    padding: 20px;
    margin-bottom: 1.5rem;
  }

  .header-avatar {
    width: 48px !important;
    height: 48px !important;
    margin-right: 12px !important;
  }

  .header-title {
    font-size: 1.5rem !important;
  }
  
  .table-header .v-card-title {
    padding: 1rem !important;
  }
  
  .activity-table :deep(th),
  .activity-table :deep(td) {
    padding: 12px 8px !important;
  }
  
  .user-cell {
    min-height: 50px;
  }
  
  .user-avatar {
    width: 32px !important;
    height: 32px !important;
  }
  
  .user-name {
    font-size: 0.875rem;
  }
  
  .user-email {
    font-size: 0.75rem;
  }
  
  .action-chip {
    font-size: 0.7rem !important;
    height: 24px !important;
  }
  
  .timestamp-cell {
    min-width: 100px;
  }
  
  .timestamp-primary,
  .timestamp-secondary {
    font-size: 0.8rem;
  }
  
  .dialog-card {
    margin: 1rem;
  }
  
  .dialog-header,
  .dialog-actions {
    padding: 1rem !important;
  }
  
  .detail-item {
    padding: 12px 1rem !important;
  }
}

@media (max-width: 480px) {
  .activity-log-container {
    padding: 0.75rem !important;
  }
  
  .header-avatar {
    width: 44px !important;
    height: 44px !important;
  }
  
  .header-title {
    font-size: 1.25rem !important;
  }
  
  .activity-table :deep(th),
  .activity-table :deep(td) {
    padding: 8px 4px !important;
    font-size: 0.8rem !important;
  }
  
  .user-name {
    font-size: 0.8rem;
  }
  
  .user-email {
    font-size: 0.7rem;
  }
  
  .action-chip {
    font-size: 0.65rem !important;
    height: 22px !important;
  }
  
  .timestamp-primary,
  .timestamp-secondary {
    font-size: 0.75rem;
  }
  
  .details-btn {
    width: 32px !important;
    height: 32px !important;
  }
  
  .dialog-card {
    margin: 0.5rem;
    max-height: none;
  }
  
  .details-grid {
    grid-template-columns: 1fr;
    gap: 0.75rem;
  }
  
  .detail-item-wrapper {
    padding: 0.75rem;
  }
  
  .value-container {
    padding: 0.5rem;
    font-size: 0.8rem;
  }
}

/* Dark mode specific adjustments */
.v-theme--dark .dialog-icon-wrapper {
  background: rgba(var(--v-theme-primary), 0.2);
}

.v-theme--dark .action-chip {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.v-theme--dark .details-list::-webkit-scrollbar-thumb {
  background: rgba(var(--v-theme-primary), 0.5);
}

/* Dark-mode header card handled above in main header-card block */

/* Animations */
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.activity-card {
  animation: fadeIn 0.6s ease-out;
}

.detail-item {
  animation: fadeIn 0.3s ease-out;
}
</style>