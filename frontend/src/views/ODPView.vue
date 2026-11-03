<template>
  <v-container fluid class="pa-4 pa-md-6">
    <!-- Header Section -->
    <div class="header-card mb-4 mb-md-6">
      <div class="header-section">
        <div class="header-content">
          <div class="d-flex flex-wrap align-center justify-space-between gap-3">
            <div class="d-flex align-center">
              <v-avatar class="me-4 elevation-4" color="primary" size="64">
                <v-icon color="white" size="32">mdi-map-marker-path</v-icon>
              </v-avatar>
              <div>
                <h1 class="text-h4 font-weight-bold mb-1">ODP &amp; GIS Network Mapping</h1>
                <p class="header-subtitle mb-0">
                  Pemetaan visual infrastruktur FTTH: Server &gt; OLT &gt; ODC &gt; ODP &gt; Pelanggan
                </p>
              </div>
            </div>

            <!-- Action Buttons -->
            <div class="d-flex flex-wrap align-center gap-2">
              <!-- KMZ Import Button -->
              <v-btn
                color="secondary"
                variant="elevated"
                size="default"
                elevation="2"
                prepend-icon="mdi-file-upload-outline"
                class="text-none font-weight-bold rounded-lg"
                @click="openKmzDialog"
              >
                Import KMZ / KML
              </v-btn>

              <!-- Toggle Table Button -->
              <v-btn
                :color="showTable ? 'primary' : 'grey-lighten-2'"
                :variant="showTable ? 'tonal' : 'outlined'"
                size="default"
                :prepend-icon="showTable ? 'mdi-table-off' : 'mdi-table'"
                class="text-none rounded-lg"
                @click="showTable = !showTable"
              >
                {{ showTable ? 'Sembunyikan Tabel' : 'Tampilkan Tabel' }}
              </v-btn>

              <!-- Refresh Button -->
              <v-btn
                variant="outlined"
                color="grey-darken-1"
                icon="mdi-refresh"
                size="default"
                class="rounded-lg"
                :loading="loading"
                @click="fetchAllData"
                title="Refresh Data"
              ></v-btn>

              <!-- Tambah ODP Button -->
              <v-btn
                color="primary"
                variant="elevated"
                size="default"
                elevation="3"
                prepend-icon="mdi-plus-circle-outline"
                class="text-none font-weight-bold rounded-lg"
                @click="openDialog()"
              >
                Tambah ODP
              </v-btn>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="content-section">
      <!-- Full-Height GIS Map Workspace -->
      <v-card elevation="3" rounded="xl" class="mb-6 overflow-hidden">
        <ODPMap
          :odps="odps"
          :olts="olts"
          :servers="servers"
          :odcs="odcs"
          :pelanggans="pelanggans"
          :kmz-preview="kmzPreviewData"
          height="720px"
          @coordinate-updated="handleCoordinateUpdated"
          @add-odp-click="handleAddOdpFromMap"
          @edit-odp="openDialog"
        />
      </v-card>

      <!-- Collapsible ODP Management Table -->
      <v-expand-transition>
        <v-card v-if="showTable" elevation="2" rounded="xl" class="mb-6">
          <v-card-title class="pa-4 d-flex flex-wrap align-center justify-space-between gap-3">
            <div class="d-flex align-center">
              <v-icon color="primary" class="me-2">mdi-format-list-bulleted</v-icon>
              <span class="text-h6 font-weight-bold">Daftar Titik ODP Terdaftar</span>
              <v-chip size="small" color="primary" variant="flat" class="ms-3 font-weight-bold">
                {{ odps.length }} ODP
              </v-chip>
            </div>

            <!-- Table Search -->
            <v-text-field
              v-model="tableSearch"
              prepend-inner-icon="mdi-magnify"
              placeholder="Cari kode ODP, alamat, OLT..."
              density="compact"
              variant="outlined"
              hide-details
              style="max-width: 300px;"
              clearable
            ></v-text-field>
          </v-card-title>

          <v-divider></v-divider>

          <v-data-table
            :headers="headers"
            :items="odps"
            :search="tableSearch"
            :loading="loading"
            loading-text="Memuat data ODP..."
            no-data-text="Belum ada data ODP"
          >
            <template v-slot:loading>
              <SkeletonLoader type="table" :rows="5" />
            </template>

            <template v-slot:item.kode_odp="{ item }">
              <div class="d-flex align-center font-weight-bold text-primary">
                <v-icon size="18" color="success" class="me-2">mdi-transmission-tower</v-icon>
                {{ item.kode_odp }}
              </div>
            </template>

            <template v-slot:item.kapasitas_port="{ item }">
              <v-progress-linear
                :model-value="item.kapasitas_port > 0 ? (item.port_terpakai / item.kapasitas_port) * 100 : 0"
                :color="getCapacityColor(item.port_terpakai, item.kapasitas_port)"
                height="20"
                rounded
              >
                <strong class="text-white text-caption">{{ item.port_terpakai }} / {{ item.kapasitas_port }}</strong>
              </v-progress-linear>
            </template>

            <template v-slot:item.olt="{ item }">
              <v-chip size="x-small" color="purple" variant="tonal" class="font-weight-medium">
                {{ item.olt?.nama_olt || '-' }}
              </v-chip>
            </template>

            <template v-slot:item.odc="{ item }">
              <v-chip v-if="item.odc" size="x-small" color="orange" variant="tonal" class="font-weight-medium">
                {{ item.odc.nama_odc }}
              </v-chip>
              <span v-else class="text-caption text-grey">-</span>
            </template>

            <template v-slot:item.coordinates="{ item }">
              <div v-if="item.latitude && item.longitude" class="text-caption font-mono">
                {{ item.latitude.toFixed(5) }}, {{ item.longitude.toFixed(5) }}
              </div>
              <v-chip v-else size="x-small" color="error" variant="text">Belum ada</v-chip>
            </template>

            <template v-slot:item.actions="{ item }">
              <div class="d-flex gap-2 justify-end">
                <v-btn
                  size="small"
                  variant="tonal"
                  color="teal"
                  @click="viewTopology(item.olt_id)"
                  prepend-icon="mdi-sitemap"
                >
                  Topologi
                </v-btn>
                <v-btn
                  size="small"
                  variant="tonal"
                  color="primary"
                  @click="openDialog(item)"
                  prepend-icon="mdi-pencil"
                >
                  Edit
                </v-btn>
                <v-btn
                  size="small"
                  variant="tonal"
                  color="error"
                  @click="openDeleteDialog(item)"
                  prepend-icon="mdi-delete"
                >
                  Hapus
                </v-btn>
              </div>
            </template>
          </v-data-table>
        </v-card>
      </v-expand-transition>

      <!-- Add / Edit ODP Dialog -->
      <v-dialog v-model="dialog" max-width="640px" persistent>
        <v-card rounded="xl">
          <v-card-title class="bg-primary text-white pa-4 d-flex align-center">
            <v-icon color="white" class="me-2">mdi-transmission-tower</v-icon>
            {{ formTitle }}
          </v-card-title>
          <v-card-text class="py-4">
            <v-text-field
              v-model="editedItem.kode_odp"
              label="Kode ODP *"
              variant="outlined"
              density="comfortable"
              class="mb-3"
              hint="Contoh: ODP-TMB-001"
              persistent-hint
            ></v-text-field>

            <v-text-field
              v-model="editedItem.alamat"
              label="Alamat / Titik Tiang ODP"
              variant="outlined"
              density="comfortable"
              class="mb-3"
            ></v-text-field>

            <v-row class="mb-1">
              <v-col cols="12" md="6">
                <v-text-field
                  v-model.number="editedItem.kapasitas_port"
                  label="Kapasitas Port"
                  type="number"
                  variant="outlined"
                  density="comfortable"
                  hint="Biasanya 8 atau 16 port"
                  persistent-hint
                ></v-text-field>
              </v-col>
              <v-col cols="12" md="6">
                <v-select
                  v-model="editedItem.olt_id"
                  :items="olts"
                  item-title="nama_olt"
                  item-value="id"
                  label="Terhubung ke OLT *"
                  variant="outlined"
                  density="comfortable"
                ></v-select>
              </v-col>
            </v-row>

            <v-row class="mb-1">
              <v-col cols="12">
                <v-select
                  v-model="editedItem.odc_id"
                  :items="odcs"
                  item-title="nama_odc"
                  item-value="id"
                  label="Terhubung ke ODC (Kabinet)"
                  variant="outlined"
                  density="comfortable"
                  clearable
                  hint="Pilih ODC jika ODP ini terhubung ke kabinet tertentu"
                  persistent-hint
                ></v-select>
              </v-col>
            </v-row>

            <!-- Coordinate Inputs -->
            <v-card variant="tonal" color="blue-grey" rounded="lg" class="pa-3 mb-2">
              <div class="text-caption font-weight-bold text-medium-emphasis mb-2">
                <v-icon size="16" class="me-1">mdi-crosshairs-gps</v-icon>
                Koordinat GIS (Bisa juga disesuaikan dengan drag marker di peta)
              </div>
              <v-row>
                <v-col cols="6">
                  <v-text-field
                    v-model.number="editedItem.latitude"
                    label="Latitude"
                    type="number"
                    step="any"
                    variant="outlined"
                    density="compact"
                    hint="Contoh: -6.238300"
                    persistent-hint
                  ></v-text-field>
                </v-col>
                <v-col cols="6">
                  <v-text-field
                    v-model.number="editedItem.longitude"
                    label="Longitude"
                    type="number"
                    step="any"
                    variant="outlined"
                    density="compact"
                    hint="Contoh: 106.975600"
                    persistent-hint
                  ></v-text-field>
                </v-col>
              </v-row>
            </v-card>
          </v-card-text>

          <v-divider></v-divider>

          <v-card-actions class="pa-4 bg-grey-lighten-4">
            <v-spacer></v-spacer>
            <v-btn variant="text" @click="closeDialog">Batal</v-btn>
            <v-btn color="primary" variant="flat" :loading="saving" class="px-6 rounded-lg font-weight-bold" @click="saveOdp">
              Simpan
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>

      <!-- Delete Confirmation Dialog -->
      <v-dialog v-model="dialogDelete" max-width="500px">
        <v-card rounded="xl">
          <v-card-title class="text-h5 text-center pt-8">Konfirmasi Hapus</v-card-title>
          <v-card-text class="text-center">
            Yakin ingin menghapus ODP <strong>{{ itemToDelete?.kode_odp }}</strong>?
            <br />
            Tindakan ini tidak dapat dibatalkan.
          </v-card-text>
          <v-card-actions class="pa-4">
            <v-spacer></v-spacer>
            <v-btn variant="text" @click="closeDeleteDialog">Batal</v-btn>
            <v-btn color="error" variant="flat" :loading="deleting" class="px-6 rounded-lg font-weight-bold" @click="confirmDelete">
              Ya, Hapus
            </v-btn>
            <v-spacer></v-spacer>
          </v-card-actions>
        </v-card>
      </v-dialog>

      <!-- KMZ / KML Import Modal Dialog -->
      <KmzImportDialog
        v-model="kmzDialog"
        :olts="olts"
        :odcs="odcs"
        @imported="handleKmzImported"
        @preview="handleKmzPreview"
      />

      <!-- Notification Snackbar -->
      <v-snackbar v-model="snackbar.show" :color="snackbar.color" timeout="4000" location="top right">
        {{ snackbar.text }}
        <template v-slot:actions>
          <v-btn variant="text" @click="snackbar.show = false">Tutup</v-btn>
        </template>
      </v-snackbar>
    </div>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import apiClient from '@/services/api';
import { useRouter } from 'vue-router';
import ODPMap, {
  type OdpNode,
  type OltNode,
  type ServerNode,
  type OdcNode,
  type PelangganNode,
} from '@/components/ODPMap.vue';
import KmzImportDialog from '@/components/KmzImportDialog.vue';
import SkeletonLoader from '@/components/SkeletonLoader.vue';
import type { KmzParseResult } from '@/utils/kmzParser';

// --- State Management ---
const router = useRouter();
const odps = ref<OdpNode[]>([]);
const olts = ref<OltNode[]>([]);
const servers = ref<ServerNode[]>([]);
const odcs = ref<OdcNode[]>([]);
const pelanggans = ref<PelangganNode[]>([]);

const loading = ref(true);
const saving = ref(false);
const deleting = ref(false);

const showTable = ref(false);
const tableSearch = ref('');

const dialog = ref(false);
const dialogDelete = ref(false);
const kmzDialog = ref(false);

const editedItem = ref<Partial<OdpNode>>({});
const itemToDelete = ref<OdpNode | null>(null);
const kmzPreviewData = ref<KmzParseResult | null>(null);

const snackbar = ref({
  show: false,
  text: '',
  color: 'success',
});

// --- Computed Properties ---
const isEditMode = computed(() => !!editedItem.value.id);
const formTitle = computed(() => (isEditMode.value ? 'Edit Titik ODP' : 'Tambah Titik ODP Baru'));

// --- Table Headers ---
const headers = [
  { title: 'Kode ODP', key: 'kode_odp' },
  { title: 'Alamat / Lokasi', key: 'alamat' },
  { title: 'Kapasitas (Terpakai/Total)', key: 'kapasitas_port', align: 'center' as const },
  { title: 'OLT', key: 'olt' },
  { title: 'ODC', key: 'odc' },
  { title: 'Koordinat', key: 'coordinates', sortable: false },
  { title: 'Aksi', key: 'actions', sortable: false, align: 'end' as const },
];

// --- Lifecycle Hooks ---
onMounted(() => {
  fetchAllData();
});

// --- API Calls ---
async function fetchAllData() {
  loading.value = true;
  try {
    // 1. Fetch Network GIS Map Overview (includes servers, olts, odcs, odps, pelanggans)
    try {
      const mapRes = await apiClient.get<any>('/network/map-overview');
      const data = mapRes.data?.data;
      if (data) {
        servers.value = data.servers || [];
        olts.value = data.olts || [];
        odcs.value = data.odcs || [];
        odps.value = data.odps || [];
        pelanggans.value = data.pelanggans || [];
        return;
      }
    } catch {
      // Fallback to separate endpoints if /network/map-overview is in transition
    }

    // Fallback: fetch individual collections
    const [odpRes, oltRes] = await Promise.all([
      apiClient.get<any>('/odp'),
      apiClient.get<any>('/olt'),
    ]);

    odps.value = Array.isArray(odpRes.data) ? odpRes.data : odpRes.data.data || [];
    olts.value = Array.isArray(oltRes.data) ? oltRes.data : oltRes.data.data || [];

    try {
      const odcRes = await apiClient.get<any>('/odc');
      odcs.value = Array.isArray(odcRes.data) ? odcRes.data : odcRes.data.data || [];
    } catch {
      odcs.value = [];
    }
  } catch (error) {
    console.error('Gagal mengambil data ODP / GIS:', error);
    showNotification('Gagal memuat data jaringan', 'error');
  } finally {
    loading.value = false;
  }
}

async function saveOdp() {
  if (!editedItem.value.kode_odp || !editedItem.value.olt_id) {
    showNotification('Kode ODP dan Target OLT wajib diisi!', 'warning');
    return;
  }

  saving.value = true;
  try {
    if (isEditMode.value) {
      await apiClient.patch(`/odp/${editedItem.value.id}`, editedItem.value);
      showNotification(`ODP ${editedItem.value.kode_odp} berhasil diperbarui`);
    } else {
      await apiClient.post('/odp', editedItem.value);
      showNotification(`ODP ${editedItem.value.kode_odp} berhasil ditambahkan`);
    }
    closeDialog();
    fetchAllData();
  } catch (error: any) {
    console.error('Gagal menyimpan ODP:', error);
    showNotification(error.response?.data?.error || 'Gagal menyimpan ODP', 'error');
  } finally {
    saving.value = false;
  }
}

async function confirmDelete() {
  if (!itemToDelete.value) return;
  deleting.value = true;
  try {
    await apiClient.delete(`/odp/${itemToDelete.value.id}`);
    showNotification(`ODP ${itemToDelete.value.kode_odp} berhasil dihapus`);
    closeDeleteDialog();
    fetchAllData();
  } catch (error) {
    console.error('Gagal menghapus ODP:', error);
    showNotification('Gagal menghapus ODP', 'error');
  } finally {
    deleting.value = false;
  }
}

// --- Map Event Handlers ---
function handleCoordinateUpdated(payload: { type: string; id: number; lat: number; lng: number }) {
  showNotification(
    `Koordinat ${payload.type.toUpperCase()} #${payload.id} berhasil diperbarui ke (${payload.lat.toFixed(5)}, ${payload.lng.toFixed(5)})`,
    'success'
  );

  // Update in local odps array if ODP
  if (payload.type === 'odp') {
    const found = odps.value.find((o) => o.id === payload.id);
    if (found) {
      found.latitude = payload.lat;
      found.longitude = payload.lng;
    }
  }
}

function handleAddOdpFromMap(coords: { lat: number; lng: number }) {
  editedItem.value = {
    kapasitas_port: 8,
    latitude: parseFloat(coords.lat.toFixed(6)),
    longitude: parseFloat(coords.lng.toFixed(6)),
    olt_id: olts.value.length > 0 ? olts.value[0].id : undefined,
  };
  dialog.value = true;
}

function handleKmzImported(count: number) {
  showNotification(`Berhasil mengimport ${count} titik ODP dari KMZ ke database!`, 'success');
  fetchAllData();
}

function handleKmzPreview(result: KmzParseResult) {
  kmzPreviewData.value = result;
  showNotification(`Menampilkan pratinjau KMZ: ${result.nodes.length} titik & ${result.lines.length} kabel`, 'info');
}

// --- Dialog Controls ---
function openKmzDialog() {
  kmzDialog.value = true;
}

function openDialog(item?: OdpNode) {
  editedItem.value = item
    ? { ...item }
    : {
        kapasitas_port: 8,
        olt_id: olts.value.length > 0 ? olts.value[0].id : undefined,
      };
  dialog.value = true;
}

function closeDialog() {
  dialog.value = false;
  editedItem.value = {};
}

function openDeleteDialog(item: OdpNode) {
  itemToDelete.value = item;
  dialogDelete.value = true;
}

function closeDeleteDialog() {
  dialogDelete.value = false;
  setTimeout(() => {
    itemToDelete.value = null;
  }, 300);
}

function viewTopology(oltId: number) {
  router.push({ name: 'TopologyView', params: { olt_id: oltId } });
}

function getCapacityColor(used = 0, total = 8) {
  const percentage = total > 0 ? (used / total) * 100 : 0;
  if (percentage >= 90) return 'error';
  if (percentage >= 70) return 'warning';
  return 'success';
}

function showNotification(text: string, color = 'success') {
  snackbar.value = {
    show: true,
    text,
    color,
  };
}
</script>

<style scoped>
/* Header Card styling */
.header-card {
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06);
  background: white;
}

.header-content {
  padding: 20px 28px;
}

.content-section {
  width: 100%;
}

.header-section {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
}

.header-section h1 {
  color: #0f172a !important;
  font-weight: 700;
}

.header-subtitle {
  color: #64748b !important;
  font-size: 0.95rem;
}

.font-mono {
  font-family: monospace;
}
</style>
