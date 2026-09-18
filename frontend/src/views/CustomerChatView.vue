<template>
  <v-container fluid class="customer-chat-container pa-2 pa-md-3">
    <!-- Header Banner / Toolbar -->
    <v-card class="chat-header-banner mb-2 elevation-0 border rounded-lg" color="surface">
      <div class="d-flex align-center justify-space-between px-3 py-2">
        <div class="d-flex align-center gap-3">
          <v-avatar color="primary" variant="tonal" size="38">
            <v-icon size="20" color="primary">mdi-forum-outline</v-icon>
          </v-avatar>
          <div>
            <div class="d-flex align-center gap-2">
              <span class="text-subtitle-1 font-weight-bold text-high-emphasis">Live Chat Pelanggan</span>
              <v-chip
                :color="isWsConnected ? 'success' : 'grey'"
                size="x-small"
                variant="tonal"
                class="font-weight-bold px-2"
              >
                <v-icon start size="8" :class="{ 'animate-pulse': isWsConnected }">mdi-circle</v-icon>
                {{ isWsConnected ? 'WebSocket Online' : 'Menghubungkan...' }}
              </v-chip>
            </div>
            <div class="text-caption text-medium-emphasis" style="font-size: 0.72rem;">
              Pusat interaksi real-time pelanggan Jakinet, Jelantik, & Jelantik Nagrak
            </div>
          </div>
        </div>

        <!-- Quick Summary Badges & Action -->
        <div class="d-flex align-center gap-2">
          <v-chip
            v-if="totalUnreadCount > 0"
            color="error"
            size="small"
            variant="flat"
            class="font-weight-bold"
          >
            <v-icon start size="14">mdi-bell-ring-outline</v-icon>
            {{ totalUnreadCount }} Pesan Baru
          </v-chip>

          <v-tooltip location="bottom" text="Muat Ulang Room">
            <template v-slot:activator="{ props }">
              <v-btn
                v-bind="props"
                icon="mdi-refresh"
                variant="tonal"
                size="small"
                density="comfortable"
                :loading="isLoadingRooms"
                @click="fetchRooms()"
              ></v-btn>
            </template>
          </v-tooltip>
        </div>
      </div>
    </v-card>

    <!-- Main Workspace (3-Panel Chat Workspace) -->
    <v-card class="chat-workspace elevation-0 border rounded-lg" color="surface">
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
              :variant="selectedBrand === filter.value ? 'flat' : 'tonal'"
              size="small"
              class="font-weight-medium cursor-pointer filter-chip"
              :prepend-icon="filter.icon"
              @click="setBrandFilter(filter.value)"
            >
              {{ filter.label }}
            </v-chip>
          </div>

          <v-divider></v-divider>

          <!-- Rooms List -->
          <div class="rooms-list-scroll flex-grow-1 overflow-y-auto">
            <div v-if="isLoadingRooms" class="pa-6 text-center">
              <v-progress-circular indeterminate color="primary" size="28"></v-progress-circular>
              <div class="text-caption text-medium-emphasis mt-2">Memuat daftar obrolan...</div>
            </div>

            <div v-else-if="filteredRooms.length === 0" class="pa-8 text-center text-medium-emphasis">
              <v-icon size="40" color="grey-lighten-1" class="mb-2">mdi-message-text-outline</v-icon>
              <div class="text-body-2 font-weight-medium">Tidak ada percakapan</div>
              <div class="text-caption">Belum ada obrolan yang cocok dengan filter</div>
            </div>

            <v-list v-else lines="two" class="pa-0">
              <v-list-item
                v-for="room in filteredRooms"
                :key="room.id"
                :active="activeRoom?.id === room.id"
                class="room-item px-3 py-2 border-b"
                :class="{ 'room-item--active': activeRoom?.id === room.id }"
                @click="selectRoom(room)"
              >
                <!-- Customer Avatar with Brand Color -->
                <template v-slot:prepend>
                  <v-avatar :color="getBrandColor(room.brand)" size="40" class="elevation-1 me-2 rounded-lg">
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

                    <span v-if="room.pelanggan?.no_telp" class="text-caption text-medium-emphasis d-flex align-center">
                      <v-icon size="12" class="me-1">mdi-phone-outline</v-icon>
                      {{ room.pelanggan.no_telp }}
                    </span>
                  </div>

                  <!-- Last message snippet & unread badge -->
                  <div class="d-flex align-center justify-space-between mt-1">
                    <span class="text-caption text-truncate text-medium-emphasis flex-grow-1">
                      {{ room.last_message_text || 'Mulai obrolan...' }}
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
            <div class="empty-chat-icon-wrap mb-3 pa-4 rounded-circle bg-primary-lighten-5">
              <v-icon size="64" color="primary">mdi-chat-processing-outline</v-icon>
            </div>
            <h3 class="text-h6 font-weight-bold text-high-emphasis">Pusat Layanan Chat Pelanggan</h3>
            <p class="text-body-2 text-medium-emphasis max-w-sm mt-1">
              Pilih salah satu percakapan di sebelah kiri untuk membaca dan membalas pesan pelanggan secara langsung.
            </p>
          </div>

          <!-- Active Room View -->
          <template v-else>
            <!-- Chat Room Header -->
            <div class="chat-room-header pa-3 border-b bg-surface d-flex align-center justify-space-between">
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
                <v-avatar :color="getBrandColor(activeRoom.brand)" size="42" class="elevation-1 rounded-lg">
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
                    <span class="text-caption font-weight-medium text-medium-emphasis d-flex align-center">
                      <v-icon size="14" color="medium-emphasis" class="me-1">mdi-phone-outline</v-icon>
                      {{ activeRoom.pelanggan?.no_telp || '-' }}
                    </span>

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

                    <span v-if="isCustomerTyping" class="text-caption text-primary font-weight-medium d-flex align-center gap-1 animate-pulse">
                      <v-icon size="12">mdi-pencil-outline</v-icon>
                      sedang mengetik...
                    </span>
                  </div>
                </div>
              </div>

              <!-- Header Action Buttons -->
              <div class="d-flex align-center gap-2">
                <v-btn
                  v-if="activeRoom.pelanggan?.no_telp"
                  variant="tonal"
                  color="success"
                  size="small"
                  prepend-icon="mdi-whatsapp"
                  class="text-none font-weight-medium d-none d-sm-inline-flex"
                  @click="openWhatsApp(activeRoom.pelanggan.no_telp)"
                >
                  WhatsApp
                </v-btn>

                <v-btn
                  variant="outlined"
                  size="small"
                  prepend-icon="mdi-account-search-outline"
                  class="text-none d-none d-md-inline-flex"
                  @click="navigateToCustomer(activeRoom.pelanggan_id)"
                >
                  Detail
                </v-btn>

                <v-tooltip location="bottom" :text="showInfoPanel ? 'Sembunyikan Profil' : 'Tampilkan Profil Pelanggan 360'">
                  <template v-slot:activator="{ props }">
                    <v-btn
                      v-bind="props"
                      :icon="showInfoPanel ? 'mdi-dock-right' : 'mdi-information-outline'"
                      variant="text"
                      size="small"
                      :color="showInfoPanel ? 'primary' : 'medium-emphasis'"
                      @click="showInfoPanel = !showInfoPanel"
                    ></v-btn>
                  </template>
                </v-tooltip>
              </div>
            </div>

            <!-- Messages Stream Area -->
            <div ref="messagesScrollContainer" class="messages-container flex-grow-1 pa-3 pa-md-4 overflow-y-auto">
              <div v-if="isLoadingMessages" class="text-center pa-8">
                <v-progress-circular indeterminate color="primary" size="32"></v-progress-circular>
                <div class="text-caption text-medium-emphasis mt-2">Memuat riwayat obrolan...</div>
              </div>

              <div v-else-if="activeMessages.length === 0" class="text-center pa-8 text-medium-emphasis">
                <v-icon size="48" color="grey-lighten-2" class="mb-2">mdi-chat-plus-outline</v-icon>
                <div class="text-body-2 font-weight-medium">Belum ada pesan dalam obrolan ini</div>
                <div class="text-caption">Ketik balasan di bawah untuk memulai percakapan dengan pelanggan</div>
              </div>

              <!-- Messages List -->
              <div v-else class="d-flex flex-column gap-2">
                <template v-for="(msg, idx) in activeMessages" :key="msg.id || msg.temp_id || idx">
                  <!-- Date Pill Separator -->
                  <div v-if="shouldShowDateHeader(idx)" class="text-center my-3">
                    <span class="date-pill px-3 py-1 rounded-pill text-caption font-weight-medium">
                      <v-icon size="12" class="me-1 opacity-70">mdi-calendar-blank-outline</v-icon>
                      {{ formatDateHeader(msg.created_at) }}
                    </span>
                  </div>

                  <!-- Message Bubble Row -->
                  <div
                    class="message-row d-flex"
                    :class="msg.sender_type === 'admin' ? 'justify-end' : 'justify-start'"
                  >
                    <div
                      class="message-bubble pa-3"
                      :class="msg.sender_type === 'admin' ? 'bubble-admin' : 'bubble-customer'"
                    >
                      <!-- Sender Header -->
                      <div class="d-flex align-center justify-space-between gap-2 mb-1">
                        <span
                          class="sender-name font-weight-bold text-caption d-flex align-center gap-1"
                          :class="msg.sender_type === 'admin' ? 'text-white' : 'text-primary'"
                        >
                          <v-icon size="12" :color="msg.sender_type === 'admin' ? 'white' : 'primary'">
                            {{ msg.sender_type === 'admin' ? 'mdi-shield-account' : 'mdi-account' }}
                          </v-icon>
                          {{ msg.sender_type === 'admin' ? (msg.sender_name || 'CS Artacom') : (activeRoom.pelanggan?.nama || 'Pelanggan') }}
                        </span>
                      </div>

                      <!-- Text Message Body -->
                      <div
                        class="message-text text-body-2"
                        :class="msg.sender_type === 'admin' ? 'text-white' : 'text-high-emphasis'"
                        style="white-space: pre-wrap; word-break: break-word; line-height: 1.45;"
                      >
                        {{ msg.message }}
                      </div>

                      <!-- Footer: Timestamp & Delivery Status Icons -->
                      <div class="message-footer d-flex align-center justify-end gap-1 mt-1">
                        <span
                          class="timestamp text-caption"
                          :class="msg.sender_type === 'admin' ? 'text-blue-lighten-4' : 'text-medium-emphasis'"
                          style="font-size: 0.68rem;"
                        >
                          {{ formatTime(msg.created_at) }}
                        </span>

                        <!-- Status Icons for CS Messages -->
                        <template v-if="msg.sender_type === 'admin'">
                          <!-- Pending -->
                          <v-icon
                            v-if="msg.status === 'pending'"
                            size="12"
                            color="blue-lighten-4"
                            title="Sedang dikirim..."
                          >mdi-clock-outline</v-icon>

                          <!-- Sent (Ceklis 1) -->
                          <v-icon
                            v-else-if="msg.status === 'sent'"
                            size="14"
                            color="blue-lighten-4"
                            title="Tersimpan di server (Ceklis 1)"
                          >mdi-check</v-icon>

                          <!-- Delivered (Ceklis 2 Abu-abu) -->
                          <v-icon
                            v-else-if="msg.status === 'delivered'"
                            size="15"
                            color="blue-lighten-4"
                            title="Terkirim ke perangkat pelanggan (Ceklis 2 Abu-abu)"
                          >mdi-check-all</v-icon>

                          <!-- Read (Ceklis 2 Biru WhatsApp) -->
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
            <div class="quick-replies-bar px-3 py-2 bg-surface border-t d-flex align-center gap-2 overflow-x-auto">
              <div class="d-flex align-center gap-1 text-caption font-weight-bold text-medium-emphasis flex-shrink-0">
                <v-icon size="14" color="primary">mdi-lightning-bolt-outline</v-icon>
                <span>Templat Cepat:</span>
              </div>
              <v-chip
                v-for="(tpl, tIdx) in quickTemplates"
                :key="tIdx"
                size="small"
                variant="tonal"
                color="primary"
                class="cursor-pointer font-weight-medium flex-shrink-0 quick-chip"
                :prepend-icon="tpl.icon"
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
                placeholder="Ketik balasan CS... (Tekan Enter untuk kirim, Shift+Enter untuk baris baru)"
                hide-details
                class="chat-input-textarea"
                @keydown.enter.exact.prevent="sendAdminMessage"
                @input="notifyAdminTyping"
              ></v-textarea>

              <v-btn
                color="primary"
                icon="mdi-send-variant"
                elevation="1"
                size="default"
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
          class="customer-info-panel border-s pa-3 pa-lg-4 bg-surface overflow-y-auto fill-height"
        >
          <div class="d-flex align-center justify-space-between pb-3 mb-3 border-b">
            <div class="d-flex align-center gap-2">
              <v-icon size="18" color="primary">mdi-card-account-details-outline</v-icon>
              <h4 class="text-subtitle-2 font-weight-bold text-high-emphasis">Profil Pelanggan 360</h4>
            </div>
            <v-btn
              icon="mdi-close"
              variant="text"
              size="x-small"
              density="compact"
              @click="showInfoPanel = false"
            ></v-btn>
          </div>

          <!-- Customer Identity Card -->
          <v-card variant="outlined" class="pa-3 rounded-lg mb-3 info-card">
            <div class="d-flex align-center gap-2 mb-3">
              <v-avatar :color="getBrandColor(activeRoom.brand)" size="36" class="rounded-lg">
                <span class="text-caption font-weight-bold text-white">
                  {{ getInitials(activeRoom.pelanggan?.nama || 'Pelanggan') }}
                </span>
              </v-avatar>
              <div>
                <div class="font-weight-bold text-body-2 text-high-emphasis">
                  {{ activeRoom.pelanggan?.nama || '-' }}
                </div>
                <div class="text-caption text-medium-emphasis">
                  ID: {{ activeRoom.pelanggan?.customer_id || ('#' + activeRoom.pelanggan_id) }}
                </div>
              </div>
            </div>

            <v-divider class="my-2"></v-divider>

            <div class="info-row mb-2">
              <div class="text-caption text-medium-emphasis d-flex align-center gap-1 mb-1">
                <v-icon size="14">mdi-shield-check-outline</v-icon>
                <span>Brand Layanan</span>
              </div>
              <v-chip :color="getBrandColor(activeRoom.brand)" size="x-small" variant="flat" class="font-weight-bold text-white">
                {{ normalizeBrandName(activeRoom.brand) }}
              </v-chip>
            </div>

            <div class="info-row mb-2">
              <div class="text-caption text-medium-emphasis d-flex align-center gap-1 mb-1">
                <v-icon size="14">mdi-phone-outline</v-icon>
                <span>Nomor Telepon</span>
              </div>
              <div class="d-flex align-center justify-space-between">
                <span class="text-body-2 font-weight-medium">{{ activeRoom.pelanggan?.no_telp || '-' }}</span>
                <v-btn
                  v-if="activeRoom.pelanggan?.no_telp"
                  icon="mdi-content-copy"
                  size="x-small"
                  variant="text"
                  density="compact"
                  title="Salin Nomor"
                  @click="copyToClipboard(activeRoom.pelanggan.no_telp)"
                ></v-btn>
              </div>
            </div>

            <div class="info-row">
              <div class="text-caption text-medium-emphasis d-flex align-center gap-1 mb-1">
                <v-icon size="14">mdi-map-marker-outline</v-icon>
                <span>Alamat / Lokasi</span>
              </div>
              <div class="text-caption text-high-emphasis">
                {{ activeRoom.pelanggan?.alamat || '-' }}
                <span v-if="activeRoom.pelanggan?.blok || activeRoom.pelanggan?.unit">
                  (Blok {{ activeRoom.pelanggan?.blok || '-' }} / Unit {{ activeRoom.pelanggan?.unit || '-' }})
                </span>
              </div>
            </div>
          </v-card>

          <!-- Subscription & Technical Info -->
          <v-card variant="outlined" class="pa-3 rounded-lg mb-3 info-card">
            <div class="d-flex align-center gap-2 mb-2">
              <v-icon size="16" color="primary">mdi-wifi</v-icon>
              <span class="text-subtitle-2 font-weight-bold">Status Layanan FTTH</span>
            </div>

            <div class="info-row mb-2">
              <div class="text-caption text-medium-emphasis d-flex align-center gap-1 mb-1">
                <v-icon size="14">mdi-speedometer</v-icon>
                <span>Paket Layanan</span>
              </div>
              <div class="font-weight-bold text-body-2 text-primary">
                {{ getCustomerPackage(activeRoom.pelanggan) }}
              </div>
            </div>

            <div class="info-row mb-2">
              <div class="text-caption text-medium-emphasis d-flex align-center gap-1 mb-1">
                <v-icon size="14">mdi-check-circle-outline</v-icon>
                <span>Status Langganan</span>
              </div>
              <v-chip
                :color="getCustomerStatus(activeRoom.pelanggan) === 'Aktif' ? 'success' : 'error'"
                size="x-small"
                variant="flat"
                class="font-weight-bold"
              >
                {{ getCustomerStatus(activeRoom.pelanggan) }}
              </v-chip>
            </div>

            <div v-if="activeRoom.pelanggan?.data_teknis?.id_pelanggan" class="info-row">
              <div class="text-caption text-medium-emphasis d-flex align-center gap-1 mb-1">
                <v-icon size="14">mdi-account-key-outline</v-icon>
                <span>Username PPPoE</span>
              </div>
              <div class="font-weight-bold text-caption font-mono bg-grey-lighten-4 pa-1 px-2 rounded">
                {{ activeRoom.pelanggan.data_teknis.id_pelanggan }}
              </div>
            </div>
          </v-card>

          <!-- Action Buttons -->
          <div class="d-flex flex-column gap-2">
            <v-btn
              block
              color="success"
              prepend-icon="mdi-whatsapp"
              variant="flat"
              class="text-none font-weight-bold"
              @click="openWhatsApp(activeRoom.pelanggan?.no_telp)"
            >
              Hubungi via WhatsApp Web
            </v-btn>
            <v-btn
              block
              variant="outlined"
              prepend-icon="mdi-account-search-outline"
              class="text-none font-weight-medium"
              @click="navigateToCustomer(activeRoom.pelanggan_id)"
            >
              Buka di Data Pelanggan
            </v-btn>
          </div>
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
let pollingTimer: any = null;

// Snackbar notification
const snackbar = ref({
  show: false,
  text: '',
  color: 'success',
});

// Brand filter options
const brandFilters = [
  { label: 'Semua', value: 'ALL', color: 'primary', icon: 'mdi-layers-outline' },
  { label: 'JAKINET', value: 'JAKINET', color: 'error', icon: 'mdi-network' },
  { label: 'JELANTIK', value: 'JELANTIK', color: 'primary', icon: 'mdi-web' },
  { label: 'JELANTIK NAGRAK', value: 'JELANTIK NAGRAK', color: 'success', icon: 'mdi-shield-home-outline' },
];

// Quick templates
const quickTemplates = [
  {
    label: 'Salam',
    icon: 'mdi-hand-wave-outline',
    text: 'Halo, selamat datang di layanan Customer Care Artacom. Ada yang bisa kami bantu?'
  },
  {
    label: 'Cek Teknis',
    icon: 'mdi-wrench-clock-outline',
    text: 'Baik pak/bu, mohon ditunggu sebentar ya. Sedang kami lakukan pengecekan ke tim teknis lapangan.'
  },
  {
    label: 'Lunas & Aktif',
    icon: 'mdi-check-decagram-outline',
    text: 'Terima kasih atas konfirmasinya. Tagihan Anda telah terverifikasi dan layanan internet sudah aktif normal kembali.'
  },
  {
    label: 'Foto Modem',
    icon: 'mdi-camera-outline',
    text: 'Bisa tolong difotokan lampu indikator (PON / LOS / Internet) yang menyala pada perangkat modem router Anda?'
  },
  {
    label: 'Restart Modem',
    icon: 'mdi-restart',
    text: 'Bisa dicoba untuk mematikan modem router selama 1-2 menit, lalu hidupkan kembali dan periksa koneksinya?'
  }
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
async function fetchRooms(silent = false) {
  if (!silent) isLoadingRooms.value = true;
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
    if (!silent) isLoadingRooms.value = false;
  }
}

// Silent polling for active room and room list to guarantee zero-refresh experience
async function pollActiveRoomSilent() {
  if (document.hidden) return;
  if (activeRoom.value && !isLoadingMessages.value) {
    try {
      const res = await apiClient.get(`/chat/messages/${activeRoom.value.id}?limit=50`);
      if (res.data?.data && Array.isArray(res.data.data)) {
        const latest = res.data.data;
        let hasNew = false;
        for (const item of latest) {
          const idx = activeMessages.value.findIndex(
            (m) => m.id === item.id || (m.temp_id && m.temp_id === item.temp_id)
          );
          if (idx === -1) {
            activeMessages.value.push(item);
            hasNew = true;
          } else {
            if (activeMessages.value[idx].status !== item.status) {
              activeMessages.value[idx].status = item.status;
            }
            if (!activeMessages.value[idx].id && item.id) {
              activeMessages.value[idx].id = item.id;
            }
          }
        }
        if (hasNew) {
          scrollToBottom();
        }
      }
    } catch (_) {}
  }
  fetchRooms(true);
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
    sender_name: authStore.user?.name || 'Admin CS',
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

  // Move active room to top of list
  const currentRoomId = activeRoom.value.id;
  const roomIdx = rooms.value.findIndex((r) => r.id === currentRoomId);
  if (roomIdx > 0) {
    const [moved] = rooms.value.splice(roomIdx, 1);
    rooms.value.unshift(moved);
  }

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
  const userName = encodeURIComponent(user?.name || 'Admin CS');

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
      // Reconnect after 3 seconds
      setTimeout(initWebSocket, 3000);
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
        const [updatedRoom] = rooms.value.splice(roomIdx, 1);
        updatedRoom.last_message_text = data.message;
        updatedRoom.last_message_at = data.created_at;
        rooms.value.unshift(updatedRoom);
      } else {
        fetchRooms(true);
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
          rooms.value[0].unread_count_admin = (rooms.value[0].unread_count_admin || 0) + 1;
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

function onVisibilityChange() {
  if (!document.hidden) {
    pollActiveRoomSilent();
  }
}

// Lifecycle hooks
onMounted(() => {
  fetchRooms();
  initWebSocket();
  // Background polling every 4s for zero-refresh guarantees
  pollingTimer = setInterval(pollActiveRoomSilent, 4000);
  document.addEventListener('visibilitychange', onVisibilityChange);
});

onUnmounted(() => {
  stopHeartbeat();
  if (pollingTimer) {
    clearInterval(pollingTimer);
    pollingTimer = null;
  }
  document.removeEventListener('visibilitychange', onVisibilityChange);
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
  height: calc(100vh - 84px);
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

.filter-chips-scroll::-webkit-scrollbar,
.quick-replies-bar::-webkit-scrollbar {
  display: none;
}

.room-item {
  transition: all 0.15s ease;
  cursor: pointer;
  border-left: 3px solid transparent;
}

.room-item:hover {
  background-color: rgba(var(--v-theme-primary), 0.04);
}

.room-item--active {
  background-color: rgba(var(--v-theme-primary), 0.08) !important;
  border-left-color: rgb(var(--v-theme-primary)) !important;
}

.messages-container {
  background-color: #f8fafc;
}

.v-theme--dark .messages-container {
  background-color: #0b1120;
}

.message-bubble {
  max-width: 72%;
  min-width: 140px;
  position: relative;
  word-wrap: break-word;
}

.bubble-customer {
  background-color: #ffffff;
  color: #1e293b;
  border-radius: 16px 16px 16px 4px !important;
  border: 1px solid rgba(226, 232, 240, 0.9);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.v-theme--dark .bubble-customer {
  background-color: #1e293b;
  color: #f8fafc;
  border-color: rgba(51, 65, 85, 0.8);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
}

.bubble-admin {
  background: linear-gradient(135deg, #1e40af, #2563eb);
  color: #ffffff;
  border-radius: 16px 16px 4px 16px !important;
  box-shadow: 0 2px 8px rgba(37, 99, 235, 0.25);
}

.date-pill {
  background-color: rgba(226, 232, 240, 0.85);
  color: #475569;
  border: 1px solid rgba(203, 213, 225, 0.6);
  font-size: 0.72rem;
  letter-spacing: 0.2px;
}

.v-theme--dark .date-pill {
  background-color: #1e293b;
  color: #94a3b8;
  border-color: #334155;
}

.quick-chip {
  transition: transform 0.1s ease, box-shadow 0.1s ease;
}

.quick-chip:hover {
  transform: translateY(-1px);
}

.info-card {
  border-color: rgba(0, 0, 0, 0.08) !important;
  background-color: rgba(var(--v-theme-surface), 0.6);
}

.v-theme--dark .info-card {
  border-color: rgba(255, 255, 255, 0.08) !important;
}

.animate-pulse {
  animation: pulse 1.8s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

/* Modern thin scrollbar */
.rooms-list-scroll::-webkit-scrollbar,
.messages-container::-webkit-scrollbar,
.customer-info-panel::-webkit-scrollbar {
  width: 5px;
}

.rooms-list-scroll::-webkit-scrollbar-thumb,
.messages-container::-webkit-scrollbar-thumb,
.customer-info-panel::-webkit-scrollbar-thumb {
  background-color: rgba(0, 0, 0, 0.12);
  border-radius: 4px;
}

.v-theme--dark .rooms-list-scroll::-webkit-scrollbar-thumb,
.v-theme--dark .messages-container::-webkit-scrollbar-thumb,
.v-theme--dark .customer-info-panel::-webkit-scrollbar-thumb {
  background-color: rgba(255, 255, 255, 0.15);
}
</style>
