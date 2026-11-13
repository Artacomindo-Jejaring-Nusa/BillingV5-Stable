<template>
  <v-dialog 
    :model-value="modelValue" 
    @update:modelValue="$emit('update:modelValue', $event)" 
    max-width="580px" 
    persistent
    transition="dialog-bottom-transition"
  >
    <v-card class="modal-card border" elevation="8">
      <!-- Clean Modal Header (Anti-Slop SaaS) -->
      <div class="modal-header d-flex align-center justify-space-between pa-5 border-b">
        <div class="d-flex align-center">
          <div class="modal-icon-wrapper me-3">
            <v-icon color="primary" size="22">
              {{ isEditMode ? 'mdi-pencil-outline' : 'mdi-domain-plus' }}
            </v-icon>
          </div>
          <div>
            <h2 class="modal-title mb-0">{{ formTitle }}</h2>
            <p class="modal-subtitle mb-0">
              {{ isEditMode ? 'Perbarui informasi dan konfigurasi brand provider' : 'Tambahkan brand provider baru ke dalam sistem' }}
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
            <!-- ID Brand Field -->
            <v-col cols="12">
              <div class="form-group mb-4">
                <label class="form-label">ID Brand <span class="text-error" v-if="!isEditMode">*</span></label>
                <v-text-field
                  v-model="localItem.id_brand"
                  :disabled="isEditMode"
                  placeholder="Contoh: jnt-01, art-02"
                  variant="outlined"
                  density="comfortable"
                  class="form-field"
                  :class="{ 'field-disabled': isEditMode }"
                  prepend-inner-icon="mdi-identifier"
                  hide-details="auto"
                >
                  <template v-slot:append-inner v-if="isEditMode">
                    <v-tooltip text="ID Brand bersifat permanen dan tidak dapat diubah">
                      <template v-slot:activator="{ props }">
                        <v-icon v-bind="props" color="grey" size="small">mdi-lock-outline</v-icon>
                      </template>
                    </v-tooltip>
                  </template>
                </v-text-field>
              </div>
            </v-col>

            <!-- Brand Name Field -->
            <v-col cols="12">
              <div class="form-group mb-4">
                <label class="form-label">Nama Brand Provider <span class="text-error">*</span></label>
                <v-text-field
                  v-model="localItem.brand"
                  placeholder="Masukkan nama resmi brand (contoh: Jelantik Fiber)"
                  variant="outlined"
                  density="comfortable"
                  class="form-field"
                  prepend-inner-icon="mdi-domain"
                  hide-details="auto"
                />
              </div>
            </v-col>

            <!-- Tax Field -->
            <v-col cols="12" sm="6">
              <div class="form-group mb-4">
                <label class="form-label">Pajak Pertambahan Nilai (PPN)</label>
                <v-text-field
                  v-model.number="localItem.pajak"
                  type="number"
                  placeholder="11.0"
                  variant="outlined"
                  density="comfortable"
                  class="form-field"
                  prepend-inner-icon="mdi-percent"
                  suffix="%"
                  hide-details="auto"
                />
              </div>
            </v-col>

            <!-- Xendit Key Field -->
            <v-col cols="12" sm="6">
              <div class="form-group mb-4">
                <label class="form-label">Key Name Xendit</label>
                <v-text-field
                  v-model="localItem.xendit_key_name"
                  placeholder="Contoh: xendit_live_key"
                  variant="outlined"
                  density="comfortable"
                  class="form-field"
                  prepend-inner-icon="mdi-key-variant"
                  hide-details="auto"
                />
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
          :disabled="!localItem.brand || !localItem.id_brand"
        >
          {{ isEditMode ? 'Perbarui Brand' : 'Simpan Brand' }}
        </v-btn>
      </div>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import type { HargaLayanan } from '@/interfaces/layanan';

const props = defineProps<{
  modelValue: boolean,
  editedItem: Partial<HargaLayanan>
}>();

const emit = defineEmits(['update:modelValue', 'save']);

const localItem = ref<Partial<HargaLayanan>>({});
const isEditMode = computed(() => !!localItem.value.id_brand && props.editedItem && !!props.editedItem.id_brand);
const formTitle = computed(() => isEditMode.value ? 'Edit Brand Provider' : 'Tambah Brand Provider');

watch(() => props.editedItem, (newVal) => {
  localItem.value = { ...newVal };
}, { immediate: true, deep: true });

function submit() {
  emit('save', localItem.value);
  emit('update:modelValue', false);
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

.field-disabled :deep(.v-field) {
  background: #f8fafc !important;
  color: #64748b !important;
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

.v-theme--dark .field-disabled :deep(.v-field) {
  background: #0f172a !important;
}

.v-theme--dark .btn-cancel {
  border-color: #475569 !important;
  color: #cbd5e1 !important;
}
</style>
