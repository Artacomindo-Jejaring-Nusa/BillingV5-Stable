<template>
  <v-container fluid class="pa-4 pa-md-6 bg-grey-lighten-4 fill-height align-start">
    <div class="w-100">
      <!-- Header Section with Gradient Background -->
      <div class="header-card mb-6">
        <div class="header-section">
          <div class="header-content">
            <div class="d-flex align-center">
              <v-avatar class="me-4 elevation-4" color="rgba(255,255,255,0.2)" size="72">
                <v-icon color="white" size="36">mdi-chart-line</v-icon>
              </v-avatar>
              <div>
                <h1 class="text-h4 font-weight-bold text-white mb-1">Laporan Keuangan</h1>
                <p class="header-subtitle mb-0">
                  Monitoring detail pendapatan, tagihan & pajak
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Navigation Tabs -->
      <v-tabs v-model="activeTab" color="primary" class="mb-4 bg-white rounded-xl modern-card border-none" density="comfortable">
        <v-tab :value="0" class="text-none font-weight-bold">
          <v-icon start>mdi-chart-line</v-icon>
          Ringkasan Finansial & Tagihan
        </v-tab>
        <v-tab :value="1" class="text-none font-weight-bold">
          <v-icon start color="indigo">mdi-brain</v-icon>
          Analisis AI & Pola Pembayaran
          <v-chip size="x-small" color="indigo" variant="flat" class="ms-2 font-weight-bold">ML Model</v-chip>
        </v-tab>
      </v-tabs>

      <!-- Filter Controls -->
      <v-card class="mb-6 modern-card border-none" elevation="0" rounded="xl">
        <v-card-text class="pa-3 pa-md-4">
          <v-row align="center" class="ga-3" no-gutters>
            <template v-if="activeTab === 0">
              <v-col cols="12" sm="6" md="auto" class="flex-grow-1">
                <v-menu v-model="menuStart" :close-on-content-click="false" location="bottom start" offset="8">
                  <template v-slot:activator="{ props }">
                    <v-text-field
                      :model-value="formatDate(startDate)"
                      label="Tanggal Awal"
                      prepend-inner-icon="mdi-calendar"
                      readonly
                      v-bind="props"
                      variant="outlined"
                      density="compact"
                      hide-details
                      class="bg-white"
                      color="primary"
                    ></v-text-field>
                  </template>
                  <v-date-picker v-model="startDate" @update:model-value="menuStart = false" color="primary"></v-date-picker>
                </v-menu>
              </v-col>
              <v-col cols="12" sm="6" md="auto" class="flex-grow-1">
                <v-menu v-model="menuEnd" :close-on-content-click="false" location="bottom start" offset="8">
                  <template v-slot:activator="{ props }">
                    <v-text-field
                      :model-value="formatDate(endDate)"
                      label="Tanggal Akhir"
                      prepend-inner-icon="mdi-calendar"
                      readonly
                      v-bind="props"
                      variant="outlined"
                      density="compact"
                      hide-details
                      class="bg-white"
                      color="primary"
                    ></v-text-field>
                  </template>
                  <v-date-picker v-model="endDate" @update:model-value="menuEnd = false" color="primary"></v-date-picker>
                </v-menu>
              </v-col>
            </template>

            <v-col cols="12" sm="4" md="auto" class="flex-grow-1">
              <v-select
                v-model="selectedLocation"
                :items="locations"
                label="Wilayah"
                prepend-inner-icon="mdi-map-marker"
                variant="outlined"
                density="compact"
                hide-details
                clearable
                class="bg-white"
                color="primary"
              ></v-select>
            </v-col>

            <v-col cols="auto">
              <v-btn
                :color="activeTab === 1 ? 'indigo' : 'primary'"
                @click="activeTab === 1 ? fetchMLInsights() : fetchReport()"
                :loading="activeTab === 1 ? isMLLoading : isReportLoading"
                height="40"
                class="text-none px-6 text-white"
                :prepend-icon="activeTab === 1 ? 'mdi-brain' : 'mdi-filter'"
                rounded="lg"
                elevation="1"
              >
                {{ activeTab === 1 ? 'Analisis Ulang ML' : 'Tampilkan' }}
              </v-btn>
            </v-col>

            <v-col v-if="activeTab === 0" cols="auto">
              <v-btn
                color="green-darken-1"
                @click="exportToExcel"
                :disabled="!reportSummary || reportSummary.total_invoices === 0 || exporting"
                height="40"
                class="text-none px-6 text-white"
                prepend-icon="mdi-microsoft-excel"
                rounded="lg"
                variant="flat"
              >
                Export
                <v-progress-circular v-if="exporting" indeterminate size="20" class="ms-2"></v-progress-circular>
              </v-btn>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>

      <!-- TAB 0: RINGKASAN FINANSIAL & TAGIHAN -->
      <div v-if="activeTab === 0">
        <!-- Loading State -->
        <div v-if="isLoading && !reportSummary" class="text-center pa-10">
          <v-progress-circular indeterminate color="primary" size="64"></v-progress-circular>
          <p class="mt-4 text-grey">Memuat data...</p>
        </div>

        <!-- DASHBOARD CONTENT -->
        <div v-if="reportSummary" class="fade-enter-active">
        
        <!-- 1. RINGKASAN KEUANGAN (Financial Summary) -->
        <div class="mb-4">
          <div class="d-flex align-center mb-2">
            <v-icon color="grey-darken-2" class="me-2">mdi-wallet-outline</v-icon>
            <span class="text-subtitle-1 font-weight-bold text-grey-darken-3">Ringkasan Keuangan</span>
          </div>
          <v-row>
            <!-- 1. TOTAL PEMASUKAN (Cash Flow) -->
            <v-col cols="12" md="4">
              <v-card class="stats-card border-top-success h-100" elevation="0">
                <v-card-text>
                  <div class="text-caption text-grey mb-1">Total Pemasukan</div>
                  <div class="text-h5 font-weight-bold text-success">
                    {{ formatCurrency(reportSummary.financial_summary?.total_pemasukan || 0) }}
                  </div>
                   <div class="text-caption text-grey mt-2">
                    Uang yang diterima sistem
                  </div>
                </v-card-text>
              </v-card>
            </v-col>

            <!-- 2. TOTAL TAGIHAN (Invoiced) -->
             <v-col cols="12" md="4">
              <v-card class="stats-card border-top-primary h-100" elevation="0">
                 <v-card-text>
                  <div class="text-caption text-grey mb-1">Total Tagihan Dicetak</div>
                  <div class="text-h5 font-weight-bold text-primary">
                    {{ formatCurrency(reportSummary.billing_summary?.total_tagihan?.total || 0) }}
                  </div>
                  <div class="text-caption text-grey mt-2">
                    Nilai invoice yang diterbitkan
                  </div>
                </v-card-text>
              </v-card>
            </v-col>

            <!-- 3. BELUM TERBAYAR (Outstanding) -->
            <v-col cols="12" md="4">
              <v-card class="stats-card border-top-warning h-100" elevation="0">
                 <v-card-text>
                  <div class="text-caption text-grey mb-1">Belum Terbayar</div>
                  <div class="text-h5 font-weight-bold text-warning darken-2">
                    {{ formatCurrency((reportSummary.billing_summary?.pending?.total || 0) + (reportSummary.billing_summary?.expired?.total || 0)) }}
                  </div>
                   <div class="text-caption text-grey mt-2">
                    Potensi pendapatan (Pending + Expired)
                  </div>
                </v-card-text>
              </v-card>
            </v-col>
          </v-row>
        </div>

        <!-- 2. RINGKASAN TAGIHAN (Billing Summary) -->
        <div class="mb-4 mt-6">
           <div class="d-flex align-center mb-2">
            <v-icon color="grey-darken-2" class="me-2">mdi-invoice-text-outline</v-icon>
            <span class="text-subtitle-1 font-weight-bold text-grey-darken-3">Ringkasan Tagihan</span>
            <v-chip size="x-small" class="ms-2" variant="flat" color="grey-lighten-3">
              {{ formatDate(startDate) }} - {{ formatDate(endDate) }}
            </v-chip>
          </div>
          
          <v-row>
            <!-- Total Tagihan -->
             <v-col cols="12" sm="6" md="3">
               <v-card class="stats-card h-100" elevation="0">
                 <v-card-text>
                   <div class="text-subtitle-2 text-grey-darken-1 mb-2">Total Tagihan</div>
                   <div class="text-h4 font-weight-black mb-4">{{ reportSummary.billing_summary?.total_tagihan?.count || 0 }}</div>
                   
                   <div class="stats-detail-row">
                      <span>Nominal:</span>
                      <span class="font-weight-medium">{{ formatCurrency(reportSummary.billing_summary?.total_tagihan?.nominal || 0) }}</span>
                   </div>
                   <div class="stats-detail-row text-error">
                      <span>Diskon:</span>
                      <span class="font-weight-medium">{{ formatCurrency(reportSummary.billing_summary?.total_tagihan?.diskon || 0) }}</span>
                   </div>
                   <div class="stats-detail-row">
                      <span>Biaya Pasang:</span>
                      <span class="font-weight-medium">{{ formatCurrency(reportSummary.billing_summary?.total_tagihan?.biaya_pasang || 0) }}</span>
                   </div>
                   <div class="stats-detail-row total mt-2 pt-2 border-t">
                      <span>Total:</span>
                      <span class="font-weight-bold text-grey-darken-3">{{ formatCurrency(reportSummary.billing_summary?.total_tagihan?.total || 0) }}</span>
                   </div>
                 </v-card-text>
               </v-card>
             </v-col>

              <!-- Lunas -->
             <v-col cols="12" sm="6" md="3">
               <v-card class="stats-card h-100" elevation="0">
                 <v-card-text>
                   <div class="text-subtitle-2 text-success mb-2">Lunas</div>
                   <div class="text-h4 font-weight-black text-success mb-4">{{ reportSummary.billing_summary?.lunas?.count || 0 }}</div>
                   
                   <div class="stats-detail-row">
                      <span>Nominal:</span>
                      <span class="font-weight-medium">{{ formatCurrency(reportSummary.billing_summary?.lunas?.nominal || 0) }}</span>
                   </div>
                   <div class="stats-detail-row text-error">
                      <span>Diskon:</span>
                      <span class="font-weight-medium">{{ formatCurrency(reportSummary.billing_summary?.lunas?.diskon || 0) }}</span>
                   </div>
                    <div class="stats-detail-row">
                      <span>Biaya Pasang:</span>
                      <span class="font-weight-medium">{{ formatCurrency(reportSummary.billing_summary?.lunas?.biaya_pasang || 0) }}</span>
                   </div>
                   <div class="stats-detail-row total mt-2 pt-2 border-t text-success">
                      <span>Total:</span>
                      <span class="font-weight-bold">{{ formatCurrency(reportSummary.billing_summary?.lunas?.total || 0) }}</span>
                   </div>
                 </v-card-text>
               </v-card>
             </v-col>

              <!-- Pending -->
             <v-col cols="12" sm="6" md="3">
               <v-card class="stats-card h-100" elevation="0">
                 <v-card-text>
                   <div class="text-subtitle-2 text-warning darken-2 mb-2">Pending</div>
                   <div class="text-h4 font-weight-black text-warning darken-2 mb-4">{{ reportSummary.billing_summary?.pending?.count || 0 }}</div>
                   
                   <div class="stats-detail-row">
                      <span>Nominal:</span>
                      <span class="font-weight-medium">{{ formatCurrency(reportSummary.billing_summary?.pending?.nominal || 0) }}</span>
                   </div>
                   <div class="stats-detail-row text-error">
                      <span>Diskon:</span>
                      <span class="font-weight-medium">{{ formatCurrency(reportSummary.billing_summary?.pending?.diskon || 0) }}</span>
                   </div>
                    <div class="stats-detail-row">
                      <span>Biaya Pasang:</span>
                      <span class="font-weight-medium">{{ formatCurrency(reportSummary.billing_summary?.pending?.biaya_pasang || 0) }}</span>
                   </div>
                   <div class="stats-detail-row total mt-2 pt-2 border-t text-warning darken-2">
                      <span>Total:</span>
                      <span class="font-weight-bold">{{ formatCurrency(reportSummary.billing_summary?.pending?.total || 0) }}</span>
                   </div>
                 </v-card-text>
               </v-card>
             </v-col>

             <!-- Expired -->
             <v-col cols="12" sm="6" md="3">
               <v-card class="stats-card h-100" elevation="0">
                 <v-card-text>
                   <div class="text-subtitle-2 text-error mb-2">Expired</div>
                   <div class="text-h4 font-weight-black text-error mb-4">{{ reportSummary.billing_summary?.expired?.count || 0 }}</div>
                   
                   <div class="stats-detail-row">
                      <span>Nominal:</span>
                      <span class="font-weight-medium">{{ formatCurrency(reportSummary.billing_summary?.expired?.nominal || 0) }}</span>
                   </div>
                   <div class="stats-detail-row text-error">
                      <span>Diskon:</span>
                      <span class="font-weight-medium">{{ formatCurrency(reportSummary.billing_summary?.expired?.diskon || 0) }}</span>
                   </div>
                    <div class="stats-detail-row">
                      <span>Biaya Pasang:</span>
                      <span class="font-weight-medium">{{ formatCurrency(reportSummary.billing_summary?.expired?.biaya_pasang || 0) }}</span>
                   </div>
                   <div class="stats-detail-row total mt-2 pt-2 border-t text-error">
                      <span>Total:</span>
                      <span class="font-weight-bold">{{ formatCurrency(reportSummary.billing_summary?.expired?.total || 0) }}</span>
                   </div>
                 </v-card-text>
               </v-card>
             </v-col>
          </v-row>
        </div>

        <!-- 3. INFORMASI PAJAK & METODE PEMBAYARAN -->
        <v-row class="mb-4">
          <!-- Informasi Pajak -->
          <v-col cols="12" md="8">
             <v-card class="stats-card h-100" elevation="0">
              <v-card-title class="text-subtitle-1 font-weight-bold text-grey-darken-3 px-4 pt-4 pb-2">
                Informasi Pajak
              </v-card-title>
              <v-card-text class="pa-0">
                <v-table density="compact" class="text-caption">
                  <thead>
                    <tr>
                      <th class="text-left font-weight-bold">Status</th>
                      <th class="text-right">PPN</th>
                      <th class="text-right">BHP</th>
                      <th class="text-right">USO</th>
                      <th class="text-right font-weight-bold">Total Pajak</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr>
                      <td class="font-weight-medium">Lunas</td>
                      <td class="text-right text-success">{{ formatCurrency(reportSummary.tax_summary?.lunas?.ppn || 0) }}</td>
                      <td class="text-right">{{ formatCurrency(reportSummary.tax_summary?.lunas?.bhp || 0) }}</td>
                      <td class="text-right">{{ formatCurrency(reportSummary.tax_summary?.lunas?.uso || 0) }}</td>
                      <td class="text-right font-weight-bold text-success">{{ formatCurrency(reportSummary.tax_summary?.lunas?.total_pajak || 0) }}</td>
                    </tr>
                    <tr>
                      <td class="font-weight-medium">Pending</td>
                       <td class="text-right text-warning">{{ formatCurrency(reportSummary.tax_summary?.pending?.ppn || 0) }}</td>
                      <td class="text-right">{{ formatCurrency(reportSummary.tax_summary?.pending?.bhp || 0) }}</td>
                      <td class="text-right">{{ formatCurrency(reportSummary.tax_summary?.pending?.uso || 0) }}</td>
                      <td class="text-right font-weight-bold text-warning">{{ formatCurrency(reportSummary.tax_summary?.pending?.total_pajak || 0) }}</td>
                    </tr>
                     <tr>
                      <td class="font-weight-medium">Expired</td>
                       <td class="text-right text-error">{{ formatCurrency(reportSummary.tax_summary?.expired?.ppn || 0) }}</td>
                      <td class="text-right">{{ formatCurrency(reportSummary.tax_summary?.expired?.bhp || 0) }}</td>
                      <td class="text-right">{{ formatCurrency(reportSummary.tax_summary?.expired?.uso || 0) }}</td>
                      <td class="text-right font-weight-bold text-error">{{ formatCurrency(reportSummary.tax_summary?.expired?.total_pajak || 0) }}</td>
                    </tr>
                     <tr class="bg-grey-lighten-4 font-weight-bold">
                      <td>TOTAL</td>
                       <td class="text-right">{{ formatCurrency(reportSummary.tax_summary?.total?.ppn || 0) }}</td>
                      <td class="text-right">{{ formatCurrency(reportSummary.tax_summary?.total?.bhp || 0) }}</td>
                      <td class="text-right">{{ formatCurrency(reportSummary.tax_summary?.total?.uso || 0) }}</td>
                      <td class="text-right">{{ formatCurrency(reportSummary.tax_summary?.total?.total_pajak || 0) }}</td>
                    </tr>
                  </tbody>
                </v-table>
                 <div class="px-4 py-2 text-caption text-grey">Harga sudah termasuk pajak</div>
              </v-card-text>
            </v-card>
          </v-col>

          <!-- Metode Pembayaran -->
          <v-col cols="12" md="4">
            <v-card class="stats-card h-100" elevation="0">
              <v-card-title class="text-subtitle-1 font-weight-bold text-grey-darken-3 px-4 pt-4 pb-2">
                Ringkasan per Metode Bayar
              </v-card-title>
              <v-card-text class="pa-0 overflow-y-auto" style="max-height: 250px;">
                <v-list density="compact">
                  <v-list-item v-for="method in reportSummary.payment_methods" :key="method.method" class="border-bottom">
                    <template v-slot:prepend>
                      <v-avatar size="32" color="grey-lighten-4" class="me-3">
                         <span class="text-caption font-weight-bold">{{ method.count }}x</span>
                      </v-avatar>
                    </template>
                    <v-list-item-title class="text-caption font-weight-bold">
                      {{ method.method }}
                    </v-list-item-title>
                    <v-list-item-subtitle class="text-caption">
                       Total: {{ formatCurrency(method.total_amount) }}
                    </v-list-item-subtitle>
                  </v-list-item>
                   <v-list-item v-if="reportSummary.payment_methods.length === 0">
                      <v-list-item-title class="text-caption text-center text-grey">
                        Belum ada data pembayaran
                      </v-list-item-title>
                   </v-list-item>
                </v-list>
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>

         <!-- 4. DATA TABLE (Keeping existing table but styled cleaner) -->
         <v-card elevation="0" rounded="lg" class="modern-card mt-6">
            <v-card-title class="pa-4 text-subtitle-1 font-weight-bold text-grey-darken-3 border-b">
              Data Tagihan (Detail)
            </v-card-title>
             <v-data-table-server
              v-model:page="currentPage"
              v-model:items-per-page="itemsPerPage"
              :headers="headers"
              :items="invoiceDetails"
              :items-length="reportSummary?.total_invoices || 0"
              :loading="isDetailsLoading"
              @update:options="handleTableOptionsUpdate"
              class="text-caption"
              density="compact"
            >
             <template v-slot:item.total_harga="{ item }">
                <span class="font-weight-medium">{{ formatCurrency(item.total_harga) }}</span>
             </template>
             <template v-slot:item.tgl_lunas="{ item }">
                {{ item.tgl_lunas ? new Date(item.tgl_lunas).toLocaleString('id-ID') : '-' }}
             </template>
             <template v-slot:item.brand="{ item }">
                <v-chip size="x-small" variant="tonal" color="primary">{{ item.brand }}</v-chip>
             </template>
            </v-data-table-server>
         </v-card>

        </div>
        <!-- Empty State -->
        <div v-else-if="!isLoading && !reportSummary" class="empty-state text-center pa-10">
          <v-icon size="64" color="grey-lighten-2" class="mb-4">mdi-chart-box-outline</v-icon>
          <p class="text-grey">Silakan klik "Tampilkan" untuk melihat laporan.</p>
        </div>
      </div>

      <!-- TAB 1: ANALISIS AI & POLA PEMBAYARAN (MACHINE LEARNING) -->
      <div v-else-if="activeTab === 1">
        <!-- Loading State for ML -->
        <div v-if="isMLLoading && !mlInsights" class="text-center pa-10">
          <v-progress-circular indeterminate color="indigo" size="64"></v-progress-circular>
          <p class="mt-4 text-grey">Memproses model Machine Learning & Clustering K-Means...</p>
        </div>

        <div v-else-if="mlInsights">
          <!-- 1. Cycle & Operational Info Banner -->
          <v-card class="mb-6 modern-card border-none bg-indigo-lighten-5 text-indigo-darken-4" elevation="0" rounded="xl">
            <v-card-text class="pa-4">
              <div class="d-flex flex-wrap align-center justify-space-between ga-3">
                <div class="d-flex align-center">
                  <v-avatar color="indigo" size="44" class="me-3 text-white elevation-2">
                    <v-icon size="24">mdi-brain</v-icon>
                  </v-avatar>
                  <div>
                    <div class="text-subtitle-1 font-weight-bold">
                      Analisis Pola Pembayaran & Prediksi Cash Flow (Model K-Means)
                    </div>
                    <div class="text-caption text-grey-darken-2">
                      Siklus Penagihan: Invoice Terbit <strong>Tgl 26/27</strong> | Jatuh Tempo <strong>Tgl 01</strong> | Batas Masa Tenggang <strong>Tgl 10</strong> | Auto-Isolir <strong>Tgl 11 00:00 WIB</strong>
                    </div>
                  </div>
                </div>
                <div class="d-flex align-center ga-2">
                  <v-chip color="indigo" variant="flat" size="small" class="font-weight-bold">
                    {{ mlInsights.cycle_info?.total_active_subscribers || 0 }} Pelanggan Teranalisis
                  </v-chip>
                  <v-btn
                    color="indigo"
                    variant="tonal"
                    size="small"
                    prepend-icon="mdi-refresh"
                    :loading="isMLLoading"
                    @click="fetchMLInsights"
                    class="text-none"
                  >
                    Refresh Model
                  </v-btn>
                </div>
              </div>
            </v-card-text>
          </v-card>

          <!-- 2. Ringkasan Proyeksi ML (4 Metric Cards) -->
          <v-row class="mb-4">
            <!-- Total Proyeksi Tagihan -->
            <v-col cols="12" sm="6" md="3">
              <v-card class="stats-card border-top-primary h-100" elevation="0">
                <v-card-text>
                  <div class="text-caption text-grey mb-1">Total Proyeksi Billing</div>
                  <div class="text-h5 font-weight-bold text-primary">
                    {{ formatCurrency(mlInsights.summary?.projected_billing || 0) }}
                  </div>
                  <div class="text-caption text-grey mt-2">
                    Target penagihan siklus bulan ini
                  </div>
                </v-card-text>
              </v-card>
            </v-col>

            <!-- Estimasi Terkumpul Awal (Tgl 26 - 5) -->
            <v-col cols="12" sm="6" md="3">
              <v-card class="stats-card border-top-success h-100" elevation="0">
                <v-card-text>
                  <div class="text-caption text-grey mb-1">Puncak Awal (s/d Tgl 5)</div>
                  <div class="text-h5 font-weight-bold text-success">
                    {{ formatCurrency(mlInsights.summary?.projected_early_collection || 0) }}
                  </div>
                  <div class="text-caption text-grey mt-2">
                    Estimasi masuk pasca terbit & awal bulan
                  </div>
                </v-card-text>
              </v-card>
            </v-col>

            <!-- Masa Tenggang (Tgl 6 - 10) -->
            <v-col cols="12" sm="6" md="3">
              <v-card class="stats-card border-top-warning h-100" elevation="0">
                <v-card-text>
                  <div class="text-caption text-grey mb-1">Masa Tenggang (Tgl 6 - 10)</div>
                  <div class="text-h5 font-weight-bold text-amber-darken-2">
                    {{ formatCurrency(mlInsights.summary?.projected_grace_collection || 0) }}
                  </div>
                  <div class="text-caption text-grey mt-2">
                    Sebelum isolir otomatis tgl 11
                  </div>
                </v-card-text>
              </v-card>
            </v-col>

            <!-- Potensi Berisiko / Rawan Isolir -->
            <v-col cols="12" sm="6" md="3">
              <v-card class="stats-card border-top-error h-100" elevation="0">
                <v-card-text>
                  <div class="text-caption text-grey mb-1">Pendapatan Berisiko (Tgl 11+)</div>
                  <div class="text-h5 font-weight-bold text-error">
                    {{ formatCurrency(mlInsights.summary?.projected_at_risk || 0) }}
                  </div>
                  <div class="text-caption text-grey mt-2">
                    Tingkat Recovery: <strong>{{ mlInsights.summary?.projected_recovery_rate || 0 }}%</strong>
                  </div>
                </v-card-text>
              </v-card>
            </v-col>
          </v-row>

          <!-- 3. Fase Proyeksi Cash Flow & Timeline Siklus 26/27 -->
          <div class="mb-6">
            <div class="d-flex align-center mb-3">
              <v-icon color="grey-darken-2" class="me-2">mdi-chart-timeline-variant-shimmer</v-icon>
              <span class="text-subtitle-1 font-weight-bold text-grey-darken-3">
                Proyeksi Cash Flow Pembayaran per Fase Siklus 26/27
              </span>
            </div>
            <v-row>
              <v-col
                v-for="(phase, idx) in mlInsights.cash_flow_forecast"
                :key="idx"
                cols="12"
                sm="6"
                md="3"
              >
                <v-card class="stats-card h-100" elevation="0">
                  <v-card-text>
                    <div class="d-flex justify-space-between align-center mb-1">
                      <span class="text-caption font-weight-bold text-grey-darken-2">{{ phase.phase }}</span>
                      <v-chip size="x-small" variant="flat" color="grey-lighten-3" class="font-weight-medium">
                        {{ phase.target_days }}
                      </v-chip>
                    </div>
                    <div class="text-h6 font-weight-black my-1">
                      {{ formatCurrency(phase.estimated_amount) }}
                    </div>
                    <div class="d-flex align-center justify-space-between text-caption text-grey mb-2">
                      <span>{{ phase.description }}</span>
                      <span class="font-weight-bold text-grey-darken-3">{{ phase.percentage }}%</span>
                    </div>
                    <v-progress-linear
                      :model-value="phase.percentage"
                      :color="idx === 0 ? 'teal' : idx === 1 ? 'primary' : idx === 2 ? 'amber-darken-2' : 'error'"
                      height="6"
                      rounded
                    ></v-progress-linear>
                  </v-card-text>
                </v-card>
              </v-col>
            </v-row>
          </div>

          <!-- 4. Segmentasi & Klaster Perilaku Pembayaran (K-Means) -->
          <div class="mb-6">
            <div class="d-flex align-center justify-space-between mb-3">
              <div class="d-flex align-center">
                <v-icon color="grey-darken-2" class="me-2">mdi-account-group-outline</v-icon>
                <span class="text-subtitle-1 font-weight-bold text-grey-darken-3">
                  Klaster Perilaku Pelanggan (Segmentasi ML K-Means)
                </span>
              </div>
              <div class="text-caption text-grey">
                Klik kartu klaster untuk memfilter daftar pelanggan di bawah
              </div>
            </div>
            <v-row>
              <v-col
                v-for="cluster in mlInsights.clusters_summary"
                :key="cluster.key"
                cols="12"
                sm="6"
                md="3"
              >
                <v-card
                  class="stats-card h-100 cursor-pointer"
                  :class="{ 'cluster-active-border': mlClusterFilter === cluster.key }"
                  :style="mlClusterFilter === cluster.key ? `border: 2px solid ${cluster.color} !important;` : ''"
                  elevation="0"
                  @click="mlClusterFilter = mlClusterFilter === cluster.key ? 'all' : cluster.key"
                >
                  <v-card-text>
                    <div class="d-flex align-center justify-space-between mb-2">
                      <v-chip size="x-small" :color="cluster.color" variant="flat" class="text-white font-weight-bold">
                        {{ cluster.risk_level }} Risk
                      </v-chip>
                      <span class="text-caption font-weight-bold" :style="`color: ${cluster.color}`">
                        {{ cluster.percentage }}% User
                      </span>
                    </div>
                    <div class="text-subtitle-2 font-weight-bold text-grey-darken-3 mb-1">
                      {{ cluster.name }}
                    </div>
                    <div class="text-h5 font-weight-black mb-1">
                      {{ cluster.count }} <span class="text-caption font-weight-regular text-grey">pelanggan</span>
                    </div>
                    <div class="text-caption text-grey">
                      Total Nominal: <strong>{{ formatCurrency(cluster.total_nominal) }}</strong>
                    </div>
                  </v-card-text>
                </v-card>
              </v-col>
            </v-row>
          </div>

          <!-- 5. Rekomendasi Tindakan Cerdas (Actionable AI Recommendations) -->
          <v-card class="mb-6 modern-card border-none bg-grey-lighten-4" elevation="0" rounded="xl">
            <v-card-text class="pa-4">
              <div class="d-flex align-center mb-2">
                <v-icon color="indigo" class="me-2">mdi-lightbulb-on-outline</v-icon>
                <span class="text-subtitle-2 font-weight-bold text-grey-darken-3">
                  Rekomendasi Tindakan Otomatis Sistem
                </span>
              </div>
              <ul class="text-caption text-grey-darken-3 ps-4 mb-0">
                <li v-for="(insight, idx) in mlInsights.actionable_insights" :key="idx" class="mb-1">
                  {{ insight }}
                </li>
              </ul>
            </v-card-text>
          </v-card>

          <!-- 6. Tabel Klasifikasi Pelanggan Berdasarkan ML -->
          <v-card elevation="0" rounded="lg" class="modern-card">
            <v-card-title class="pa-4 d-flex flex-wrap align-center justify-space-between border-b ga-3">
              <div class="d-flex align-center">
                <v-icon color="grey-darken-2" class="me-2">mdi-format-list-bulleted-type</v-icon>
                <span class="text-subtitle-1 font-weight-bold text-grey-darken-3">
                  Daftar Klasifikasi Pelanggan & Pola Bayar
                </span>
              </div>
              <div class="d-flex flex-wrap align-center ga-3">
                <v-select
                  v-model="mlClusterFilter"
                  :items="clusterFilterOptions"
                  item-title="title"
                  item-value="value"
                  label="Filter Klaster"
                  variant="outlined"
                  density="compact"
                  hide-details
                  style="min-width: 220px;"
                  color="indigo"
                ></v-select>
                <v-text-field
                  v-model="mlSearch"
                  placeholder="Cari nama / no. telp..."
                  prepend-inner-icon="mdi-magnify"
                  variant="outlined"
                  density="compact"
                  hide-details
                  clearable
                  style="min-width: 220px;"
                  color="indigo"
                ></v-text-field>
              </div>
            </v-card-title>

            <v-data-table
              :headers="mlCustomerHeaders"
              :items="filteredClassifiedCustomers"
              :items-per-page="10"
              class="text-caption"
              density="compact"
            >
              <template v-slot:item.customer_name="{ item }">
                <div class="font-weight-bold">{{ item.customer_name }}</div>
                <div class="text-grey text-caption">{{ item.no_telp || '-' }}</div>
              </template>

              <template v-slot:item.monthly_bill="{ item }">
                <span class="font-weight-medium">{{ formatCurrency(item.monthly_bill) }}</span>
              </template>

              <template v-slot:item.cluster_name="{ item }">
                <v-chip size="x-small" :color="item.cluster_color" variant="tonal" class="font-weight-bold">
                  {{ item.cluster_name }}
                </v-chip>
              </template>

              <template v-slot:item.avg_payment_day="{ item }">
                <span v-if="item.avg_payment_day > 0" class="font-weight-medium">
                  Tgl {{ Math.round(item.avg_payment_day) }}
                </span>
                <span v-else class="text-grey">-</span>
              </template>

              <template v-slot:item.days_to_pay_avg="{ item }">
                <span v-if="item.days_to_pay_avg > 0">
                  {{ item.days_to_pay_avg }} hari
                </span>
                <span v-else class="text-grey">-</span>
              </template>

              <template v-slot:item.risk_score="{ item }">
                <v-chip
                  size="x-small"
                  :color="item.risk_score >= 70 ? 'error' : item.risk_score >= 40 ? 'warning' : 'success'"
                  variant="flat"
                  class="font-weight-bold"
                >
                  {{ item.risk_score }} / 100
                </v-chip>
              </template>

              <template v-slot:item.recommendation="{ item }">
                <span class="text-grey-darken-2">{{ item.recommendation }}</span>
              </template>
            </v-data-table>
          </v-card>
        </div>
        <div v-else class="empty-state text-center pa-10">
          <v-icon size="64" color="grey-lighten-2" class="mb-4">mdi-brain</v-icon>
          <p class="text-grey">Belum ada data analisis Machine Learning. Klik tombol di bawah untuk memproses.</p>
          <v-btn color="indigo" prepend-icon="mdi-play" @click="fetchMLInsights" class="text-none mt-2">
            Mulai Analisis Machine Learning
          </v-btn>
        </div>
      </div>

    </div>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue';
import { useDisplay } from 'vuetify';
import { debounce } from 'lodash-es';
import apiClient from '@/services/api';

// --- ML Interfaces ---
interface MLClusterSummary {
  key: string;
  name: string;
  count: number;
  percentage: number;
  total_nominal: number;
  color: string;
  risk_level: string;
}

interface MLCashFlowPhase {
  phase: string;
  description: string;
  estimated_amount: number;
  percentage: number;
  target_days: string;
  status_class: string;
}

interface MLClassifiedCustomer {
  customer_id: number;
  customer_name: string;
  no_telp: string;
  brand: string;
  monthly_bill: number;
  cluster_key: string;
  cluster_name: string;
  cluster_color: string;
  risk_score: number;
  risk_level: string;
  avg_payment_day: number;
  days_to_pay_avg: number;
  paid_ratio: number;
  recommendation: string;
}

interface MLRevenueInsightResponse {
  status: string;
  cycle_info: {
    invoice_issuance_date: string;
    due_date: string;
    grace_period_cutoff: string;
    total_active_subscribers: number;
  };
  summary: {
    total_customers: number;
    projected_billing: number;
    projected_early_collection: number;
    projected_grace_collection: number;
    projected_at_risk: number;
    projected_recovery_rate: number;
  };
  clusters_summary: MLClusterSummary[];
  cash_flow_forecast: MLCashFlowPhase[];
  classified_customers: MLClassifiedCustomer[];
  actionable_insights: string[];
}

// --- Interfaces ---
interface InvoiceReportItem {
  invoice_number: string;
  pelanggan_nama: string;
  tgl_lunas: string;
  total_harga: number;
  metode: string;
  alamat?: string;
  brand: string;
}

interface BillStat {
  count: number;
  nominal: number;
  diskon: number;
  biaya_pasang: number;
  total: number;
}

interface TaxStat {
  ppn: number;
  bhp: number;
  uso: number;
  total_pajak: number;
}

interface PaymentMethodStat {
  method: string;
  count: number;
  total_amount: number;
  pajak: number;
  diskon: number;
}

interface RevenueReportResponse {
  total_pendapatan: number;
  total_invoices: number;
  financial_summary: {
    total_pemasukan: number;
    total_pengeluaran: number;
    saldo_akhir: number;
  };
  billing_summary: {
    total_tagihan: BillStat;
    lunas: BillStat;
    pending: BillStat;
    expired: BillStat;
  };
  tax_summary: {
    total: TaxStat;
    lunas: TaxStat;
    pending: TaxStat;
    expired: TaxStat;
  };
  payment_methods: PaymentMethodStat[];
  rincian_invoice: InvoiceReportItem[];
}

// --- Data ---
const display = useDisplay();
const activeTab = ref(0);
const startDate = ref(new Date(new Date().getFullYear(), new Date().getMonth(), 1));
const endDate = ref(new Date());
const menuStart = ref(false);
const menuEnd = ref(false);
const isReportLoading = ref(false);
const isDetailsLoading = ref(false);
const isLoading = computed(() => isReportLoading.value || isDetailsLoading.value);

const reportSummary = ref<RevenueReportResponse | null>(null);
const invoiceDetails = ref<InvoiceReportItem[]>([]);

// --- ML State ---
const mlInsights = ref<MLRevenueInsightResponse | null>(null);
const isMLLoading = ref(false);
const mlClusterFilter = ref('all');
const mlSearch = ref('');

const clusterFilterOptions = [
  { title: 'Semua Klaster', value: 'all' },
  { title: '🟢 Early Birds (Tepat Waktu)', value: 'early_ontime' },
  { title: '🔵 Masa Tenggang (2-10)', value: 'grace_period' },
  { title: '🟡 Terlambat Kronis (Rawan Isolir)', value: 'chronic_late' },
  { title: '🔴 At-Risk / Potensi Churn', value: 'at_risk_churn' },
];

const mlCustomerHeaders = [
  { title: 'Pelanggan', key: 'customer_name' },
  { title: 'Brand', key: 'brand' },
  { title: 'Tagihan Bulanan', key: 'monthly_bill', align: 'end' },
  { title: 'Klaster Pola Bayar', key: 'cluster_name' },
  { title: 'Rata-rata Tgl Bayar', key: 'avg_payment_day', align: 'center' },
  { title: 'Waktu Bayar', key: 'days_to_pay_avg', align: 'center' },
  { title: 'Skor Risiko', key: 'risk_score', align: 'center' },
  { title: 'Rekomendasi Tindakan', key: 'recommendation' },
] as const;

const filteredClassifiedCustomers = computed(() => {
  if (!mlInsights.value || !mlInsights.value.classified_customers) return [];
  let list = mlInsights.value.classified_customers;
  if (mlClusterFilter.value !== 'all') {
    list = list.filter(c => c.cluster_key === mlClusterFilter.value);
  }
  if (mlSearch.value) {
    const q = mlSearch.value.toLowerCase();
    list = list.filter(c =>
      c.customer_name.toLowerCase().includes(q) ||
      (c.no_telp && c.no_telp.toLowerCase().includes(q))
    );
  }
  return list;
});

const currentPage = ref(1);
const itemsPerPage = ref(10);
const selectedLocation = ref<string | null>(null);
const locations = ref<string[]>([]);
const selectedBrand = ref<string | null>(null);
const brandOptions = ref<any[]>([]);
const exporting = ref(false);

// --- Table Headers ---
const headers = [
  { title: 'No. Invoice', key: 'invoice_number' },
  { title: 'Pelanggan', key: 'pelanggan_nama' },
  { title: 'Brand', key: 'brand' },
  { title: 'Alamat', key: 'alamat' },
  { title: 'Tgl Bayar', key: 'tgl_lunas' },
  { title: 'Metode', key: 'metode' },
  { title: 'Total', key: 'total_harga', align: 'end' },
] as const;

// --- Methods ---
function formatDate(date: Date): string {
  return date.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' });
}

const formatCurrency = (value: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency', currency: 'IDR', minimumFractionDigits: 0,
  }).format(value);
};

async function fetchLocations() {
  try {
    const response = await apiClient.get('/pelanggan/lokasi/unik');
    locations.value = response.data;
  } catch (err) { console.error(err); }
}

async function fetchBrandOptions() {
  try {
    const response = await apiClient.get('/harga_layanan');
    brandOptions.value = response.data;
  } catch (err) { console.error(err); }
}

function toISODateString(date: Date): string {
  const year = date.getFullYear();
  const month = (date.getMonth() + 1).toString().padStart(2, '0');
  const day = date.getDate().toString().padStart(2, '0');
  return `${year}-${month}-${day}`;
}

async function fetchReport() {
  isReportLoading.value = true;
  reportSummary.value = null;
  invoiceDetails.value = [];
  currentPage.value = 1;
  try {
    const params = {
      start_date: toISODateString(startDate.value),
      end_date: toISODateString(endDate.value),
      ...(selectedLocation.value && { alamat: selectedLocation.value }),
      ...(selectedBrand.value && { id_brand: selectedBrand.value }),
    };

    const response = await apiClient.get<RevenueReportResponse>('/reports/revenue', { params });
    reportSummary.value = response.data;

    await fetchInvoiceDetails({ page: 1, itemsPerPage: itemsPerPage.value, sortBy: [] });

  } catch (error) {
    console.error("Fetch report failed", error);
  } finally {
    isReportLoading.value = false;
  }
}

const lastParams = ref({});
const debouncedFetchInvoiceDetails = debounce(async (options: { page: number, itemsPerPage: number, sortBy: any[] }) => {
  if (!reportSummary.value || reportSummary.value.total_invoices === 0) return;

  const params = {
    start_date: toISODateString(startDate.value),
    end_date: toISODateString(endDate.value),
    alamat: selectedLocation.value || undefined,
    id_brand: selectedBrand.value || undefined,
    skip: (options.page - 1) * options.itemsPerPage,
    limit: options.itemsPerPage,
  };

  const paramsKey = JSON.stringify(params);
  if (paramsKey === JSON.stringify(lastParams.value) && invoiceDetails.value.length > 0) return;
  lastParams.value = { ...params };

  isDetailsLoading.value = true;
  try {
    const response = await apiClient.get('/reports/revenue/details', { params });
    invoiceDetails.value = response.data;
  } catch (error) {
    console.error(error);
    lastParams.value = {};
  } finally {
    isDetailsLoading.value = false;
  }
}, 300);

async function fetchInvoiceDetails(options: { page: number, itemsPerPage: number, sortBy: any[] }) {
  await debouncedFetchInvoiceDetails(options);
}

async function handleTableOptionsUpdate(options: { page: number, itemsPerPage: number, sortBy: any[] }) {
  if (!isDetailsLoading.value && reportSummary.value) {
    await fetchInvoiceDetails(options);
  }
}

async function exportToExcel() {
    if (!reportSummary.value) return;
    exporting.value = true;
    try {
        const XLSX = await import('xlsx');
        
        // Export Logic from Invoice Details
         const params = {
            start_date: toISODateString(startDate.value),
            end_date: toISODateString(endDate.value),
            alamat: selectedLocation.value || undefined,
            id_brand: selectedBrand.value || undefined,
        };

        const response = await apiClient.get<InvoiceReportItem[]>('/reports/revenue/details', { params });
        const allData = response.data;
        
         const dataToExport = allData.map((item: any) => ({
          "Nomor Invoice": item.invoice_number,
          "Nama Pelanggan": item.pelanggan_nama,
          "Brand": item.brand || "",
          "Alamat": item.alamat || "",
          "Tanggal Bayar": item.tgl_lunas ? new Date(item.tgl_lunas).toLocaleString() : '-',
          "Metode Pembayaran": item.metode || "",
          "Jumlah (Rp)": item.total_harga
        }));

        const worksheet = XLSX.utils.json_to_sheet(dataToExport);
        const workbook = XLSX.utils.book_new();
        XLSX.utils.book_append_sheet(workbook, worksheet, "Laporan");
        XLSX.writeFile(workbook, "Laporan_Keuangan.xlsx");
    } catch(e) {
        console.error(e);
        alert("Gagal export data");
    } finally {
        exporting.value = false;
    }
}

async function fetchMLInsights() {
  isMLLoading.value = true;
  try {
    const params = {
      ...(selectedLocation.value && { location: selectedLocation.value }),
      ...(selectedBrand.value && { brand: selectedBrand.value }),
    };
    const response = await apiClient.get<MLRevenueInsightResponse>('/reports/revenue/ml-insights', { params });
    mlInsights.value = response.data;
  } catch (error) {
    console.error("Fetch ML insights failed", error);
  } finally {
    isMLLoading.value = false;
  }
}

watch(activeTab, (newTab) => {
  if (newTab === 1 && !mlInsights.value) {
    fetchMLInsights();
  }
});

onMounted(async () => {
  await Promise.all([fetchLocations(), fetchBrandOptions(), fetchReport(), fetchMLInsights()]);
});
</script>

<style scoped>
/* Header Card Styles */
.header-card {
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  background: white;
}

.header-section {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  border-radius: 8px;
  position: relative;
  overflow: hidden;
}

.header-section::before {
  display: none;
}

.header-content {
  padding: 24px 32px;
  position: relative;
  z-index: 1;
}

.header-subtitle {
  color: #64748b !important;
  font-size: 0.95rem;
  opacity: 1;
}

.modern-card {
  border: 1px solid #e0e0e0;
  background-color: white;
}

.stats-card {
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  background-color: white;
  transition: box-shadow 0.2s;
}
.stats-card:hover {
  box-shadow: 0 4px 12px rgba(0,0,0,0.05);
}

.border-top-success { border-top: 4px solid #4CAF50; }
.border-top-error { border-top: 4px solid #F44336; }
.border-top-primary { border-top: 4px solid #1976D2; }

.stats-detail-row {
  display: flex;
  justify-content: space-between;
  font-size: 0.8rem;
  color: #616161;
  margin-bottom: 4px;
}

.stats-detail-row.total {
  font-size: 0.9rem;
}

.cursor-pointer {
  cursor: pointer;
}

.cluster-active-border {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12) !important;
}

.border-top-warning { border-top: 4px solid #FFA000; }

/* Optional: Scrollbar for payment methods */
.overflow-y-auto::-webkit-scrollbar {
  width: 4px;
}
.overflow-y-auto::-webkit-scrollbar-track {
  background: #f1f1f1; 
}
.overflow-y-auto::-webkit-scrollbar-thumb {
  background: #bdbdbd; 
  border-radius: 4px;
}
</style>
