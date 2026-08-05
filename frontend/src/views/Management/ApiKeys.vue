<template>
  <v-container fluid class="pa-4 pa-md-6">
    <!-- Modern Header with Gradient Background -->
    <div class="header-card mb-4 mb-md-6">
      <div class="header-section">
        <div class="header-content d-flex flex-column flex-sm-row align-start align-sm-center justify-space-between gap-4">
          <div class="d-flex align-center">
            <v-avatar class="me-4 elevation-4" color="primary" size="80">
              <v-icon color="white" size="40">mdi-key-chain</v-icon>
            </v-avatar>
            <div>
              <h1 class="text-h4 font-weight-bold text-white mb-2">Integrasi API Key</h1>
              <p class="header-subtitle mb-0">
                Kelola API Key untuk mengintegrasikan Billing dengan portal pelanggan atau aplikasi pihak ketiga lainnya.
              </p>
            </div>
          </div>
          <v-btn
            color="white"
            variant="flat"
            rounded="lg"
            size="large"
            class="generate-btn text-primary font-weight-bold elevation-2"
            prepend-icon="mdi-plus"
            @click="openGenerateDialog"
          >
            Generate API Key
          </v-btn>
        </div>
      </div>
    </div>

    <!-- Info Banner -->
    <v-alert
      type="info"
      variant="tonal"
      class="mb-6 rounded-xl border-s-4"
      icon="mdi-shield-lock-outline"
      border="start"
    >
      <strong>Informasi Keamanan:</strong> Raw token hanya akan ditampilkan <strong>satu kali</strong> saat dibuat. Sistem kami hanya menyimpan hasil enkripsi satu arah (hash SHA-256) demi keamanan. Mohon simpan API Key Anda di tempat yang aman setelah dibuat.
    </v-alert>

    <!-- Main Content Card -->
    <v-card rounded="xl" elevation="2" class="main-card">
      <v-card-title class="d-flex align-center pa-6">
        <v-icon color="primary" class="me-2" size="28">mdi-key-variant</v-icon>
        <span class="text-h6 font-weight-bold">Daftar API Key Aktif</span>
        <v-spacer></v-spacer>
        <!-- Search Field -->
        <v-text-field
          v-model="search"
          prepend-inner-icon="mdi-magnify"
          label="Cari API Key"
          variant="outlined"
          density="compact"
          hide-details
          flat
          rounded="lg"
          style="max-width: 300px"
        ></v-text-field>
      </v-card-title>

      <v-divider></v-divider>

      <!-- Data Table -->
      <v-data-table
        :headers="headers"
        :items="apiKeys"
        :search="search"
        :loading="loading"
        hover
        class="api-keys-table pa-2"
        no-data-text="Tidak ada API Key yang terdaftar."
        loading-text="Memuat daftar API Key..."
      >
        <!-- Status Column -->
        <template v-slot:item.is_active="{ item }">
          <v-switch
            :model-value="item.is_active"
            color="success"
            hide-details
            inset
            density="compact"
            class="status-switch"
            :loading="statusLoadingId === item.id"
            @update:model-value="toggleKeyStatus(item)"
          ></v-switch>
        </template>

        <!-- Created At Column -->
        <template v-slot:item.created_at="{ item }">
          <span class="text-body-2 text-medium-emphasis">
            {{ formatDate(item.created_at) }}
          </span>
        </template>

        <!-- Prefix Column -->
        <template v-slot:item.prefix="{ item }">
          <v-chip size="small" variant="tonal" color="primary" class="font-mono">
            {{ item.prefix }}...
          </v-chip>
        </template>

        <!-- Role Column -->
        <template v-slot:item.role_name="{ item }">
          <v-chip size="small" variant="flat" color="blue-grey-lighten-4" class="text-blue-grey-darken-3 font-weight-medium">
            {{ item.role_name }}
          </v-chip>
        </template>

        <!-- Actions Column -->
        <template v-slot:item.actions="{ item }">
          <v-tooltip location="top" text="Hapus/Revoke API Key">
            <template v-slot:activator="{ props }">
              <v-btn
                v-bind="props"
                icon="mdi-trash-can-outline"
                variant="text"
                color="error"
                density="comfortable"
                @click="confirmDelete(item)"
              ></v-btn>
            </template>
          </v-tooltip>
        </template>
      </v-data-table>
    </v-card>

    <!-- Dialog Generate API Key -->
    <v-dialog v-model="generateDialog" max-width="500px" persistent>
      <v-card rounded="xl" class="pa-4">
        <v-card-title class="text-h5 font-weight-bold d-flex align-center">
          <v-icon color="primary" class="me-2">mdi-key-plus</v-icon>
          Generate API Key Baru
        </v-card-title>

        <v-card-text class="pt-4">
          <v-form ref="form" v-model="formValid">
            <v-text-field
              v-model="newKeyForm.name"
              label="Nama/Deskripsi Integrasi"
              placeholder="Contoh: Flutter Customer Portal"
              variant="outlined"
              rounded="lg"
              class="mb-4"
              :rules="[v => !!v || 'Nama integrasi wajib diisi']"
            ></v-text-field>

            <v-select
              v-model="newKeyForm.role_id"
              :items="roles"
              item-title="name"
              item-value="id"
              label="Hak Akses (Role)"
              placeholder="Pilih Role untuk API Key ini"
              variant="outlined"
              rounded="lg"
              :loading="rolesLoading"
              :rules="[v => !!v || 'Hak akses (Role) wajib dipilih']"
            ></v-select>
          </v-form>
        </v-card-text>

        <v-card-actions class="px-6 pb-6">
          <v-spacer></v-spacer>
          <v-btn
            variant="text"
            rounded="lg"
            @click="generateDialog = false"
            :disabled="generating"
          >
            Batal
          </v-btn>
          <v-btn
            color="primary"
            variant="flat"
            rounded="lg"
            :loading="generating"
            :disabled="!formValid"
            @click="submitGenerate"
          >
            Generate Key
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Dialog Tampilan API Key Sukses (Raw Token Show Once) -->
    <v-dialog v-model="showResultDialog" max-width="550px" persistent>
      <v-card rounded="xl" class="pa-4">
        <v-card-title class="text-h5 font-weight-bold text-success d-flex align-center">
          <v-icon color="success" class="me-2">mdi-check-circle</v-icon>
          API Key Berhasil Dibuat
        </v-card-title>

        <v-card-text class="pt-4">
          <p class="text-body-2 text-medium-emphasis mb-4">
            API Key Anda berhasil digenerate. Salin sekarang dan simpan di tempat yang aman. Kunci ini tidak akan ditampilkan lagi demi keamanan.
          </p>

          <!-- Raw Key Display Area -->
          <v-alert
            color="success"
            variant="tonal"
            class="pa-4 rounded-lg d-flex align-center justify-space-between mb-4 border-dashed"
          >
            <div class="d-flex flex-column" style="width: 85%">
              <span class="text-caption text-success font-weight-bold uppercase mb-1">Raw API Key Token</span>
              <span class="text-subtitle-1 font-mono break-all font-weight-bold">{{ generatedRawKey }}</span>
            </div>
            <v-spacer></v-spacer>
            <v-btn
              icon="mdi-content-copy"
              variant="text"
              color="success"
              @click="copyToClipboard"
            ></v-btn>
          </v-alert>

          <v-alert
            type="warning"
            variant="tonal"
            class="rounded-lg"
            icon="mdi-alert-outline"
          >
            Setelah Anda menutup jendela dialog ini, Anda tidak akan bisa melihat kunci ini lagi. Pastikan Anda telah menyalinnya!
          </v-alert>
        </v-card-text>

        <v-card-actions class="px-6 pb-6">
          <v-spacer></v-spacer>
          <v-btn
            color="success"
            variant="flat"
            rounded="lg"
            class="px-6"
            @click="closeResultDialog"
          >
            Saya Sudah Menyalinnya
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Dialog Konfirmasi Delete -->
    <v-dialog v-model="deleteDialog" max-width="450px">
      <v-card rounded="xl" class="pa-4">
        <v-card-title class="text-h5 font-weight-bold d-flex align-center">
          <v-icon color="error" class="me-2">mdi-alert-circle-outline</v-icon>
          Hapus/Revoke API Key?
        </v-card-title>

        <v-card-text class="pt-4">
          Apakah Anda yakin ingin menghapus API Key <strong>"{{ keyToDelete?.name }}"</strong>?
          <br /><br />
          Tindakan ini tidak bisa dibatalkan dan semua integrasi yang menggunakan kunci ini akan <strong>terputus seketika</strong>.
        </v-card-text>

        <v-card-actions class="px-6 pb-6">
          <v-spacer></v-spacer>
          <v-btn
            variant="text"
            rounded="lg"
            @click="deleteDialog = false"
            :disabled="deleting"
          >
            Batal
          </v-btn>
          <v-btn
            color="error"
            variant="flat"
            rounded="lg"
            :loading="deleting"
            @click="submitDelete"
          >
            Hapus Kunci
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Global Snackbar -->
    <v-snackbar
      v-model="snackbar.show"
      :color="snackbar.color"
      timeout="4000"
      location="top right"
      rounded="lg"
      elevation="4"
    >
      <v-icon start class="me-2">{{ snackbar.icon }}</v-icon>
      {{ snackbar.text }}
      <template v-slot:actions>
        <v-btn variant="text" @click="snackbar.show = false">Tutup</v-btn>
      </template>
    </v-snackbar>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import apiClient from '@/services/api';

// State Variables
const apiKeys = ref<any[]>([]);
const roles = ref<any[]>([]);
const search = ref('');
const loading = ref(false);
const rolesLoading = ref(false);
const generating = ref(false);
const deleting = ref(false);
const statusLoadingId = ref<number | null>(null);

// Dialogs Control
const generateDialog = ref(false);
const showResultDialog = ref(false);
const deleteDialog = ref(false);

// Forms
const formValid = ref(false);
const form = ref<any>(null);
const newKeyForm = ref({
  name: '',
  role_id: null as number | null
});
const generatedRawKey = ref('');
const keyToDelete = ref<any>(null);

// Snackbar
const snackbar = ref({
  show: false,
  text: '',
  color: 'success',
  icon: 'mdi-check-circle'
});

// Table Headers
const headers = [
  { title: 'Nama Integrasi', key: 'name', align: 'start' as const, sortable: true },
  { title: 'Prefix Key', key: 'prefix', align: 'start' as const, sortable: false },
  { title: 'Hak Akses (Role)', key: 'role_name', align: 'start' as const, sortable: true },
  { title: 'Status', key: 'is_active', align: 'center' as const, sortable: false },
  { title: 'Tanggal Dibuat', key: 'created_at', align: 'start' as const, sortable: true },
  { title: 'Aksi', key: 'actions', align: 'center' as const, sortable: false }
];

// Helper Functions
const showMessage = (text: string, color = 'success', icon = 'mdi-check-circle') => {
  snackbar.value = { show: true, text, color, icon };
};

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-';
  const date = new Date(dateStr);
  return date.toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
};

// API Calls
const fetchApiKeys = async () => {
  loading.value = true;
  try {
    const response = await apiClient.get('/api-keys');
    apiKeys.value = response.data.data || [];
  } catch (error: any) {
    showMessage(error.response?.data?.error || 'Gagal memuat API Key', 'error', 'mdi-alert-circle');
  } finally {
    loading.value = false;
  }
};

const fetchRoles = async () => {
  rolesLoading.value = true;
  try {
    const response = await apiClient.get('/roles');
    roles.value = response.data.data || [];
  } catch (error: any) {
    showMessage(error.response?.data?.error || 'Gagal memuat daftar role', 'error', 'mdi-alert-circle');
  } finally {
    rolesLoading.value = false;
  }
};

const openGenerateDialog = () => {
  newKeyForm.value = { name: '', role_id: null };
  if (form.value) form.value.resetValidation();
  generateDialog.value = true;
  fetchRoles();
};

const submitGenerate = async () => {
  if (!formValid.value) return;
  generating.value = true;
  try {
    const response = await apiClient.post('/api-keys', newKeyForm.value);
    generatedRawKey.value = response.data.data.raw_key;
    generateDialog.value = false;
    showResultDialog.value = true;
    fetchApiKeys();
  } catch (error: any) {
    showMessage(error.response?.data?.error || 'Gagal membuat API Key', 'error', 'mdi-alert-circle');
  } finally {
    generating.value = false;
  }
};

const closeResultDialog = () => {
  showResultDialog.value = false;
  generatedRawKey.value = '';
  showMessage('API Key berhasil disimpan!', 'success', 'mdi-check-circle');
};

const toggleKeyStatus = async (item: any) => {
  statusLoadingId.value = item.id;
  try {
    await apiClient.patch(`/api-keys/${item.id}/toggle`);
    item.is_active = !item.is_active;
    showMessage(`Status API Key "${item.name}" berhasil diubah.`, 'success', 'mdi-check-circle');
  } catch (error: any) {
    showMessage(error.response?.data?.error || 'Gagal mengubah status API Key', 'error', 'mdi-alert-circle');
  } finally {
    statusLoadingId.value = null;
  }
};

const confirmDelete = (item: any) => {
  keyToDelete.value = item;
  deleteDialog.value = true;
};

const submitDelete = async () => {
  if (!keyToDelete.value) return;
  deleting.value = true;
  try {
    await apiClient.delete(`/api-keys/${keyToDelete.value.id}`);
    deleteDialog.value = false;
    showMessage(`API Key "${keyToDelete.value.name}" berhasil dihapus/revoked.`, 'success', 'mdi-check-circle');
    fetchApiKeys();
  } catch (error: any) {
    showMessage(error.response?.data?.error || 'Gagal menghapus API Key', 'error', 'mdi-alert-circle');
  } finally {
    deleting.value = false;
    keyToDelete.value = null;
  }
};

const copyToClipboard = () => {
  if (!generatedRawKey.value) return;
  navigator.clipboard.writeText(generatedRawKey.value);
  showMessage('Kunci disalin ke clipboard!', 'info', 'mdi-content-copy');
};

// Lifecycle Hooks
onMounted(() => {
  fetchApiKeys();
});
</script>

<style scoped>
.font-mono {
  font-family: monospace, Courier, monospace;
}

.header-card {
  background: linear-gradient(135deg, #6366f1 0%, #a855f7 100%);
  border-radius: 24px;
  overflow: hidden;
}

.header-section {
  padding: 40px;
  position: relative;
}

.header-subtitle {
  color: rgba(255, 255, 255, 0.85);
  font-size: 1.05rem;
}

.main-card {
  overflow: hidden;
}

.break-all {
  word-break: break-all;
}

/* Animations */
@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.animate-pulse {
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}
</style>
