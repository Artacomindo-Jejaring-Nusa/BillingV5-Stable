<template>
  <v-dialog 
    :model-value="modelValue" 
    @update:modelValue="$emit('update:modelValue', $event)" 
    max-width="600px" 
    persistent
    transition="dialog-bottom-transition"
  >
    <v-card class="modal-card border" elevation="8">
      <!-- Clean Modal Header (Anti-Slop SaaS) -->
      <div class="modal-header d-flex align-center justify-space-between pa-5 border-b">
        <div class="d-flex align-center">
          <div class="modal-icon-wrapper me-3">
            <v-icon color="primary" size="22">
              {{ isEditMode ? 'mdi-package-variant' : 'mdi-package-variant-plus' }}
            </v-icon>
          </div>
          <div>
            <h2 class="modal-title mb-0">{{ formTitle }}</h2>
            <p class="modal-subtitle mb-0">
              {{ isEditMode ? 'Ubah parameter dan tarif paket layanan' : 'Buat paket layanan internet baru untuk brand ini' }}
            </p>
          </div>
        </div>
        <v-btn
          icon="mdi-close"
          variant="text"
          size="small"
          density="comfortable"
          class="close-btn"
          @click="$emit('update:modelValue', false)"
        ></v-btn>
      </div>

      <!-- Form Content -->
      <v-card-text class="pa-5 pa-sm-6">
        <v-container fluid class="pa-0">
          <v-row dense>
            <!-- Package Name Field -->
            <v-col cols="12">
              <div class="form-group mb-4">
                <label class="form-label">Nama Paket Layanan <span class="text-error">*</span></label>
                <v-text-field
                  v-model="localItem.nama_paket"
                  placeholder="Contoh: Paket Home Regular, Business Ultra"
                  variant="outlined"
                  density="comfortable"
                  class="form-field"
                  prepend-inner-icon="mdi-package-variant-closed"
                  hide-details="auto"
                />
              </div>
            </v-col>

            <!-- Speed Field -->
            <v-col cols="12" sm="6">
              <div class="form-group mb-4">
                <label class="form-label">Bandwidth / Kecepatan <span class="text-error">*</span></label>
                <v-text-field
                  v-model.number="localItem.kecepatan"
                  type="number"
                  placeholder="Contoh: 20, 50, 100"
                  variant="outlined"
                  density="comfortable"
                  class="form-field"
                  prepend-inner-icon="mdi-speedometer"
                  suffix="Mbps"
                  hide-details="auto"
                >
                  <template v-slot:append-inner v-if="localItem.kecepatan">
                    <span class="speed-badge-pill">
                      {{ getSpeedLabel(localItem.kecepatan) }}
                    </span>
                  </template>
                </v-text-field>
              </div>
            </v-col>

            <!-- Price Field -->
            <v-col cols="12" sm="6">
              <div class="form-group mb-4">
                <label class="form-label">Tarif Bulanan (IDR) <span class="text-error">*</span></label>
                <v-text-field
                  v-model.number="localItem.harga"
                  type="number"
                  placeholder="Contoh: 150000"
                  variant="outlined"
                  density="comfortable"
                  class="form-field"
                  prepend-inner-icon="mdi-cash"
                  prefix="Rp"
                  hide-details="auto"
                />
              </div>
            </v-col>

            <!-- Package Preview Summary Card -->
            <v-col cols="12" v-if="localItem.nama_paket || localItem.kecepatan || localItem.harga">
              <div class="preview-box pa-4 rounded-lg border mt-1">
                <div class="d-flex align-center justify-space-between mb-3 border-b pb-2">
                  <div class="d-flex align-center gap-2">
                    <v-icon size="16" color="primary">mdi-eye-outline</v-icon>
                    <span class="text-caption font-weight-bold text-uppercase preview-label">Ringkasan Paket</span>
                  </div>
                  <span v-if="localItem.kecepatan" class="preview-speed-chip">
                    {{ localItem.kecepatan }} Mbps ({{ getSpeedLabel(localItem.kecepatan) }})
                  </span>
                </div>

                <div class="d-flex flex-column gap-2">
                  <div class="d-flex align-center justify-space-between">
                    <span class="text-caption text-medium-emphasis">Nama Paket:</span>
                    <span class="text-body-2 font-weight-bold">{{ localItem.nama_paket || '-' }}</span>
                  </div>
                  <div class="d-flex align-center justify-space-between">
                    <span class="text-caption text-medium-emphasis">Harga Bulanan:</span>
                    <span class="text-subtitle-1 font-weight-bold preview-price">
                      {{ localItem.harga ? formatCurrency(localItem.harga) : 'Rp 0' }} <span class="text-caption text-medium-emphasis font-weight-normal">/bulan</span>
                    </span>
                  </div>
                </div>
              </div>
            </v-col>
          </v-row>
        </v-container>
      </v-card-text>

      <!-- Modal Actions Footer -->
      <div class="modal-footer d-flex align-center justify-end pa-4 px-6 border-t gap-2">
        <v-btn
          variant="outlined"
          color="secondary"
          class="text-none font-weight-medium btn-cancel"
          @click="$emit('update:modelValue', false)"
        >
          Batal
        </v-btn>
        <v-btn
          color="primary"
          class="text-none font-weight-bold btn-submit"
          elevation="0"
          @click="submit"
          :disabled="!isFormValid"
        >
          {{ isEditMode ? 'Perbarui Paket' : 'Simpan Paket' }}
        </v-btn>
      </div>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import type { PaketLayanan } from '@/interfaces/layanan';

const props = defineProps<{
  modelValue: boolean,
  editedItem: Partial<PaketLayanan>,
  brandId?: string
}>();

const emit = defineEmits(['update:modelValue', 'save']);

const localItem = ref<Partial<PaketLayanan>>({});

const isEditMode = computed(() => !!localItem.value.id);
const formTitle = computed(() => isEditMode.value ? 'Edit Paket Layanan' : 'Tambah Paket Layanan');

const isFormValid = computed(() => {
  return !!(localItem.value.nama_paket &&
         localItem.value.kecepatan &&
         localItem.value.harga);
});

watch(() => props.editedItem, (newVal) => {
  localItem.value = { ...newVal };
}, { immediate: true, deep: true });

watch(() => props.brandId, (newBrandId) => {
  if (newBrandId && !localItem.value.id_brand) {
    localItem.value.id_brand = newBrandId.trim();
  }
}, { immediate: true });

function submit() {
  const itemToSave = { ...localItem.value };
  if (itemToSave.id_brand) {
    itemToSave.id_brand = itemToSave.id_brand.trim();
  }
  emit('save', itemToSave);
  emit('update:modelValue', false);
}

function getSpeedColor(speed: number | undefined): string {
  if (!speed) return 'grey';
  if (speed < 25) return 'orange';
  if (speed < 100) return 'blue';
  if (speed < 500) return 'green';
  return 'purple';
}

function getSpeedLabel(speed: number | undefined): string {
  if (!speed) return '';
  if (speed < 25) return 'Basic';
  if (speed < 100) return 'Fast';
  if (speed < 500) return 'Ultra';
  return 'Extreme';
}

function formatCurrency(value: number | undefined): string {
  if (!value || isNaN(value)) return "Rp 0";
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0
  }).format(value);
}
</script>

<style scoped>
.gap-2 {
  gap: 0.5rem;
}

.modal-card {
  border-radius: 12px !important;
  background: #ffffff;
  overflow: hidden;
}

.modal-header {
  background: #ffffff;
}

.modal-icon-wrapper {
  width: 42px;
  height: 42px;
  border-radius: 10px;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.modal-title {
  font-size: 1.125rem;
  font-weight: 700;
  color: #0f172a;
  line-height: 1.25;
}

.modal-subtitle {
  font-size: 0.8125rem;
  color: #64748b;
  margin-top: 2px;
}

.close-btn {
  color: #64748b !important;
  border-radius: 6px !important;
}

.close-btn:hover {
  background: #f1f5f9 !important;
  color: #0f172a !important;
}

/* Form Styling */
.form-group {
  position: relative;
}

.form-label {
  display: block;
  font-size: 0.8125rem;
  font-weight: 600;
  color: #334155;
  margin-bottom: 6px;
}

.speed-badge-pill {
  font-size: 0.6875rem;
  font-weight: 700;
  color: #1d4ed8;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  padding: 2px 6px;
  border-radius: 4px;
}

/* Preview Box */
.preview-box {
  background: #f8fafc;
  border-color: #e2e8f0 !important;
}

.preview-label {
  color: #64748b;
  letter-spacing: 0.05em;
}

.preview-speed-chip {
  font-size: 0.75rem;
  font-weight: 600;
  color: #1d4ed8;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  padding: 2px 8px;
  border-radius: 6px;
}

.preview-price {
  color: #047857;
  font-variant-numeric: tabular-nums;
}

/* Footer Buttons */
.modal-footer {
  background: #f8fafc;
}

.btn-cancel {
  border-color: #cbd5e1 !important;
  color: #475569 !important;
  border-radius: 8px !important;
  padding: 0 16px !important;
  height: 38px !important;
}

.btn-cancel:hover {
  background: #f1f5f9 !important;
  border-color: #94a3b8 !important;
}

.btn-submit {
  background-color: #0f172a !important;
  color: #ffffff !important;
  border-radius: 8px !important;
  padding: 0 20px !important;
  height: 38px !important;
}

.btn-submit:hover {
  background-color: #1e293b !important;
}

/* Dark Mode Adjustments */
.v-theme--dark .modal-card,
.v-theme--dark .modal-header {
  background: #1e293b !important;
  color: #f8fafc !important;
}

.v-theme--dark .modal-footer {
  background: #0f172a !important;
  border-color: #334155 !important;
}

.v-theme--dark .modal-title {
  color: #f8fafc !important;
}

.v-theme--dark .form-label {
  color: #cbd5e1 !important;
}

.v-theme--dark .modal-icon-wrapper {
  background: #0f172a !important;
  border-color: #334155 !important;
}

.v-theme--dark .preview-box {
  background: #0f172a !important;
  border-color: #334155 !important;
}

.v-theme--dark .btn-cancel {
  border-color: #475569 !important;
  color: #cbd5e1 !important;
}
</style>
