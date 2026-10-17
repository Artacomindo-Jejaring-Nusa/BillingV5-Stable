<template>
  <v-container fluid class="pa-4 pa-md-6">
    <!-- Header Section -->
    <v-card class="pa-4 pa-md-6 mb-6 rounded-xl border bg-surface" elevation="0">
      <div class="d-flex flex-column flex-md-row align-start align-md-center justify-space-between gap-4">
        <div>
          <div class="d-flex align-center gap-2 mb-1">
            <v-icon color="primary" size="26">mdi-tag-multiple</v-icon>
            <h1 class="text-h5 font-weight-bold">Brand & Paket Layanan</h1>
          </div>
          <p class="text-body-2 text-medium-emphasis mb-0">
            Kelola brand provider internet dan paket kecepatan beserta tarifnya secara terpadu
          </p>
        </div>

        <div class="d-flex align-center gap-2 flex-wrap">
          <v-chip variant="outlined" color="primary" size="small" class="font-weight-medium">
            <v-icon start size="16">mdi-domain</v-icon>
            {{ brands.length }} Brand Aktif
          </v-chip>
          <v-chip variant="outlined" color="success" size="small" class="font-weight-medium">
            <v-icon start size="16">mdi-package-variant</v-icon>
            {{ totalPackagesCount }} Paket Layanan
          </v-chip>
          <v-btn
            color="primary"
            variant="flat"
            prepend-icon="mdi-plus"
            @click="openBrandDialog()"
            class="text-none font-weight-bold rounded-lg px-4"
          >
            Tambah Brand
          </v-btn>
        </div>
      </div>

      <!-- Quick Filter / Search Bar -->
      <div class="mt-4 pt-4 border-t d-flex align-center gap-3">
        <v-text-field
          v-model="searchQuery"
          placeholder="Cari nama paket, kecepatan (contoh: 20 Mbps), atau brand..."
          variant="outlined"
          density="compact"
          prepend-inner-icon="mdi-magnify"
          clearable
          hide-details
          class="search-input max-w-md"
        ></v-text-field>
        <span v-if="searchQuery" class="text-caption text-medium-emphasis">
          Menampilkan hasil pencarian untuk "{{ searchQuery }}"
        </span>
      </div>
    </v-card>

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
    <v-card
      v-else-if="brands.length === 0"
      class="rounded-xl border pa-8 text-center bg-surface"
      elevation="0"
    >
      <v-icon size="48" color="grey" class="mb-3">mdi-domain-off</v-icon>
      <h3 class="text-h6 font-weight-bold mb-1">Belum Ada Brand Provider</h3>
      <p class="text-body-2 text-medium-emphasis mb-4">
        Tambahkan brand provider pertama Anda untuk mulai mengatur paket layanan.
      </p>
      <v-btn
        color="primary"
        prepend-icon="mdi-plus"
        @click="openBrandDialog()"
        class="text-none font-weight-bold rounded-lg"
      >
        Tambah Brand Baru
      </v-btn>
    </v-card>

    <!-- Grouped Brand Cards List -->
    <div v-else class="brand-groups-list">
      <v-card
        v-for="item in brandPackagesMap"
        :key="item.brand.id_brand"
        class="brand-group-card rounded-xl border bg-surface mb-5"
        elevation="0"
      >
        <!-- Brand Header Bar -->
        <div class="brand-card-header pa-4 pa-sm-5 d-flex flex-wrap align-center justify-space-between gap-3 border-b">
          <div class="d-flex align-center gap-3">
            <v-avatar size="44" color="primary" class="font-weight-bold text-subtitle-2 rounded-lg text-white">
              {{ item.brand.brand.substring(0, 2).toUpperCase() }}
            </v-avatar>
            <div>
              <div class="d-flex align-center gap-2 flex-wrap">
                <span class="text-h6 font-weight-bold">{{ item.brand.brand }}</span>
                <v-chip size="x-small" variant="tonal" color="blue-grey" class="font-weight-bold">
                  {{ item.brand.id_brand }}
                </v-chip>
                <v-chip
                  size="x-small"
                  variant="tonal"
                  :color="item.brand.pajak > 10 ? 'orange' : 'green'"
                  class="font-weight-medium"
                >
                  Pajak {{ item.brand.pajak }}%
                </v-chip>
                <v-chip
                  size="x-small"
                  variant="outlined"
                  color="purple"
                  prepend-icon="mdi-key"
                  class="font-weight-medium"
                >
                  Gateway: {{ item.brand.xendit_key_name }}
                </v-chip>
              </div>
              <div class="text-caption text-medium-emphasis mt-1">
                {{ item.totalCount }} paket tersedia untuk brand ini
              </div>
            </div>
          </div>

          <!-- Brand Actions -->
          <div class="d-flex align-center gap-2">
            <v-btn
              size="small"
              variant="tonal"
              color="primary"
              prepend-icon="mdi-plus"
              class="text-none font-weight-bold rounded-lg"
              @click="openPackageDialog(undefined, item.brand)"
            >
              Tambah Paket
            </v-btn>
            <v-btn
              size="small"
              variant="text"
              color="default"
              icon="mdi-pencil-outline"
              class="rounded-lg"
              @click="openBrandDialog(item.brand)"
              title="Edit Brand"
            ></v-btn>
            <v-btn
              size="small"
              variant="text"
              color="error"
              icon="mdi-delete-outline"
              class="rounded-lg"
              @click="openDeleteBrandDialog(item.brand)"
              title="Hapus Brand"
            ></v-btn>
          </div>
        </div>

        <!-- Packages Table inside Brand Card -->
        <div class="table-responsive">
          <v-table class="packages-table bg-transparent" density="comfortable" hover>
            <thead>
              <tr class="table-header-row">
                <th class="text-left font-weight-bold text-caption text-medium-emphasis text-uppercase" style="width: 35%;">
                  Nama Paket
                </th>
                <th class="text-center font-weight-bold text-caption text-medium-emphasis text-uppercase" style="width: 20%;">
                  Kecepatan
                </th>
                <th class="text-end font-weight-bold text-caption text-medium-emphasis text-uppercase" style="width: 30%;">
                  Harga Paket (per Bulan)
                </th>
                <th class="text-center font-weight-bold text-caption text-medium-emphasis text-uppercase" style="width: 15%;">
                  Aksi
                </th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="pkg in item.packages" :key="pkg.id" class="package-row">
                <td class="py-3">
                  <div class="d-flex align-center gap-2">
                    <v-avatar size="28" color="primary" variant="tonal" class="rounded-md">
                      <v-icon size="16" color="primary">mdi-wifi</v-icon>
                    </v-avatar>
                    <div>
                      <div class="font-weight-medium text-body-2">{{ pkg.nama_paket }}</div>
                    </div>
                  </div>
                </td>
                <td class="text-center py-3">
                  <v-chip size="small" variant="tonal" color="info" prepend-icon="mdi-speedometer" class="font-weight-medium">
                    {{ pkg.kecepatan }} Mbps
                  </v-chip>
                </td>
                <td class="text-end py-3">
                  <span class="font-weight-bold text-subtitle-2 text-success">
                    {{ formatCurrency(pkg.harga) }}
                  </span>
                  <span class="text-caption text-medium-emphasis"> /bln</span>
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
                          icon="mdi-pencil"
                          class="rounded-lg"
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
                          icon="mdi-delete"
                          class="rounded-lg"
                          @click="openDeletePackageDialog(pkg)"
                        ></v-btn>
                      </template>
                    </v-tooltip>
                  </div>
                </td>
              </tr>
              <!-- Empty state within this specific brand -->
              <tr v-if="item.packages.length === 0">
                <td colspan="4" class="text-center py-6 text-medium-emphasis">
                  <div class="d-flex flex-column align-center justify-center py-2">
                    <v-icon size="32" color="grey" class="mb-1">mdi-package-variant-closed</v-icon>
                    <span class="text-body-2 mb-2">
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
      </v-card>
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
.gap-2 {
  gap: 0.5rem;
}
.gap-3 {
  gap: 0.75rem;
}
.gap-4 {
  gap: 1rem;
}

.search-input {
  max-width: 480px;
}

.brand-group-card {
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.brand-group-card:hover {
  border-color: rgba(var(--v-theme-primary), 0.35) !important;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.04) !important;
}

.brand-card-header {
  background: rgba(var(--v-theme-on-surface), 0.015);
}

.table-header-row th {
  background: rgba(var(--v-theme-on-surface), 0.02) !important;
  letter-spacing: 0.5px;
}

.package-row:hover {
  background-color: rgba(var(--v-theme-primary), 0.03) !important;
}

.table-responsive {
  width: 100%;
  overflow-x: auto;
}

@media (max-width: 600px) {
  .search-input {
    max-width: 100%;
  }
}
</style>