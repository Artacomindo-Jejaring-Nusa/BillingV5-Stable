<template>
  <v-dialog :model-value="modelValue" @update:model-value="$emit('update:modelValue', $event)" max-width="960px" persistent scrollable>
    <v-card rounded="xl" class="elevation-12">
      <!-- Header -->
      <v-card-title class="pa-4 bg-primary d-flex align-center">
        <v-avatar color="white" size="36" class="me-3">
          <v-icon color="primary">mdi-map-marker-path</v-icon>
        </v-avatar>
        <div>
          <div class="text-h6 font-weight-bold text-white">Import Modul KMZ / KML</div>
          <div class="text-caption text-white opacity-80">Pemetaan otomatis Server &gt; OLT &gt; ODC &gt; ODP &gt; Pelanggan</div>
        </div>
        <v-spacer></v-spacer>
        <v-btn icon="mdi-close" variant="text" color="white" size="small" @click="close"></v-btn>
      </v-card-title>

      <v-divider></v-divider>

      <!-- Content -->
      <v-card-text class="pa-4 pa-md-6" style="max-height: 70vh;">
        <!-- Step 1: Upload Dropzone -->
        <div v-if="!parsedResult" class="text-center">
          <div
            class="dropzone-box pa-8 rounded-xl border-dashed d-flex flex-column align-center justify-center cursor-pointer"
            :class="{ 'dropzone-active': isDragging }"
            @dragover.prevent="isDragging = true"
            @dragleave.prevent="isDragging = false"
            @drop.prevent="handleDrop"
            @click="triggerFileInput"
          >
            <input
              ref="fileInputRef"
              type="file"
              accept=".kmz,.kml"
              style="display: none"
              @change="handleFileChange"
            />
            <v-avatar color="primary-lighten-4" size="72" class="mb-4">
              <v-icon color="primary" size="36">mdi-cloud-upload-outline</v-icon>
            </v-avatar>
            <h3 class="text-h6 font-weight-bold mb-1">Pilih atau Drag File KMZ / KML</h3>
            <p class="text-body-2 text-medium-emphasis mb-4">
              Mendukung file .kmz (Google Earth) dan .kml dengan layer placemark dan linestring
            </p>
            <v-btn color="primary" variant="flat" prepend-icon="mdi-file-find-outline" class="rounded-lg text-none px-6">
              Pilih File dari Komputer
            </v-btn>
          </div>

          <div v-if="parsing" class="mt-6 text-center">
            <v-progress-circular indeterminate color="primary" size="48"></v-progress-circular>
            <p class="text-body-2 text-medium-emphasis mt-2">Mengekstrak dan memproses struktur KMZ/KML...</p>
          </div>

          <v-alert v-if="errorMessage" type="error" variant="tonal" class="mt-4 rounded-lg" closable>
            {{ errorMessage }}
          </v-alert>
        </div>

        <!-- Step 2: Parsed Result Preview & Target Selection -->
        <div v-else>
          <!-- Summary Banner -->
          <v-card variant="tonal" color="primary" rounded="lg" class="pa-4 mb-4">
            <div class="d-flex flex-wrap align-center justify-space-between gap-3">
              <div class="d-flex align-center">
                <v-icon color="primary" size="28" class="me-3">mdi-file-check-outline</v-icon>
                <div>
                  <div class="font-weight-bold text-subtitle-1">{{ parsedResult.fileName }}</div>
                  <div class="text-caption text-medium-emphasis">
                    Ditemukan {{ parsedResult.nodes.length }} titik koordinat dan {{ parsedResult.lines.length }} jalur kabel
                  </div>
                </div>
              </div>
              <v-btn size="small" variant="text" color="error" prepend-icon="mdi-refresh" @click="resetUpload">
                Ganti File
              </v-btn>
            </div>

            <!-- Chips Breakdown -->
            <div class="d-flex flex-wrap gap-2 mt-3">
              <v-chip size="small" color="indigo" variant="flat">
                <v-icon start size="16">mdi-server-network</v-icon>
                Server: {{ parsedResult.summary.servers }}
              </v-chip>
              <v-chip size="small" color="purple" variant="flat">
                <v-icon start size="16">mdi-router-wireless</v-icon>
                OLT: {{ parsedResult.summary.olts }}
              </v-chip>
              <v-chip size="small" color="warning" variant="flat">
                <v-icon start size="16">mdi-inbox-multiple</v-icon>
                ODC: {{ parsedResult.summary.odcs }}
              </v-chip>
              <v-chip size="small" color="success" variant="flat">
                <v-icon start size="16">mdi-transmission-tower</v-icon>
                ODP: {{ parsedResult.summary.odps }}
              </v-chip>
              <v-chip size="small" color="cyan" variant="flat">
                <v-icon start size="16">mdi-account-network</v-icon>
                Pelanggan: {{ parsedResult.summary.pelanggans }}
              </v-chip>
              <v-chip size="small" color="grey-darken-1" variant="flat">
                <v-icon start size="16">mdi-vector-polyline</v-icon>
                Kabel / Line: {{ parsedResult.summary.lines }}
              </v-chip>
            </div>
          </v-card>

          <!-- OLT / ODC Assignment Target -->
          <v-row class="mb-2">
            <v-col cols="12" md="6">
              <v-select
                v-model="targetOltId"
                :items="olts"
                item-title="nama_olt"
                item-value="id"
                label="Target OLT (Wajib untuk ODP yang diimport)"
                variant="outlined"
                density="comfortable"
                hint="ODP yang belum terhubung ke OLT akan diarahkan ke OLT ini"
                persistent-hint
                :rules="[(v) => !!v || 'Harap pilih target OLT']"
              ></v-select>
            </v-col>
            <v-col cols="12" md="6">
              <v-select
                v-model="targetOdcId"
                :items="odcs"
                item-title="nama_odc"
                item-value="id"
                label="Target ODC (Opsional)"
                variant="outlined"
                density="comfortable"
                clearable
                hint="Pilih ODC jika ODP berada di bawah naungan kabinet tertentu"
                persistent-hint
              ></v-select>
            </v-col>
          </v-row>

          <!-- Selection Controls -->
          <div class="d-flex align-center justify-space-between mb-2">
            <div class="text-subtitle-2 font-weight-bold">
              Daftar Titik ODP untuk Diimport ({{ selectedCount }} dari {{ odpNodes.length }} terpilih)
            </div>
            <div class="d-flex gap-2">
              <v-btn size="x-small" variant="text" @click="selectAll(true)">Pilih Semua</v-btn>
              <v-btn size="x-small" variant="text" color="error" @click="selectAll(false)">Batalkan Semua</v-btn>
            </div>
          </div>

          <!-- Table of Nodes -->
          <v-table density="compact" class="border rounded-lg" height="280px" fixed-header>
            <thead>
              <tr>
                <th style="width: 50px;">
                  <v-checkbox-btn
                    :model-value="isAllSelected"
                    :indeterminate="isPartiallySelected"
                    @update:model-value="toggleAll"
                  ></v-checkbox-btn>
                </th>
                <th>Tipe</th>
                <th>Kode / Nama Placemark</th>
                <th>Kapasitas Port</th>
                <th>Latitude</th>
                <th>Longitude</th>
                <th>Keterangan</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="node in odpNodes" :key="node.id" :class="{ 'bg-blue-lighten-5': node.selected }">
                <td>
                  <v-checkbox-btn v-model="node.selected"></v-checkbox-btn>
                </td>
                <td>
                  <v-chip size="x-small" :color="getTypeColor(node.type)" variant="flat" class="text-uppercase font-weight-bold">
                    {{ node.type }}
                  </v-chip>
                </td>
                <td class="font-weight-medium">
                  <v-text-field
                    v-model="node.name"
                    density="compact"
                    variant="plain"
                    hide-details
                    class="pa-0 font-weight-bold"
                  ></v-text-field>
                </td>
                <td style="width: 100px;">
                  <v-text-field
                    v-model.number="node.kapasitas_port"
                    type="number"
                    density="compact"
                    variant="outlined"
                    hide-details
                    style="max-width: 80px;"
                  ></v-text-field>
                </td>
                <td class="text-caption font-mono">{{ node.lat.toFixed(6) }}</td>
                <td class="text-caption font-mono">{{ node.lng.toFixed(6) }}</td>
                <td class="text-caption text-truncate" style="max-width: 150px;">{{ node.description || '-' }}</td>
              </tr>
              <tr v-if="odpNodes.length === 0">
                <td colspan="7" class="text-center py-4 text-medium-emphasis">
                  Tidak ditemukan titik ODP di dalam file KMZ ini
                </td>
              </tr>
            </tbody>
          </v-table>
        </div>
      </v-card-text>

      <v-divider></v-divider>

      <!-- Actions -->
      <v-card-actions class="pa-4 bg-grey-lighten-4">
        <v-btn
          v-if="parsedResult"
          variant="tonal"
          color="secondary"
          prepend-icon="mdi-eye-outline"
          @click="previewOnMap"
        >
          Pratinjau di Peta
        </v-btn>
        <v-spacer></v-spacer>
        <v-btn variant="text" @click="close">Batal</v-btn>
        <v-btn
          v-if="parsedResult"
          color="primary"
          variant="flat"
          prepend-icon="mdi-cloud-upload"
          :loading="importing"
          :disabled="selectedCount === 0 || !targetOltId"
          class="px-6 rounded-lg font-weight-bold"
          @click="submitImport"
        >
          Import {{ selectedCount }} Titik ke Peta
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { parseKmzOrKml, type KmzParseResult, type KmzNode } from '@/utils/kmzParser';
import apiClient from '@/services/api';

const props = defineProps<{
  modelValue: boolean;
  olts: Array<{ id: number; nama_olt: string }>;
  odcs?: Array<{ id: number; nama_odc: string; kode_odc: string }>;
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', val: boolean): void;
  (e: 'imported', count: number): void;
  (e: 'preview', result: KmzParseResult): void;
}>();

const fileInputRef = ref<HTMLInputElement | null>(null);
const isDragging = ref(false);
const parsing = ref(false);
const importing = ref(false);
const errorMessage = ref('');
const parsedResult = ref<KmzParseResult | null>(null);

const targetOltId = ref<number | null>(null);
const targetOdcId = ref<number | null>(null);

const odpNodes = computed(() => {
  if (!parsedResult.value) return [];
  // Show ODPs, ODCs, or all points as importable targets
  return parsedResult.value.nodes.filter(
    (n) => n.type === 'odp' || n.type === 'odc' || n.type === 'server' || n.type === 'olt'
  );
});

const selectedCount = computed(() => {
  return odpNodes.value.filter((n) => n.selected).length;
});

const isAllSelected = computed(() => {
  return odpNodes.value.length > 0 && odpNodes.value.every((n) => n.selected);
});

const isPartiallySelected = computed(() => {
  const selected = odpNodes.value.filter((n) => n.selected).length;
  return selected > 0 && selected < odpNodes.value.length;
});

function triggerFileInput() {
  fileInputRef.value?.click();
}

function handleFileChange(event: Event) {
  const input = event.target as HTMLInputElement;
  if (input.files && input.files[0]) {
    processFile(input.files[0]);
  }
}

function handleDrop(event: DragEvent) {
  isDragging.value = false;
  if (event.dataTransfer?.files && event.dataTransfer.files[0]) {
    processFile(event.dataTransfer.files[0]);
  }
}

async function processFile(file: File) {
  const name = file.name.toLowerCase();
  if (!name.endsWith('.kmz') && !name.endsWith('.kml')) {
    errorMessage.value = 'Format file tidak didukung! Harap unggah file .kmz atau .kml';
    return;
  }

  parsing.value = true;
  errorMessage.value = '';

  try {
    const result = await parseKmzOrKml(file);
    parsedResult.value = result;

    // Set default OLT if available and not selected
    if (!targetOltId.value && props.olts && props.olts.length > 0) {
      targetOltId.value = props.olts[0].id;
    }
  } catch (err: any) {
    errorMessage.value = err.message || 'Gagal memproses file KMZ/KML';
  } finally {
    parsing.value = false;
  }
}

function resetUpload() {
  parsedResult.value = null;
  errorMessage.value = '';
  if (fileInputRef.value) fileInputRef.value.value = '';
}

function selectAll(val: boolean) {
  odpNodes.value.forEach((n) => (n.selected = val));
}

function toggleAll(val: boolean | null) {
  const target = val === true;
  selectAll(target);
}

function previewOnMap() {
  if (parsedResult.value) {
    emit('preview', parsedResult.value);
    close();
  }
}

async function submitImport() {
  if (!targetOltId.value) {
    errorMessage.value = 'Silakan pilih target OLT terlebih dahulu';
    return;
  }

  const selected = odpNodes.value.filter((n) => n.selected);
  if (selected.length === 0) return;

  importing.value = true;
  errorMessage.value = '';

  try {
    const items = selected.map((n) => ({
      kode_odp: n.name,
      alamat: n.description || `ODP ${n.name}`,
      kapasitas_port: n.kapasitas_port || 8,
      latitude: n.lat,
      longitude: n.lng,
      olt_id: targetOltId.value!,
      odc_id: targetOdcId.value || undefined,
    }));

    const response = await apiClient.post('/odp/batch-import', { items });
    const count = response.data?.count || items.length;
    emit('imported', count);
    close();
  } catch (err: any) {
    errorMessage.value = err.response?.data?.error || err.message || 'Gagal mengimport data ODP';
  } finally {
    importing.value = false;
  }
}

function close() {
  emit('update:modelValue', false);
  resetUpload();
}

function getTypeColor(type: string) {
  switch (type) {
    case 'server': return 'indigo';
    case 'olt': return 'purple';
    case 'odc': return 'warning';
    case 'odp': return 'success';
    case 'pelanggan': return 'cyan';
    default: return 'grey';
  }
}
</script>

<style scoped>
.dropzone-box {
  border: 2px dashed #93c5fd;
  background: #f8fafc;
  transition: all 0.2s ease;
}
.dropzone-box:hover,
.dropzone-active {
  border-color: #3b82f6;
  background: #eff6ff;
  transform: translateY(-2px);
}
.font-mono {
  font-family: monospace;
}
</style>
