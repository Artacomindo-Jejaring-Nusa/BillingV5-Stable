<template>
  <v-container fluid class="pa-4 pa-md-6">
    <!-- Header Section -->
    <div class="header-card mb-4 mb-md-6">
      <div class="d-flex flex-column flex-md-row align-start align-md-center justify-space-between gap-4">
        <div class="d-flex align-center header-info">
          <div class="header-avatar-wrapper">
            <v-avatar class="header-avatar" color="transparent" size="50">
              <v-icon color="primary" size="28">mdi-tag-multiple</v-icon>
            </v-avatar>
          </div>
          <div class="ml-4">
            <h1 class="header-title">Brand & Paket Layanan</h1>
            <p class="header-subtitle">
              Kelola brand provider internet dan paket kecepatan beserta tarifnya secara terpadu
            </p>
          </div>
        </div>

        <div class="d-flex align-center gap-2 flex-wrap header-actions">
          <div class="stat-badge">
            <v-icon size="16" class="mr-1 stat-badge-icon">mdi-domain</v-icon>
            <span class="stat-badge-value">{{ brands.length }}</span>
            <span class="stat-badge-label">Brand Aktif</span>
          </div>
          <div class="stat-badge">
            <v-icon size="16" class="mr-1 stat-badge-icon">mdi-package-variant</v-icon>
            <span class="stat-badge-value">{{ totalPackagesCount }}</span>
            <span class="stat-badge-label">Paket Layanan</span>
          </div>
          <v-btn
            color="primary"
            prepend-icon="mdi-plus"
            @click="openBrandDialog()"
            class="primary-btn text-none font-weight-bold"
            elevation="0"
          >
            Tambah Brand
          </v-btn>
        </div>
      </div>
    </div>

    <!-- Quick Filter / Search Bar Card -->
    <div class="filter-card mb-5">
      <div class="d-flex flex-column flex-sm-row align-start align-sm-center justify-space-between gap-3">
        <div class="filter-search-wrapper w-100">
          <v-text-field
            v-model="searchQuery"
            placeholder="Cari nama paket, kecepatan (contoh: 20 Mbps), atau brand..."
            variant="outlined"
            density="compact"
            prepend-inner-icon="mdi-magnify"
            clearable
            hide-details
            class="search-input"
          ></v-text-field>
        </div>
        <div v-if="searchQuery" class="filter-result-indicator text-caption text-medium-emphasis text-no-wrap">
          Menampilkan hasil untuk: <strong>"{{ searchQuery }}"</strong>
        </div>
      </div>
    </div>

    <!-- Loading Skeleton -->
    <div v-if="brandLoading || packageLoading">
      <v-skeleton-loader
        v-for="i in 3"
        :key="i"
        type="card"
        class="mb-4 rounded-xl border"
      />
    </div>

    <!-- Empty State if no brands exist -->
    <div v-else-if="brands.length === 0" class="empty-state-card mb-6">
      <div class="empty-icon-box mb-3">
        <v-icon size="36" color="slate-400">mdi-domain-off</v-icon>
      </div>
      <h3 class="empty-state-title mb-1">Belum Ada Brand Provider</h3>
      <p class="empty-state-desc mb-5">
        Tambahkan brand provider pertama Anda untuk mulai mengatur paket kecepatan dan tarif layanan internet.
      </p>
      <v-btn
        color="primary"
        prepend-icon="mdi-plus"
        @click="openBrandDialog()"
        class="primary-btn text-none font-weight-bold"
        elevation="0"
      >
        Tambah Brand Baru
      </v-btn>
    </div>

    <!-- Grouped Brand Cards List -->
    <div v-else class="brand-groups-list">
      <div
        v-for="item in brandPackagesMap"
        :key="item.brand.id_brand"
        class="brand-group-card mb-5"
      >
        <!-- Brand Header Bar -->
        <div class="brand-card-header pa-4 pa-sm-5 d-flex flex-wrap align-center justify-space-between gap-3">
          <div class="d-flex align-center gap-3">
            <div class="brand-monogram">
              {{ item.brand.brand ? item.brand.brand.substring(0, 2).toUpperCase() : 'BR' }}
            </div>
            <div>
              <div class="d-flex align-center gap-2 flex-wrap">
                <span class="brand-title">{{ item.brand.brand }}</span>
                <span class="meta-pill id-pill">
                  {{ item.brand.id_brand }}
                </span>
                <span
                  class="meta-pill tax-pill"
                  :class="item.brand.pajak > 10 ? 'tax-standard' : 'tax-low'"
                >
                  PPN {{ item.brand.pajak }}%
                </span>
                <span v-if="item.brand.xendit_key_name" class="meta-pill gateway-pill">
                  <v-icon size="13" class="mr-1">mdi-key-variant</v-icon>
                  {{ item.brand.xendit_key_name }}
                </span>
              </div>
              <div class="brand-subtitle mt-1">
                {{ item.totalCount }} paket tersedia untuk brand ini
              </div>
            </div>
          </div>

          <!-- Brand Actions -->
          <div class="d-flex align-center gap-2">
            <v-btn
              size="small"
              variant="outlined"
              color="primary"
              prepend-icon="mdi-plus"
              class="text-none font-weight-bold btn-add-package"
              @click="openPackageDialog(undefined, item.brand)"
            >
              Tambah Paket
            </v-btn>
            <v-tooltip text="Ubah Brand" location="top">
              <template v-slot:activator="{ props }">
                <v-btn
                  v-bind="props"
                  size="small"
                  variant="outlined"
                  class="icon-action-btn"
                  icon="mdi-pencil-outline"
                  @click="openBrandDialog(item.brand)"
                ></v-btn>
              </template>
            </v-tooltip>
            <v-tooltip text="Hapus Brand" location="top">
              <template v-slot:activator="{ props }">
                <v-btn
                  v-bind="props"
                  size="small"
                  variant="outlined"
                  color="error"
                  class="icon-action-btn icon-delete-btn"
                  icon="mdi-delete-outline"
                  @click="openDeleteBrandDialog(item.brand)"
                ></v-btn>
              </template>
            </v-tooltip>
          </div>
        </div>

        <!-- Packages Table inside Brand Card -->
        <div class="table-responsive">
          <v-table class="packages-table bg-transparent" density="comfortable" hover>
            <thead>
              <tr class="table-header-row">
                <th class="text-left font-weight-bold text-caption text-uppercase" style="width: 35%;">
                  Nama Paket
                </th>
                <th class="text-center font-weight-bold text-caption text-uppercase" style="width: 20%;">
                  Kecepatan
                </th>
                <th class="text-end font-weight-bold text-caption text-uppercase" style="width: 30%;">
                  Harga Paket (per Bulan)
                </th>
                <th class="text-center font-weight-bold text-caption text-uppercase" style="width: 15%;">
                  Aksi
                </th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="pkg in item.packages" :key="pkg.id" class="package-row">
                <td class="py-3">
                  <div class="d-flex align-center gap-2">
                    <div class="package-icon-box">
                      <v-icon size="16" color="primary">mdi-wifi</v-icon>
                    </div>
                    <div>
                      <div class="package-name">{{ pkg.nama_paket }}</div>
                    </div>
                  </div>
                </td>
                <td class="text-center py-3">
                  <span class="speed-pill">
                    <v-icon size="14" class="mr-1">mdi-speedometer</v-icon>
                    {{ pkg.kecepatan }} Mbps
                  </span>
                </td>
                <td class="text-end py-3">
                  <span class="price-value">
                    {{ formatCurrency(pkg.harga) }}
                  </span>
                  <span class="price-period"> /bln</span>
                </td>
                <td class="text-center py-3">
                  <div class="d-flex justify-center gap-1">
                    <v-tooltip text="Ubah Paket" location="top">
                      <template v-slot:activator="{ props }">
                        <v-btn
                          v-bind="props"
                          size="small"
                          variant="text"
                          color="primary"
                          icon="mdi-pencil-outline"
                          class="row-action-btn"
                          @click="openPackageDialog(pkg, item.brand)"
                        ></v-btn>
                      </template>
                    </v-tooltip>
                    <v-tooltip text="Hapus Paket" location="top">
                      <template v-slot:activator="{ props }">
                        <v-btn
                          v-bind="props"
                          size="small"
                          variant="text"
                          color="error"
                          icon="mdi-delete-outline"
                          class="row-action-btn row-delete-btn"
                          @click="openDeletePackageDialog(pkg)"
                        ></v-btn>
                      </template>
                    </v-tooltip>
                  </div>
                </td>
              </tr>
              <!-- Empty state within this specific brand -->
              <tr v-if="item.packages.length === 0">
                <td colspan="4" class="text-center py-8">
                  <div class="d-flex flex-column align-center justify-center py-3">
                    <v-icon size="32" color="slate-400" class="mb-2">mdi-package-variant-closed</v-icon>
                    <span class="text-body-2 text-medium-emphasis mb-3">
                      {{ searchQuery ? 'Tidak ada paket yang cocok dengan pencarian Anda' : 'Belum ada paket layanan untuk brand ini' }}
                    </span>
                    <v-btn
                      v-if="!searchQuery"
                      size="small"
                      variant="outlined"
                      color="primary"
                      prepend-icon="mdi-plus"
                      class="text-none font-weight-medium rounded-lg"
                      @click="openPackageDialog(undefined, item.brand)"
                    >
                      Tambah Paket Pertama
                    </v-btn>
                  </div>
                </td>
              </tr>
            </tbody>
          </v-table>
        </div>
      </div>
    </div>

    <!-- Dialogs -->
    <HargaLayananDialog v-model="dialogBrand" :edited-item="editedBrand" @save="saveBrand" />
    <PaketLayananDialog
      v-model="dialogPackage"
      :edited-item="editedPackage"
      :brand-id="selectedBrand?.id_brand"
      @save="savePackage"
    />
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import apiClient from '@/services/api';
import type { HargaLayanan, PaketLayanan } from '@/interfaces/layanan';

import HargaLayananDialog from '@/components/dialogs/HargaLayananDialog.vue';
import PaketLayananDialog from '@/components/dialogs/PaketLayananDialog.vue';

// ===== STATE MANAGEMENT =====
const brands = ref<HargaLayanan[]>([]);
const packages = ref<PaketLayanan[]>([]);
const searchQuery = ref('');

const brandLoading = ref(true);
const packageLoading = ref(true);

const selectedBrand = ref<HargaLayanan | null>(null);

// Dialogs visibility
const dialogBrand = ref(false);
const dialogPackage = ref(false);

const editedBrand = ref<Partial<HargaLayanan>>({});
const editedPackage = ref<Partial<PaketLayanan>>({});

let editedBrandIndex = -1;
let editedPackageIndex = -1;

// ===== COMPUTED PROPERTIES =====
const brandPackagesMap = computed(() => {
  const query = searchQuery.value.toLowerCase().trim();
  return brands.value.map(brand => {
    const brandPkgs = packages.value.filter((p: PaketLayanan) => p.id_brand === brand.id_brand);
    let filtered = brandPkgs;
    if (query) {
      filtered = brandPkgs.filter((p: PaketLayanan) =>
        p.nama_paket.toLowerCase().includes(query) ||
        String(p.kecepatan).includes(query) ||
        brand.brand.toLowerCase().includes(query) ||
        brand.id_brand.toLowerCase().includes(query)
      );
    }
    return {
      brand,
      packages: filtered,
      totalCount: brandPkgs.length
    };
  });
});

const totalPackagesCount = computed(() => packages.value.length);

// ===== LIFECYCLE HOOKS =====
onMounted(() => {
  fetchBrands();
  fetchPackages();
});

async function fetchBrands() {
  brandLoading.value = true;
  try {
    const response = await apiClient.get<HargaLayanan[]>('/harga_layanan');
    brands.value = response.data.map(brand => ({
      ...brand,
      id_brand: brand.id_brand.trim()
    }));
  } finally {
    brandLoading.value = false;
  }
}

async function fetchPackages() {
  packageLoading.value = true;
  try {
    const response = await apiClient.get<PaketLayanan[]>('/paket_layanan');
    packages.value = response.data;
  } finally {
    packageLoading.value = false;
  }
}

// Brand Dialog Logic
function openBrandDialog(item?: HargaLayanan) {
  editedBrandIndex = item ? brands.value.findIndex(b => b.id_brand === item.id_brand) : -1;
  editedBrand.value = item ? { ...item } : { pajak: 11.0 };
  dialogBrand.value = true;
}

async function saveBrand(item: HargaLayanan) {
  if (editedBrandIndex > -1) {
    await apiClient.patch(`/harga_layanan/${item.id_brand}`, item);
  } else {
    await apiClient.post('/harga_layanan', item);
  }
  fetchBrands();
}

// Package Dialog Logic
function openPackageDialog(item?: PaketLayanan, brand?: HargaLayanan) {
  if (brand) {
    selectedBrand.value = brand;
  } else if (item) {
    selectedBrand.value = brands.value.find(b => b.id_brand === item.id_brand) || null;
  } else if (brands.value.length > 0 && !selectedBrand.value) {
    selectedBrand.value = brands.value[0];
  }
  editedPackageIndex = item ? packages.value.findIndex(p => p.id === item.id) : -1;
  editedPackage.value = item ? { ...item } : {};
  dialogPackage.value = true;
}

async function savePackage(item: PaketLayanan) {
  if (editedPackageIndex > -1) {
    await apiClient.patch(`/paket_layanan/${item.id}`, item);
  } else {
    item.id_brand = selectedBrand.value!.id_brand.trim();
    await apiClient.post('/paket_layanan', item);
  }
  fetchPackages();
}

// Format Currency
function formatCurrency(value: number) {
  if (isNaN(value)) return "Rp 0";
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(value);
}

// Delete Functions
async function deleteBrand(id_brand: string) {
  try {
    await apiClient.delete(`/harga_layanan/${id_brand}`);
    fetchBrands();
    fetchPackages();
    if (selectedBrand.value && selectedBrand.value.id_brand === id_brand) {
      selectedBrand.value = null;
    }
  } catch (error) {
    console.error('Error deleting brand:', error);
    alert('Gagal menghapus brand. Pastikan tidak ada pelanggan yang menggunakan brand ini.');
  }
}

function openDeleteBrandDialog(item: HargaLayanan) {
  if (confirm(`Apakah Anda yakin ingin menghapus brand "${item.brand}"? Semua paket layanan terkait juga akan terhapus.`)) {
    deleteBrand(item.id_brand);
  }
}

async function deletePackage(paketId: number) {
  try {
    await apiClient.delete(`/paket_layanan/${paketId}`);
    fetchPackages();
  } catch (error) {
    console.error('Error deleting package:', error);
    alert('Gagal menghapus paket. Pastikan tidak ada pelanggan yang menggunakan paket ini.');
  }
}

function openDeletePackageDialog(item: PaketLayanan) {
  if (confirm(`Apakah Anda yakin ingin menghapus paket "${item.nama_paket}"?`)) {
    deletePackage(item.id!);
  }
}
</script>

<style scoped>
.gap-1 {
  gap: 0.25rem;
}
.gap-2 {
  gap: 0.5rem;
}
.gap-3 {
  gap: 0.75rem;
}
.gap-4 {
  gap: 1rem;
}

/* ==========================================================================
   Header Card (Standard Clean SaaS Style)
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

.header-info {
  width: auto;
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

/* Stat Badges */
.stat-badge {
  display: inline-flex;
  align-items: center;
  padding: 6px 12px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 0.8125rem;
}

.stat-badge-icon {
  color: #64748b;
}

.stat-badge-value {
  font-weight: 700;
  color: #0f172a;
  margin-right: 4px;
}

.stat-badge-label {
  color: #64748b;
}

/* Primary Button */
.primary-btn {
  background-color: #0f172a !important;
  color: #ffffff !important;
  border-radius: 8px !important;
  padding: 0 18px !important;
  height: 40px !important;
  letter-spacing: normal;
  transition: all 0.2s ease;
}

.primary-btn:hover {
  background-color: #1e293b !important;
}

/* ==========================================================================
   Filter / Search Card
   ========================================================================== */
.filter-card {
  background: #ffffff;
  border-radius: 8px;
  padding: 16px 20px;
  border: 1px solid #e2e8f0;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.search-input {
  max-width: 460px;
}

.filter-result-indicator {
  color: #64748b;
  font-size: 0.8125rem;
}

/* ==========================================================================
   Empty State Card
   ========================================================================== */
.empty-state-card {
  background: #ffffff;
  border: 1px dashed #cbd5e1;
  border-radius: 12px;
  padding: 48px 24px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.empty-icon-box {
  width: 64px;
  height: 64px;
  border-radius: 16px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-state-title {
  font-size: 1.125rem;
  font-weight: 700;
  color: #0f172a;
}

.empty-state-desc {
  font-size: 0.875rem;
  color: #64748b;
  max-width: 440px;
  margin-left: auto;
  margin-right: auto;
  line-height: 1.5;
}

/* ==========================================================================
   Brand Group Card
   ========================================================================== */
.brand-group-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  overflow: hidden;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.brand-group-card:hover {
  border-color: #cbd5e1;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.04);
}

.brand-card-header {
  background: #ffffff;
  border-bottom: 1px solid #f1f5f9;
}

.brand-monogram {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  color: #0f172a;
  font-weight: 700;
  font-size: 0.9375rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  letter-spacing: 0.5px;
}

.brand-title {
  font-size: 1.125rem;
  font-weight: 700;
  color: #0f172a;
  line-height: 1.3;
}

.brand-subtitle {
  font-size: 0.8125rem;
  color: #64748b;
}

/* Meta Pills */
.meta-pill {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 600;
  line-height: 1.4;
}

.id-pill {
  background: #f1f5f9;
  color: #475569;
  border: 1px solid #e2e8f0;
}

.tax-pill.tax-standard {
  background: #ecfdf5;
  color: #047857;
  border: 1px solid #a7f3d0;
}

.tax-pill.tax-low {
  background: #fffbeb;
  color: #b45309;
  border: 1px solid #fde68a;
}

.gateway-pill {
  background: #f8fafc;
  color: #64748b;
  border: 1px solid #e2e8f0;
}

/* Brand Action Buttons */
.btn-add-package {
  border-color: #cbd5e1 !important;
  color: #0f172a !important;
  background: #ffffff !important;
  border-radius: 8px !important;
  height: 34px !important;
}

.btn-add-package:hover {
  background: #f8fafc !important;
  border-color: #94a3b8 !important;
}

.icon-action-btn {
  border-color: #e2e8f0 !important;
  color: #64748b !important;
  background: #ffffff !important;
  border-radius: 8px !important;
}

.icon-action-btn:hover {
  background: #f1f5f9 !important;
  color: #0f172a !important;
  border-color: #cbd5e1 !important;
}

.icon-delete-btn:hover {
  background: #fef2f2 !important;
  color: #ef4444 !important;
  border-color: #fecaca !important;
}

/* ==========================================================================
   Packages Table
   ========================================================================== */
.table-responsive {
  width: 100%;
  overflow-x: auto;
}

.table-header-row th {
  background: #f8fafc !important;
  color: #64748b !important;
  letter-spacing: 0.05em;
  font-size: 0.75rem !important;
  border-bottom: 1px solid #e2e8f0 !important;
  height: 44px !important;
}

.package-row td {
  border-bottom: 1px solid #f1f5f9 !important;
}

.package-row:hover td {
  background: #f8fafc !important;
}

.package-icon-box {
  width: 30px;
  height: 30px;
  border-radius: 8px;
  background: #eff6ff;
  border: 1px solid #dbeafe;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.package-name {
  font-weight: 600;
  font-size: 0.875rem;
  color: #0f172a;
}

.speed-pill {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  background: #eff6ff;
  color: #1d4ed8;
  border: 1px solid #bfdbfe;
  border-radius: 6px;
  font-size: 0.8125rem;
  font-weight: 600;
}

.price-value {
  font-size: 0.9375rem;
  font-weight: 700;
  color: #047857;
  font-variant-numeric: tabular-nums;
}

.price-period {
  font-size: 0.75rem;
  color: #64748b;
}

.row-action-btn {
  border-radius: 6px !important;
  transition: all 0.2s ease;
}

.row-action-btn:hover {
  background: #f1f5f9 !important;
}

.row-delete-btn:hover {
  background: #fef2f2 !important;
}

/* ==========================================================================
   Dark Theme Adjustments
   ========================================================================== */
.v-theme--dark .header-card,
.v-theme--dark .filter-card,
.v-theme--dark .brand-group-card,
.v-theme--dark .brand-card-header {
  background: #1e293b !important;
  border-color: #334155 !important;
  color: #f8fafc !important;
}

.v-theme--dark .header-title,
.v-theme--dark .brand-title,
.v-theme--dark .package-name,
.v-theme--dark .stat-badge-value,
.v-theme--dark .empty-state-title {
  color: #f8fafc !important;
}

.v-theme--dark .header-avatar-wrapper,
.v-theme--dark .brand-monogram,
.v-theme--dark .empty-icon-box {
  background: #0f172a !important;
  border-color: #334155 !important;
  color: #f8fafc !important;
}

.v-theme--dark .stat-badge,
.v-theme--dark .id-pill,
.v-theme--dark .gateway-pill {
  background: #0f172a !important;
  border-color: #334155 !important;
  color: #94a3b8 !important;
}

.v-theme--dark .table-header-row th {
  background: #0f172a !important;
  color: #94a3b8 !important;
  border-color: #334155 !important;
}

.v-theme--dark .package-row td {
  border-color: #334155 !important;
}

.v-theme--dark .package-row:hover td {
  background: rgba(255, 255, 255, 0.03) !important;
}

.v-theme--dark .btn-add-package,
.v-theme--dark .icon-action-btn {
  background: #1e293b !important;
  border-color: #334155 !important;
  color: #f8fafc !important;
}

.v-theme--dark .package-icon-box {
  background: rgba(59, 130, 246, 0.15) !important;
  border-color: rgba(59, 130, 246, 0.3) !important;
}

.v-theme--dark .empty-state-card {
  background: #1e293b !important;
  border-color: #334155 !important;
}

@media (max-width: 600px) {
  .search-input {
    max-width: 100%;
  }
}
</style>
