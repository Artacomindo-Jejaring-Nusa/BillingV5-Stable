<template>
  <div class="customer-chat-wrapper pa-3">
    <!-- Main Workspace Container -->
    <v-card class="chat-workspace elevation-1 border rounded-xl" color="surface">
      <!-- ================= PANEL 1: ROOMS SIDEBAR ================= -->
      <aside
        class="rooms-sidebar border-e"
        :class="{ 'd-none d-md-flex': activeRoom && isMobile }"
      >
        <!-- Sidebar Header: Title, Online Status & Refresh -->
        <div class="sidebar-header px-4 py-3 border-b d-flex align-center justify-space-between bg-surface">
          <div class="d-flex align-center gap-2">
            <v-avatar color="primary" variant="tonal" size="34" class="rounded-lg">
              <v-icon size="18" color="primary">mdi-forum-outline</v-icon>
            </v-avatar>
            <div>
              <span class="text-subtitle-2 font-weight-bold text-high-emphasis">Chat Pelanggan</span>
              <div class="d-flex align-center gap-1">
                <v-icon size="8" :color="isWsConnected ? 'success' : 'grey'" :class="{ 'animate-pulse': isWsConnected }">
                  mdi-circle
                </v-icon>
                <span class="text-caption text-medium-emphasis" style="font-size: 0.7rem;">
                  {{ isWsConnected ? 'Online' : 'Menghubungkan...' }}
                </span>
              </div>
            </div>
          </div>

          <div class="d-flex align-center gap-1">
            <v-badge
              v-if="totalUnreadCount > 0"
              :content="totalUnreadCount"
              color="error"
              inline
              class="me-1"
            ></v-badge>

            <v-tooltip location="bottom" text="Muat Ulang">
              <template v-slot:activator="{ props }">
                <v-btn
                  v-bind="props"
                  icon="mdi-refresh"
                  variant="text"
                  size="small"
                  density="compact"
                  :loading="isLoadingRooms"
                  @click="fetchRooms()"
                ></v-btn>
              </template>
            </v-tooltip>
          </div>
        </div>

        <!-- Search Box -->
        <div class="px-3 pt-3 pb-2">
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

        <!-- Status Filter Tabs (Aktif / Selesai / Semua) -->
        <div class="px-3 pb-2 d-flex gap-1">
          <v-btn
            size="x-small"
            :variant="selectedStatus === 'open' ? 'flat' : 'tonal'"
            :color="selectedStatus === 'open' ? 'primary' : 'default'"
            class="flex-grow-1 text-none font-weight-bold rounded-pill"
            style="font-size: 0.72rem; height: 26px;"
            @click="selectedStatus = 'open'"
          >
            Aktif
            <span v-if="openRoomsCount > 0" class="ms-1 px-1.5 py-0.2 rounded-pill bg-primary-lighten-1 text-white font-weight-bold" style="font-size: 0.65rem;">
              {{ openRoomsCount }}
            </span>
          </v-btn>
          <v-btn
            size="x-small"
            :variant="selectedStatus === 'closed' ? 'flat' : 'tonal'"
            :color="selectedStatus === 'closed' ? 'primary' : 'default'"
            class="flex-grow-1 text-none font-weight-bold rounded-pill"
            style="font-size: 0.72rem; height: 26px;"
            @click="selectedStatus = 'closed'"
          >
            Selesai
            <span v-if="closedRoomsCount > 0" class="ms-1 px-1.5 py-0.2 rounded-pill bg-grey-darken-1 text-white font-weight-bold" style="font-size: 0.65rem;">
              {{ closedRoomsCount }}
            </span>
          </v-btn>
          <v-btn
            size="x-small"
            :variant="selectedStatus === 'ALL' ? 'flat' : 'tonal'"
            :color="selectedStatus === 'ALL' ? 'primary' : 'default'"
            class="flex-grow-1 text-none font-weight-bold rounded-pill"
            style="font-size: 0.72rem; height: 26px;"
            @click="selectedStatus = 'ALL'"
          >
            Semua
          </v-btn>
        </div>

        <!-- Brand Filter Chips (Compact & Clean) -->
        <div class="px-3 pb-2 brand-filter-bar d-flex gap-1.5 overflow-x-auto">
          <v-chip
            v-for="filter in brandFilters"
            :key="filter.value"
            :color="selectedBrand === filter.value ? filter.color : undefined"
            :variant="selectedBrand === filter.value ? 'flat' : 'tonal'"
            size="small"
            class="font-weight-medium cursor-pointer brand-filter-chip flex-shrink-0"
            @click="setBrandFilter(filter.value)"
          >
            {{ filter.label }}
          </v-chip>
        </div>

        <v-divider></v-divider>

        <!-- Rooms List -->
        <div class="rooms-list-scroll flex-grow-1 overflow-y-auto pa-2">
          <div v-if="isLoadingRooms" class="pa-8 text-center">
            <v-progress-circular indeterminate color="primary" size="26"></v-progress-circular>
            <div class="text-caption text-medium-emphasis mt-2">Memuat percakapan...</div>
          </div>

          <div v-else-if="filteredRooms.length === 0" class="pa-8 text-center text-medium-emphasis">
            <v-icon size="36" color="grey-lighten-1" class="mb-2">mdi-message-text-outline</v-icon>
            <div class="text-body-2 font-weight-medium">Tidak ada percakapan</div>
            <div class="text-caption">
              {{ selectedStatus === 'closed' ? 'Belum ada percakapan yang selesai' : 'Belum ada obrolan aktif' }}
            </div>
          </div>

          <div v-else class="d-flex flex-column gap-1.5">
            <div
              v-for="room in filteredRooms"
              :key="room.id"
              class="room-card px-3 py-2.5 rounded-lg cursor-pointer"
              :class="{ 'room-card--active': activeRoom?.id === room.id }"
              @click="selectRoom(room)"
            >
              <div class="d-flex align-start gap-3">
                <!-- Customer Avatar with Brand Color -->
                <v-avatar :color="getBrandColor(room.brand)" size="38" class="elevation-1 rounded-circle flex-shrink-0 mt-0.5">
                  <span class="text-caption font-weight-bold text-white">
                    {{ getInitials(room.pelanggan?.nama || 'Pelanggan') }}
                  </span>
                </v-avatar>

                <!-- Details -->
                <div class="flex-grow-1 min-w-0">
                  <!-- Name & Timestamp Row -->
                  <div class="d-flex align-center justify-space-between mb-1">
                    <span class="font-weight-bold text-truncate text-body-2 text-high-emphasis flex-grow-1 me-2" style="max-width: 160px;">
                      {{ room.pelanggan?.nama || 'Pelanggan #' + room.pelanggan_id }}
                    </span>
                    <span class="text-caption text-medium-emphasis flex-shrink-0 font-weight-medium" style="font-size: 0.7rem;">
                      {{ formatTimestamp(room.last_message_at) }}
                    </span>
                  </div>

                  <!-- Brand Pill & Phone Row -->
                  <div class="d-flex align-center gap-1.5 mb-1.5 flex-wrap">
                    <v-chip
                      :color="getBrandColor(room.brand)"
                      size="x-small"
                      variant="flat"
                      class="px-1.5 font-weight-bold text-white"
                      style="font-size: 0.62rem; height: 18px;"
                    >
                      {{ normalizeBrandName(room.brand) }}
                    </v-chip>

                    <span v-if="room.pelanggan?.no_telp" class="text-caption text-medium-emphasis d-flex align-center" style="font-size: 0.72rem;">
                      <v-icon size="11" class="me-0.5">mdi-phone-outline</v-icon>
                      {{ room.pelanggan.no_telp }}
                    </span>
                  </div>

                  <!-- Last message snippet & status/unread badge Row -->
                  <div class="d-flex align-center justify-space-between">
                    <span class="text-caption text-truncate text-medium-emphasis flex-grow-1 me-2" style="font-size: 0.75rem;">
                      {{ room.last_message_text || 'Mulai obrolan...' }}
                    </span>
                    <v-chip
                      v-if="room.status === 'closed'"
                      size="x-small"
                      variant="tonal"
                      color="grey"
                      class="font-weight-medium flex-shrink-0 px-1.5"
                      style="font-size: 0.65rem; height: 18px;"
                    >
                      Selesai
                    </v-chip>
                    <v-badge
                      v-else-if="room.unread_count_admin > 0"
                      :content="room.unread_count_admin"
                      color="error"
                      inline
                      class="flex-shrink-0"
                    ></v-badge>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </aside>

      <!-- ================= PANEL 2: ACTIVE CHAT CONVERSATION ================= -->
      <main
        class="chat-conversation-panel"
        :class="{ 'd-none d-md-flex': !activeRoom && isMobile }"
      >
        <!-- Empty State When No Room is Selected -->
        <div v-if="!activeRoom" class="d-flex flex-column align-center justify-center fill-height pa-8 text-center bg-surface">
          <div class="empty-chat-icon-wrap mb-4 pa-5 rounded-circle bg-primary-lighten-5">
            <v-icon size="56" color="primary">mdi-chat-processing-outline</v-icon>
          </div>
          <h3 class="text-h6 font-weight-bold text-high-emphasis">Pusat Layanan Chat Pelanggan</h3>
          <p class="text-body-2 text-medium-emphasis max-w-sm mt-1">
            Pilih salah satu percakapan di sebelah kiri untuk membaca dan membalas pesan pelanggan secara langsung.
          </p>
        </div>

        <!-- Active Room View -->
        <template v-else>
          <!-- Chat Room Header -->
          <header class="chat-room-header px-4 py-2.5 border-b bg-surface d-flex align-center justify-space-between">
            <div class="d-flex align-center gap-3 min-w-0">
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
              <v-avatar :color="getBrandColor(activeRoom.brand)" size="42" class="elevation-1 rounded-circle flex-shrink-0">
                <span class="text-subtitle-2 font-weight-bold text-white">
                  {{ getInitials(activeRoom.pelanggan?.nama || 'Pelanggan') }}
                </span>
              </v-avatar>

              <!-- Customer Details Header -->
              <div class="min-w-0">
                <div class="d-flex align-center gap-2 flex-wrap">
                  <h3 class="text-subtitle-1 font-weight-bold mb-0 text-truncate text-high-emphasis">
                    {{ activeRoom.pelanggan?.nama || 'Pelanggan #' + activeRoom.pelanggan_id }}
                  </h3>

                  <!-- Brand Badge -->
                  <v-chip
                    :color="getBrandColor(activeRoom.brand)"
                    size="x-small"
                    variant="flat"
                    class="font-weight-bold text-white px-2"
                  >
                    {{ normalizeBrandName(activeRoom.brand) }}
                  </v-chip>
                </div>

                <!-- Phone Number & Typing Indicator -->
                <div class="d-flex align-center gap-2 mt-0.5">
                  <span class="text-caption text-medium-emphasis d-flex align-center">
                    <v-icon size="13" color="medium-emphasis" class="me-1">mdi-phone-outline</v-icon>
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
                        size="x-small"
                        color="medium-emphasis"
                        @click="copyToClipboard(activeRoom.pelanggan.no_telp)"
                      ></v-btn>
                    </template>
                  </v-tooltip>

                  <span v-if="isCustomerTyping" class="text-caption text-primary font-weight-medium d-flex align-center gap-1 animate-pulse ms-2">
                    <v-icon size="12">mdi-pencil-outline</v-icon>
                    sedang mengetik...
                  </span>
                </div>
              </div>
            </div>

            <!-- Header Action Buttons -->
            <div class="d-flex align-center gap-2 flex-shrink-0 ms-3">
              <!-- Close / Selesai Conversation Button -->
              <v-btn
                v-if="activeRoom.status !== 'closed'"
                variant="tonal"
                color="success"
                size="small"
                prepend-icon="mdi-check-circle-outline"
                class="text-none font-weight-bold rounded-pill"
                :loading="isUpdatingStatus"
                @click="closeActiveRoom"
              >
                Tutup Percakapan
              </v-btn>
              <div v-else class="d-flex align-center gap-1.5">
                <v-chip
                  color="grey"
                  variant="tonal"
                  size="small"
                  class="font-weight-bold"
                  prepend-icon="mdi-check-all"
                >
                  Selesai
                </v-chip>
                <v-btn
                  variant="tonal"
                  color="primary"
                  size="small"
                  prepend-icon="mdi-lock-open-outline"
                  class="text-none font-weight-bold rounded-pill"
                  :loading="isUpdatingStatus"
                  @click="reopenActiveRoom"
                >
                  Buka Kembali
                </v-btn>
              </div>

              <!-- Sound Toggle -->
              <v-tooltip location="bottom" :text="isSoundEnabled ? 'Nonaktifkan Notifikasi Suara' : 'Aktifkan Notifikasi Suara'">
                <template v-slot:activator="{ props }">
                  <v-btn
                    v-bind="props"
                    :icon="isSoundEnabled ? 'mdi-volume-high' : 'mdi-volume-off'"
                    variant="text"
                    size="small"
                    :color="isSoundEnabled ? 'primary' : 'medium-emphasis'"
                    @click="isSoundEnabled = !isSoundEnabled"
                  ></v-btn>
                </template>
              </v-tooltip>

              <v-btn
                v-if="activeRoom.pelanggan?.no_telp"
                variant="tonal"
                color="success"
                size="small"
                prepend-icon="mdi-whatsapp"
                class="text-none font-weight-bold rounded-pill d-none d-sm-inline-flex"
                @click="openWhatsApp(activeRoom.pelanggan.no_telp)"
              >
                WhatsApp
              </v-btn>

              <v-btn
                variant="tonal"
                size="small"
                prepend-icon="mdi-account-details-outline"
                class="text-none font-weight-medium rounded-pill d-none d-md-inline-flex"
                @click="navigateToCustomer(activeRoom.pelanggan_id)"
              >
                Detail Pelanggan
              </v-btn>

              <v-tooltip location="bottom" :text="showInfoPanel ? 'Tutup Panel Profil' : 'Lihat Profil Pelanggan 360'">
                <template v-slot:activator="{ props }">
                  <v-btn
                    v-bind="props"
                    :icon="showInfoPanel ? 'mdi-close' : 'mdi-information-outline'"
                    variant="tonal"
                    size="small"
                    :color="showInfoPanel ? 'primary' : 'default'"
                    @click="showInfoPanel = !showInfoPanel"
                  ></v-btn>
                </template>
              </v-tooltip>

              <!-- Close / Exit Chat Room (Esc) -->
              <v-tooltip location="bottom" text="Keluar dari Obrolan (Esc)">
                <template v-slot:activator="{ props }">
                  <v-btn
                    v-bind="props"
                    icon="mdi-close"
                    variant="text"
                    size="small"
                    color="medium-emphasis"
                    @click="activeRoom = null"
                  ></v-btn>
                </template>
              </v-tooltip>
            </div>
          </header>

          <!-- Messages Stream Area -->
          <div ref="messagesScrollContainer" class="messages-container flex-grow-1 px-6 py-4 overflow-y-auto">
            <div v-if="isLoadingMessages" class="text-center pa-8">
              <v-progress-circular indeterminate color="primary" size="30"></v-progress-circular>
              <div class="text-caption text-medium-emphasis mt-2">Memuat riwayat obrolan...</div>
            </div>

            <div v-else-if="activeMessages.length === 0" class="text-center pa-8 text-medium-emphasis">
              <v-icon size="44" color="grey-lighten-2" class="mb-2">mdi-chat-plus-outline</v-icon>
              <div class="text-body-2 font-weight-medium">Belum ada pesan dalam obrolan ini</div>
              <div class="text-caption">Ketik balasan di bawah untuk memulai percakapan</div>
            </div>

            <!-- Messages List -->
            <div v-else class="d-flex flex-column">
              <!-- Closed Notice at Top of Stream if Room is Closed -->
              <div v-if="activeRoom.status === 'closed'" class="text-center my-3">
                <span class="d-inline-flex align-center gap-1.5 px-4 py-2 rounded-pill bg-grey-lighten-4 text-caption text-medium-emphasis border">
                  <v-icon size="16" color="success">mdi-check-circle-outline</v-icon>
                  Percakapan ini telah selesai. Kirim pesan baru untuk membuka kembali secara otomatis.
                </span>
              </div>

              <template v-for="(msg, idx) in activeMessages" :key="msg.id || msg.temp_id || idx">
                <!-- Date Pill Separator -->
                <div v-if="shouldShowDateHeader(idx)" class="text-center my-3">
                  <span class="date-pill px-3 py-1 rounded-pill text-caption font-weight-medium">
                    {{ formatDateHeader(msg.created_at) }}
                  </span>
                </div>

                <!-- Message Bubble Row -->
                <div
                  class="message-row d-flex mb-3.5"
                  :class="msg.sender_type === 'admin' ? 'justify-end' : 'justify-start'"
                >
                  <div
                    class="message-bubble py-2.5 px-4"
                    :class="msg.sender_type === 'admin' ? 'bubble-admin' : 'bubble-customer'"
                  >
                    <!-- Image Attachment if present -->
                    <div v-if="msg.attachment_url || msg.message_type === 'image'" class="mb-1.5" style="min-width: 200px;">
                      <v-img
                        :src="getFullMediaUrl(msg.attachment_url)"
                        width="260"
                        min-height="160"
                        max-height="320"
                        aspect-ratio="1"
                        class="rounded-lg cursor-pointer elevation-1 bg-grey-lighten-3"
                        cover
                        @click="openImageLightbox(msg.attachment_url)"
                      >
                        <template v-slot:placeholder>
                          <div class="d-flex align-center justify-center fill-height" style="min-height: 160px; width: 260px;">
                            <v-progress-circular indeterminate color="primary" size="28"></v-progress-circular>
                          </div>
                        </template>
                        <template v-slot:error>
                          <div class="d-flex flex-column align-center justify-center fill-height pa-4 text-caption text-medium-emphasis bg-grey-lighten-3 rounded-lg" style="min-height: 140px; width: 260px;">
                            <v-icon size="28" color="grey-darken-1" class="mb-1">mdi-image-broken-variant</v-icon>
                            <span class="font-weight-medium text-center">Gagal memuat gambar</span>
                          </div>
                        </template>
                      </v-img>
                    </div>

                    <!-- Text Message Body (Clean, without redundant sender name on 1-on-1 chat) -->
                    <div
                      v-if="msg.message"
                      class="message-text"
                      :class="msg.sender_type === 'admin' ? 'text-white' : 'text-high-emphasis'"
                      style="white-space: pre-wrap; word-break: break-word; overflow-wrap: anywhere; line-height: 1.55; font-size: 0.92rem;"
                    >
                      {{ msg.message }}
                    </div>

                    <!-- Footer: Timestamp & Delivery Status Icons -->
                    <div class="message-footer d-flex align-center justify-end gap-1.5 mt-1.5">
                      <span
                        class="timestamp"
                        :class="msg.sender_type === 'admin' ? 'text-blue-lighten-4' : 'text-medium-emphasis'"
                        style="font-size: 0.7rem;"
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
                          title="Terkirim ke perangkat (Ceklis 2 Abu-abu)"
                        >mdi-check-all</v-icon>

                        <!-- Read (Ceklis 2 Biru WhatsApp) -->
                        <v-icon
                          v-else-if="msg.status === 'read'"
                          size="15"
                          color="cyan-accent-2"
                          title="Telah dibaca (Ceklis 2 Biru)"
                        >mdi-check-all</v-icon>
                      </template>
                    </div>
                  </div>
                </div>
              </template>
            </div>
          </div>

          <!-- Quick Template Replies Bar (Sleek & Scrollable) -->
          <div class="quick-replies-bar px-4 py-2 bg-surface border-t d-flex align-center gap-2 overflow-x-auto">
            <v-icon size="14" color="amber-darken-2" class="flex-shrink-0">mdi-lightning-bolt</v-icon>
            <span class="text-caption font-weight-bold text-medium-emphasis flex-shrink-0">Templat:</span>
            <div class="d-flex align-center gap-1.5 flex-nowrap overflow-x-auto py-0.5">
              <v-chip
                v-for="(tpl, tIdx) in quickTemplates"
                :key="tIdx"
                size="small"
                variant="tonal"
                color="primary"
                class="cursor-pointer font-weight-medium flex-shrink-0 quick-chip px-3"
                :prepend-icon="tpl.icon"
                @click="useTemplate(tpl)"
              >
                {{ tpl.label }}
              </v-chip>
            </div>
          </div>

          <!-- Bottom Chat Input Bar -->
          <footer class="chat-input-bar px-4 py-3 bg-surface border-t">
            <div class="d-flex align-end gap-2" style="position: relative;">
              <!-- Emoji Picker Toggle -->
              <div style="position: relative;">
                <v-tooltip location="top" text="Pilih Emoji">
                  <template v-slot:activator="{ props }">
                    <v-btn
                      v-bind="props"
                      icon="mdi-emoticon-happy-outline"
                      variant="text"
                      size="small"
                      color="medium-emphasis"
                      class="flex-shrink-0 mb-1"
                      @click="showEmojiPicker = !showEmojiPicker"
                    ></v-btn>
                  </template>
                </v-tooltip>

                <!-- Emoji Picker Popup -->
                <div v-if="showEmojiPicker" class="emoji-picker-popup">
                  <EmojiPicker
                    :native="true"
                    :disable-skin-tones="true"
                    :display-recent="true"
                    @select="onSelectEmoji"
                  />
                </div>
              </div>

              <!-- Hidden File Input for Image Upload -->
              <input
                ref="fileInputRef"
                type="file"
                accept="image/*"
                class="d-none"
                @change="onImageSelected"
              />

              <!-- Image Upload Button -->
              <v-tooltip location="top" text="Kirim Gambar">
                <template v-slot:activator="{ props }">
                  <v-btn
                    v-bind="props"
                    icon="mdi-image-outline"
                    variant="text"
                    size="small"
                    color="medium-emphasis"
                    class="flex-shrink-0 mb-1"
                    :loading="isUploadingImage"
                    @click="triggerImageSelect"
                  ></v-btn>
                </template>
              </v-tooltip>

              <!-- Text Input -->
              <v-textarea
                ref="chatInputRef"
                v-model="inputMessage"
                rows="1"
                auto-grow
                max-rows="4"
                density="compact"
                variant="outlined"
                rounded="xl"
                placeholder="Ketik balasan CS... (Enter untuk kirim, Shift+Enter baris baru)"
                hide-details
                class="chat-input-textarea flex-grow-1"
                @keydown.enter.exact.prevent="sendAdminMessage"
                @input="notifyAdminTyping"
                @focus="showEmojiPicker = false"
              ></v-textarea>

              <!-- Send Button -->
              <v-btn
                color="primary"
                icon="mdi-send"
                elevation="1"
                size="default"
                class="flex-shrink-0 mb-1"
                :disabled="!inputMessage.trim()"
                @click="sendAdminMessage"
                title="Kirim Pesan"
              ></v-btn>
            </div>
          </footer>
        </template>
      </main>

      <!-- ================= PANEL 3: CONTACT 360 (CUSTOMER PROFILE) ================= -->
      <aside
        v-if="showInfoPanel && activeRoom"
        class="customer-info-panel border-s d-flex flex-column bg-surface overflow-y-auto"
      >
        <!-- Panel Header -->
        <div class="px-4 py-3 border-b d-flex align-center justify-space-between flex-shrink-0">
          <div class="d-flex align-center gap-2">
            <v-icon size="18" color="primary">mdi-card-account-details-outline</v-icon>
            <h4 class="text-subtitle-2 font-weight-bold text-high-emphasis">Profil Pelanggan</h4>
          </div>
          <v-btn
            icon="mdi-close"
            variant="text"
            size="small"
            density="compact"
            @click="showInfoPanel = false"
          ></v-btn>
        </div>

        <div class="pa-4 d-flex flex-column gap-4">
          <!-- Profile Hero -->
          <div class="text-center pb-1">
            <v-avatar :color="getBrandColor(activeRoom.brand)" size="56" class="elevation-2 mb-2">
              <span class="text-h6 font-weight-bold text-white">
                {{ getInitials(activeRoom.pelanggan?.nama || 'Pelanggan') }}
              </span>
            </v-avatar>
            <h3 class="text-subtitle-1 font-weight-bold text-high-emphasis mb-0">
              {{ activeRoom.pelanggan?.nama || '-' }}
            </h3>
            <div class="text-caption text-medium-emphasis">
              ID: <strong>{{ activeRoom.pelanggan?.customer_id || ('#' + activeRoom.pelanggan_id) }}</strong>
            </div>
            <div class="mt-2">
              <v-chip :color="getBrandColor(activeRoom.brand)" size="x-small" variant="flat" class="font-weight-bold text-white px-2">
                {{ normalizeBrandName(activeRoom.brand) }}
              </v-chip>
            </div>
          </div>

          <v-divider></v-divider>

          <!-- Section: Kontak & Lokasi -->
          <div>
            <div class="text-overline text-medium-emphasis mb-2 font-weight-bold">Kontak & Lokasi</div>
            <v-card variant="tonal" class="pa-3 rounded-lg d-flex flex-column gap-2.5">
              <!-- Phone -->
              <div class="d-flex align-center justify-space-between">
                <div class="d-flex align-center gap-2 text-body-2">
                  <v-icon size="16" color="primary">mdi-phone-outline</v-icon>
                  <span class="font-weight-medium">{{ activeRoom.pelanggan?.no_telp || '-' }}</span>
                </div>
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

              <v-divider></v-divider>

              <!-- Alamat -->
              <div>
                <div class="d-flex align-center gap-2 text-caption text-medium-emphasis mb-1">
                  <v-icon size="14">mdi-map-marker-outline</v-icon>
                  <span>Alamat Pemasangan:</span>
                </div>
                <div class="text-caption text-high-emphasis font-weight-medium ps-5">
                  {{ activeRoom.pelanggan?.alamat || '-' }}
                  <span v-if="activeRoom.pelanggan?.blok || activeRoom.pelanggan?.unit" class="text-primary font-weight-bold d-block mt-0.5">
                    Blok {{ activeRoom.pelanggan?.blok || '-' }} / Unit {{ activeRoom.pelanggan?.unit || '-' }}
                  </span>
                </div>
              </div>
            </v-card>
          </div>

          <!-- Section: Layanan FTTH -->
          <div>
            <div class="text-overline text-medium-emphasis mb-2 font-weight-bold">Layanan FTTH</div>
            <v-card variant="tonal" class="pa-3 rounded-lg d-flex flex-column gap-2.5">
              <div class="d-flex align-center justify-space-between">
                <div class="d-flex align-center gap-2 text-caption text-medium-emphasis">
                  <v-icon size="16" color="primary">mdi-speedometer</v-icon>
                  <span>Paket:</span>
                </div>
                <span class="text-caption font-weight-bold text-primary">
                  {{ getCustomerPackage(activeRoom.pelanggan) }}
                </span>
              </div>

              <v-divider></v-divider>

              <div class="d-flex align-center justify-space-between">
                <div class="d-flex align-center gap-2 text-caption text-medium-emphasis">
                  <v-icon size="16" color="primary">mdi-shield-check-outline</v-icon>
                  <span>Status:</span>
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

              <template v-if="activeRoom.pelanggan?.data_teknis?.id_pelanggan">
                <v-divider></v-divider>
                <div class="d-flex align-center justify-space-between">
                  <div class="d-flex align-center gap-2 text-caption text-medium-emphasis">
                    <v-icon size="16" color="primary">mdi-account-key-outline</v-icon>
                    <span>PPPoE:</span>
                  </div>
                  <code class="text-caption font-weight-bold px-1.5 py-0.5 rounded bg-surface">
                    {{ activeRoom.pelanggan.data_teknis.id_pelanggan }}
                  </code>
                </div>
              </template>
            </v-card>
          </div>

          <!-- Action Buttons -->
          <div class="d-flex flex-column gap-2 pt-1">
            <v-btn
              block
              color="success"
              prepend-icon="mdi-whatsapp"
              variant="flat"
              class="text-none font-weight-bold rounded-lg"
              @click="openWhatsApp(activeRoom.pelanggan?.no_telp)"
            >
              Chat via WhatsApp Web
            </v-btn>
            <v-btn
              block
              variant="tonal"
              prepend-icon="mdi-account-search-outline"
              class="text-none font-weight-medium rounded-lg"
              @click="navigateToCustomer(activeRoom.pelanggan_id)"
            >
              Buka Data Pelanggan
            </v-btn>
          </div>
        </div>
      </aside>
    </v-card>

    <!-- Dialog: Image Preview & Caption Before Send -->
    <v-dialog v-model="imageUploadDialog" max-width="480" persistent>
      <v-card class="rounded-xl overflow-hidden">
        <v-card-title class="d-flex align-center justify-space-between px-4 py-3 border-b">
          <span class="text-subtitle-1 font-weight-bold">Kirim Gambar</span>
          <v-btn icon="mdi-close" variant="text" size="small" @click="cancelImageUpload"></v-btn>
        </v-card-title>
        <v-card-text class="pa-4 text-center">
          <div class="image-preview-box rounded-lg overflow-hidden mb-3 bg-grey-lighten-4 d-flex align-center justify-center" style="max-height: 280px; min-height: 180px;">
            <img
              v-if="selectedImagePreviewUrl"
              :src="selectedImagePreviewUrl"
              alt="Preview"
              style="max-width: 100%; max-height: 280px; object-fit: contain;"
            />
          </div>
          <v-text-field
            v-model="imageCaption"
            placeholder="Tambah keterangan gambar... (opsional)"
            variant="outlined"
            density="compact"
            hide-details
            rounded="lg"
            prepend-inner-icon="mdi-format-text"
            @keydown.enter.prevent="sendImageMessage"
          ></v-text-field>
        </v-card-text>
        <v-card-actions class="px-4 pb-4 pt-0 d-flex justify-end gap-2">
          <v-btn variant="text" rounded="pill" :disabled="isUploadingImage" @click="cancelImageUpload">
            Batal
          </v-btn>
          <v-btn
            color="primary"
            variant="flat"
            rounded="pill"
            prepend-icon="mdi-send"
            :loading="isUploadingImage"
            @click="sendImageMessage"
          >
            Kirim
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Dialog: Fullscreen Image Lightbox -->
    <v-dialog v-model="previewImageDialog" max-width="850">
      <v-card class="rounded-xl overflow-hidden bg-black" elevation="8">
        <div class="d-flex align-center justify-space-between px-4 py-2 bg-grey-darken-4 text-white">
          <span class="text-caption">Pratinjau Gambar</span>
          <div class="d-flex align-center gap-1">
            <v-btn
              icon="mdi-open-in-new"
              variant="text"
              size="small"
              color="white"
              title="Buka di tab baru"
              @click="openInNewTab(lightboxImageUrl)"
            ></v-btn>
            <v-btn
              icon="mdi-close"
              variant="text"
              size="small"
              color="white"
              @click="previewImageDialog = false"
            ></v-btn>
          </div>
        </div>
        <div class="pa-2 d-flex align-center justify-center bg-grey-darken-4" style="min-height: 300px; max-height: 80vh;">
          <img
            v-if="lightboxImageUrl"
            :src="lightboxImageUrl"
            alt="Fullscreen Preview"
            style="max-width: 100%; max-height: 75vh; object-fit: contain;"
          />
        </div>
      </v-card>
    </v-dialog>

    <!-- Global Snackbar -->
    <v-snackbar v-model="snackbar.show" :color="snackbar.color" :timeout="3000" location="top right">
      {{ snackbar.text }}
    </v-snackbar>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue';
import { useDisplay } from 'vuetify';
import { useRouter } from 'vue-router';
import apiClient from '@/services/api';
import { useAuthStore } from '@/stores/auth';
import EmojiPicker from 'vue3-emoji-picker';
import 'vue3-emoji-picker/css';
import chatNotificationSound from '@/assets/chat-notification.mp3';

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
const selectedStatus = ref<'open' | 'closed' | 'ALL'>('open');
const isUpdatingStatus = ref(false);
// Default false so chat has full spacious width when opened!
const showInfoPanel = ref(false);
const inputMessage = ref('');
const isCustomerTyping = ref(false);
const showEmojiPicker = ref(false);
const isSoundEnabled = ref(true);
const chatInputRef = ref<any>(null);
const isWsConnected = ref(false);

const messagesScrollContainer = ref<HTMLElement | null>(null);

// Image upload & preview state
const fileInputRef = ref<HTMLInputElement | null>(null);
const isUploadingImage = ref(false);
const imageUploadDialog = ref(false);
const selectedImageFile = ref<File | null>(null);
const selectedImagePreviewUrl = ref<string | null>(null);
const imageCaption = ref('');

// Lightbox state
const previewImageDialog = ref(false);
const lightboxImageUrl = ref('');

// Audio notification instance (preloaded)
let notificationAudio: HTMLAudioElement | null = null;

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

// Brand filter options with clean compact labels that never truncate
const brandFilters = [
  { label: 'Semua', value: 'ALL', color: 'primary' },
  { label: 'Jakinet', value: 'JAKINET', color: 'error' },
  { label: 'Jelantik', value: 'JELANTIK', color: 'primary' },
  { label: 'Nagrak', value: 'JELANTIK NAGRAK', color: 'success' },
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
  },
  {
    label: 'Tutup & Terima Kasih',
    icon: 'mdi-hand-heart-outline',
    text: 'Terima kasih telah menghubungi Customer Care Artacom. Jika tidak ada hal lain yang ditanyakan, percakapan ini akan kami tutup. Selamat beraktivitas!'
  }
];

// Computed unread total
const totalUnreadCount = computed(() => {
  return rooms.value.reduce((acc, r) => acc + (r.unread_count_admin || 0), 0);
});

// Computed open and closed room counts
const openRoomsCount = computed(() => {
  return rooms.value.filter((r) => !r.status || r.status === 'open').length;
});

const closedRoomsCount = computed(() => {
  return rooms.value.filter((r) => r.status === 'closed').length;
});

// Filtered rooms
const filteredRooms = computed(() => {
  let list = rooms.value;

  // Filter Status
  if (selectedStatus.value === 'open') {
    list = list.filter((r) => !r.status || r.status === 'open');
  } else if (selectedStatus.value === 'closed') {
    list = list.filter((r) => r.status === 'closed');
  }

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

// Silent polling as fallback only if WebSocket is offline
async function pollActiveRoomSilent() {
  if (document.hidden) return;
  // Jika WebSocket sedang online, tidak perlu polling HTTP untuk menghemat resource
  if (isWsConnected.value) return;

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

  // Auto reopen room if it was closed
  if (activeRoom.value.status === 'closed') {
    activeRoom.value.status = 'open';
    const rIdx = rooms.value.findIndex((r) => r.id === activeRoom.value.id);
    if (rIdx !== -1) {
      rooms.value[rIdx].status = 'open';
    }
  }

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

// Close & Reopen Room Actions
async function closeActiveRoom() {
  if (!activeRoom.value) return;
  const roomId = activeRoom.value.id;
  const custName = activeRoom.value.pelanggan?.nama || 'Pelanggan';
  isUpdatingStatus.value = true;
  try {
    await apiClient.post(`/chat/rooms/${roomId}/close`);
    activeRoom.value.status = 'closed';
    const roomIdx = rooms.value.findIndex((r) => r.id === roomId);
    if (roomIdx !== -1) {
      rooms.value[roomIdx].status = 'closed';
    }
    sendWsEvent('close_room', { room_id: roomId });
    showSnackbar(`Percakapan dengan ${custName} telah diselesaikan.`, 'success');
  } catch (err: any) {
    showSnackbar('Gagal menutup percakapan: ' + (err.response?.data?.error || err.message), 'error');
  } finally {
    isUpdatingStatus.value = false;
  }
}

async function reopenActiveRoom() {
  if (!activeRoom.value) return;
  const roomId = activeRoom.value.id;
  isUpdatingStatus.value = true;
  try {
    await apiClient.post(`/chat/rooms/${roomId}/reopen`);
    activeRoom.value.status = 'open';
    const roomIdx = rooms.value.findIndex((r) => r.id === roomId);
    if (roomIdx !== -1) {
      rooms.value[roomIdx].status = 'open';
    }
    sendWsEvent('reopen_room', { room_id: roomId });
    showSnackbar('Percakapan telah dibuka kembali.', 'info');
  } catch (err: any) {
    showSnackbar('Gagal membuka percakapan: ' + (err.response?.data?.error || err.message), 'error');
  } finally {
    isUpdatingStatus.value = false;
  }
}

function useTemplate(tpl: any) {
  inputMessage.value = tpl.text;
}

// Emoji picker handler
function onSelectEmoji(emoji: any) {
  inputMessage.value += emoji.i;
  showEmojiPicker.value = false;
}

// Sound notification
function playNotificationSound() {
  if (!isSoundEnabled.value) return;
  try {
    if (!notificationAudio) {
      notificationAudio = new Audio(chatNotificationSound);
      notificationAudio.volume = 0.6;
    }
    notificationAudio.currentTime = 0;
    notificationAudio.play().catch(() => {
      // Browser may block autoplay until user interacts
    });
  } catch (_) {}
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
      // Catch up on missed messages/rooms after reconnect
      fetchRooms(true);
      if (activeRoom.value) {
        pollActiveRoomSilent();
      }
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
  } catch (err) {
    console.error('Failed to init WebSocket:', err);
    isWsConnected.value = false;
  }
}

function handleWsIncoming(payload: any) {
  const { event, data } = payload;
  if (!event || !data) return;

  switch (event) {
    case 'ack_sent':
    case 'ack_message': {
      // Sent (Ceklis 1)
      const index = activeMessages.value.findIndex(
        (m) => (data.temp_id && m.temp_id === data.temp_id) || (data.id && m.id === data.id)
      );
      if (index !== -1) {
        activeMessages.value[index].id = data.id;
        activeMessages.value[index].status = data.status || 'sent';
      }
      break;
    }

    case 'message_error': {
      showSnackbar('Gagal mengirim pesan: ' + (data.error || 'Terjadi kesalahan'), 'error');
      const index = activeMessages.value.findIndex(
        (m) => data.temp_id && m.temp_id === data.temp_id
      );
      if (index !== -1) {
        activeMessages.value[index].status = 'error';
      }
      break;
    }

    case 'new_message': {
      const roomIdx = rooms.value.findIndex((r) => r.id === data.room_id);
      if (roomIdx !== -1) {
        rooms.value[roomIdx].last_message_text = data.message;
        rooms.value[roomIdx].last_message_at = data.created_at;
        rooms.value[roomIdx].status = 'open'; // Auto reopen room on new message!
        if (roomIdx > 0) {
          const [moved] = rooms.value.splice(roomIdx, 1);
          rooms.value.unshift(moved);
        }
      } else {
        fetchRooms(true);
      }

      // Play notification sound for incoming customer messages
      if (data.sender_type === 'customer') {
        playNotificationSound();
      }

      if (activeRoom.value && activeRoom.value.id === data.room_id) {
        activeRoom.value.status = 'open';
        const exists = activeMessages.value.some(
          (m) => (data.id && m.id === data.id) || (data.temp_id && m.temp_id === data.temp_id)
        );
        if (!exists) {
          activeMessages.value.push(data);
          scrollToBottom();
        }

        if (data.sender_type === 'customer') {
          apiClient.post(`/chat/messages/${data.room_id}/read?reader_type=admin`).catch(() => {});
          sendWsEvent('read_room', { room_id: data.room_id });
        }
      } else {
        if (roomIdx !== -1) {
          rooms.value[0].unread_count_admin = (rooms.value[0].unread_count_admin || 0) + 1;
        }
        showSnackbar(`Pesan baru dari ${data.sender_name || 'Pelanggan'}`, 'info');
      }
      break;
    }

    case 'room_status_update': {
      const roomIdx = rooms.value.findIndex((r) => r.id === data.room_id);
      if (roomIdx !== -1) {
        rooms.value[roomIdx].status = data.status;
      }
      if (activeRoom.value && activeRoom.value.id === data.room_id) {
        activeRoom.value.status = data.status;
      }
      break;
    }

    case 'message_status_update': {
      // Delivered or Read
      const index = activeMessages.value.findIndex(
        (m) => m.id === data.id || (data.temp_id && m.temp_id === data.temp_id)
      );
      if (index !== -1) {
        activeMessages.value[index].status = data.status;
      }
      break;
    }

    case 'room_read': {
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

// Media URL helper
function getFullMediaUrl(url?: string): string {
  if (!url) return '';
  if (url.startsWith('http://') || url.startsWith('https://')) return url;
  if (import.meta.env.DEV) {
    const protocol = window.location.protocol;
    const hostname = window.location.hostname;
    return `${protocol}//${hostname}:8000${url.startsWith('/') ? '' : '/'}${url}`;
  }
  return url;
}

// Image upload handlers
function triggerImageSelect() {
  if (fileInputRef.value) {
    fileInputRef.value.value = '';
    fileInputRef.value.click();
  }
}

// Helper: Compress large image using HTML5 Canvas before uploading
async function compressImage(file: File, maxDimension = 1920, quality = 0.82): Promise<File> {
  // If not an image or is a GIF (preserve animation) or already small (< 400KB), return as is
  if (!file.type.startsWith('image/') || file.type === 'image/gif' || file.size < 400 * 1024) {
    return file;
  }

  return new Promise((resolve) => {
    const img = new Image();
    const reader = new FileReader();

    reader.onload = (e) => {
      img.onload = () => {
        let width = img.width;
        let height = img.height;

        // Calculate aspect ratio downscaling if larger than maxDimension
        if (width > maxDimension || height > maxDimension) {
          if (width > height) {
            height = Math.round((height * maxDimension) / width);
            width = maxDimension;
          } else {
            width = Math.round((width * maxDimension) / height);
            height = maxDimension;
          }
        }

        const canvas = document.createElement('canvas');
        canvas.width = width;
        canvas.height = height;
        const ctx = canvas.getContext('2d');
        if (!ctx) {
          resolve(file);
          return;
        }

        ctx.drawImage(img, 0, 0, width, height);

        canvas.toBlob(
          (blob) => {
            if (!blob || blob.size >= file.size) {
              resolve(file);
            } else {
              const newName = file.name.replace(/\.[^/.]+$/, '') + '.jpg';
              const compressedFile = new File([blob], newName, {
                type: 'image/jpeg',
                lastModified: Date.now(),
              });
              resolve(compressedFile);
            }
          },
          'image/jpeg',
          quality
        );
      };
      img.onerror = () => resolve(file);
      img.src = e.target?.result as string;
    };
    reader.onerror = () => resolve(file);
    reader.readAsDataURL(file);
  });
}

function onImageSelected(e: Event) {
  const target = e.target as HTMLInputElement;
  if (!target.files || target.files.length === 0) return;

  const file = target.files[0];
  if (!file.type.startsWith('image/')) {
    showSnackbar('Hanya file gambar yang didukung (JPG, PNG, WEBP, GIF)', 'error');
    return;
  }

  // Max 25MB
  if (file.size > 25 * 1024 * 1024) {
    showSnackbar('Ukuran gambar maksimal 25MB', 'error');
    return;
  }

  selectedImageFile.value = file;
  selectedImagePreviewUrl.value = URL.createObjectURL(file);
  imageCaption.value = '';
  imageUploadDialog.value = true;
}

function cancelImageUpload() {
  imageUploadDialog.value = false;
  if (selectedImagePreviewUrl.value) {
    URL.revokeObjectURL(selectedImagePreviewUrl.value);
    selectedImagePreviewUrl.value = null;
  }
  selectedImageFile.value = null;
  imageCaption.value = '';
}

async function sendImageMessage() {
  if (!selectedImageFile.value || !activeRoom.value) return;

  isUploadingImage.value = true;
  try {
    // Automatically compress large image before upload for ultra-fast, lightweight transfer
    const fileToUpload = await compressImage(selectedImageFile.value);

    const formData = new FormData();
    formData.append('file', fileToUpload);

    const uploadRes = await apiClient.post('/uploads/chat', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });

    const fileUrl = uploadRes.data?.file_url;
    if (!fileUrl) {
      throw new Error('Gagal mendapatkan URL gambar');
    }

    const captionText = imageCaption.value.trim();
    const tempId = `admin_img_${Date.now()}`;

    // Auto reopen room if it was closed
    if (activeRoom.value.status === 'closed') {
      activeRoom.value.status = 'open';
      const rIdx = rooms.value.findIndex((r) => r.id === activeRoom.value.id);
      if (rIdx !== -1) {
        rooms.value[rIdx].status = 'open';
      }
    }

    const localMsg = {
      id: null,
      room_id: activeRoom.value.id,
      sender_type: 'admin',
      sender_name: authStore.user?.name || 'Admin CS',
      message: captionText,
      message_type: 'image',
      attachment_url: fileUrl,
      status: 'pending',
      created_at: new Date().toISOString(),
      temp_id: tempId,
    };

    activeMessages.value.push(localMsg);
    activeRoom.value.last_message_text = captionText || '[Gambar]';
    activeRoom.value.last_message_at = new Date().toISOString();

    // Move to top
    const currentRoomId = activeRoom.value.id;
    const roomIdx = rooms.value.findIndex((r) => r.id === currentRoomId);
    if (roomIdx > 0) {
      const [moved] = rooms.value.splice(roomIdx, 1);
      rooms.value.unshift(moved);
    }

    cancelImageUpload();
    scrollToBottom();

    // Send via WebSocket
    sendWsEvent('send_message', {
      room_id: activeRoom.value.id,
      message: captionText,
      message_type: 'image',
      attachment_url: fileUrl,
      temp_id: tempId,
    });
  } catch (err: any) {
    showSnackbar('Gagal mengirim gambar: ' + (err.response?.data?.error || err.message), 'error');
  } finally {
    isUploadingImage.value = false;
  }
}

function openImageLightbox(url?: string) {
  if (!url) return;
  lightboxImageUrl.value = getFullMediaUrl(url);
  previewImageDialog.value = true;
}

function openInNewTab(url?: string) {
  if (!url) return;
  window.open(url, '_blank');
}

// Esc Key Handler to cleanly exit room
function handleKeyDown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    if (previewImageDialog.value) {
      previewImageDialog.value = false;
      return;
    }
    if (imageUploadDialog.value) {
      cancelImageUpload();
      return;
    }
    if (showEmojiPicker.value) {
      showEmojiPicker.value = false;
      return;
    }
    if (showInfoPanel.value) {
      showInfoPanel.value = false;
      return;
    }
    if (activeRoom.value) {
      activeRoom.value = null;
    }
  }
}

// Lifecycle hooks
onMounted(() => {
  fetchRooms();
  initWebSocket();
  // Fallback background polling (20 detik) hanya jika WebSocket offline
  pollingTimer = setInterval(pollActiveRoomSilent, 20000);
  document.addEventListener('visibilitychange', onVisibilityChange);
  window.addEventListener('keydown', handleKeyDown);
});

onUnmounted(() => {
  stopHeartbeat();
  if (pollingTimer) {
    clearInterval(pollingTimer);
    pollingTimer = null;
  }
  document.removeEventListener('visibilitychange', onVisibilityChange);
  window.removeEventListener('keydown', handleKeyDown);
  clearTimeout(typingClearTimer);
  clearTimeout(searchDebounceTimer);
  if (ws) {
    ws.close();
    ws = null;
  }
});
</script>

<style scoped>
.customer-chat-wrapper {
  height: calc(100vh - 150px);
  max-height: calc(100vh - 150px);
  box-sizing: border-box;
}

.chat-workspace {
  display: flex;
  height: 100%;
  width: 100%;
  overflow: hidden;
  position: relative;
}

/* Gap Utilities for Vuetify Flexbox */
.gap-1 { gap: 4px; }
.gap-1\.5 { gap: 6px; }
.gap-2 { gap: 8px; }
.gap-2\.5 { gap: 10px; }
.gap-3 { gap: 12px; }
.gap-4 { gap: 16px; }

/* ================= PANEL 1: SIDEBAR ================= */
.rooms-sidebar {
  width: 320px;
  min-width: 300px;
  max-width: 340px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: rgb(var(--v-theme-surface));
}

.brand-filter-bar::-webkit-scrollbar,
.quick-replies-bar::-webkit-scrollbar {
  display: none;
}

.brand-filter-chip {
  height: 26px;
  font-size: 0.72rem;
  letter-spacing: 0.2px;
}

.room-card {
  transition: background-color 0.15s ease, transform 0.1s ease;
  border: 1px solid transparent;
  border-left: 3px solid transparent;
}

.room-card:hover {
  background-color: rgba(var(--v-theme-primary), 0.05);
}

.room-card--active {
  background-color: rgba(var(--v-theme-primary), 0.08) !important;
  border-color: rgba(var(--v-theme-primary), 0.2) !important;
  border-left-color: rgb(var(--v-theme-primary)) !important;
}

/* ================= PANEL 2: CONVERSATION ================= */
.chat-conversation-panel {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: #f8fafc;
}

.v-theme--dark .chat-conversation-panel {
  background-color: #0b1120;
}

.messages-container {
  background-color: transparent;
}

.message-bubble {
  max-width: 72%;
  min-width: 130px;
  position: relative;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.bubble-customer {
  background-color: #ffffff;
  color: #1e293b;
  border-radius: 16px 16px 16px 4px !important;
  border: 1px solid rgba(226, 232, 240, 0.9);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.v-theme--dark .bubble-customer {
  background-color: #1e293b;
  color: #f8fafc;
  border-color: rgba(51, 65, 85, 0.8);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
}

.bubble-admin {
  background: linear-gradient(135deg, #1d4ed8, #2563eb);
  color: #ffffff;
  border-radius: 16px 16px 4px 16px !important;
  box-shadow: 0 2px 6px rgba(37, 99, 235, 0.25);
}

.date-pill {
  background-color: rgba(226, 232, 240, 0.9);
  color: #475569;
  border: 1px solid rgba(203, 213, 225, 0.7);
  font-size: 0.7rem;
  letter-spacing: 0.2px;
}

.v-theme--dark .date-pill {
  background-color: #1e293b;
  color: #94a3b8;
  border-color: #334155;
}

.quick-chip {
  transition: transform 0.1s ease;
}

.quick-chip:hover {
  transform: translateY(-1px);
}

/* ================= EMOJI PICKER ================= */
.emoji-picker-popup {
  position: absolute;
  bottom: 48px;
  left: 0;
  z-index: 100;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.18);
  border-radius: 12px;
  overflow: hidden;
}

/* ================= PANEL 3: CONTACT 360 ================= */
.customer-info-panel {
  width: 320px;
  min-width: 300px;
  max-width: 340px;
  flex-shrink: 0;
  height: 100%;
  background-color: rgb(var(--v-theme-surface));
}

.animate-pulse {
  animation: pulse 1.8s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

/* Slim modern scrollbars */
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
