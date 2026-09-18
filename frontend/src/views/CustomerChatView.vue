<template>
  <v-container fluid class="customer-chat-container pa-2 pa-md-4">
    <!-- Header Banner -->
    <v-card class="chat-header-banner mb-3 elevation-1 rounded-lg" color="surface">
      <div class="d-flex align-center justify-space-between px-4 py-3">
        <div class="d-flex align-center">
          <v-avatar color="primary" variant="tonal" size="44" class="me-3">
            <v-icon size="24" color="primary">mdi-forum-outline</v-icon>
          </v-avatar>
          <div>
            <div class="d-flex align-center gap-2">
              <h2 class="text-h6 font-weight-bold mb-0">Live Chat Pelanggan</h2>
              <v-chip
                :color="isWsConnected ? 'success' : 'grey'"
                size="x-small"
                variant="flat"
                class="font-weight-bold px-2"
              >
                <v-icon start size="10" :class="{ 'animate-pulse': isWsConnected }">mdi-circle</v-icon>
                {{ isWsConnected ? 'WebSocket Online' : 'Connecting...' }}
              </v-chip>
            </div>
            <p class="text-caption text-medium-emphasis mb-0">
              Pusat interaksi real-time pelanggan Jakinet, Jelantik, & Jelantik Nagrak
            </p>
          </div>
        </div>

        <!-- Quick Summary Badges -->
        <div class="d-none d-sm-flex align-center gap-2">
          <v-chip
            v-if="totalUnreadCount > 0"
            color="error"
            size="small"
            variant="elevated"
            class="font-weight-bold"
          >
            <v-icon start size="14">mdi-bell-ring</v-icon>
            {{ totalUnreadCount }} Pesan Baru
          </v-chip>
          <v-btn
            icon="mdi-refresh"
            variant="text"
            size="small"
            :loading="isLoadingRooms"
            @click="fetchRooms"
            title="Muat Ulang Room"
          ></v-btn>
        </div>
      </div>
    </v-card>

    <!-- Main Workspace (3-Panel WhatsApp Web Style) -->
    <v-card class="chat-workspace elevation-2 rounded-lg" color="surface">
      <v-row no-gutters class="fill-height">
        <!-- ================= PANEL 1: ROOMS LIST ================= -->
        <v-col
          cols="12"
          md="4"
          lg="3"
          class="rooms-sidebar border-e d-flex flex-column fill-height"
          :class="{ 'd-none d-md-flex': activeRoom && isMobile }"
        >
          <!-- Search Box -->
          <div class="pa-3 pb-2">
            <v-text-field
              v-model="searchQuery"
              placeholder="Cari nama, no telp, brand..."
              prepend-inner-icon="mdi-magnify"
              variant="outlined"
              density="compact"
              hide-details
              clearable
              rounded="lg"
              class="search-input"
              @update:model-value="onSearchDebounced"
            ></v-text-field>
          </div>

          <!-- Brand Filter Chips -->
          <div class="px-3 pb-2 filter-chips-scroll d-flex gap-1 overflow-x-auto">
            <v-chip
              v-for="filter in brandFilters"
              :key="filter.value"
              :color="selectedBrand === filter.value ? filter.color : undefined"
              :variant="selectedBrand === filter.value ? 'flat' : 'outlined'"
              size="x-small"
              class="font-weight-bold cursor-pointer"
              @click="setBrandFilter(filter.value)"
            >
              {{ filter.label }}
            </v-chip>
          </div>

          <v-divider></v-divider>

          <!-- Rooms List -->
          <div class="rooms-list-scroll flex-grow-1 overflow-y-auto">
            <div v-if="isLoadingRooms" class="pa-6 text-center">
              <v-progress-circular indeterminate color="primary" size="32"></v-progress-circular>
              <div class="text-caption text-medium-emphasis mt-2">Memuat daftar chat...</div>
            </div>

            <div v-else-if="filteredRooms.length === 0" class="pa-8 text-center text-medium-emphasis">
              <v-icon size="48" color="grey-lighten-1" class="mb-2">mdi-message-text-outline</v-icon>
              <div class="text-body-2 font-weight-medium">Tidak ada percakapan</div>
              <div class="text-caption">Belum ada pesan yang cocok dengan filter</div>
            </div>

            <v-list v-else lines="two" class="pa-0">
              <v-list-item
                v-for="room in filteredRooms"
                :key="room.id"
                :active="activeRoom?.id === room.id"
                class="room-item px-3 py-2 border-b"
                @click="selectRoom(room)"
              >
                <!-- Customer Avatar with Brand Color -->
                <template v-slot:prepend>
                  <v-avatar :color="getBrandColor(room.brand)" size="42" class="elevation-1">
                    <span class="text-subtitle-2 font-weight-bold text-white">
                      {{ getInitials(room.pelanggan?.nama || 'Pelanggan') }}
                    </span>
                  </v-avatar>
                </template>

                <!-- Room Info -->
                <v-list-item-title class="d-flex align-center justify-space-between mb-1">
                  <span class="font-weight-bold text-truncate text-body-2 text-high-emphasis">
                    {{ room.pelanggan?.nama || 'Pelanggan #' + room.pelanggan_id }}
                  </span>
                  <span class="text-caption text-medium-emphasis" style="font-size: 0.72rem;">
                    {{ formatTimestamp(room.last_message_at) }}
                  </span>
                </v-list-item-title>

                <v-list-item-subtitle class="d-flex flex-column gap-1">
                  <!-- Brand Badge & Phone -->
                  <div class="d-flex align-center gap-1">
                    <v-chip
                      :color="getBrandColor(room.brand)"
                      size="x-small"
                      variant="flat"
                      class="px-1 font-weight-bold text-white"
                      style="font-size: 0.65rem; height: 18px;"
                    >
                      {{ normalizeBrandName(room.brand) }}
                    </v-chip>

                    <span v-if="room.pelanggan?.no_telp" class="text-caption text-medium-emphasis">
                      📞 {{ room.pelanggan.no_telp }}
                    </span>
                  </div>

                  <!-- Last message snippet & unread badge -->
                  <div class="d-flex align-center justify-space-between mt-1">
                    <span class="text-caption text-truncate text-medium-emphasis flex-grow-1">
                      {{ room.last_message_text || 'Mulai percakapan...' }}
                    </span>
                    <v-badge
                      v-if="room.unread_count_admin > 0"
                      :content="room.unread_count_admin"
                      color="error"
                      inline
                      class="ms-2"
                    ></v-badge>
                  </div>
                </v-list-item-subtitle>
              </v-list-item>
            </v-list>
          </div>
        </v-col>

        <!-- ================= PANEL 2: ACTIVE CHAT CONVERSATION ================= -->
        <v-col
          cols="12"
          :md="showInfoPanel ? 5 : 8"
          :lg="showInfoPanel ? 6 : 9"
          class="chat-conversation-panel d-flex flex-column fill-height"
          :class="{ 'd-none d-md-flex': !activeRoom && isMobile }"
        >
          <!-- Empty State When No Room is Selected -->
          <div v-if="!activeRoom" class="d-flex flex-column align-center justify-center fill-height pa-8 text-center">
            <div class="empty-chat-icon-wrap mb-4">
              <v-icon size="80" color="primary" class="opacity-40">mdi-chat-processing-outline</v-icon>
            </div>
            <h3 class="text-h6 font-weight-bold text-high-emphasis">Pusat Layanan Chat Pelanggan</h3>
            <p class="text-body-2 text-medium-emphasis max-w-sm mt-2">
              Pilih salah satu percakapan di sebelah kiri untuk mulai membaca dan membalas pesan pelanggan secara langsung.
            </p>
          </div>

          <!-- Active Room View -->
          <template v-else>
            <!-- Chat Room Header -->
            <div class="chat-room-header pa-3 border-b bg-surface d-flex align-center justify-space-between elevation-1">
              <div class="d-flex align-center gap-3">
                <!-- Back button on mobile -->
                <v-btn
                  v-if="isMobile"
                  icon="mdi-arrow-left"
                  variant="text"
                  size="small"
                  @click="activeRoom = null"
                  class="me-1"
                ></v-btn>

                <!-- Avatar -->
                <v-avatar :color="getBrandColor(activeRoom.brand)" size="42" class="elevation-1">
                  <span class="text-subtitle-2 font-weight-bold text-white">
                    {{ getInitials(activeRoom.pelanggan?.nama || 'Pelanggan') }}
                  </span>
                </v-avatar>

                <!-- Customer Details Header -->
                <div>
                  <div class="d-flex align-center gap-2">
                    <h3 class="text-subtitle-1 font-weight-bold mb-0 text-high-emphasis">
                      {{ activeRoom.pelanggan?.nama || 'Pelanggan #' + activeRoom.pelanggan_id }}
                    </h3>

                    <!-- BRAND BADGE (JAKINET / JELANTIK / JELANTIK NAGRAK) -->
                    <v-chip
                      :color="getBrandColor(activeRoom.brand)"
                      size="small"
                      variant="flat"
                      class="font-weight-bold text-white px-2"
                    >
                      <v-icon start size="12">mdi-shield-check</v-icon>
                      {{ normalizeBrandName(activeRoom.brand) }}
                    </v-chip>
                  </div>

                  <!-- Phone Number with Click-To-WhatsApp & Copy -->
                  <div class="d-flex align-center gap-2 mt-1">
                    <span class="text-caption font-weight-medium text-medium-emphasis">
                      <v-icon size="14" color="medium-emphasis" class="me-1">mdi-phone</v-icon>
                      {{ activeRoom.pelanggan?.no_telp || '-' }}
                    </span>

                    <v-tooltip location="bottom" text="Kirim pesan WhatsApp resmi">
                      <template v-slot:activator="{ props }">
                        <v-btn
                          v-bind="props"
                          v-if="activeRoom.pelanggan?.no_telp"
                          icon="mdi-whatsapp"
                          variant="text"
                          color="success"
                          density="compact"
                          size="small"
                          @click="openWhatsApp(activeRoom.pelanggan.no_telp)"
                        ></v-btn>
                      </template>
                    </v-tooltip>

                    <v-tooltip location="bottom" text="Salin nomor telepon">
                      <template v-slot:activator="{ props }">
                        <v-btn
                          v-bind="props"
                          v-if="activeRoom.pelanggan?.no_telp"
                          icon="mdi-content-copy"
                          variant="text"
                          density="compact"
                          size="small"
                          @click="copyToClipboard(activeRoom.pelanggan.no_telp)"
                        ></v-btn>
                      </template>
                    </v-tooltip>

                    <span v-if="isCustomerTyping" class="text-caption text-primary font-italic animate-pulse">
                      • sedang mengetik...
                    </span>
                  </div>
                </div>
              </div>

              <!-- Header Action Buttons -->
              <div class="d-flex align-center gap-1">
                <v-btn
                  icon="mdi-information-outline"
                  variant="text"
                  :color="showInfoPanel ? 'primary' : 'medium-emphasis'"
                  @click="showInfoPanel = !showInfoPanel"
                  title="Lihat Detail Profil Pelanggan 360"
                ></v-btn>
              </div>
            </div>

            <!-- Messages Stream Area -->
            <div ref="messagesScrollContainer" class="messages-container flex-grow-1 pa-4 overflow-y-auto">
              <div v-if="isLoadingMessages" class="text-center pa-8">
                <v-progress-circular indeterminate color="primary"></v-progress-circular>
                <div class="text-caption text-medium-emphasis mt-2">Memuat riwayat obrolan...</div>
              </div>

              <div v-else-if="activeMessages.length === 0" class="text-center pa-8 text-medium-emphasis">
                <v-icon size="48" color="grey-lighten-2" class="mb-2">mdi-chat-plus-outline</v-icon>
                <div class="text-body-2">Belum ada pesan dalam room ini</div>
                <div class="text-caption">Ketik pesan di bawah untuk memulai obrolan dengan pelanggan</div>
              </div>

              <!-- Messages List -->
              <div v-else class="d-flex flex-column gap-2">
                <template v-for="(msg, idx) in activeMessages" :key="msg.id || msg.temp_id || idx">
                  <!-- Date Pill Separator -->
                  <div v-if="shouldShowDateHeader(idx)" class="text-center my-3">
                    <span class="date-pill px-3 py-1 rounded-pill text-caption font-weight-medium bg-grey-lighten-3 text-grey-darken-2">
                      {{ formatDateHeader(msg.created_at) }}
                    </span>
                  </div>

                  <!-- Message Bubble -->
                  <div
                    class="message-row d-flex"
                    :class="msg.sender_type === 'admin' ? 'justify-end' : 'justify-start'"
                  >
                    <div
                      class="message-bubble rounded-lg pa-3 elevation-1"
                      :class="msg.sender_type === 'admin' ? 'bubble-admin bg-primary text-white' : 'bubble-customer bg-white border text-grey-darken-4'"
                    >
                      <!-- Sender Header -->
                      <div class="d-flex align-center justify-space-between gap-2 mb-1">
                        <span
                          class="sender-name font-weight-bold text-caption"
                          :class="msg.sender_type === 'admin' ? 'text-blue-lighten-4' : 'text-primary'"
                        >
                          {{ msg.sender_type === 'admin' ? 'CS Artacom' : (activeRoom.pelanggan?.nama || 'Pelanggan') }}
                        </span>
                      </div>

                      <!-- Text Message Body -->
                      <div class="message-text text-body-2" style="white-space: pre-wrap; word-break: break-word;">
                        {{ msg.message }}
                      </div>

                      <!-- Footer: Timestamp & WhatsApp Checkmarks -->
                      <div class="message-footer d-flex align-center justify-end gap-1 mt-1">
                        <span
                          class="timestamp text-caption"
                          :class="msg.sender_type === 'admin' ? 'text-blue-lighten-4' : 'text-medium-emphasis'"
                          style="font-size: 0.68rem;"
                        >
                          {{ formatTime(msg.created_at) }}
                        </span>

                        <!-- Checkmark Status (Khusus Pesan CS) -->
                        <template v-if="msg.sender_type === 'admin'">
                          <!-- 🕒 Pending -->
                          <v-icon
                            v-if="msg.status === 'pending'"
                            size="12"
                            color="blue-lighten-3"
                            title="Sedang dikirim..."
                          >mdi-clock-outline</v-icon>

                          <!-- ✔️ Sent (Ceklis 1) -->
                          <v-icon
                            v-else-if="msg.status === 'sent'"
                            size="14"
                            color="blue-lighten-3"
                            title="Terkirim ke server (Ceklis 1)"
                          >mdi-check</v-icon>

                          <!-- ✔️✔️ Delivered (Ceklis 2 Abu-abu) -->
                          <v-icon
                            v-else-if="msg.status === 'delivered'"
                            size="15"
                            color="blue-lighten-3"
                            title="Terkirim ke perangkat pelanggan (Ceklis 2 Abu-abu)"
                          >mdi-check-all</v-icon>

                          <!-- ✔️✔️ Read (Ceklis 2 Biru WhatsApp) -->
                          <v-icon
                            v-else-if="msg.status === 'read'"
                            size="15"
                            color="cyan-accent-2"
                            title="Telah dibaca oleh pelanggan (Ceklis 2 Biru)"
                          >mdi-check-all</v-icon>
                        </template>
                      </div>
                    </div>
                  </div>
                </template>
              </div>
            </div>

            <!-- Quick Template Replies Bar -->
            <div class="quick-replies-bar px-3 py-1 bg-surface border-t d-flex align-center gap-1 overflow-x-auto">
              <span class="text-caption text-medium-emphasis font-weight-bold me-1">Cepat:</span>
              <v-chip
                v-for="(tpl, tIdx) in quickTemplates"
                :key="tIdx"
                size="x-small"
                variant="outlined"
                color="primary"
                class="cursor-pointer"
                @click="useTemplate(tpl)"
              >
                {{ tpl.label }}
              </v-chip>
            </div>

            <!-- Bottom Chat Input Bar -->
            <div class="chat-input-bar pa-3 bg-surface border-t d-flex align-center gap-2">
              <v-textarea
                v-model="inputMessage"
                rows="1"
                auto-grow
                max-rows="4"
                density="compact"
                variant="outlined"
                rounded="lg"
                placeholder="Ketik balasan CS... (Tekan Enter untuk kirim)"
                hide-details
                @keydown.enter.exact.prevent="sendAdminMessage"
                @input="notifyAdminTyping"
              ></v-textarea>

              <v-btn
                color="primary"
                icon="mdi-send"
                elevation="2"
                :disabled="!inputMessage.trim()"
                @click="sendAdminMessage"
                title="Kirim Pesan"
              ></v-btn>
            </div>
          </template>
        </v-col>

        <!-- ================= PANEL 3: CONTACT 360 (CUSTOMER PROFILE) ================= -->
        <v-col
          v-if="showInfoPanel && activeRoom"
          cols="12"
          md="3"
          lg="3"
          class="customer-info-panel border-s pa-4 bg-surface overflow-y-auto fill-height"
        >
          <div class="d-flex align-center justify-space-between mb-3">
            <h4 class="text-subtitle-2 font-weight-bold text-high-emphasis">Profil Pelanggan 360</h4>
            <v-btn
              icon="mdi-close"
              variant="text"
              size="x-small"
              @click="showInfoPanel = false"
            ></v-btn>
          </div>

          <!-- Customer Identity Card -->
          <v-card variant="outlined" class="pa-3 rounded-lg mb-3">
            <div class="text-caption text-medium-emphasis">Nama Lengkap</div>
            <div class="font-weight-bold text-body-2 mb-2">
              {{ activeRoom.pelanggan?.nama || '-' }}
            </div>

            <div class="text-caption text-medium-emphasis">Customer ID</div>
            <div class="font-weight-bold text-body-2 mb-2">
              {{ activeRoom.pelanggan?.customer_id || 'ID #' + activeRoom.pelanggan_id }}
            </div>

            <div class="text-caption text-medium-emphasis">Brand Layanan</div>
            <div class="mb-2">
              <v-chip :color="getBrandColor(activeRoom.brand)" size="small" variant="flat" class="font-weight-bold text-white">
                {{ normalizeBrandName(activeRoom.brand) }}
              </v-chip>
            </div>

            <div class="text-caption text-medium-emphasis">Nomor Telepon</div>
            <div class="font-weight-bold text-body-2 mb-2 d-flex align-center gap-1">
              <span>{{ activeRoom.pelanggan?.no_telp || '-' }}</span>
              <v-btn
                v-if="activeRoom.pelanggan?.no_telp"
                icon="mdi-content-copy"
                size="x-small"
                variant="text"
                @click="copyToClipboard(activeRoom.pelanggan.no_telp)"
              ></v-btn>
            </div>

            <div class="text-caption text-medium-emphasis">Alamat / Lokasi</div>
            <div class="text-body-2">
              {{ activeRoom.pelanggan?.alamat || '-' }}
              <span v-if="activeRoom.pelanggan?.blok || activeRoom.pelanggan?.unit">
                (Blok {{ activeRoom.pelanggan?.blok || '-' }} / Unit {{ activeRoom.pelanggan?.unit || '-' }})
              </span>
            </div>
          </v-card>

          <!-- Subscription & Technical Info -->
          <v-card variant="outlined" class="pa-3 rounded-lg mb-3">
            <div class="text-subtitle-2 font-weight-bold mb-2">Status Layanan FTTH</div>

            <div class="text-caption text-medium-emphasis">Paket Layanan</div>
            <div class="font-weight-bold text-body-2 mb-2 text-primary">
              {{ getCustomerPackage(activeRoom.pelanggan) }}
            </div>

            <div class="text-caption text-medium-emphasis">Status Langganan</div>
            <div class="mb-2">
              <v-chip
                :color="getCustomerStatus(activeRoom.pelanggan) === 'Aktif' ? 'success' : 'error'"
                size="x-small"
                variant="flat"
                class="font-weight-bold"
              >
                {{ getCustomerStatus(activeRoom.pelanggan) }}
              </v-chip>
            </div>

            <div v-if="activeRoom.pelanggan?.data_teknis?.id_pelanggan" class="text-caption text-medium-emphasis">Username PPPoE</div>
            <div v-if="activeRoom.pelanggan?.data_teknis?.id_pelanggan" class="font-weight-bold text-body-2 mb-2 font-mono">
              {{ activeRoom.pelanggan.data_teknis.id_pelanggan }}
            </div>
          </v-card>

          <!-- Actions -->
          <v-btn
            block
            color="success"
            prepend-icon="mdi-whatsapp"
            class="text-none mb-2 font-weight-bold"
            @click="openWhatsApp(activeRoom.pelanggan?.no_telp)"
          >
            Hubungi via WhatsApp Web
          </v-btn>
          <v-btn
            block
            variant="outlined"
            prepend-icon="mdi-account-search"
            class="text-none"
            @click="navigateToCustomer(activeRoom.pelanggan_id)"
          >
            Buka di Data Pelanggan
          </v-btn>
        </v-col>
      </v-row>
    </v-card>

    <!-- Global Snackbar -->
    <v-snackbar v-model="snackbar.show" :color="snackbar.color" :timeout="3000" location="top right">
      {{ snackbar.text }}
    </v-snackbar>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue';
import { useDisplay } from 'vuetify';
import { useRouter } from 'vue-router';
import apiClient from '@/services/api';
import { useAuthStore } from '@/stores/auth';

const router = useRouter();
const { mobile } = useDisplay();
const authStore = useAuthStore();

const isMobile = computed(() => mobile.value);

// --- State ---
const rooms = ref<any[]>([]);
const activeRoom = ref<any | null>(null);
const activeMessages = ref<any[]>([]);
const isLoadingRooms = ref(false);
const isLoadingMessages = ref(false);
const searchQuery = ref('');
const selectedBrand = ref('ALL');
const showInfoPanel = ref(true);
const inputMessage = ref('');
const isCustomerTyping = ref(false);
const isWsConnected = ref(false);

const messagesScrollContainer = ref<HTMLElement | null>(null);

let ws: WebSocket | null = null;
let heartbeatTimer: any = null;
let typingClearTimer: any = null;
let searchDebounceTimer: any = null;

// Snackbar notification
const snackbar = ref({
  show: false,
  text: '',
  color: 'success',
});

// Brand filter options
const brandFilters = [
  { label: 'Semua', value: 'ALL', color: 'primary' },
  { label: 'JAKINET', value: 'JAKINET', color: 'error' },
  { label: 'JELANTIK', value: 'JELANTIK', color: 'primary' },
  { label: 'JELANTIK NAGRAK', value: 'JELANTIK NAGRAK', color: 'success' },
];

// Quick templates
const quickTemplates = [
  { label: '👋 Salam', text: 'Halo, selamat datang di layanan Customer Care Artacom. Ada yang bisa kami bantu?' },
  { label: '⏳ Cek Teknis', text: 'Baik pak/bu, mohon ditunggu sebentar ya. Sedang kami lakukan pengecekan ke tim teknis lapangan.' },
  { label: '✅ Lunas/Aktif', text: 'Terima kasih atas konfirmasinya. Tagihan Anda telah terverifikasi dan layanan internet sudah aktif normal kembali.' },
  { label: '📸 Foto Modem', text: 'Bisa tolong difotokan lampu indikator (PON / LOS / Internet) yang menyala pada perangkat modem router Anda?' },
];

// Computed unread total
const totalUnreadCount = computed(() => {
  return rooms.value.reduce((acc, r) => acc + (r.unread_count_admin || 0), 0);
});

// Filtered rooms
const filteredRooms = computed(() => {
  let list = rooms.value;

  // Filter Brand
  if (selectedBrand.value !== 'ALL') {
    list = list.filter((r) => normalizeBrandName(r.brand) === selectedBrand.value);
  }

  // Search Filter
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase().trim();
    list = list.filter((r) => {
      const name = (r.pelanggan?.nama || '').toLowerCase();
      const phone = (r.pelanggan?.no_telp || '').toLowerCase();
      const brand = (r.brand || '').toLowerCase();
      const lastMsg = (r.last_message_text || '').toLowerCase();
      return name.includes(q) || phone.includes(q) || brand.includes(q) || lastMsg.includes(q);
    });
  }

  return list;
});

// Brand helpers
function normalizeBrandName(brandStr?: string): string {
  if (!brandStr) return 'JAKINET';
  const b = brandStr.toUpperCase();
  if (b.includes('NAGRAK')) return 'JELANTIK NAGRAK';
  if (b.includes('JELANTIK')) return 'JELANTIK';
  return 'JAKINET';
}

function getBrandColor(brandStr?: string): string {
  const norm = normalizeBrandName(brandStr);
  if (norm === 'JELANTIK NAGRAK') return 'success'; // Emerald Green
  if (norm === 'JELANTIK') return 'primary'; // Deep Blue
  return 'error'; // Jakinet Red
}

function getInitials(name: string): string {
  if (!name) return 'P';
  const parts = name.trim().split(' ');
  if (parts.length >= 2) {
    return (parts[0][0] + parts[1][0]).toUpperCase();
  }
  return name.substring(0, 2).toUpperCase();
}

function setBrandFilter(brandVal: string) {
  selectedBrand.value = brandVal;
}

function onSearchDebounced() {
  clearTimeout(searchDebounceTimer);
  searchDebounceTimer = setTimeout(() => {
    // search filter handled by computed filteredRooms
  }, 300);
}

// REST: Fetch Rooms
async function fetchRooms() {
  isLoadingRooms.value = true;
  try {
    const res = await apiClient.get('/chat/rooms', {
      params: {
        page: 1,
        page_size: 100,
      },
    });
    if (res.data?.data) {
      rooms.value = res.data.data;
    }
  } catch (e: any) {
    console.error('Fetch chat rooms error:', e);
  } finally {
    isLoadingRooms.value = false;
  }
}

// Select Room
async function selectRoom(room: any) {
  activeRoom.value = room;
  isLoadingMessages.value = true;
  activeMessages.value = [];

  try {
    const res = await apiClient.get(`/chat/messages/${room.id}?limit=100`);
    if (res.data?.data) {
      activeMessages.value = res.data.data;
    }

    // Mark as read on server via REST & WebSocket
    if (room.unread_count_admin > 0) {
      room.unread_count_admin = 0;
      apiClient.post(`/chat/messages/${room.id}/read?reader_type=admin`).catch(() => {});
      sendWsEvent('read_room', { room_id: room.id });
    }

    scrollToBottom();
  } catch (e: any) {
    console.error('Fetch messages error:', e);
  } finally {
    isLoadingMessages.value = false;
  }
}

// Send Admin Message
function sendAdminMessage() {
  const text = inputMessage.value.trim();
  if (!text || !activeRoom.value) return;

  const tempId = `admin_temp_${Date.now()}`;
  const localMsg = {
    id: null,
    room_id: activeRoom.value.id,
    sender_type: 'admin',
    sender_name: authStore.user?.nama || 'Admin CS',
    message: text,
    message_type: 'text',
    status: 'pending', // 🕒 Pending
    created_at: new Date().toISOString(),
    temp_id: tempId,
  };

  // Optimistic UI
  activeMessages.value.push(localMsg);
  activeRoom.value.last_message_text = text;
  activeRoom.value.last_message_at = new Date().toISOString();
  inputMessage.value = '';

  scrollToBottom();

  // Send via WebSocket
  sendWsEvent('send_message', {
    room_id: activeRoom.value.id,
    message: text,
    temp_id: tempId,
    message_type: 'text',
  });
}

function useTemplate(tpl: any) {
  inputMessage.value = tpl.text;
}

function notifyAdminTyping() {
  if (!activeRoom.value) return;
  sendWsEvent('typing', {
    room_id: activeRoom.value.id,
    is_typing: inputMessage.value.trim().length > 0,
  });
}

// ================= WEBSOCKET CONNECTION =================
function initWebSocket() {
  const user = authStore.user;
  const userId = user?.id || 1;
  const userName = encodeURIComponent(user?.nama || 'Admin CS');

  // Build WebSocket URL
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  let wsUrl = '';
  if (import.meta.env.DEV) {
    const hostname = window.location.hostname;
    wsUrl = `${protocol}//${hostname}:8000/api/v1/chat/ws?role=admin&user_id=${userId}&name=${userName}`;
  } else {
    wsUrl = `${protocol}//${window.location.host}/api/v1/chat/ws?role=admin&user_id=${userId}&name=${userName}`;
  }

  try {
    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      isWsConnected.value = true;
      startHeartbeat();
    };

    ws.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data);
        handleWsIncoming(payload);
      } catch (err) {
        console.warn('WS JSON parse error:', err);
      }
    };

    ws.onclose = () => {
      isWsConnected.value = false;
      stopHeartbeat();
      // Reconnect after 5 seconds
      setTimeout(initWebSocket, 5000);
    };

    ws.onerror = () => {
      isWsConnected.value = false;
    };
  } catch (e) {
    console.error('WS Connection error:', e);
  }
}

function handleWsIncoming(payload: any) {
  const event = payload.event;
  const data = payload.data;

  switch (event) {
    case 'ack_sent': {
      // Ceklis 1 Abu-abu
      const index = activeMessages.value.findIndex(
        (m) => m.temp_id === data.temp_id || m.id === data.id
      );
      if (index !== -1) {
        activeMessages.value[index].id = data.id;
        activeMessages.value[index].status = 'sent';
      }
      break;
    }

    case 'new_message': {
      const roomIdx = rooms.value.findIndex((r) => r.id === data.room_id);
      if (roomIdx !== -1) {
        rooms.value[roomIdx].last_message_text = data.message;
        rooms.value[roomIdx].last_message_at = data.created_at;
      }

      if (activeRoom.value && activeRoom.value.id === data.room_id) {
        // Message is in the open room
        const exists = activeMessages.value.some(
          (m) => (data.id && m.id === data.id) || (data.temp_id && m.temp_id === data.temp_id)
        );
        if (!exists) {
          activeMessages.value.push(data);
          scrollToBottom();
          // Mark as read immediately
          sendWsEvent('read_room', { room_id: data.room_id });
          apiClient.post(`/chat/messages/${data.room_id}/read?reader_type=admin`).catch(() => {});
        }
      } else {
        // Message is in another room -> increase unread count
        if (roomIdx !== -1) {
          rooms.value[roomIdx].unread_count_admin = (rooms.value[roomIdx].unread_count_admin || 0) + 1;
        }
        showSnackbar(`Pesan baru dari ${data.sender_name || 'Pelanggan'}`, 'info');
      }
      break;
    }

    case 'message_status_update': {
      // Delivered (Ceklis 2 Abu-abu) or Read (Ceklis 2 Biru)
      const index = activeMessages.value.findIndex(
        (m) => m.id === data.id || (data.temp_id && m.temp_id === data.temp_id)
      );
      if (index !== -1) {
        activeMessages.value[index].status = data.status;
      }
      break;
    }

    case 'room_read': {
      // All messages in room read by customer -> turn all CS messages into 'read' (Ceklis 2 Biru)
      if (activeRoom.value && activeRoom.value.id === data.room_id) {
        activeMessages.value.forEach((m) => {
          if (m.sender_type === 'admin') {
            m.status = 'read';
          }
        });
      }
      break;
    }

    case 'typing_indicator': {
      if (activeRoom.value && activeRoom.value.id === data.room_id) {
        isCustomerTyping.value = !!data.is_typing;
        clearTimeout(typingClearTimer);
        if (data.is_typing) {
          typingClearTimer = setTimeout(() => {
            isCustomerTyping.value = false;
          }, 3000);
        }
      }
      break;
    }
  }
}

function sendWsEvent(event: string, data: any) {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ event, data }));
  }
}

function startHeartbeat() {
  stopHeartbeat();
  heartbeatTimer = setInterval(() => {
    sendWsEvent('ping', {});
  }, 25000);
}

function stopHeartbeat() {
  if (heartbeatTimer) {
    clearInterval(heartbeatTimer);
    heartbeatTimer = null;
  }
}

// Formatting helpers
function formatTime(dateStr?: string): string {
  if (!dateStr) return '';
  const d = new Date(dateStr);
  return d.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
}

function formatTimestamp(dateStr?: string): string {
  if (!dateStr) return '';
  const d = new Date(dateStr);
  const now = new Date();
  if (d.toDateString() === now.toDateString()) {
    return formatTime(dateStr);
  }
  return d.toLocaleDateString('id-ID', { day: '2-digit', month: 'short' });
}

function shouldShowDateHeader(index: number): boolean {
  if (index === 0) return true;
  const current = new Date(activeMessages.value[index].created_at);
  const prev = new Date(activeMessages.value[index - 1].created_at);
  return current.toDateString() !== prev.toDateString();
}

function formatDateHeader(dateStr?: string): string {
  if (!dateStr) return '';
  const d = new Date(dateStr);
  const today = new Date();
  if (d.toDateString() === today.toDateString()) return 'Hari Ini';
  const yesterday = new Date(today);
  yesterday.setDate(yesterday.getDate() - 1);
  if (d.toDateString() === yesterday.toDateString()) return 'Kemarin';
  return d.toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' });
}

function scrollToBottom() {
  nextTick(() => {
    if (messagesScrollContainer.value) {
      messagesScrollContainer.value.scrollTop = messagesScrollContainer.value.scrollHeight + 100;
    }
  });
}

function copyToClipboard(text?: string) {
  if (!text) return;
  navigator.clipboard.writeText(text);
  showSnackbar(`Nomor ${text} berhasil disalin`, 'success');
}

function openWhatsApp(phone?: string) {
  if (!phone) return;
  let clean = phone.replace(/[^0-9]/g, '');
  if (clean.startsWith('0')) {
    clean = '62' + clean.substring(1);
  }
  window.open(`https://wa.me/${clean}`, '_blank');
}

function navigateToCustomer(pelangganId: number) {
  router.push(`/pelanggan?id=${pelangganId}`);
}

function getCustomerPackage(p?: any): string {
  if (!p) return '-';
  if (p.harga_layanan?.nama_layanan) return p.harga_layanan.nama_layanan;
  if (p.layanan) return p.layanan;
  return 'Internet FTTH';
}

function getCustomerStatus(p?: any): string {
  if (!p) return 'Aktif';
  if (p.langganan && p.langganan.length > 0) {
    return p.langganan[0].status || 'Aktif';
  }
  return 'Aktif';
}

function showSnackbar(text: string, color = 'success') {
  snackbar.value = { show: true, text, color };
}

// Lifecycle hooks
onMounted(() => {
  fetchRooms();
  initWebSocket();
});

onUnmounted(() => {
  stopHeartbeat();
  clearTimeout(typingClearTimer);
  clearTimeout(searchDebounceTimer);
  if (ws) {
    ws.close();
    ws = null;
  }
});
</script>

<style scoped>
.customer-chat-container {
  height: calc(100vh - 90px);
  display: flex;
  flex-direction: column;
}

.chat-workspace {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.rooms-sidebar {
  background-color: rgb(var(--v-theme-surface));
}

.filter-chips-scroll::-webkit-scrollbar {
  display: none;
}

.room-item {
  transition: background-color 0.15s ease;
  cursor: pointer;
}

.room-item:hover {
  background-color: rgba(var(--v-theme-primary), 0.04);
}

.messages-container {
  background-color: #f1f5f9;
}

.v-theme--dark .messages-container {
  background-color: #0f172a;
}

.message-bubble {
  max-width: 75%;
  min-width: 140px;
}

.bubble-customer {
  border-bottom-left-radius: 2px !important;
}

.bubble-admin {
  border-bottom-right-radius: 2px !important;
}

.quick-replies-bar::-webkit-scrollbar {
  display: none;
}

.animate-pulse {
  animation: pulse 1.8s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}
</style>
