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

        <!-- Search & Brand Filter Row -->
        <div class="px-3 pt-3 pb-2 d-flex align-center gap-1.5">
          <v-text-field
            v-model="searchQuery"
            placeholder="Cari chat..."
            prepend-inner-icon="mdi-magnify"
            variant="outlined"
            density="compact"
            hide-details
            clearable
            rounded="lg"
            class="search-input flex-grow-1"
            @update:model-value="onSearchDebounced"
          ></v-text-field>

          <!-- Brand Filter Dropdown Menu (Eliminates messy extra chip row) -->
          <v-menu location="bottom end">
            <template v-slot:activator="{ props }">
              <v-btn
                v-bind="props"
                variant="tonal"
                :color="selectedBrand === 'ALL' ? 'default' : getBrandColor(selectedBrand)"
                size="small"
                class="text-none font-weight-bold px-2 rounded-lg flex-shrink-0"
                style="height: 40px; min-width: 40px;"
                title="Filter Channel / Brand"
              >
                <v-icon size="16">mdi-filter-variant</v-icon>
                <span class="ms-1 d-none d-sm-inline" style="font-size: 0.74rem;">
                  {{ getBrandShortLabel(selectedBrand) }}
                </span>
                <v-icon size="12" class="ms-0.5">mdi-chevron-down</v-icon>
              </v-btn>
            </template>
            <v-list density="compact" class="py-1 rounded-lg elevation-3">
              <v-list-item
                v-for="filter in brandFilters"
                :key="filter.value"
                :value="filter.value"
                :active="selectedBrand === filter.value"
                @click="setBrandFilter(filter.value)"
              >
                <template v-slot:prepend>
                  <v-icon size="14" :color="filter.color || 'primary'">mdi-circle-medium</v-icon>
                </template>
                <v-list-item-title class="text-caption font-weight-medium">
                  {{ filter.label }}
                </v-list-item-title>
              </v-list-item>
            </v-list>
          </v-menu>
        </div>

        <!-- Omnichannel Status Tabs (Terbuka | Selesai | Semua) -->
        <div class="px-3 pb-2">
          <div class="omnichannel-status-tabs">
            <button
              v-for="tab in statusTabs"
              :key="tab.value"
              type="button"
              class="omnichannel-tab-btn"
              :class="{ 'omnichannel-tab-btn--active': selectedStatusTab === tab.value }"
              @click="selectedStatusTab = tab.value"
            >
              <span>{{ tab.label }}</span>
              <span class="omnichannel-tab-count">{{ tab.count }}</span>
            </button>
          </div>
        </div>

        <!-- Assignment Segment Bar (3 Segments: Semua Tim, Saya, Unassigned) -->
        <div class="px-3 pb-2">
          <div class="assignment-filter-bar">
            <button
              v-for="tab in assignmentTabs"
              :key="tab.value"
              type="button"
              class="assignment-tab-btn"
              :class="{ 'assignment-tab-btn--active': selectedAssignment === tab.value }"
              @click="selectedAssignment = tab.value"
            >
              <v-icon size="12" class="me-1">{{ tab.icon }}</v-icon>
              <span>{{ tab.label }}</span>
              <span
                v-if="tab.count !== undefined && tab.count > 0"
                class="assignment-tab-count ms-1"
              >
                {{ tab.count }}
              </span>
            </button>
          </div>
        </div>

        <v-divider></v-divider>

        <!-- Rooms List -->
        <div class="rooms-list-scroll flex-grow-1 overflow-y-auto">
          <div v-if="isLoadingRooms" class="pa-8 text-center">
            <v-progress-circular indeterminate color="primary" size="26"></v-progress-circular>
            <div class="text-caption text-medium-emphasis mt-2">Memuat percakapan...</div>
          </div>

          <div v-else-if="filteredRooms.length === 0" class="pa-8 text-center text-medium-emphasis">
            <v-icon size="36" color="grey-lighten-1" class="mb-2">mdi-message-text-outline</v-icon>
            <div class="text-body-2 font-weight-medium">Tidak ada percakapan</div>
            <div class="text-caption">
              {{ selectedStatusTab === 'resolved' ? 'Belum ada percakapan yang selesai' : (selectedStatusTab === 'open' ? 'Belum ada obrolan aktif' : 'Belum ada riwayat percakapan') }}
            </div>
          </div>

          <div v-else class="d-flex flex-column">
            <div
              v-for="room in filteredRooms"
              :key="room.id"
              class="room-card px-3 py-3 cursor-pointer"
              :class="{ 'room-card--active': activeRoom?.id === room.id }"
              @click="selectRoom(room)"
            >
              <div class="d-flex align-start gap-3">
                <!-- Customer Avatar -->
                <div class="position-relative flex-shrink-0 mt-0.5">
                  <v-avatar :color="getBrandColor(room.brand)" size="38" class="rounded-circle">
                    <span class="text-caption font-weight-bold text-white">
                      {{ getInitials(room.pelanggan?.nama || 'Pelanggan') }}
                    </span>
                  </v-avatar>
                  <span
                    v-if="room.status !== 'closed'"
                    class="position-absolute rounded-circle border border-white bg-success"
                    style="width: 9px; height: 9px; bottom: 0; right: 0;"
                  ></span>
                </div>

                <!-- Details -->
                <div class="flex-grow-1 min-w-0">
                  <!-- Row 1: Name & Timestamp -->
                  <div class="d-flex align-baseline justify-space-between mb-1">
                    <span class="room-title text-truncate flex-grow-1 me-2 font-weight-bold text-high-emphasis" style="font-size: 0.85rem; line-height: 1.25;">
                      {{ room.pelanggan?.nama || 'Pelanggan #' + room.pelanggan_id }}
                    </span>
                    <span class="room-time text-caption text-medium-emphasis flex-shrink-0" style="font-size: 0.7rem; font-variant-numeric: tabular-nums;">
                      {{ formatTimestamp(room.last_message_at) }}
                    </span>
                  </div>

                  <!-- Row 2: Last message snippet & unread badge -->
                  <div class="d-flex align-center justify-space-between mb-1.5">
                    <span class="room-snippet text-caption text-truncate flex-grow-1 me-2 text-medium-emphasis" style="font-size: 0.77rem; line-height: 1.35;">
                      {{ room.last_message_text || 'Mulai obrolan...' }}
                    </span>
                    <span
                      v-if="room.unread_count_admin > 0"
                      class="room-unread-badge flex-shrink-0"
                    >
                      {{ room.unread_count_admin > 99 ? '99+' : room.unread_count_admin }}
                    </span>
                  </div>

                  <!-- Row 3: Unified Meta Row -->
                  <div class="d-flex align-center gap-1.5 flex-nowrap overflow-hidden">
                    <!-- Brand Pill -->
                    <span class="room-brand-chip">
                      {{ getBrandShortLabel(room.brand) }}
                    </span>

                    <!-- Resolved Tag (Muted text tag) -->
                    <span
                      v-if="room.status === 'closed'"
                      class="room-subtle-tag text-caption"
                    >
                      &bull; Resolved
                    </span>

                    <!-- Assignment Tag -->
                    <span
                      v-if="room.assigned_admin_id === authStore.user?.id"
                      class="room-assign-chip room-assign-chip--me d-inline-flex align-center gap-0.5"
                    >
                      <v-icon size="10">mdi-account-check</v-icon>
                      Saya
                    </span>
                    <span
                      v-else-if="!room.assigned_admin_id"
                      class="room-assign-chip room-assign-chip--unassigned"
                    >
                      Unassigned
                    </span>
                    <span
                      v-else
                      class="room-assign-chip room-assign-chip--other text-truncate"
                      style="max-width: 80px;"
                    >
                      {{ room.assigned_admin?.nama || 'Admin' }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Omnichannel Sidebar Bottom Summary Footer -->
        <div class="sidebar-footer px-4 py-2 border-t bg-surface d-flex align-center justify-space-between text-caption font-weight-medium text-medium-emphasis">
          <div class="d-flex align-center gap-1.5">
            <v-icon size="13" color="deep-purple">mdi-account-check</v-icon>
            <span>Assigned: <strong class="text-high-emphasis">{{ assignedRoomsCount }}</strong></span>
          </div>
          <div class="d-flex align-center gap-1.5">
            <v-icon size="13" color="amber-darken-3">mdi-account-clock-outline</v-icon>
            <span>Unassigned: <strong class="text-high-emphasis">{{ unassignedRoomsCount }}</strong></span>
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
            <!-- Header Action Buttons (Streamlined & Compact to fit on 1 line at 100% zoom) -->
            <div class="d-flex align-center gap-1.5 flex-shrink-0 ms-2">
              <!-- Assignment Status & Action -->
              <template v-if="activeRoom.status !== 'closed'">
                <!-- If assigned to logged-in user: compact chip with unassign button -->
                <div v-if="activeRoom.assigned_admin_id === authStore.user?.id" class="d-flex align-center gap-1">
                  <v-chip
                    color="primary"
                    variant="flat"
                    size="small"
                    class="font-weight-bold text-white px-2"
                    prepend-icon="mdi-account-check"
                    style="height: 28px; font-size: 0.72rem;"
                  >
                    Saya
                  </v-chip>
                  <v-tooltip location="bottom" text="Lepas Penugasan">
                    <template v-slot:activator="{ props }">
                      <v-btn
                        v-bind="props"
                        variant="tonal"
                        color="grey-darken-1"
                        size="small"
                        icon="mdi-account-minus-outline"
                        density="compact"
                        style="height: 28px; width: 28px;"
                        :loading="isAssigning"
                        @click="unassignActiveRoom"
                      ></v-btn>
                    </template>
                  </v-tooltip>
                </div>

                <!-- If assigned to someone else -->
                <div v-else-if="activeRoom.assigned_admin_id" class="d-flex align-center gap-1">
                  <v-chip
                    color="deep-purple"
                    variant="tonal"
                    size="small"
                    class="font-weight-bold px-2 text-truncate"
                    prepend-icon="mdi-account-outline"
                    style="height: 28px; font-size: 0.72rem; max-width: 100px;"
                  >
                    {{ activeRoom.assigned_admin?.nama || 'Admin' }}
                  </v-chip>
                  <v-btn
                    variant="flat"
                    color="primary"
                    size="small"
                    class="text-none font-weight-bold rounded-pill px-2.5"
                    style="height: 28px; font-size: 0.72rem;"
                    :loading="isAssigning"
                    @click="assignRoomToMe"
                  >
                    Ambil Alih
                  </v-btn>
                </div>

                <!-- If unassigned -->
                <div v-else class="d-flex align-center gap-1">
                  <v-btn
                    variant="flat"
                    color="primary"
                    size="small"
                    prepend-icon="mdi-hand-back-left-outline"
                    class="text-none font-weight-bold rounded-pill px-2.5"
                    style="height: 28px; font-size: 0.72rem;"
                    :loading="isAssigning"
                    @click="assignRoomToMe"
                  >
                    Ambil Alih
                  </v-btn>
                </div>
              </template>

              <!-- Trouble Ticket Quick Open Tooltip Button -->
              <v-tooltip location="bottom" text="Buat Trouble Ticket">
                <template v-slot:activator="{ props }">
                  <v-btn
                    v-bind="props"
                    icon="mdi-ticket-alert-outline"
                    variant="tonal"
                    color="primary"
                    size="small"
                    density="compact"
                    style="height: 28px; width: 28px;"
                    @click="openTroubleTicketPanel"
                  ></v-btn>
                </template>
              </v-tooltip>

              <!-- Resolve / Selesaikan Conversation Button -->
              <v-btn
                v-if="activeRoom.status !== 'closed'"
                variant="tonal"
                color="deep-purple"
                size="small"
                prepend-icon="mdi-check-circle-outline"
                class="text-none font-weight-bold rounded-lg px-2.5"
                style="height: 28px; font-size: 0.72rem;"
                :loading="isUpdatingStatus"
                @click="openCloseRoomDialog"
              >
                Resolve
              </v-btn>
              <v-btn
                v-else
                variant="tonal"
                color="primary"
                size="small"
                prepend-icon="mdi-lock-open-outline"
                class="text-none font-weight-bold rounded-lg px-2.5"
                style="height: 28px; font-size: 0.72rem;"
                :loading="isUpdatingStatus"
                @click="reopenActiveRoom"
              >
                Buka Kembali
              </v-btn>

              <!-- WhatsApp Quick Open Tooltip Button -->
              <v-tooltip location="bottom" text="Buka di WhatsApp">
                <template v-slot:activator="{ props }">
                  <v-btn
                    v-bind="props"
                    v-if="activeRoom.pelanggan?.no_telp"
                    variant="tonal"
                    color="success"
                    size="small"
                    icon="mdi-whatsapp"
                    density="compact"
                    style="height: 28px; width: 28px;"
                    @click="openWhatsApp(activeRoom.pelanggan.no_telp)"
                  ></v-btn>
                </template>
              </v-tooltip>

              <!-- Sound Toggle -->
              <v-tooltip location="bottom" :text="isSoundEnabled ? 'Suara Notifikasi Aktif' : 'Suara Notifikasi Mati'">
                <template v-slot:activator="{ props }">
                  <v-btn
                    v-bind="props"
                    :icon="isSoundEnabled ? 'mdi-volume-high' : 'mdi-volume-off'"
                    variant="text"
                    size="small"
                    density="compact"
                    style="height: 28px; width: 28px;"
                    :color="isSoundEnabled ? 'primary' : 'medium-emphasis'"
                    @click="isSoundEnabled = !isSoundEnabled"
                  ></v-btn>
                </template>
              </v-tooltip>

              <!-- Panel 3 (Profil 360 / Ticket) Toggle Button -->
              <v-tooltip location="bottom" :text="showInfoPanel ? 'Tutup Panel Samping' : 'Panel Profil & Trouble Ticket'">
                <template v-slot:activator="{ props }">
                  <v-btn
                    v-bind="props"
                    :icon="showInfoPanel ? 'mdi-close' : 'mdi-dock-right'"
                    variant="tonal"
                    size="small"
                    density="compact"
                    style="height: 28px; width: 28px;"
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
                    density="compact"
                    style="height: 28px; width: 28px;"
                    color="medium-emphasis"
                    @click="activeRoom = null"
                  ></v-btn>
                </template>
              </v-tooltip>
            </div>
          </header>

          <!-- Messages Stream Area -->
          <div
            ref="messagesScrollContainer"
            class="messages-container flex-grow-1 overflow-y-auto chat-messages-stream"
          >
            <div v-if="isLoadingMessages" class="text-center pa-8">
              <v-progress-circular indeterminate color="primary" size="30"></v-progress-circular>
              <div class="text-caption text-medium-emphasis mt-2">Memuat riwayat obrolan...</div>
            </div>

            <!-- Empty state if no messages -->
            <div
              v-else-if="activeMessages.length === 0"
              class="d-flex flex-column align-center justify-center fill-height text-center text-medium-emphasis pa-6"
            >
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
                <!-- Date Separator -->
                <div v-if="shouldShowDateHeader(idx)" class="date-divider my-3">
                  <span class="date-divider-text">{{ formatDateHeader(msg.created_at) }}</span>
                </div>

                <!-- Message Bubble Row -->
                <div
                  class="message-row d-flex"
                  :class="msg.sender_type === 'admin' ? 'justify-end' : 'justify-start'"
                >
                  <div
                    class="message-bubble"
                    :class="{
                      'bubble-admin': msg.sender_type === 'admin',
                      'bubble-customer': msg.sender_type === 'customer',
                      'bubble-system': msg.sender_type === 'system'
                    }"
                  >
                    <!-- Header for System / AI message -->
                    <div v-if="msg.sender_type === 'system'" class="d-flex align-center gap-1.5 mb-1.5 text-primary">
                      <v-icon size="14" color="primary">mdi-robot-outline</v-icon>
                      <span class="text-caption font-weight-bold" style="font-size: 0.75rem;">Asisten Virtual AI</span>
                    </div>
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

          <!-- Quick Template Replies Bar (Only shown when room is Open) -->
          <div
            v-if="activeRoom.status !== 'closed'"
            class="quick-replies-bar px-3 py-1.5 bg-surface border-t d-flex align-center gap-2 overflow-x-auto"
          >
            <v-icon size="14" color="amber-darken-2" class="flex-shrink-0">mdi-lightning-bolt</v-icon>
            <span class="text-caption font-weight-bold text-medium-emphasis flex-shrink-0" style="font-size: 0.72rem;">Templat:</span>
            <div class="d-flex align-center gap-1.5 flex-nowrap overflow-x-auto py-0.5 flex-grow-1 quick-replies-scroll">
              <v-chip
                v-for="(tpl, tIdx) in quickTemplates"
                :key="tpl.id || tpl.shortcut || tIdx"
                size="x-small"
                variant="tonal"
                color="primary"
                class="cursor-pointer font-weight-medium flex-shrink-0 quick-chip px-2.5"
                style="height: 24px; font-size: 0.72rem;"
                :prepend-icon="tpl.icon || 'mdi-message-text-outline'"
                @click="useTemplate(tpl)"
              >
                {{ tpl.title || tpl.label }}
              </v-chip>
            </div>
            <!-- Kelola Template Button -->
            <v-btn
              variant="text"
              size="x-small"
              prepend-icon="mdi-cog-outline"
              color="medium-emphasis"
              class="flex-shrink-0 text-caption font-weight-medium px-1.5"
              style="font-size: 0.7rem;"
              @click="openManageTemplates"
            >
              Kelola
            </v-btn>
          </div>

          <!-- Reopen Notice Banner when Room is Resolved (Clean & Slim) -->
          <div
            v-if="activeRoom.status === 'closed'"
            class="px-4 py-2 bg-grey-lighten-4 border-t d-flex align-center justify-space-between text-caption text-medium-emphasis"
          >
            <div class="d-flex align-center gap-2">
              <v-icon size="16" color="success">mdi-check-circle-outline</v-icon>
              <span style="font-size: 0.75rem;">
                Percakapan ini berstatus <strong>Resolved (Selesai)</strong>. Mengetik balasan akan otomatis membuka obrolan kembali.
              </span>
            </div>
            <v-btn
              variant="tonal"
              size="x-small"
              color="primary"
              class="text-none font-weight-bold rounded-pill px-3"
              style="height: 24px; font-size: 0.7rem;"
              prepend-icon="mdi-lock-open-outline"
              :loading="isUpdatingStatus"
              @click="reopenActiveRoom"
            >
              Buka Kembali
            </v-btn>
          </div>

          <!-- Bottom Chat Input Bar -->
          <footer class="chat-input-bar px-4 py-3 bg-surface border-t" style="position: relative;">
            <!-- Omnichannel Reply / Note Mode Toggle -->
            <div class="d-flex align-center justify-space-between mb-2">
              <div class="d-flex align-center gap-1 bg-grey-lighten-4 pa-1 rounded-pill">
                <v-btn
                  size="x-small"
                  :variant="inputMode === 'reply' ? 'flat' : 'text'"
                  :color="inputMode === 'reply' ? 'primary' : 'medium-emphasis'"
                  class="text-none font-weight-bold rounded-pill px-3"
                  style="font-size: 0.72rem; height: 24px;"
                  @click="inputMode = 'reply'"
                >
                  <v-icon size="12" class="me-1">mdi-reply</v-icon>
                  Balas Pelanggan
                </v-btn>
                <v-btn
                  size="x-small"
                  :variant="inputMode === 'note' ? 'flat' : 'text'"
                  :color="inputMode === 'note' ? 'amber-darken-3' : 'medium-emphasis'"
                  class="text-none font-weight-bold rounded-pill px-3"
                  style="font-size: 0.72rem; height: 24px;"
                  @click="inputMode = 'note'"
                >
                  <v-icon size="12" class="me-1">mdi-note-text-outline</v-icon>
                  Catatan Internal
                </v-btn>
              </div>
            </div>

            <!-- Slash Command Floating Autocomplete Popover -->
            <div v-if="showSlashMenu && filteredSlashTemplates.length > 0" class="slash-popup-menu elevation-4">
              <div class="slash-popup-header d-flex align-center justify-space-between px-3 py-1.5 border-b bg-slate-50">
                <div class="d-flex align-center gap-1.5 text-caption font-weight-bold text-medium-emphasis">
                  <v-icon size="14" color="primary">mdi-lightning-bolt</v-icon>
                  <span>Template Pintasan</span>
                  <span class="text-caption text-medium-emphasis">({{ filteredSlashTemplates.length }})</span>
                </div>
                <span class="text-caption text-medium-emphasis">[Enter] pilih &bull; [&uarr;&darr;] navigasi &bull; [Esc] tutup</span>
              </div>
              <div class="slash-popup-list py-1">
                <div
                  v-for="(tpl, idx) in filteredSlashTemplates"
                  :key="tpl.id || tpl.shortcut || idx"
                  class="slash-popup-item px-3 py-2 d-flex align-start gap-2 cursor-pointer"
                  :class="{ 'slash-item-active': selectedSlashIndex === idx }"
                  @mouseenter="selectedSlashIndex = idx"
                  @click="applySlashTemplate(tpl)"
                >
                  <v-chip size="x-small" color="primary" variant="flat" class="font-weight-bold flex-shrink-0 mt-0.5">
                    /{{ tpl.shortcut }}
                  </v-chip>
                  <div class="flex-grow-1 min-w-0">
                    <div class="text-caption font-weight-bold text-high-emphasis">{{ tpl.title || tpl.label }}</div>
                    <div class="text-caption text-medium-emphasis text-truncate" style="max-width: 500px;">
                      {{ tpl.content || tpl.text }}
                    </div>
                  </div>
                </div>
              </div>
            </div>

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
                placeholder="Ketik balasan CS... (Ketik / untuk template, Enter untuk kirim)"
                hide-details
                class="chat-input-textarea flex-grow-1"
                @keydown="handleChatInputKeyDown"
                @input="onChatInput"
                @focus="showEmojiPicker = false"
              ></v-textarea>

              <!-- Send Button -->
              <v-btn
                color="primary"
                icon="mdi-send"
                elevation="1"
                size="default"
                class="flex-shrink-0 mb-1 chat-send-btn"
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
        <!-- Panel Header with Tab Switcher -->
        <div class="right-panel-header px-4 border-b d-flex align-center justify-space-between flex-shrink-0 bg-surface">
          <div class="right-panel-tabs d-flex align-center gap-4">
            <button
              type="button"
              class="panel-tab-btn"
              :class="{ 'panel-tab-btn--active': rightPanelTab === 'profile' }"
              @click="rightPanelTab = 'profile'"
            >
              <v-icon size="15" class="me-1">mdi-account-details-outline</v-icon>
              <span>Profil 360</span>
            </button>
            <button
              type="button"
              class="panel-tab-btn"
              :class="{ 'panel-tab-btn--active': rightPanelTab === 'ticket' }"
              @click="rightPanelTab = 'ticket'"
            >
              <v-icon size="15" class="me-1">mdi-ticket-alert-outline</v-icon>
              <span>Trouble Ticket</span>
              <span
                v-if="customerTickets.length > 0"
                class="tab-count-badge ms-1"
              >
                {{ customerTickets.length }}
              </span>
            </button>
          </div>

          <v-btn
            icon="mdi-close"
            variant="text"
            size="small"
            density="compact"
            color="medium-emphasis"
            @click="showInfoPanel = false"
          ></v-btn>
        </div>

        <!-- TAB 1: CUSTOMER PROFILE 360 -->
        <div v-if="rightPanelTab === 'profile'" class="pa-4 d-flex flex-column gap-3.5">
          <!-- Profile Hero -->
          <div class="text-center pb-1">
            <v-avatar :color="getBrandColor(activeRoom.brand)" size="52" class="elevation-1 mb-2">
              <span class="text-subtitle-1 font-weight-bold text-white">
                {{ getInitials(activeRoom.pelanggan?.nama || 'Pelanggan') }}
              </span>
            </v-avatar>
            <h3 class="text-subtitle-2 font-weight-bold text-high-emphasis mb-0">
              {{ activeRoom.pelanggan?.nama || '-' }}
            </h3>
            <div class="text-caption text-medium-emphasis mt-0.5" style="font-size: 0.72rem;">
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
            <div class="text-overline text-medium-emphasis mb-1.5 font-weight-bold" style="font-size: 0.68rem; letter-spacing: 0.8px;">KONTAK & LOKASI</div>
            <div class="info-group-box pa-3 rounded-lg border d-flex flex-column gap-2 bg-surface">
              <!-- Phone -->
              <div class="d-flex align-center justify-space-between">
                <div class="d-flex align-center gap-2 text-body-2" style="font-size: 0.82rem;">
                  <v-icon size="15" color="primary">mdi-phone-outline</v-icon>
                  <span class="font-weight-medium">{{ activeRoom.pelanggan?.no_telp || '-' }}</span>
                </div>
                <v-btn
                  v-if="activeRoom.pelanggan?.no_telp"
                  icon="mdi-content-copy"
                  size="x-small"
                  variant="text"
                  density="compact"
                  color="medium-emphasis"
                  title="Salin Nomor"
                  @click="copyToClipboard(activeRoom.pelanggan.no_telp)"
                ></v-btn>
              </div>

              <v-divider></v-divider>

              <!-- Alamat -->
              <div>
                <div class="d-flex align-center gap-1.5 text-caption text-medium-emphasis mb-1" style="font-size: 0.72rem;">
                  <v-icon size="13">mdi-map-marker-outline</v-icon>
                  <span>Alamat Pemasangan:</span>
                </div>
                <div class="text-caption text-high-emphasis font-weight-medium ps-4" style="font-size: 0.75rem; line-height: 1.4;">
                  {{ activeRoom.pelanggan?.alamat || '-' }}
                  <span v-if="activeRoom.pelanggan?.blok || activeRoom.pelanggan?.unit" class="text-primary font-weight-bold d-block mt-0.5">
                    Blok {{ activeRoom.pelanggan?.blok || '-' }} / Unit {{ activeRoom.pelanggan?.unit || '-' }}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- Section: Layanan FTTH -->
          <div>
            <div class="text-overline text-medium-emphasis mb-1.5 font-weight-bold" style="font-size: 0.68rem; letter-spacing: 0.8px;">LAYANAN FTTH</div>
            <div class="info-group-box pa-3 rounded-lg border d-flex flex-column gap-2 bg-surface">
              <div class="d-flex align-center justify-space-between">
                <div class="d-flex align-center gap-1.5 text-caption text-medium-emphasis" style="font-size: 0.72rem;">
                  <v-icon size="15" color="primary">mdi-speedometer</v-icon>
                  <span>Paket:</span>
                </div>
                <span class="text-caption font-weight-bold text-primary" style="font-size: 0.76rem;">
                  {{ getCustomerPackage(activeRoom.pelanggan) }}
                </span>
              </div>

              <v-divider></v-divider>

              <div class="d-flex align-center justify-space-between">
                <div class="d-flex align-center gap-1.5 text-caption text-medium-emphasis" style="font-size: 0.72rem;">
                  <v-icon size="15" color="primary">mdi-shield-check-outline</v-icon>
                  <span>Status:</span>
                </div>
                <v-chip
                  :color="getCustomerStatus(activeRoom.pelanggan) === 'Aktif' ? 'success' : 'error'"
                  size="x-small"
                  variant="flat"
                  class="font-weight-bold"
                  style="height: 20px; font-size: 0.68rem;"
                >
                  {{ getCustomerStatus(activeRoom.pelanggan) }}
                </v-chip>
              </div>

              <template v-if="activeRoom.pelanggan?.data_teknis?.id_pelanggan">
                <v-divider></v-divider>
                <div class="d-flex align-center justify-space-between">
                  <div class="d-flex align-center gap-1.5 text-caption text-medium-emphasis" style="font-size: 0.72rem;">
                    <v-icon size="15" color="primary">mdi-account-key-outline</v-icon>
                    <span>PPPoE:</span>
                  </div>
                  <code class="text-caption font-weight-bold px-1.5 py-0.5 rounded border bg-panel" style="font-size: 0.72rem;">
                    {{ activeRoom.pelanggan.data_teknis.id_pelanggan }}
                  </code>
                </div>
              </template>
            </div>
          </div>

          <!-- Action Buttons -->
          <div class="d-flex flex-column gap-2 pt-1">
            <v-btn
              block
              color="success"
              prepend-icon="mdi-whatsapp"
              variant="flat"
              class="text-none font-weight-bold rounded-lg"
              style="height: 36px; font-size: 0.8rem;"
              @click="openWhatsApp(activeRoom.pelanggan?.no_telp)"
            >
              Chat via WhatsApp Web
            </v-btn>
            <v-btn
              block
              variant="tonal"
              prepend-icon="mdi-account-search-outline"
              class="text-none font-weight-medium rounded-lg"
              style="height: 36px; font-size: 0.8rem;"
              @click="navigateToCustomer(activeRoom.pelanggan_id)"
            >
              Buka Data Pelanggan
            </v-btn>
          </div>
        </div>

        <!-- TAB 2: TROUBLE TICKET & AUTO-FILL -->
        <div v-else class="pa-4 d-flex flex-column gap-3.5">
          <!-- Asisten Pintar Auto-Fill Banner -->
          <div class="smart-ticket-card pa-3 rounded-lg border">
            <div class="d-flex align-center gap-2 mb-2">
              <v-avatar color="primary" size="26" variant="tonal" class="rounded-circle">
                <v-icon size="14" color="primary">mdi-creation</v-icon>
              </v-avatar>
              <div>
                <div class="text-caption font-weight-bold text-high-emphasis" style="font-size: 0.78rem;">Asisten Tiket Cerdas</div>
                <div class="text-caption text-medium-emphasis" style="font-size: 0.68rem;">Deteksi otomatis keluhan pelanggan dari percakapan</div>
              </div>
            </div>
            <v-btn
              block
              color="primary"
              variant="flat"
              size="small"
              class="text-none font-weight-bold rounded-lg"
              style="height: 30px; font-size: 0.74rem;"
              prepend-icon="mdi-lightning-bolt"
              :loading="isAutoFilling"
              @click="autoFillTicketFromChat"
            >
              ⚡ Auto-Fill dari Chat
            </v-btn>
          </div>

          <!-- Trouble Ticket Form -->
          <div class="d-flex flex-column gap-2.5">
            <div class="text-overline text-medium-emphasis font-weight-bold" style="font-size: 0.68rem; letter-spacing: 0.8px;">FORMULIR TROUBLE TICKET</div>

            <!-- Kategori Kendala -->
            <div>
              <label class="text-caption font-weight-bold text-medium-emphasis mb-1 d-block" style="font-size: 0.72rem;">
                Kategori Gangguan *
              </label>
              <v-select
                v-model="ticketForm.category"
                :items="ticketCategories"
                item-title="title"
                item-value="value"
                density="compact"
                variant="outlined"
                rounded="lg"
                hide-details
              ></v-select>
            </div>

            <!-- Judul Tiket -->
            <div>
              <label class="text-caption font-weight-bold text-medium-emphasis mb-1 d-block" style="font-size: 0.72rem;">
                Judul Tiket *
              </label>
              <v-text-field
                v-model="ticketForm.title"
                placeholder="Contoh: Kabel Fiber Optic Putus / WiFi Lemot"
                density="compact"
                variant="outlined"
                rounded="lg"
                hide-details
              ></v-text-field>
            </div>

            <!-- Tingkat Prioritas -->
            <div>
              <label class="text-caption font-weight-bold text-medium-emphasis mb-1 d-block" style="font-size: 0.72rem;">
                Tingkat Prioritas *
              </label>
              <v-select
                v-model="ticketForm.priority"
                :items="ticketPriorities"
                item-title="title"
                item-value="value"
                density="compact"
                variant="outlined"
                rounded="lg"
                hide-details
              ></v-select>
            </div>

            <!-- Rincian Keluhan -->
            <div>
              <label class="text-caption font-weight-bold text-medium-emphasis mb-1 d-block" style="font-size: 0.72rem;">
                Rincian / Deskripsi Keluhan *
              </label>
              <v-textarea
                v-model="ticketForm.description"
                placeholder="Tuliskan keluhan atau laporan gangguan pelanggan..."
                rows="3"
                density="compact"
                variant="outlined"
                rounded="lg"
                auto-grow
                hide-details
              ></v-textarea>
            </div>

            <!-- Ringkasan Info Pelanggan -->
            <div class="info-group-box pa-2.5 rounded-lg border bg-surface">
              <div class="d-flex align-center justify-space-between mb-1">
                <span class="text-caption text-medium-emphasis font-weight-medium">Pelanggan:</span>
                <span class="text-caption font-weight-bold text-high-emphasis">{{ activeRoom.pelanggan?.nama }}</span>
              </div>
              <div class="d-flex align-center justify-space-between mb-1">
                <span class="text-caption text-medium-emphasis font-weight-medium">Paket:</span>
                <span class="text-caption font-weight-bold text-primary">{{ getCustomerPackage(activeRoom.pelanggan) }}</span>
              </div>
              <div v-if="activeRoom.pelanggan?.alamat" class="text-caption text-medium-emphasis d-flex align-center">
                <v-icon size="12" class="me-1">mdi-map-marker-outline</v-icon>
                <span>{{ activeRoom.pelanggan.alamat }}</span>
              </div>
            </div>

            <!-- Submit Button -->
            <v-btn
              block
              color="primary"
              variant="flat"
              size="default"
              prepend-icon="mdi-ticket-alert-outline"
              class="text-none font-weight-bold rounded-lg mt-1"
              style="height: 38px; font-size: 0.82rem;"
              :loading="isSubmittingTicket"
              @click="submitTroubleTicket"
            >
              Terbitkan Trouble Ticket
            </v-btn>
          </div>

          <v-divider></v-divider>

          <!-- Riwayat Tiket Pelanggan -->
          <div>
            <div class="d-flex align-center justify-space-between mb-2">
              <span class="text-overline font-weight-bold text-medium-emphasis" style="font-size: 0.68rem; letter-spacing: 0.8px;">RIWAYAT TIKET PELANGGAN</span>
              <span class="text-caption text-medium-emphasis">({{ customerTickets.length }})</span>
            </div>

            <div v-if="isLoadingCustomerTickets" class="text-center py-4">
              <v-progress-circular indeterminate size="20" color="primary"></v-progress-circular>
            </div>

            <div v-else-if="customerTickets.length === 0" class="text-caption text-medium-emphasis text-center py-3 border rounded-lg bg-surface">
              Belum ada riwayat tiket gangguan.
            </div>

            <div v-else class="d-flex flex-column gap-2">
              <div
                v-for="t in customerTickets"
                :key="t.id"
                class="ticket-history-card pa-2.5 rounded-lg border bg-surface"
              >
                <div class="d-flex align-center justify-space-between mb-1">
                  <span class="text-caption font-weight-bold text-primary" style="font-size: 0.76rem;">{{ t.ticket_number }}</span>
                  <v-chip size="x-small" :color="getTicketStatusColor(t.status)" variant="tonal" class="font-weight-bold px-1.5" style="height: 18px; font-size: 0.65rem;">
                    {{ formatTicketStatus(t.status) }}
                  </v-chip>
                </div>
                <div class="text-caption font-weight-medium text-truncate text-high-emphasis mb-1" style="font-size: 0.78rem;">
                  {{ t.title }}
                </div>
                <div class="d-flex align-center justify-space-between text-caption text-medium-emphasis" style="font-size: 0.68rem;">
                  <span>{{ t.category }}</span>
                  <span>{{ formatDate(t.created_at) }}</span>
                </div>
              </div>
            </div>
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

    <!-- Dialog: Selesaikan Percakapan & Kirim Pesan Penutup -->
    <v-dialog v-model="closeRoomDialog" max-width="520">
      <v-card class="rounded-xl overflow-hidden" elevation="8">
        <v-card-title class="d-flex align-center justify-space-between px-5 py-4 border-b">
          <div class="d-flex align-center gap-2">
            <v-avatar color="success" size="32" variant="tonal">
              <v-icon size="18" color="success">mdi-check-circle-outline</v-icon>
            </v-avatar>
            <span class="text-subtitle-1 font-weight-bold">Selesaikan Percakapan</span>
          </div>
          <v-btn icon="mdi-close" variant="text" size="small" @click="closeRoomDialog = false"></v-btn>
        </v-card-title>
        <v-card-text class="pa-5">
          <p class="text-body-2 text-medium-emphasis mb-3">
            Tandai sesi percakapan dengan <strong class="text-high-emphasis">{{ activeRoom?.pelanggan?.nama || 'Pelanggan' }}</strong> sebagai selesai.
          </p>

          <v-checkbox
            v-model="sendClosingMessage"
            label="Kirim pesan penutup otomatis ke pelanggan"
            color="primary"
            density="compact"
            hide-details
            class="mb-3 font-weight-medium"
          ></v-checkbox>

          <div v-if="sendClosingMessage" class="mt-1">
            <label class="text-caption font-weight-bold text-medium-emphasis mb-1 d-block">
              Pesan Penutup (dapat diedit):
            </label>
            <v-textarea
              v-model="closingMessageText"
              rows="3"
              variant="outlined"
              density="compact"
              rounded="lg"
              auto-grow
              hide-details
              placeholder="Tulis pesan penutup..."
            ></v-textarea>
          </div>
        </v-card-text>
        <v-card-actions class="px-5 pb-5 pt-0 d-flex justify-end gap-2">
          <v-btn variant="text" rounded="pill" :disabled="isUpdatingStatus" @click="closeRoomDialog = false">
            Batal
          </v-btn>
          <v-btn
            color="success"
            variant="flat"
            rounded="pill"
            prepend-icon="mdi-check-circle"
            :loading="isUpdatingStatus"
            @click="confirmCloseRoom"
          >
            Selesaikan & Tutup
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Dialog: Sukses Terbitkan Trouble Ticket -->
    <v-dialog v-model="ticketSuccessDialog" max-width="480">
      <v-card class="rounded-xl overflow-hidden" elevation="8">
        <v-card-title class="d-flex align-center justify-space-between px-5 py-4 border-b bg-surface">
          <div class="d-flex align-center gap-2">
            <v-avatar color="success" size="32" variant="tonal">
              <v-icon size="18" color="success">mdi-check-circle-outline</v-icon>
            </v-avatar>
            <span class="text-subtitle-1 font-weight-bold">Tiket Gangguan Berhasil Dibuat</span>
          </div>
          <v-btn icon="mdi-close" variant="text" size="small" @click="ticketSuccessDialog = false"></v-btn>
        </v-card-title>
        <v-card-text class="pa-5 text-center">
          <div class="text-h6 font-weight-bold text-primary mb-1">
            #{{ createdTicketResult?.ticket_number || '-' }}
          </div>
          <div class="text-body-2 font-weight-medium text-high-emphasis mb-2">
            {{ createdTicketResult?.title || '-' }}
          </div>
          <p class="text-caption text-medium-emphasis mb-4">
            Tiket gangguan telah masuk ke sistem dan tim teknisi dapat segera memonitor serta menindaklanjuti.
          </p>

          <v-alert
            type="info"
            variant="tonal"
            density="compact"
            class="text-start text-caption mb-0 rounded-lg"
          >
            Kirimkan nomor tiket ke ruang obrolan agar pelanggan mengetahui bahwa kendalanya sudah resmi tercatat.
          </v-alert>
        </v-card-text>
        <v-card-actions class="px-5 pb-5 pt-0 d-flex justify-end gap-2">
          <v-btn variant="text" rounded="pill" @click="ticketSuccessDialog = false">
            Tutup
          </v-btn>
          <v-btn
            color="primary"
            variant="flat"
            rounded="pill"
            prepend-icon="mdi-send"
            @click="sendTicketNoticeToChat"
          >
            Kirim No. Tiket ke Chat
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Dialog: Kelola Template Live Chat -->
    <v-dialog v-model="manageTemplatesDialog" max-width="750" scrollable>
      <v-card rounded="xl" class="border">
        <v-card-title class="d-flex align-center justify-space-between pa-5 border-b bg-surface">
          <div class="d-flex align-center gap-2">
            <v-avatar color="primary" variant="tonal" size="36">
              <v-icon size="20" color="primary">mdi-lightning-bolt</v-icon>
            </v-avatar>
            <div>
              <div class="text-subtitle-1 font-weight-bold text-high-emphasis">Konfigurasi Template Live Chat</div>
              <div class="text-caption text-medium-emphasis">Kelola pesan cepat &amp; pintasan balasan CS</div>
            </div>
          </div>
          <div class="d-flex align-center gap-2">
            <v-btn
              color="primary"
              variant="flat"
              size="small"
              prepend-icon="mdi-plus"
              class="font-weight-bold"
              @click="openAddTemplateDialog"
            >
              Tambah Template
            </v-btn>
            <v-btn icon="mdi-close" variant="text" size="small" @click="manageTemplatesDialog = false"></v-btn>
          </div>
        </v-card-title>

        <v-card-text class="pa-0">
          <v-table hover density="comfortable">
            <thead>
              <tr class="bg-slate-50">
                <th style="width: 140px;">Pintasan (/)</th>
                <th style="width: 180px;">Judul</th>
                <th>Isi Pesan</th>
                <th style="width: 100px;" class="text-center">Aksi</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="quickTemplates.length === 0">
                <td colspan="4" class="text-center py-6 text-medium-emphasis">
                  Belum ada template. Klik "Tambah Template" untuk membuatnya.
                </td>
              </tr>
              <tr v-for="tpl in quickTemplates" :key="tpl.id || tpl.shortcut">
                <td>
                  <v-chip size="x-small" color="primary" variant="tonal" class="font-weight-bold font-mono">
                    /{{ tpl.shortcut }}
                  </v-chip>
                </td>
                <td class="font-weight-medium">
                  <div class="d-flex align-center gap-1.5">
                    <v-icon size="16" color="medium-emphasis">{{ tpl.icon || 'mdi-message-text-outline' }}</v-icon>
                    <span>{{ tpl.title || tpl.label }}</span>
                  </div>
                </td>
                <td class="text-body-2 text-medium-emphasis py-2" style="max-width: 280px;">
                  <div style="white-space: pre-wrap; font-size: 0.8125rem; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;">
                    {{ tpl.content || tpl.text }}
                  </div>
                </td>
                <td class="text-center">
                  <div class="d-flex align-center justify-center gap-1">
                    <v-btn
                      icon="mdi-pencil-outline"
                      size="x-small"
                      variant="text"
                      color="primary"
                      title="Edit Template"
                      @click="openEditTemplateDialog(tpl)"
                    ></v-btn>
                    <v-btn
                      icon="mdi-delete-outline"
                      size="x-small"
                      variant="text"
                      color="error"
                      title="Hapus Template"
                      @click="deleteTemplate(tpl)"
                    ></v-btn>
                  </div>
                </td>
              </tr>
            </tbody>
          </v-table>
        </v-card-text>

        <v-divider></v-divider>
        <v-card-actions class="pa-4 bg-surface justify-end">
          <v-btn variant="outlined" size="small" @click="manageTemplatesDialog = false">Tutup</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Dialog: Form Tambah / Edit Template -->
    <v-dialog v-model="formTemplateDialog" max-width="500">
      <v-card rounded="xl" class="border">
        <v-card-title class="pa-5 border-b bg-surface">
          <span class="text-subtitle-1 font-weight-bold">
            {{ editingTemplateId ? 'Edit Template Chat' : 'Tambah Template Chat Baru' }}
          </span>
        </v-card-title>
        <v-card-text class="pa-5">
          <v-form ref="templateFormRef" @submit.prevent="saveTemplate">
            <v-text-field
              v-model="templateForm.shortcut"
              label="Pintasan (Shortcut)"
              placeholder="contoh: salam, cekteknis, promo"
              prefix="/"
              variant="outlined"
              density="compact"
              class="mb-3"
              hint="Ketikkan kata ini setelah tanda slash (/) di kolom chat"
              persistent-hint
              :rules="[v => !!v || 'Shortcut wajib diisi']"
            ></v-text-field>

            <v-text-field
              v-model="templateForm.title"
              label="Judul Template"
              placeholder="contoh: Salam Pembuka"
              variant="outlined"
              density="compact"
              class="mb-3"
              :rules="[v => !!v || 'Judul template wajib diisi']"
            ></v-text-field>

            <v-textarea
              v-model="templateForm.content"
              label="Isi Pesan Balasan"
              placeholder="Tuliskan template balasan CS di sini..."
              variant="outlined"
              density="compact"
              rows="4"
              class="mb-3"
              :rules="[v => !!v || 'Isi pesan tidak boleh kosong']"
            ></v-textarea>

            <v-text-field
              v-model="templateForm.icon"
              label="Ikon MDI (Opsional)"
              placeholder="mdi-message-text-outline"
              variant="outlined"
              density="compact"
              prepend-inner-icon="mdi-emoticon-outline"
              hint="Nama ikon Material Design Icons, misal: mdi-hand-wave-outline"
            ></v-text-field>
          </v-form>
        </v-card-text>
        <v-divider></v-divider>
        <v-card-actions class="pa-4 bg-surface justify-end gap-2">
          <v-btn variant="outlined" size="small" @click="formTemplateDialog = false">Batal</v-btn>
          <v-btn
            color="primary"
            variant="flat"
            size="small"
            :loading="isSavingTemplate"
            @click="saveTemplate"
          >
            Simpan Template
          </v-btn>
        </v-card-actions>
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
import apiClient, { troubleTicketAPI } from '@/services/api';
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
const isUpdatingStatus = ref(false);

// Omnichannel Status Tabs: 'all' | 'open' | 'resolved'
type StatusFilter = 'all' | 'open' | 'resolved';
// Default to 'open' (Terbuka) so resolved conversations do NOT crowd the view on initial load
const selectedStatusTab = ref<StatusFilter>('open');

// Assignment Filter: 'all' | 'mine' | 'unassigned'
type AssignmentFilter = 'all' | 'mine' | 'unassigned' | 'assigned';
const selectedAssignment = ref<AssignmentFilter>('all');

// Chat Input Mode: 'reply' | 'note'
const inputMode = ref<'reply' | 'note'>('reply');

const roomCounts = ref({
  all: 0,
  mine: 0,
  unassigned: 0,
  assigned: 0,
  closed: 0,
});
const isAssigning = ref(false);

// Panel 3 (Profil 360 & Trouble Ticket) State
const showInfoPanel = ref(false);
const rightPanelTab = ref<'profile' | 'ticket'>('profile');
const isAutoFilling = ref(false);
const isSubmittingTicket = ref(false);
const ticketSuccessDialog = ref(false);
const createdTicketResult = ref<any | null>(null);
const customerTickets = ref<any[]>([]);
const isLoadingCustomerTickets = ref(false);

const ticketForm = ref({
  category: 'slow_connection',
  priority: 'medium',
  title: '',
  description: '',
});

const ticketCategories = [
  { title: 'Kabel Fiber Optic Bermasalah / Putus', value: 'cable_issue' },
  { title: 'Internet Mati Total / Tidak Ada Sinyal (LOS)', value: 'no_connection' },
  { title: 'Koneksi Lambat / Lemot', value: 'slow_connection' },
  { title: 'Koneksi Tidak Stabil / Putus-Nyambung', value: 'intermittent' },
  { title: 'Modem / Router / ONU Bermasalah', value: 'onu_issue' },
  { title: 'Gangguan Server / OLT', value: 'olt_issue' },
  { title: 'Kendala PPPoE / Mikrotik', value: 'mikrotik_issue' },
  { title: 'Kerusakan Hardware / Adaptor', value: 'hardware_issue' },
  { title: 'Lainnya', value: 'other' },
];

const ticketPriorities = [
  { title: 'Rendah (Low)', value: 'low' },
  { title: 'Sedang (Medium)', value: 'medium' },
  { title: 'Tinggi (High)', value: 'high' },
  { title: 'Kritis (Critical)', value: 'critical' },
];

interface StatusTabItem {
  label: string;
  value: StatusFilter;
  count: number;
}

const statusTabs = computed<StatusTabItem[]>(() => [
  {
    label: 'Terbuka',
    value: 'open',
    count: openRoomsCount.value,
  },
  {
    label: 'Selesai',
    value: 'resolved',
    count: closedRoomsCount.value,
  },
  {
    label: 'Semua',
    value: 'all',
    count: rooms.value.length,
  },
]);

interface AssignmentTabItem {
  label: string;
  value: AssignmentFilter;
  icon: string;
  color: string;
  count?: number;
}

// 3 segments fit cleanly without horizontal scrollbars
const assignmentTabs = computed<AssignmentTabItem[]>(() => [
  {
    label: 'Semua Tim',
    value: 'all',
    icon: 'mdi-account-group-outline',
    color: 'default',
  },
  {
    label: 'Saya',
    value: 'mine',
    icon: 'mdi-account-outline',
    color: 'primary',
    count: mineRoomsCount.value,
  },
  {
    label: 'Unassigned',
    value: 'unassigned',
    icon: 'mdi-account-clock-outline',
    color: 'amber-darken-3',
    count: unassignedRoomsCount.value,
  },
]);

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

// Close conversation dialog state
const closeRoomDialog = ref(false);
const sendClosingMessage = ref(true);
const closingMessageText = ref('');

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

// Quick templates (Dynamic loaded from API)
const quickTemplates = ref<any[]>([
  {
    shortcut: 'salam',
    title: 'Salam',
    label: 'Salam',
    icon: 'mdi-hand-wave-outline',
    content: 'Halo, selamat datang di layanan Customer Care Artacom. Ada yang bisa kami bantu?',
    text: 'Halo, selamat datang di layanan Customer Care Artacom. Ada yang bisa kami bantu?'
  },
  {
    shortcut: 'cekteknis',
    title: 'Cek Teknis',
    label: 'Cek Teknis',
    icon: 'mdi-wrench-clock-outline',
    content: 'Baik pak/bu, mohon ditunggu sebentar ya. Sedang kami lakukan pengecekan ke tim teknis lapangan.',
    text: 'Baik pak/bu, mohon ditunggu sebentar ya. Sedang kami lakukan pengecekan ke tim teknis lapangan.'
  },
  {
    shortcut: 'lunas',
    title: 'Lunas & Aktif',
    label: 'Lunas & Aktif',
    icon: 'mdi-check-decagram-outline',
    content: 'Terima kasih atas konfirmasinya. Tagihan Anda telah terverifikasi dan layanan internet sudah aktif normal kembali.',
    text: 'Terima kasih atas konfirmasinya. Tagihan Anda telah terverifikasi dan layanan internet sudah aktif normal kembali.'
  },
  {
    shortcut: 'fotomodem',
    title: 'Foto Modem',
    label: 'Foto Modem',
    icon: 'mdi-camera-outline',
    content: 'Bisa tolong difotokan lampu indikator (PON / LOS / Internet) yang menyala pada perangkat modem router Anda?',
    text: 'Bisa tolong difotokan lampu indikator (PON / LOS / Internet) yang menyala pada perangkat modem router Anda?'
  },
  {
    shortcut: 'restart',
    title: 'Restart Modem',
    label: 'Restart Modem',
    icon: 'mdi-restart',
    content: 'Bisa dicoba untuk mematikan modem router selama 1-2 menit, lalu hidupkan kembali dan periksa koneksinya?',
    text: 'Bisa dicoba untuk mematikan modem router selama 1-2 menit, lalu hidupkan kembali dan periksa koneksinya?'
  },
  {
    shortcut: 'tutup',
    title: 'Tutup & Terima Kasih',
    label: 'Tutup & Terima Kasih',
    icon: 'mdi-hand-heart-outline',
    content: 'Terima kasih telah menghubungi Customer Care Artacom. Jika tidak ada hal lain yang ditanyakan, percakapan ini akan kami tutup. Selamat beraktivitas!',
    text: 'Terima kasih telah menghubungi Customer Care Artacom. Jika tidak ada hal lain yang ditanyakan, percakapan ini akan kami tutup. Selamat beraktivitas!'
  }
]);

// Slash command autocomplete state
const showSlashMenu = ref(false);
const slashQuery = ref('');
const selectedSlashIndex = ref(0);

const filteredSlashTemplates = computed(() => {
  const q = slashQuery.value.trim().toLowerCase();
  if (!q) return quickTemplates.value;
  return quickTemplates.value.filter((t: any) => {
    const s = (t.shortcut || '').toLowerCase();
    const title = (t.title || t.label || '').toLowerCase();
    const c = (t.content || t.text || '').toLowerCase();
    return s.includes(q) || title.includes(q) || c.includes(q);
  });
});

// Manage templates state
const manageTemplatesDialog = ref(false);
const formTemplateDialog = ref(false);
const editingTemplateId = ref<number | null>(null);
const isSavingTemplate = ref(false);
const templateFormRef = ref<any>(null);
const templateForm = ref({
  shortcut: '',
  title: '',
  content: '',
  icon: 'mdi-message-text-outline',
  sort_order: 0,
});

// Computed unread total
const totalUnreadCount = computed(() => {
  return rooms.value.reduce((acc, r) => acc + (r.unread_count_admin || 0), 0);
});

// Computed open, closed, mine, unassigned, assigned counts
const openRoomsCount = computed(() => {
  return rooms.value.filter((r) => !r.status || r.status === 'open').length;
});

const closedRoomsCount = computed(() => {
  return rooms.value.filter((r) => r.status === 'closed').length;
});

const mineRoomsCount = computed(() => {
  const currentUserId = authStore.user?.id;
  return rooms.value.filter((r) => r.assigned_admin_id === currentUserId).length;
});

const unassignedRoomsCount = computed(() => {
  return rooms.value.filter((r) => !r.assigned_admin_id || r.assigned_admin_id === 0).length;
});

const assignedRoomsCount = computed(() => {
  return rooms.value.filter((r) => !!r.assigned_admin_id && r.assigned_admin_id > 0).length;
});

// Filtered rooms
const filteredRooms = computed(() => {
  let list = rooms.value;

  // 1. Filter Status (All, Open, Resolved)
  if (selectedStatusTab.value === 'open') {
    list = list.filter((r) => !r.status || r.status === 'open');
  } else if (selectedStatusTab.value === 'resolved') {
    list = list.filter((r) => r.status === 'closed');
  }

  // 2. Filter Assignment
  const currentUserId = authStore.user?.id;
  if (selectedAssignment.value === 'mine') {
    list = list.filter((r) => r.assigned_admin_id === currentUserId);
  } else if (selectedAssignment.value === 'unassigned') {
    list = list.filter((r) => !r.assigned_admin_id || r.assigned_admin_id === 0);
  } else if (selectedAssignment.value === 'assigned') {
    list = list.filter((r) => !!r.assigned_admin_id && r.assigned_admin_id > 0);
  }

  // 3. Filter Brand
  if (selectedBrand.value !== 'ALL') {
    list = list.filter((r) => normalizeBrandName(r.brand) === selectedBrand.value);
  }

  // 4. Search Filter
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase().trim();
    list = list.filter((r) => {
      const name = (r.pelanggan?.nama || '').toLowerCase();
      const phone = (r.pelanggan?.no_telp || '').toLowerCase();
      const brand = (r.brand || '').toLowerCase();
      const lastMsg = (r.last_message_text || '').toLowerCase();
      const adminName = (r.assigned_admin?.nama || '').toLowerCase();
      return name.includes(q) || phone.includes(q) || brand.includes(q) || lastMsg.includes(q) || adminName.includes(q);
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

function getBrandShortLabel(brandStr?: string): string {
  if (!brandStr || brandStr === 'ALL') return 'Semua Brand';
  const norm = normalizeBrandName(brandStr);
  if (norm === 'JELANTIK NAGRAK') return 'Nagrak';
  if (norm === 'JELANTIK') return 'Jelantik';
  return 'Jakinet';
}

function getBrandBgClass(brandStr?: string): string {
  const norm = normalizeBrandName(brandStr);
  if (norm === 'JELANTIK NAGRAK') return 'bg-success';
  if (norm === 'JELANTIK') return 'bg-primary';
  return 'bg-error';
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
  fetchRoomCounts();
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
  fetchRoomCounts();
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

  // Load previous trouble tickets for this customer
  if (room.pelanggan_id) {
    loadCustomerTickets(room.pelanggan_id);
  }

  // Reset ticket form with default values
  ticketForm.value = {
    category: 'slow_connection',
    priority: 'medium',
    title: '',
    description: '',
  };

  try {
    const res = await apiClient.get(`/chat/messages/${room.id}?limit=200`);
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
  let text = inputMessage.value.trim();
  if (!text || !activeRoom.value) return;

  if (inputMode.value === 'note') {
    text = `📝 [Catatan Internal] ${text}`;
  }

  // Auto reopen room if it was closed
  if (activeRoom.value.status === 'closed') {
    activeRoom.value.status = 'open';
    const rIdx = rooms.value.findIndex((r) => r.id === activeRoom.value.id);
    if (rIdx !== -1) {
      rooms.value[rIdx].status = 'open';
    }
  }

  // Auto assign room to me if not yet assigned to an admin
  if (!activeRoom.value.assigned_admin_id && authStore.user?.id) {
    activeRoom.value.assigned_admin_id = authStore.user.id;
    activeRoom.value.assigned_admin = {
      id: authStore.user.id,
      nama: authStore.user.name || authStore.user.username || 'Saya',
    };
    const rIdx = rooms.value.findIndex((r) => r.id === activeRoom.value.id);
    if (rIdx !== -1) {
      rooms.value[rIdx].assigned_admin_id = authStore.user.id;
      rooms.value[rIdx].assigned_admin = activeRoom.value.assigned_admin;
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
function openCloseRoomDialog() {
  if (!activeRoom.value) return;
  const brand = (activeRoom.value.brand || '').toUpperCase();
  let brandName = 'Artacom';
  if (brand.includes('JELANTIK')) {
    brandName = 'Jelantik';
  } else if (brand.includes('JAKINET')) {
    brandName = 'Jakinet';
  }

  closingMessageText.value = `Terima kasih telah menghubungi Customer Care ${brandName}. Senang dapat membantu Anda. Percakapan ini kami tandai selesai. Jika Anda membutuhkan bantuan kembali di kemudian hari, silakan kirimkan pesan kepada kami. Semoga hari Anda menyenangkan! 🙏`;
  sendClosingMessage.value = true;
  closeRoomDialog.value = true;
}

async function confirmCloseRoom() {
  if (!activeRoom.value) return;
  const roomId = activeRoom.value.id;
  const custName = activeRoom.value.pelanggan?.nama || 'Pelanggan';
  isUpdatingStatus.value = true;

  try {
    // 1. Kirim pesan penutup otomatis jika dipilih
    if (sendClosingMessage.value && closingMessageText.value.trim()) {
      const msgText = closingMessageText.value.trim();
      const tempId = `admin_close_${Date.now()}`;

      const localMsg = {
        id: null,
        room_id: roomId,
        sender_type: 'admin',
        sender_name: authStore.user?.name || 'Admin CS',
        message: msgText,
        message_type: 'text',
        status: 'pending',
        created_at: new Date().toISOString(),
        temp_id: tempId,
      };

      activeMessages.value.push(localMsg);
      activeRoom.value.last_message_text = msgText;
      activeRoom.value.last_message_at = new Date().toISOString();

      // Pindahkan room ke paling atas
      const roomIdx = rooms.value.findIndex((r) => r.id === roomId);
      if (roomIdx > 0) {
        const [moved] = rooms.value.splice(roomIdx, 1);
        rooms.value.unshift(moved);
      }

      sendWsEvent('send_message', {
        room_id: roomId,
        message: msgText,
        message_type: 'text',
        temp_id: tempId,
      });

      scrollToBottom();
    }

    // 2. Tutup room via API & WebSocket
    await apiClient.post(`/chat/rooms/${roomId}/close`);
    activeRoom.value.status = 'closed';
    const roomIdx = rooms.value.findIndex((r) => r.id === roomId);
    if (roomIdx !== -1) {
      rooms.value[roomIdx].status = 'closed';
    }
    sendWsEvent('close_room', { room_id: roomId });

    closeRoomDialog.value = false;
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

// Quick templates methods & API integration
async function fetchTemplates() {
  try {
    const res = await apiClient.get('/chat/templates');
    if (res.data?.data && Array.isArray(res.data.data) && res.data.data.length > 0) {
      quickTemplates.value = res.data.data;
    }
  } catch (err: any) {
    console.warn('Gagal memuat template dari backend, menggunakan default:', err);
  }
}

function useTemplate(tpl: any) {
  const content = tpl.content || tpl.text || '';
  inputMessage.value = content;
  nextTick(() => {
    if (chatInputRef.value) {
      chatInputRef.value.focus();
    }
  });
}

function onChatInput() {
  notifyAdminTyping();
  checkSlashCommand();
}

function checkSlashCommand() {
  const text = inputMessage.value;
  const lastSlashIndex = text.lastIndexOf('/');
  if (lastSlashIndex !== -1) {
    if (lastSlashIndex === 0 || text[lastSlashIndex - 1] === ' ' || text[lastSlashIndex - 1] === '\n') {
      const query = text.slice(lastSlashIndex + 1);
      if (!/\s/.test(query)) {
        slashQuery.value = query.toLowerCase();
        showSlashMenu.value = true;
        selectedSlashIndex.value = 0;
        return;
      }
    }
  }
  showSlashMenu.value = false;
}

function handleChatInputKeyDown(e: KeyboardEvent) {
  if (showSlashMenu.value && filteredSlashTemplates.value.length > 0) {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      selectedSlashIndex.value = (selectedSlashIndex.value + 1) % filteredSlashTemplates.value.length;
      return;
    }
    if (e.key === 'ArrowUp') {
      e.preventDefault();
      selectedSlashIndex.value = (selectedSlashIndex.value - 1 + filteredSlashTemplates.value.length) % filteredSlashTemplates.value.length;
      return;
    }
    if (e.key === 'Enter' || e.key === 'Tab') {
      if (!e.shiftKey) {
        e.preventDefault();
        applySlashTemplate(filteredSlashTemplates.value[selectedSlashIndex.value]);
        return;
      }
    }
    if (e.key === 'Escape') {
      e.preventDefault();
      showSlashMenu.value = false;
      return;
    }
  }

  // Normal Enter sends message
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    sendAdminMessage();
  }
}

function applySlashTemplate(tpl: any) {
  if (!tpl) return;
  const content = tpl.content || tpl.text || '';
  const text = inputMessage.value;
  const lastSlashIndex = text.lastIndexOf('/');
  if (lastSlashIndex !== -1) {
    inputMessage.value = text.slice(0, lastSlashIndex) + content;
  } else {
    inputMessage.value = content;
  }
  showSlashMenu.value = false;
  nextTick(() => {
    if (chatInputRef.value) {
      chatInputRef.value.focus();
    }
  });
}

function openManageTemplates() {
  manageTemplatesDialog.value = true;
}

function openAddTemplateDialog() {
  editingTemplateId.value = null;
  templateForm.value = {
    shortcut: '',
    title: '',
    content: '',
    icon: 'mdi-message-text-outline',
    sort_order: quickTemplates.value.length + 1,
  };
  formTemplateDialog.value = true;
}

function openEditTemplateDialog(tpl: any) {
  editingTemplateId.value = tpl.id || null;
  templateForm.value = {
    shortcut: tpl.shortcut || '',
    title: tpl.title || tpl.label || '',
    content: tpl.content || tpl.text || '',
    icon: tpl.icon || 'mdi-message-text-outline',
    sort_order: tpl.sort_order || 0,
  };
  formTemplateDialog.value = true;
}

async function saveTemplate() {
  if (!templateForm.value.shortcut.trim() || !templateForm.value.content.trim()) {
    showSnackbar('Shortcut dan Isi Pesan wajib diisi', 'error');
    return;
  }

  isSavingTemplate.value = true;
  try {
    const payload = {
      shortcut: templateForm.value.shortcut.replace(/^\//, '').trim().toLowerCase(),
      title: templateForm.value.title.trim() || templateForm.value.shortcut.trim(),
      content: templateForm.value.content.trim(),
      icon: templateForm.value.icon?.trim() || 'mdi-message-text-outline',
      sort_order: templateForm.value.sort_order || 0,
    };

    if (editingTemplateId.value) {
      await apiClient.put(`/chat/templates/${editingTemplateId.value}`, payload);
      showSnackbar('Template berhasil diperbarui', 'success');
    } else {
      await apiClient.post('/chat/templates', payload);
      showSnackbar('Template baru berhasil ditambahkan', 'success');
    }

    formTemplateDialog.value = false;
    await fetchTemplates();
  } catch (err: any) {
    showSnackbar('Gagal menyimpan template: ' + (err.response?.data?.error || err.message), 'error');
  } finally {
    isSavingTemplate.value = false;
  }
}

async function deleteTemplate(tpl: any) {
  if (!confirm(`Hapus template "/${tpl.shortcut}"?`)) return;

  try {
    if (tpl.id) {
      await apiClient.delete(`/chat/templates/${tpl.id}`);
    }
    quickTemplates.value = quickTemplates.value.filter((t: any) => t.id !== tpl.id && t.shortcut !== tpl.shortcut);
    showSnackbar('Template berhasil dihapus', 'success');
  } catch (err: any) {
    showSnackbar('Gagal menghapus template: ' + (err.response?.data?.error || err.message), 'error');
  }
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

    case 'human_handover_requested': {
      if (activeRoom.value?.id === data.room_id) {
        showSnackbar(`Perhatian: Pelanggan ${data.pelanggan_name || ''} meminta bantuan Customer Support manusia!`, 'warning');
      }
      playNotificationSound();
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
      fetchRoomCounts();
      break;
    }

    case 'room_assigned': {
      const roomIdx = rooms.value.findIndex((r) => r.id === data.room_id);
      if (roomIdx !== -1) {
        rooms.value[roomIdx].assigned_admin_id = data.admin_id;
        if (data.room?.assigned_admin) {
          rooms.value[roomIdx].assigned_admin = data.room.assigned_admin;
        }
      }
      if (activeRoom.value && activeRoom.value.id === data.room_id) {
        activeRoom.value.assigned_admin_id = data.admin_id;
        if (data.room?.assigned_admin) {
          activeRoom.value.assigned_admin = data.room.assigned_admin;
        }
      }
      fetchRoomCounts();
      break;
    }

    case 'room_unassigned': {
      const roomIdx = rooms.value.findIndex((r) => r.id === data.room_id);
      if (roomIdx !== -1) {
        rooms.value[roomIdx].assigned_admin_id = null;
        rooms.value[roomIdx].assigned_admin = null;
      }
      if (activeRoom.value && activeRoom.value.id === data.room_id) {
        activeRoom.value.assigned_admin_id = null;
        activeRoom.value.assigned_admin = null;
      }
      fetchRoomCounts();
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

// ================= MEKARI QONTAK INBOX ASSIGNMENT & TROUBLE TICKET METHODS =================
async function fetchRoomCounts() {
  try {
    const adminId = authStore.user?.id || 0;
    const res = await apiClient.get('/chat/rooms/counts', {
      params: {
        admin_id: adminId,
        brand: selectedBrand.value !== 'ALL' ? selectedBrand.value : undefined,
      },
    });
    if (res.data?.data) {
      roomCounts.value = res.data.data;
    }
  } catch (err) {
    console.error('Failed to fetch room counts:', err);
  }
}

async function assignRoomToMe() {
  if (!activeRoom.value) return;
  const adminId = authStore.user?.id;
  if (!adminId) {
    showSnackbar('ID Admin tidak ditemukan', 'error');
    return;
  }
  isAssigning.value = true;
  try {
    const roomId = activeRoom.value.id;
    await apiClient.post(`/chat/rooms/${roomId}/assign`, { admin_id: adminId });
    activeRoom.value.assigned_admin_id = adminId;
    activeRoom.value.assigned_admin = {
      id: adminId,
      nama: authStore.user?.name || 'Admin CS',
      email: authStore.user?.email || '',
    };
    const roomIdx = rooms.value.findIndex((r) => r.id === roomId);
    if (roomIdx !== -1) {
      rooms.value[roomIdx].assigned_admin_id = adminId;
      rooms.value[roomIdx].assigned_admin = activeRoom.value.assigned_admin;
    }
    showSnackbar('Percakapan berhasil diambil alih.', 'success');
    fetchRoomCounts();
  } catch (err: any) {
    showSnackbar('Gagal mengambil alih percakapan: ' + (err.response?.data?.error || err.message), 'error');
  } finally {
    isAssigning.value = false;
  }
}

async function unassignActiveRoom() {
  if (!activeRoom.value) return;
  isAssigning.value = true;
  try {
    const roomId = activeRoom.value.id;
    await apiClient.post(`/chat/rooms/${roomId}/unassign`);
    activeRoom.value.assigned_admin_id = null;
    activeRoom.value.assigned_admin = null;
    const roomIdx = rooms.value.findIndex((r) => r.id === roomId);
    if (roomIdx !== -1) {
      rooms.value[roomIdx].assigned_admin_id = null;
      rooms.value[roomIdx].assigned_admin = null;
    }
    showSnackbar('Penugasan percakapan telah dilepas.', 'info');
    fetchRoomCounts();
  } catch (err: any) {
    showSnackbar('Gagal melepas penugasan: ' + (err.response?.data?.error || err.message), 'error');
  } finally {
    isAssigning.value = false;
  }
}

function openTroubleTicketPanel() {
  showInfoPanel.value = true;
  rightPanelTab.value = 'ticket';
}

async function autoFillTicketFromChat() {
  if (!activeRoom.value) return;
  isAutoFilling.value = true;
  try {
    const res = await apiClient.post(`/chat/rooms/${activeRoom.value.id}/auto-fill-ticket`);
    if (res.data?.data) {
      const d = res.data.data;
      ticketForm.value.category = d.category || 'cable_issue';
      ticketForm.value.priority = d.priority || 'medium';
      ticketForm.value.title = d.title || 'Kendala Layanan Pelanggan';
      ticketForm.value.description = d.description || '';
      showSnackbar('Formulir berhasil diisi otomatis dari riwayat chat!', 'success');
    }
  } catch (err: any) {
    showSnackbar('Gagal menganalisis chat: ' + (err.response?.data?.error || err.message), 'error');
  } finally {
    isAutoFilling.value = false;
  }
}

async function submitTroubleTicket() {
  if (!activeRoom.value) return;
  if (!ticketForm.value.title.trim()) {
    showSnackbar('Judul tiket wajib diisi', 'error');
    return;
  }
  if (!ticketForm.value.description.trim()) {
    showSnackbar('Rincian deskripsi keluhan wajib diisi', 'error');
    return;
  }

  isSubmittingTicket.value = true;
  try {
    const payload = {
      pelanggan_id: Number(activeRoom.value.pelanggan_id),
      title: ticketForm.value.title.trim(),
      description: ticketForm.value.description.trim(),
      category: ticketForm.value.category,
      priority: ticketForm.value.priority,
      data_teknis_id: activeRoom.value.pelanggan?.data_teknis?.id
        ? Number(activeRoom.value.pelanggan.data_teknis.id)
        : null,
    };

    const res = await troubleTicketAPI.createTicket(payload);
    const createdTicket = res.data?.data;
    createdTicketResult.value = createdTicket;
    ticketSuccessDialog.value = true;

    // Reset form
    ticketForm.value = {
      category: 'slow_connection',
      priority: 'medium',
      title: '',
      description: '',
    };

    // Reload customer tickets
    loadCustomerTickets(activeRoom.value.pelanggan_id);
    showSnackbar(`Trouble Ticket #${createdTicket?.ticket_number || ''} berhasil diterbitkan!`, 'success');
  } catch (err: any) {
    showSnackbar('Gagal membuat Trouble Ticket: ' + (err.response?.data?.error || err.message), 'error');
  } finally {
    isSubmittingTicket.value = false;
  }
}

async function sendTicketNoticeToChat() {
  if (!activeRoom.value || !createdTicketResult.value) return;
  const tNo = createdTicketResult.value.ticket_number;
  const custName = activeRoom.value.pelanggan?.nama || 'Bapak/Ibu';
  const notice = `Halo ${custName}, keluhan Anda telah kami catat dengan nomor Trouble Ticket #${tNo}. Tim teknisi kami segera memproses dan melakukan pengecekan. Mohon ditunggu ya. Terima kasih! 🙏`;

  sendAdminTextMessage(notice);
  ticketSuccessDialog.value = false;
}

function sendAdminTextMessage(text: string) {
  if (!text || !activeRoom.value) return;

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
    status: 'pending',
    created_at: new Date().toISOString(),
    temp_id: tempId,
  };

  activeMessages.value.push(localMsg);
  activeRoom.value.last_message_text = text;
  activeRoom.value.last_message_at = new Date().toISOString();

  const currentRoomId = activeRoom.value.id;
  const roomIdx = rooms.value.findIndex((r) => r.id === currentRoomId);
  if (roomIdx > 0) {
    const [moved] = rooms.value.splice(roomIdx, 1);
    rooms.value.unshift(moved);
  }

  scrollToBottom();

  sendWsEvent('send_message', {
    room_id: activeRoom.value.id,
    message: text,
    temp_id: tempId,
    message_type: 'text',
  });
}

async function loadCustomerTickets(pelangganId: number | string) {
  if (!pelangganId) return;
  isLoadingCustomerTickets.value = true;
  try {
    const res = await troubleTicketAPI.getTickets({
      pelanggan_id: pelangganId,
      pageSize: 5,
    });
    if (res.data?.data) {
      customerTickets.value = Array.isArray(res.data.data) ? res.data.data : [];
    }
  } catch (err) {
    console.error('Failed to load customer tickets:', err);
  } finally {
    isLoadingCustomerTickets.value = false;
  }
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return '-';
  const d = new Date(dateStr);
  return d.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' });
}

function formatTicketStatus(status?: string): string {
  const map: Record<string, string> = {
    open: 'Open',
    in_progress: 'Diproses',
    pending_customer: 'Pending Pelanggan',
    pending_vendor: 'Pending Vendor',
    resolved: 'Selesai',
    closed: 'Ditutup',
    cancelled: 'Dibatalkan',
  };
  return map[status || ''] || status || 'Open';
}

function getTicketStatusColor(status?: string): string {
  const map: Record<string, string> = {
    open: 'error',
    in_progress: 'warning',
    pending_customer: 'info',
    pending_vendor: 'secondary',
    resolved: 'success',
    closed: 'grey',
    cancelled: 'grey-darken-2',
  };
  return map[status || ''] || 'primary';
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
    if (closeRoomDialog.value) {
      closeRoomDialog.value = false;
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
  fetchTemplates();
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
/* =====================================================================
   DESIGN TOKENS — Minimalist SaaS Enterprise Communication System
   Palette: Clean slate neutrals, precise borders, solid signal-teal accent (#0e7c8c),
   and functional status colors. No heavy gradients, no bloated cards.
   ===================================================================== */
.customer-chat-wrapper {
  --ink: #0f172a;
  --ink-soft: #475569;
  --muted: #94a3b8;
  --surface: #ffffff;
  --panel: #f8fafc;
  --panel-alt: #f1f5f9;
  --border: #e2e8f0;
  --border-soft: #edf2f7;
  --accent: #0e7c8c;
  --accent-deep: #0a5866;
  --accent-soft: #e2f3f4;
  --accent-hover: #0c6977;
  --amber: #d97706;
  --amber-soft: #fef3c7;
  --radius-sm: 6px;
  --radius-md: 8px;
  --radius-lg: 12px;
  --shadow-sm: 0 1px 2px rgba(15, 23, 42, 0.05);
  --shadow-md: 0 4px 12px rgba(15, 23, 42, 0.06);
  --shadow-ring: 0 0 0 2px rgba(14, 124, 140, 0.16);

  height: calc(100vh - 150px);
  max-height: calc(100vh - 150px);
  box-sizing: border-box;
  color: var(--ink);
}

.chat-workspace {
  display: flex;
  height: 100%;
  width: 100%;
  overflow: hidden;
  position: relative;
  border-color: var(--border) !important;
  box-shadow: var(--shadow-md) !important;
}

/* Gap Utilities */
.gap-1 { gap: 4px; }
.gap-1\.5 { gap: 6px; }
.gap-2 { gap: 8px; }
.gap-2\.5 { gap: 10px; }
.gap-3 { gap: 12px; }
.gap-3\.5 { gap: 14px; }
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
  background-color: var(--surface);
  border-color: var(--border) !important;
}

.sidebar-header {
  border-color: var(--border) !important;
}

.search-input :deep(.v-field) {
  border-radius: var(--radius-md);
  background-color: var(--panel);
  box-shadow: none;
}

.search-input :deep(.v-field__outline) {
  color: var(--border);
  opacity: 1;
}

.search-input :deep(.v-field--focused .v-field__outline) {
  color: var(--accent);
}

/* Omnichannel Status Tabs (Terbuka | Selesai | Semua) */
.omnichannel-status-tabs {
  background-color: var(--panel);
  border-radius: var(--radius-md);
  padding: 3px;
  display: flex;
  gap: 3px;
  border: 1px solid var(--border-soft);
}

.omnichannel-tab-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  padding: 5px 6px;
  border-radius: var(--radius-sm);
  font-size: 0.74rem;
  font-weight: 600;
  color: var(--ink-soft);
  background: transparent;
  border: none;
  cursor: pointer;
  transition: all 0.15s ease;
}

.omnichannel-tab-btn:hover:not(.omnichannel-tab-btn--active) {
  background-color: var(--surface);
  color: var(--ink);
}

.omnichannel-tab-btn--active {
  background-color: var(--surface) !important;
  color: var(--accent) !important;
  box-shadow: var(--shadow-sm);
}

.omnichannel-tab-count {
  font-size: 0.65rem;
  padding: 0 5px;
  border-radius: 999px;
  font-weight: 700;
  background-color: var(--border-soft);
  color: var(--ink-soft);
  line-height: 1.3;
}

.omnichannel-tab-btn--active .omnichannel-tab-count {
  background-color: var(--accent-soft);
  color: var(--accent);
}

/* Segmented assignment bar */
.assignment-filter-bar {
  background-color: var(--panel);
  border-radius: var(--radius-md);
  padding: 3px;
  display: flex;
  gap: 3px;
  border: 1px solid var(--border-soft);
  overflow: hidden;
  scrollbar-width: none;
  -ms-overflow-style: none;
}

.assignment-filter-bar::-webkit-scrollbar,
.quick-replies-scroll::-webkit-scrollbar,
.quick-replies-bar::-webkit-scrollbar {
  display: none !important;
}

.quick-replies-scroll {
  scrollbar-width: none;
  -ms-overflow-style: none;
}

.assignment-tab-btn {
  flex: 1 1 0px;
  min-width: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 5px 2px;
  border-radius: var(--radius-sm);
  font-size: 0.72rem;
  font-weight: 600;
  color: var(--ink-soft);
  background: transparent;
  border: none;
  cursor: pointer;
  transition: all 0.15s ease;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.assignment-tab-btn:hover:not(.assignment-tab-btn--active) {
  background-color: var(--surface);
  color: var(--ink);
}

.assignment-tab-btn--active {
  background-color: var(--surface) !important;
  color: var(--accent) !important;
  box-shadow: var(--shadow-sm);
}

.assignment-tab-count {
  font-size: 0.65rem;
  padding: 0 5px;
  border-radius: 999px;
  font-weight: 700;
  background-color: var(--border-soft);
  color: var(--ink-soft);
  line-height: 1.3;
}

.assignment-tab-btn--active .assignment-tab-count {
  background-color: var(--accent-soft);
  color: var(--accent);
}

/* Room Items */
.room-card {
  padding: 12px 14px;
  transition: background-color 0.12s ease;
  border-left: 3px solid transparent;
  border-bottom: 1px solid var(--border-soft);
  background-color: var(--surface);
}

.room-card:hover {
  background-color: var(--panel);
}

.room-card--active {
  background-color: #eaf6f7 !important;
  border-left-color: var(--accent) !important;
}

.v-theme--dark .room-card--active {
  background-color: #123239 !important;
  border-left-color: var(--accent) !important;
}

.room-title {
  color: var(--ink);
}

.room-time {
  color: var(--muted);
}

.room-snippet {
  color: var(--ink-soft);
}

.room-unread-badge {
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  background-color: #ef4444;
  color: #ffffff;
  border-radius: 999px;
  font-size: 0.65rem;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
}

.room-brand-chip {
  font-size: 0.65rem;
  font-weight: 600;
  padding: 1.5px 6px;
  border-radius: 4px;
  background-color: #475569;
  color: #ffffff;
  letter-spacing: 0.15px;
  line-height: 1.25;
}

.v-theme--dark .room-brand-chip {
  background-color: #334155;
  color: #f1f5f9;
}

.room-subtle-tag {
  font-size: 0.7rem;
  font-weight: 500;
  color: var(--muted);
  white-space: nowrap;
}

.room-assign-chip {
  font-size: 0.64rem;
  font-weight: 600;
  padding: 1.5px 6px;
  border-radius: 4px;
  line-height: 1.25;
  white-space: nowrap;
}

.room-assign-chip--me {
  background-color: var(--accent-soft);
  color: var(--accent);
}

.room-assign-chip--unassigned {
  background-color: var(--amber-soft);
  color: var(--amber);
}

.room-assign-chip--other {
  background-color: #f3e8ff;
  color: #7e22ce;
}

.v-theme--dark .room-assign-chip--other {
  background-color: #3b0764;
  color: #d8b4fe;
}

.sidebar-footer {
  border-color: var(--border-soft) !important;
  background-color: var(--panel) !important;
  height: 38px;
}

/* ================= PANEL 2: CONVERSATION ================= */
.chat-conversation-panel {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: var(--panel);
}

.chat-room-header {
  border-color: var(--border) !important;
  box-shadow: var(--shadow-sm);
  position: relative;
  z-index: 2;
}

.chat-messages-stream {
  padding: 18px 22px;
  background-color: var(--panel);
  display: flex;
  flex-direction: column;
  gap: 4px;
}

/* Clean Line-Through Date Divider */
.date-divider {
  display: flex;
  align-items: center;
  text-align: center;
  width: 100%;
}

.date-divider::before,
.date-divider::after {
  content: '';
  flex: 1;
  border-bottom: 1px solid var(--border);
}

.date-divider-text {
  padding: 0 12px;
  font-size: 0.72rem;
  font-weight: 500;
  color: var(--muted);
  background: transparent;
}

.message-row {
  margin-bottom: 10px;
  width: 100%;
}

.message-bubble {
  max-width: 66%;
  min-width: 130px;
  position: relative;
  word-break: break-word;
  overflow-wrap: anywhere;
  padding: 9px 14px;
  box-shadow: var(--shadow-sm);
}

/* Customer message: crisp white card with subtle 1px border */
.bubble-customer {
  background-color: var(--surface);
  color: var(--ink);
  border-radius: 4px 14px 14px 14px !important;
  border: 1px solid var(--border);
  margin-left: 4px;
}

/* System / AI message: subtle teal tint */
.bubble-system {
  background-color: var(--accent-soft);
  color: var(--ink);
  border-radius: 4px 14px 14px 14px !important;
  border: 1px dashed rgba(14, 124, 140, 0.35);
  margin-left: 4px;
}

/* Admin / CS message: solid flat signal-teal (no heavy gradient) */
.bubble-admin {
  background-color: var(--accent) !important;
  color: #ffffff;
  border-radius: 14px 4px 14px 14px !important;
  margin-right: 4px;
}

/* ================= QUICK REPLIES BAR ================= */
.quick-replies-bar {
  flex-shrink: 0 !important;
  min-height: 38px !important;
  border-color: var(--border) !important;
}

.quick-chip {
  transition: border-color 0.12s ease, background-color 0.12s ease;
  border-radius: 999px;
}

.quick-chip:hover {
  border-color: var(--accent) !important;
}

/* ================= SLASH POPUP MENU ================= */
.slash-popup-menu {
  position: absolute;
  bottom: 100%;
  left: 16px;
  right: 16px;
  margin-bottom: 8px;
  max-height: 280px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  box-shadow: 0 10px 25px -5px rgba(15, 23, 42, 0.12), 0 4px 8px -2px rgba(15, 23, 42, 0.06);
  z-index: 100;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.slash-popup-header {
  border-color: var(--border) !important;
  background-color: var(--panel) !important;
}

.slash-popup-list {
  overflow-y: auto;
  max-height: 230px;
}

.slash-popup-item {
  transition: background-color 0.1s ease;
  border-radius: var(--radius-sm);
  margin: 0 4px;
}

.slash-popup-item:hover,
.slash-item-active {
  background-color: var(--accent-soft) !important;
}

/* ================= INPUT BAR ================= */
.chat-input-bar {
  border-color: var(--border) !important;
}

.chat-input-textarea :deep(.v-field) {
  border-radius: var(--radius-md);
  background-color: var(--panel);
}

.chat-input-textarea :deep(.v-field__outline) {
  color: var(--border);
  opacity: 1;
}

.chat-input-textarea :deep(.v-field--focused) {
  box-shadow: var(--shadow-ring);
}

.chat-input-textarea :deep(.v-field--focused .v-field__outline) {
  color: var(--accent);
}

/* Solid Flat Accent Send Button */
.chat-send-btn {
  background-color: var(--accent) !important;
  color: #ffffff !important;
  transition: opacity 0.12s ease, transform 0.12s ease;
}

.chat-send-btn:hover:not(.v-btn--disabled) {
  background-color: var(--accent-hover) !important;
  transform: translateY(-1px);
}

/* ================= EMOJI PICKER ================= */
.emoji-picker-popup {
  position: absolute;
  bottom: 48px;
  left: 0;
  z-index: 100;
  box-shadow: 0 14px 34px rgba(15, 23, 42, 0.18);
  border-radius: var(--radius-md);
  overflow: hidden;
  border: 1px solid var(--border);
}

/* ================= PANEL 3: CONTACT 360 & TROUBLE TICKET ================= */
.customer-info-panel {
  width: 320px;
  min-width: 300px;
  max-width: 340px;
  flex-shrink: 0;
  height: 100%;
  background-color: var(--surface);
  border-color: var(--border) !important;
}

.right-panel-header {
  height: 44px;
  border-color: var(--border) !important;
}

/* Minimalist Underline Tabs for Right Panel */
.panel-tab-btn {
  display: inline-flex;
  align-items: center;
  padding: 10px 2px;
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--ink-soft);
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  cursor: pointer;
  transition: all 0.15s ease;
}

.panel-tab-btn:hover {
  color: var(--ink);
}

.panel-tab-btn--active {
  color: var(--accent) !important;
  border-bottom-color: var(--accent) !important;
}

.tab-count-badge {
  font-size: 0.65rem;
  font-weight: 700;
  background-color: #ef4444;
  color: #ffffff;
  padding: 1px 6px;
  border-radius: 999px;
  line-height: 1.2;
}

/* Reusable clean border card container for info rows */
.info-group-box {
  border-color: var(--border) !important;
  background-color: var(--surface);
}

/* Smart Auto-fill Ticket Assistant Card */
.smart-ticket-card {
  background-color: rgba(14, 124, 140, 0.05);
  border-color: rgba(14, 124, 140, 0.22) !important;
}

/* Ticket History Card */
.ticket-history-card {
  border-color: var(--border) !important;
  background-color: var(--surface);
  transition: background-color 0.12s ease;
}

.ticket-history-card:hover {
  background-color: var(--panel);
}

/* Slim scrollbars */
.rooms-list-scroll::-webkit-scrollbar,
.messages-container::-webkit-scrollbar,
.customer-info-panel::-webkit-scrollbar {
  width: 5px;
}

.rooms-list-scroll::-webkit-scrollbar-thumb,
.messages-container::-webkit-scrollbar-thumb,
.customer-info-panel::-webkit-scrollbar-thumb {
  background-color: rgba(14, 124, 140, 0.2);
  border-radius: 4px;
}

.rooms-list-scroll::-webkit-scrollbar-thumb:hover,
.messages-container::-webkit-scrollbar-thumb:hover,
.customer-info-panel::-webkit-scrollbar-thumb:hover {
  background-color: rgba(14, 124, 140, 0.38);
}

.animate-pulse {
  animation: pulse 1.8s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

/* ================= DARK THEME OVERRIDES ================= */
.v-theme--dark .customer-chat-wrapper {
  --ink: #f8fafc;
  --ink-soft: #cbd5e1;
  --muted: #64748b;
  --surface: #0f172a;
  --panel: #0b1120;
  --panel-alt: #1e293b;
  --border: #1e293b;
  --border-soft: #172033;
}

.v-theme--dark .chat-conversation-panel {
  background-color: #0b1120;
}

.v-theme--dark .bubble-customer {
  background-color: #1e293b;
  color: #f8fafc;
  border-color: #334155;
}

.v-theme--dark .bubble-system {
  background-color: #123239;
  color: #e6f6f7;
  border-color: #1c5e68;
}

.v-theme--dark .bubble-admin {
  background-color: var(--accent) !important;
  color: #ffffff;
}

.v-theme--dark .date-divider::before,
.v-theme--dark .date-divider::after {
  border-color: #334155;
}

.v-theme--dark .info-group-box,
.v-theme--dark .ticket-history-card {
  background-color: #0f172a;
  border-color: #1e293b !important;
}

.v-theme--dark .rooms-list-scroll::-webkit-scrollbar-thumb,
.v-theme--dark .messages-container::-webkit-scrollbar-thumb,
.v-theme--dark .customer-info-panel::-webkit-scrollbar-thumb {
  background-color: rgba(255, 255, 255, 0.15);
}
</style>