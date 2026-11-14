<template>
  <v-container fluid class="pa-4 pa-md-6">
    <!-- ===== HEADER SECTION (Artacom Pure Light Standard) ===== -->
    <div class="header-card mb-4 mb-md-6">
      <div class="d-flex flex-column flex-md-row align-start align-md-center justify-space-between gap-4">
        <div class="d-flex align-center">
          <div class="header-avatar-box">
            <v-icon color="white" size="22">mdi-router-wireless</v-icon>
          </div>
          <div class="ml-4">
            <div class="d-flex align-center gap-2">
              <h1 class="header-title mb-0">OLT Management</h1>
              <v-chip size="x-small" color="primary" variant="tonal" class="font-weight-bold ml-1">
                FTTH Core
              </v-chip>
            </div>
            <p class="header-subtitle mb-0 mt-0.5">
              Monitoring gateway OLT, manajemen port PON, dan diagnostik telemetri ONU real-time
            </p>
          </div>
        </div>
        <div class="d-flex align-center gap-2">
          <v-btn
            color="primary"
            variant="flat"
            elevation="0"
            rounded="pill"
            height="38"
            prepend-icon="mdi-plus"
            class="text-none font-weight-bold px-5"
            @click="openDialog()"
          >
            Tambah OLT
          </v-btn>
        </div>
      </div>
    </div>

    <!-- ===== LOADING STATE ===== -->
    <div v-if="loading" class="d-flex justify-center align-center py-12">
      <v-progress-circular indeterminate color="primary" size="56"></v-progress-circular>
    </div>

    <!-- ===== EMPTY STATE ===== -->
    <div v-else-if="olts.length === 0" class="text-center py-12 bg-white rounded-xl border">
      <v-icon size="64" color="slate-300" class="mb-3">mdi-server-network-off</v-icon>
      <h3 class="text-subtitle-1 font-weight-bold text-slate-800 mb-1">Belum Ada Gateway OLT</h3>
      <p class="text-caption text-slate-500 mb-4">Tambahkan OLT baru untuk mengaktifkan pemantauan PON dan ONU.</p>
      <v-btn
        color="primary"
        variant="flat"
        rounded="pill"
        class="text-none font-weight-bold"
        @click="openDialog()"
      >
        Tambah OLT Baru
      </v-btn>
    </div>

    <!-- ===== SERVER CARDS GRID ===== -->
    <v-row v-else>
      <v-col
        v-for="item in olts"
        :key="item.id"
        cols="12"
        sm="6"
        md="4"
        xl="3"
      >
        <v-card
          class="server-card h-100 d-flex flex-column"
          elevation="0"
        >
          <v-card-text class="pa-4 flex-grow-1">
            <!-- Card Top Bar: Icon + Title + Menu -->
            <div class="d-flex justify-space-between align-start mb-3">
              <div class="d-flex align-center">
                <div class="card-icon-box me-3">
                  <v-icon color="slate-700" size="22">mdi-router-wireless</v-icon>
                </div>
                <div>
                  <h3 class="server-title text-truncate mb-0">
                    {{ item.nama_olt }}
                  </h3>
                  <div class="d-flex align-center mt-1">
                    <v-icon size="14" color="primary" class="me-1">mdi-ip-network</v-icon>
                    <span class="font-mono text-caption text-slate-600 font-weight-medium">
                      {{ item.ip_address }}
                    </span>
                  </div>
                </div>
              </div>

              <!-- Action Menu -->
              <v-menu location="bottom end">
                <template v-slot:activator="{ props }">
                  <v-btn
                    icon
                    variant="text"
                    density="comfortable"
                    color="grey-darken-1"
                    v-bind="props"
                  >
                    <v-icon size="20">mdi-dots-vertical</v-icon>
                  </v-btn>
                </template>
                <v-list density="compact" elevation="2" rounded="lg" class="py-1 border">
                  <v-list-item @click="openDialog(item)" value="edit" class="px-4">
                    <template v-slot:prepend>
                      <v-icon size="small" color="primary" class="me-3">mdi-pencil</v-icon>
                    </template>
                    <v-list-item-title class="font-weight-medium">Edit Konfigurasi</v-list-item-title>
                  </v-list-item>
                  
                  <v-divider class="my-1"></v-divider>
                  
                  <v-list-item @click="openDeleteDialog(item)" value="delete" class="px-4 text-error">
                    <template v-slot:prepend>
                      <v-icon size="small" color="error" class="me-3">mdi-delete</v-icon>
                    </template>
                    <v-list-item-title class="font-weight-medium text-error">Hapus OLT</v-list-item-title>
                  </v-list-item>
                </v-list>
              </v-menu>
            </div>

            <v-divider class="my-3"></v-divider>

            <!-- Details Rows -->
            <div class="d-flex align-center justify-space-between mb-2">
              <span class="text-caption text-slate-500 font-weight-medium">Tipe OLT</span>
              <v-chip
                size="x-small"
                :color="item.tipe_olt?.toLowerCase() === 'hsgq' ? 'teal-darken-1' : 'indigo-darken-1'"
                variant="tonal"
                class="font-weight-bold"
              >
                {{ item.tipe_olt }}
              </v-chip>
            </div>

            <div class="d-flex align-center justify-space-between mb-2">
              <span class="text-caption text-slate-500 font-weight-medium">Username</span>
              <span class="text-caption font-mono font-weight-medium text-slate-700">
                {{ item.username || '-' }}
              </span>
            </div>

            <div class="bg-slate-50 pa-2 rounded-lg border mt-2">
              <div class="d-flex align-center justify-space-between">
                <span class="text-caption text-slate-500 d-flex align-center">
                  <v-icon size="14" class="me-1" color="slate-400">mdi-router</v-icon>
                  Mikrotik:
                </span>
                <span class="text-caption font-weight-bold text-slate-800 text-truncate ms-2">
                  {{ getMikrotikName(item.mikrotik_server_id) }}
                </span>
              </div>
            </div>
          </v-card-text>

          <!-- Card Actions -->
          <div class="pa-3 pt-0 d-flex gap-2">
            <v-btn
              height="36"
              rounded="lg"
              variant="flat"
              class="flex-grow-1 font-weight-bold text-none btn-explorer"
              elevation="0"
              @click="openExplorer(item)"
              prepend-icon="mdi-lan-connect"
            >
              PON Explorer
            </v-btn>
            <v-btn
              height="36"
              rounded="lg"
              :loading="testingConnectionId === item.id"
              @click="testConnection(item)"
              :color="connectionStatus[item.id] ? 'success' : 'default'"
              :variant="connectionStatus[item.id] ? 'flat' : 'outlined'"
              class="font-weight-medium text-none"
              elevation="0"
            >
              <v-icon :icon="connectionStatus[item.id] ? 'mdi-check-circle' : 'mdi-connection'" size="16" class="me-1"></v-icon>
              {{ connectionStatus[item.id] ? 'Terhubung' : 'Test' }}
            </v-btn>
          </div>
        </v-card>
      </v-col>
    </v-row>

    <!-- ===== DIALOG ADD / EDIT OLT ===== -->
    <v-dialog v-model="dialog" max-width="500px" persistent transition="dialog-bottom-transition">
      <v-card rounded="xl" class="overflow-hidden border">
        <!-- Header -->
        <div class="pa-4 px-6 bg-white d-flex align-center justify-space-between border-b">
          <div class="d-flex align-center">
            <div class="dialog-icon-box box-slate me-3">
              <v-icon size="22" color="slate-800">{{ isEditMode ? 'mdi-pencil' : 'mdi-plus-box' }}</v-icon>
            </div>
            <div>
              <h3 class="text-subtitle-1 font-weight-bold text-slate-900 mb-0">{{ formTitle }}</h3>
              <p class="text-caption text-slate-500 mb-0">Konfigurasi parameter OLT & integrasi Mikrotik</p>
            </div>
          </div>
          <v-btn icon size="small" variant="text" color="slate-600" @click="closeDialog">
            <v-icon size="18">mdi-close</v-icon>
          </v-btn>
        </div>
        
        <v-card-text class="pt-4 px-6 pb-6">
          <v-form @submit.prevent="saveOLT" ref="form">
            <v-row dense>
              <v-col cols="12">
                <p class="text-caption text-slate-500 font-weight-bold mb-2 text-uppercase">Informasi Utama</p>
              </v-col>
              <v-col cols="12">
                <v-text-field 
                  v-model="editedItem.nama_olt" 
                  label="Nama OLT" 
                  placeholder="Contoh: OLT-Parama"
                  variant="outlined" 
                  density="comfortable"
                  color="primary"
                  prepend-inner-icon="mdi-format-title"
                  :rules="[rules.required]"
                  class="mb-1"
                ></v-text-field>
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field 
                  v-model="editedItem.ip_address" 
                  label="IP Address" 
                  placeholder="192.168.100.9"
                  variant="outlined" 
                  density="comfortable"
                  color="primary"
                  prepend-inner-icon="mdi-ip-network"
                  :rules="[rules.required, rules.ip]"
                ></v-text-field>
              </v-col>
              <v-col cols="12" md="6">
                <v-select 
                  v-model="editedItem.tipe_olt" 
                  :items="oltTypes" 
                  label="Tipe Perangkat" 
                  variant="outlined" 
                  density="comfortable"
                  color="primary"
                  prepend-inner-icon="mdi-server"
                  :rules="[rules.required]"
                ></v-select>
              </v-col>

              <v-col cols="12" class="mt-2">
                <p class="text-caption text-slate-500 font-weight-bold mb-2 text-uppercase">Kredensial & Integrasi</p>
              </v-col>

              <v-col cols="12">
                <v-select
                  v-model="editedItem.mikrotik_server_id"
                  :items="mikrotikList"
                  item-title="name"
                  item-value="id"
                  label="Mikrotik Server Terkait"
                  variant="outlined"
                  density="comfortable"
                  color="primary"
                  prepend-inner-icon="mdi-router"
                  :loading="loadingMikrotiks"
                  :rules="[rules.required]"
                  no-data-text="Tidak ada server Mikrotik"
                ></v-select>
              </v-col>

              <v-col cols="12" md="6">
                <v-text-field 
                  v-model="editedItem.username" 
                  label="Username" 
                  placeholder="root"
                  variant="outlined" 
                  density="comfortable"
                  color="primary"
                  prepend-inner-icon="mdi-account"
                ></v-text-field>
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field 
                  v-model="editedItem.password" 
                  label="Password" 
                  type="password" 
                  variant="outlined" 
                  density="comfortable"
                  color="primary"
                  prepend-inner-icon="mdi-lock"
                  :placeholder="isEditMode ? '••••••••' : 'Password OLT'"
                  :hint="isEditMode ? 'Kosongkan jika tidak diubah' : ''"
                  persistent-hint
                  :rules="isEditMode ? [] : [rules.required]"
                ></v-text-field>
              </v-col>
            </v-row>

            <div class="d-flex gap-3 mt-4">
              <v-btn 
                variant="outlined" 
                color="slate-600" 
                rounded="lg"
                height="40"
                class="flex-grow-1 text-none font-weight-medium"
                @click="closeDialog"
              >
                Batal
              </v-btn>
              <v-btn 
                color="primary" 
                type="submit" 
                rounded="lg"
                height="40"
                class="flex-grow-1 text-none font-weight-bold"
                :loading="saving"
              >
                {{ isEditMode ? 'Simpan Perubahan' : 'Tambah OLT' }}
              </v-btn>
            </div>
          </v-form>
        </v-card-text>
      </v-card>
    </v-dialog>

    <!-- ===== DIALOG DELETE CONFIRMATION ===== -->
    <v-dialog v-model="dialogDelete" max-width="420px">
      <v-card rounded="xl" class="pa-4 border" elevation="2">
        <div class="d-flex align-center mb-3">
          <div class="dialog-icon-box box-rose me-3">
            <v-icon color="error" size="22">mdi-delete-alert-outline</v-icon>
          </div>
          <div>
            <h3 class="text-subtitle-1 font-weight-bold text-slate-900 mb-0">Hapus OLT</h3>
            <p class="text-caption text-slate-500 mb-0">Tindakan ini tidak dapat dibatalkan</p>
          </div>
        </div>
        <v-card-text class="pa-0 py-2">
          <p class="text-body-2 text-slate-700">
            Apakah Anda yakin ingin menghapus gateway OLT <strong>{{ itemToDelete?.nama_olt }}</strong> ({{ itemToDelete?.ip_address }})?
          </p>
        </v-card-text>
        <v-card-actions class="pa-0 pt-3 gap-2">
          <v-spacer></v-spacer>
          <v-btn variant="outlined" rounded="lg" color="slate-600" class="text-none font-weight-medium px-4" @click="closeDeleteDialog">
            Batal
          </v-btn>
          <v-btn variant="flat" rounded="lg" color="error" class="text-none font-weight-bold px-4" :loading="deleting" @click="confirmDelete">
            Ya, Hapus
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- ===== DIALOG PON EXPLORER & TELEMETRY LIVE ===== -->
    <v-dialog v-model="dialogExplorer" max-width="1100px" scrollable transition="dialog-bottom-transition">
      <v-card rounded="xl" class="overflow-hidden border">
        <!-- Explorer Header -->
        <div class="pa-4 px-5 bg-white d-flex align-center justify-space-between border-b">
          <div class="d-flex align-center">
            <div class="dialog-icon-box box-slate me-3">
              <v-icon size="22" color="slate-800">mdi-lan-connect</v-icon>
            </div>
            <div>
              <div class="d-flex align-center gap-2">
                <h2 class="text-subtitle-1 font-weight-bold text-slate-900 mb-0">{{ activeOlt?.nama_olt }}</h2>
                <v-chip size="x-small" :color="isHSGQ ? 'teal-darken-1' : 'indigo-darken-1'" variant="tonal" class="font-weight-bold">
                  {{ activeOlt?.tipe_olt }}
                </v-chip>
              </div>
              <p class="text-caption text-slate-500 mb-0 font-mono">
                IP: {{ activeOlt?.ip_address }} &bull; {{ isHSGQ ? 'HSGQ EPON Telnet Gateway' : 'ZTE C320 Gateway (Direct Telnet / SNMP)' }}
              </p>
            </div>
          </div>
          <v-btn icon variant="text" density="comfortable" color="slate-600" @click="dialogExplorer = false">
            <v-icon size="20">mdi-close</v-icon>
          </v-btn>
        </div>

        <!-- Navigation Tabs -->
        <v-tabs v-model="explorerTab" color="primary" density="comfortable" class="border-b bg-slate-50 px-3">
          <v-tab value="pon" prepend-icon="mdi-access-point-network" class="text-none font-weight-bold">
            PON & Telemetri ONU
          </v-tab>
          <v-tab value="uplinks" prepend-icon="mdi-server-network" class="text-none font-weight-bold">
            Kartu & Uplink Port
          </v-tab>
        </v-tabs>

        <v-card-text class="pa-4 pa-md-5 bg-slate-50">
          <!-- TAB 1: PON & ONUs -->
          <div v-if="explorerTab === 'pon'">
            <!-- Filter Bar -->
            <div class="filter-bar pa-3 rounded-lg border bg-white mb-4 d-flex align-center flex-wrap gap-3">
              <div v-if="!isHSGQ" style="min-width: 140px;" class="flex-grow-1 flex-sm-grow-0">
                <v-select
                  v-model="selectedBoard"
                  :items="availableBoards"
                  label="Board / Slot"
                  variant="outlined"
                  density="compact"
                  hide-details
                  prefix="Board "
                  @update:model-value="loadPonONUs"
                ></v-select>
              </div>
              <div style="min-width: 140px;" class="flex-grow-1 flex-sm-grow-0">
                <v-select
                  v-model="selectedPon"
                  :items="availablePons"
                  label="PON Port"
                  variant="outlined"
                  density="compact"
                  hide-details
                  prefix="PON "
                  @update:model-value="loadPonONUs"
                ></v-select>
              </div>

              <v-spacer class="d-none d-md-block"></v-spacer>

              <!-- Action Buttons -->
              <div class="d-flex align-center gap-2 flex-wrap">
                <v-btn
                  variant="tonal"
                  color="primary"
                  density="comfortable"
                  rounded="pill"
                  class="text-none font-weight-bold px-4"
                  prepend-icon="mdi-numeric-positive-1"
                  @click="openEmptyIdsDialog"
                  :loading="loadingEmptyIds"
                >
                  Cek Slot Kosong
                </v-btn>
                <v-btn
                  variant="outlined"
                  color="slate-600"
                  density="comfortable"
                  rounded="pill"
                  class="text-none font-weight-medium px-4"
                  prepend-icon="mdi-refresh"
                  @click="refreshPonData"
                  :loading="loadingPonData"
                >
                  Refresh Data
                </v-btn>
              </div>
            </div>

            <!-- Loading Indicator -->
            <v-progress-linear
              v-if="loadingPonData"
              indeterminate
              color="primary"
              class="mb-4 rounded-pill"
            ></v-progress-linear>

            <!-- Summary Status Badges -->
            <div class="d-flex align-center flex-wrap gap-2 mb-3">
              <div class="counter-badge counter-total">
                <span class="label">Total ONU:</span>
                <span class="value">{{ onusList.length }}</span>
              </div>
              <div class="counter-badge counter-online">
                <span class="dot dot-emerald"></span>
                <span class="label">Online:</span>
                <span class="value">{{ countOnline }}</span>
              </div>
              <div class="counter-badge counter-offline">
                <span class="dot dot-rose"></span>
                <span class="label">Offline:</span>
                <span class="value">{{ countOffline }}</span>
              </div>
            </div>

            <!-- ONUs Data Table -->
            <v-card rounded="lg" elevation="0" border class="overflow-hidden bg-white">
              <v-table density="comfortable" hover class="onu-table">
                <thead>
                  <tr>
                    <th class="text-caption font-weight-bold text-slate-600">ONU ID</th>
                    <th class="text-caption font-weight-bold text-slate-600">NAMA / DESKRIPSI</th>
                    <th class="text-caption font-weight-bold text-slate-600">TIPE MODEM</th>
                    <th class="text-caption font-weight-bold text-slate-600">SERIAL NUMBER / MAC</th>
                    <th class="text-caption font-weight-bold text-slate-600 text-center">RX POWER (SIGNAL)</th>
                    <th class="text-caption font-weight-bold text-slate-600 text-center">STATUS</th>
                    <th class="text-caption font-weight-bold text-slate-600 text-center">DETAIL</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="!loadingPonData && onusList.length === 0">
                    <td colspan="7" class="text-center py-8 text-slate-400">
                      <v-icon size="40" color="slate-300" class="mb-2">mdi-router-wireless-off</v-icon>
                      <div class="text-body-2 font-weight-medium">Tidak ada ONU terdaftar pada Board {{ selectedBoard }} PON {{ selectedPon }}</div>
                    </td>
                  </tr>
                  <tr v-for="onu in onusList" :key="onu.onu_id">
                    <td>
                      <span class="onu-id-badge font-mono">
                        #{{ onu.onu_id }}
                      </span>
                    </td>
                    <td>
                      <div class="font-weight-bold text-slate-900 text-body-2">{{ onu.name || '-' }}</div>
                    </td>
                    <td>
                      <v-chip v-if="onu.onu_type" size="x-small" variant="tonal" color="indigo" class="font-weight-bold">
                        {{ onu.onu_type }}
                      </v-chip>
                      <span v-else class="text-caption text-slate-400">-</span>
                    </td>
                    <td>
                      <div class="d-flex align-center font-mono text-caption">
                        <span class="font-weight-medium text-slate-700 me-1">{{ onu.serial_number }}</span>
                        <v-btn
                          icon
                          size="x-small"
                          variant="text"
                          density="compact"
                          color="slate-400"
                          @click="copyText(onu.serial_number)"
                          title="Salin Serial Number / MAC"
                        >
                          <v-icon size="14">mdi-content-copy</v-icon>
                        </v-btn>
                      </div>
                    </td>
                    <td class="text-center">
                      <v-chip
                        v-if="onu.rx_power"
                        size="small"
                        :color="getRxColor(onu.rx_power)"
                        variant="flat"
                        class="font-weight-bold font-mono"
                      >
                        {{ formatDbm(onu.rx_power) }}
                      </v-chip>
                      <v-btn
                        v-else
                        size="x-small"
                        variant="tonal"
                        color="primary"
                        rounded="pill"
                        class="text-none font-weight-medium px-3"
                        prepend-icon="mdi-signal-distance-variant"
                        @click="openOnuDetail(onu)"
                      >
                        Cek Redaman
                      </v-btn>
                    </td>
                    <td class="text-center">
                      <span :class="getStatusBadgeClass(onu.status)">
                        <span class="status-dot"></span>
                        {{ onu.status || 'Offline' }}
                      </span>
                    </td>
                    <td class="text-center">
                      <v-btn
                        size="small"
                        variant="text"
                        color="primary"
                        density="comfortable"
                        prepend-icon="mdi-information-outline"
                        class="text-none font-weight-bold"
                        @click="openOnuDetail(onu)"
                      >
                        Diagnosa
                      </v-btn>
                    </td>
                  </tr>
                </tbody>
              </v-table>
            </v-card>
          </div>

          <!-- TAB 2: UPLINKS & CARDS -->
          <div v-else-if="explorerTab === 'uplinks'">
            <v-progress-linear v-if="loadingUplinks" indeterminate color="primary" class="mb-4 rounded-pill"></v-progress-linear>

            <v-row v-if="uplinksData">
              <!-- Line Cards -->
              <v-col cols="12" md="5">
                <v-card rounded="lg" elevation="0" border class="pa-4 h-100 bg-white">
                  <div class="d-flex align-center mb-3">
                    <v-icon color="slate-700" class="me-2" size="20">mdi-credit-card-chip-outline</v-icon>
                    <h3 class="text-subtitle-2 font-weight-bold text-slate-900 mb-0">Kartu Terpasang (Line Cards)</h3>
                  </div>
                  <v-list density="compact" class="pa-0 bg-transparent">
                    <v-list-item
                      v-for="(card, i) in uplinksData.cards"
                      :key="i"
                      class="px-0 py-2 border-b"
                    >
                      <template v-slot:prepend>
                        <v-chip size="x-small" color="slate-800" variant="flat" class="font-weight-bold me-2">
                          Slot {{ card.slot }}
                        </v-chip>
                      </template>
                      <v-list-item-title class="font-weight-medium text-body-2 text-slate-800">{{ card.type }}</v-list-item-title>
                      <v-list-item-subtitle class="text-caption text-slate-500">Role: {{ card.role }}</v-list-item-subtitle>
                    </v-list-item>
                  </v-list>
                </v-card>
              </v-col>

              <!-- Uplink Ports -->
              <v-col cols="12" md="7">
                <v-card rounded="lg" elevation="0" border class="pa-4 h-100 bg-white">
                  <div class="d-flex align-center mb-3">
                    <v-icon color="slate-700" class="me-2" size="20">mdi-ethernet</v-icon>
                    <h3 class="text-subtitle-2 font-weight-bold text-slate-900 mb-0">Port Uplink Ethernet</h3>
                  </div>
                  <v-table density="compact" class="onu-table">
                    <thead>
                      <tr>
                        <th class="text-caption font-weight-bold text-slate-600">PORT</th>
                        <th class="text-caption font-weight-bold text-slate-600">KIND</th>
                        <th class="text-caption font-weight-bold text-slate-600 text-center">ADMIN</th>
                        <th class="text-caption font-weight-bold text-slate-600 text-center">LINK STATUS</th>
                        <th class="text-caption font-weight-bold text-slate-600 text-right">SPEED</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="(port, i) in uplinksData.ports" :key="i">
                        <td class="font-mono font-weight-bold text-caption text-slate-800">{{ port.name }}</td>
                        <td><v-chip size="x-small" variant="tonal" color="indigo">{{ port.kind }}</v-chip></td>
                        <td class="text-center">
                          <v-chip size="x-small" :color="port.admin_status === 'up' ? 'success' : 'grey'" variant="flat">
                            {{ port.admin_status }}
                          </v-chip>
                        </td>
                        <td class="text-center">
                          <v-chip size="x-small" :color="port.oper_status === 'up' ? 'success' : 'error'" variant="flat" class="font-weight-bold">
                            {{ port.oper_status }}
                          </v-chip>
                        </td>
                        <td class="text-right font-mono text-caption text-slate-700">{{ port.speed_mbps }} Mbps</td>
                      </tr>
                    </tbody>
                  </v-table>
                </v-card>
              </v-col>
            </v-row>
          </div>
        </v-card-text>
      </v-card>
    </v-dialog>

    <!-- ===== DIALOG DETAIL TELEMETRI ONU ===== -->
    <v-dialog v-model="dialogOnuDetail" max-width="560px" transition="dialog-bottom-transition">
      <v-card rounded="xl" class="overflow-hidden border">
        <!-- Header -->
        <div class="pa-4 px-5 bg-white d-flex align-center justify-space-between border-b">
          <div class="d-flex align-center">
            <div class="dialog-icon-box box-slate me-3">
              <v-icon size="22" color="slate-800">mdi-router-wireless</v-icon>
            </div>
            <div>
              <h3 class="text-subtitle-1 font-weight-bold text-slate-900 mb-0">
                {{ selectedOnuDetail?.name || 'Detail Telemetri ONU' }}
              </h3>
              <p class="text-caption text-slate-500 mb-0 font-mono">
                SN/MAC: {{ selectedOnuDetail?.serial_number }}
              </p>
            </div>
          </div>
          <v-btn icon size="small" variant="text" color="slate-600" @click="dialogOnuDetail = false">
            <v-icon size="18">mdi-close</v-icon>
          </v-btn>
        </div>

        <v-card-text class="pa-4 pa-sm-5 bg-slate-50">
          <div v-if="loadingOnuDetail" class="text-center py-8">
            <v-progress-circular indeterminate color="primary" size="44"></v-progress-circular>
            <p class="text-caption text-slate-500 mt-3 mb-0">Mengambil telemetri real-time dari OLT...</p>
          </div>

          <div v-else-if="selectedOnuDetail">
            <!-- Optical Rx Signal Highlight Card -->
            <div class="rx-signal-card pa-4 rounded-xl mb-3 text-center border" :class="getRxBannerClass(selectedOnuDetail.rx_power)">
              <div class="d-flex align-center justify-center gap-2 mb-1">
                <span class="text-caption font-weight-bold text-uppercase letter-spacing-wide">Optical Rx Power</span>
                <v-chip size="x-small" :color="getRxColor(selectedOnuDetail.rx_power)" variant="flat" class="font-weight-bold">
                  {{ getRxQualityLabel(selectedOnuDetail.rx_power) }}
                </v-chip>
              </div>
              <div class="text-h4 font-weight-bold font-mono my-1">
                {{ formatDbm(selectedOnuDetail.rx_power) }}
              </div>
              <div class="text-caption font-weight-medium text-slate-600">
                Tx Power: <strong>{{ formatDbm(selectedOnuDetail.tx_power) }}</strong>
              </div>
            </div>

            <!-- DMI SFP Transceiver Diagnostics (Suhu, Voltase, Arus Bias) -->
            <div v-if="parsedDiagnostics" class="bg-white pa-3 rounded-xl border mb-3">
              <div class="d-flex align-center justify-space-between mb-2 pb-1 border-b">
                <span class="text-caption font-weight-bold text-slate-800 d-flex align-center">
                  <v-icon size="16" color="primary" class="me-1">mdi-chip</v-icon>
                  Diagnostik Modul Optik (DMI)
                </span>
                <span class="text-caption text-slate-500 font-mono text-xs">Real-Time Sensor</span>
              </div>
              <v-row dense>
                <v-col cols="4">
                  <div class="dmi-item text-center pa-2 rounded-lg bg-slate-50 border">
                    <div class="text-caption text-slate-500 mb-0.5">Suhu Modul</div>
                    <div class="text-subtitle-2 font-weight-bold font-mono text-slate-800">
                      {{ parsedDiagnostics.temperature || '-' }}
                    </div>
                  </div>
                </v-col>
                <v-col cols="4">
                  <div class="dmi-item text-center pa-2 rounded-lg bg-slate-50 border">
                    <div class="text-caption text-slate-500 mb-0.5">Tegangan</div>
                    <div class="text-subtitle-2 font-weight-bold font-mono text-slate-800">
                      {{ parsedDiagnostics.voltage || '-' }}
                    </div>
                  </div>
                </v-col>
                <v-col cols="4">
                  <div class="dmi-item text-center pa-2 rounded-lg bg-slate-50 border">
                    <div class="text-caption text-slate-500 mb-0.5">Arus Bias</div>
                    <div class="text-subtitle-2 font-weight-bold font-mono text-slate-800">
                      {{ parsedDiagnostics.biasCurrent || '-' }}
                    </div>
                  </div>
                </v-col>
              </v-row>
            </div>

            <!-- Detail Metrics Grid -->
            <v-row dense class="g-2">
              <v-col cols="6">
                <div class="metric-box pa-3 bg-white rounded-lg border">
                  <div class="text-caption text-slate-500">Jarak Kabel Optik (FO)</div>
                  <div class="text-subtitle-2 font-weight-bold font-mono text-slate-800">
                    {{ selectedOnuDetail.gpon_optical_distance ? selectedOnuDetail.gpon_optical_distance + ' Meter' : '-' }}
                  </div>
                </div>
              </v-col>
              <v-col cols="6">
                <div class="metric-box pa-3 bg-white rounded-lg border">
                  <div class="text-caption text-slate-500">Status Modem</div>
                  <div class="text-subtitle-2 font-weight-bold" :class="selectedOnuDetail.status?.toLowerCase() === 'online' ? 'text-success' : 'text-error'">
                    {{ selectedOnuDetail.status || '-' }}
                  </div>
                </div>
              </v-col>
              <v-col cols="6">
                <div class="metric-box pa-3 bg-white rounded-lg border">
                  <div class="text-caption text-slate-500">Tipe Perangkat</div>
                  <div class="text-subtitle-2 font-weight-bold text-slate-800">
                    {{ selectedOnuDetail.onu_type || 'EPON' }}
                  </div>
                </div>
              </v-col>
              <v-col cols="6">
                <div class="metric-box pa-3 bg-white rounded-lg border">
                  <div class="text-caption text-slate-500">IP Management ONU</div>
                  <div class="text-subtitle-2 font-weight-bold font-mono text-slate-800">
                    {{ selectedOnuDetail.ip_address || '-' }}
                  </div>
                </div>
              </v-col>
              <v-col cols="12">
                <div class="metric-box pa-3 bg-white rounded-lg border">
                  <div class="text-caption text-slate-500">Uptime Terakhir</div>
                  <div class="text-body-2 font-weight-medium text-slate-800 font-mono">
                    {{ selectedOnuDetail.uptime || '-' }}
                  </div>
                </div>
              </v-col>
              <v-col cols="12" v-if="!parsedDiagnostics && selectedOnuDetail.description">
                <div class="metric-box pa-3 bg-white rounded-lg border">
                  <div class="text-caption text-slate-500">Deskripsi / DMI</div>
                  <div class="text-body-2 font-mono text-slate-800">
                    {{ selectedOnuDetail.description }}
                  </div>
                </div>
              </v-col>
              <v-col cols="12">
                <div class="metric-box pa-3 bg-white rounded-lg border">
                  <div class="text-caption text-slate-500">Alasan Offline / Diagnosa</div>
                  <div class="text-body-2 font-weight-bold" :class="selectedOnuDetail.offline_reason?.toLowerCase().includes('power') ? 'text-amber-600' : 'text-slate-800'">
                    {{ selectedOnuDetail.offline_reason || 'Normal / Tidak Ada Error' }}
                  </div>
                </div>
              </v-col>
            </v-row>
          </div>
        </v-card-text>
      </v-card>
    </v-dialog>

    <!-- ===== DIALOG SLOT ONU ID BEBAS (KOSONG) ===== -->
    <v-dialog v-model="dialogEmptyIds" max-width="480px">
      <v-card rounded="xl" class="overflow-hidden border">
        <div class="pa-4 px-5 bg-white d-flex align-center justify-space-between border-b">
          <div class="d-flex align-center">
            <div class="dialog-icon-box box-slate me-3">
              <v-icon size="22" color="slate-800">mdi-numeric-positive-1</v-icon>
            </div>
            <div>
              <h3 class="text-subtitle-1 font-weight-bold text-slate-900 mb-0">Slot ONU ID Bebas (Kosong)</h3>
              <p class="text-caption text-slate-500 mb-0">Board {{ selectedBoard }} &bull; PON {{ selectedPon }}</p>
            </div>
          </div>
          <v-btn icon size="small" variant="text" color="slate-600" @click="dialogEmptyIds = false">
            <v-icon size="18">mdi-close</v-icon>
          </v-btn>
        </div>
        <v-card-text class="pa-4 bg-slate-50">
          <p class="text-body-2 text-slate-600 mb-3">
            Daftar nomor ONU ID yang belum terpakai dan siap dialokasikan untuk modem pelanggan baru:
          </p>
          <div v-if="emptyOnuIds.length > 0" class="d-flex flex-wrap gap-2 max-h-56 overflow-y-auto pa-3 bg-white rounded-lg border">
            <span
              v-for="id in emptyOnuIds"
              :key="id"
              class="empty-id-pill font-mono"
            >
              #{{ id }}
            </span>
          </div>
          <div v-else class="text-center py-6 text-slate-400 bg-white rounded-lg border">
            <v-icon size="36" color="slate-300" class="mb-2">mdi-check-all</v-icon>
            <div class="text-body-2 font-weight-medium">Semua slot ONU ID pada port ini sudah penuh.</div>
          </div>
        </v-card-text>
        <v-card-actions class="pa-3 px-4 bg-white border-t">
          <v-spacer></v-spacer>
          <v-btn color="slate-700" variant="outlined" rounded="lg" class="text-none font-weight-medium px-4" @click="dialogEmptyIds = false">
            Tutup
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- ===== GLOBAL SNACKBAR NOTIFICATION ===== -->
    <v-snackbar
      v-model="snackbar.show"
      :color="snackbar.color"
      :timeout="4000"
      location="top right"
      variant="flat"
      elevation="4"
      rounded="lg"
    >
      <div class="d-flex align-center">
        <v-icon
          :icon="snackbar.color === 'success' ? 'mdi-check-circle' : 'mdi-alert-circle'"
          class="me-3"
          size="22"
        ></v-icon>
        <div class="font-weight-medium text-body-2">{{ snackbar.text }}</div>
      </div>
    </v-snackbar>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import apiClient from '@/services/api';

// --- INTERFACES ---
interface OLT {
  id: number;
  nama_olt: string;
  ip_address: string;
  tipe_olt: string;
  username?: string;
  mikrotik_server_id?: number | null; 
}

interface MikrotikSelectItem {
  id: number;
  name: string;
}

interface ZTEONUInfo {
  board: number;
  pon: number;
  onu_id: number;
  name: string;
  onu_type: string;
  serial_number: string;
  rx_power: string;
  status: string;
}

interface ZTEONUDetail {
  board: number;
  pon: number;
  onu_id: number;
  name: string;
  description: string;
  onu_type: string;
  serial_number: string;
  rx_power: string;
  tx_power: string;
  status: string;
  ip_address: string;
  last_online: string;
  last_offline: string;
  uptime: string;
  last_down_time_duration: string;
  offline_reason: string;
  gpon_optical_distance: string;
}

interface ZTECardInfo {
  slot: number;
  type: string;
  role: string;
}

interface ZTEPortInfo {
  name: string;
  shelf: number;
  slot: number;
  port: number;
  kind: string;
  admin_status: string;
  oper_status: string;
  speed_mbps: number;
}

interface ZTEUplinksData {
  cards: ZTECardInfo[];
  ports: ZTEPortInfo[];
}

// --- STATE MANAGEMENT ---
const olts = ref<OLT[]>([]);
const mikrotikList = ref<MikrotikSelectItem[]>([]);
const loading = ref(true);
const loadingMikrotiks = ref(false);
const saving = ref(false);
const deleting = ref(false);
const testingConnectionId = ref<number | null>(null);
const connectionStatus = ref<Record<number, boolean>>({});

const dialog = ref(false);
const dialogDelete = ref(false);

const editedItem = ref<Partial<OLT> & { password?: string }>({});
const itemToDelete = ref<OLT | null>(null);
const snackbar = ref({ show: false, text: '', color: 'success' });

// --- PON EXPLORER STATE ---
const dialogExplorer = ref(false);
const activeOlt = ref<OLT | null>(null);
const explorerTab = ref<'pon' | 'uplinks'>('pon');
const selectedBoard = ref(1);
const selectedPon = ref(1);
const onusList = ref<ZTEONUInfo[]>([]);
const uplinksData = ref<ZTEUplinksData | null>(null);
const loadingPonData = ref(false);
const loadingUplinks = ref(false);

const emptyOnuIds = ref<number[]>([]);
const dialogEmptyIds = ref(false);
const loadingEmptyIds = ref(false);

const selectedOnuDetail = ref<ZTEONUDetail | null>(null);
const dialogOnuDetail = ref(false);
const loadingOnuDetail = ref(false);

// --- COMPUTED PROPERTIES ---
const isEditMode = computed(() => !!editedItem.value.id);
const formTitle = computed(() => isEditMode.value ? 'Edit Gateway OLT' : 'Tambah Gateway OLT');
const countOnline = computed(() => onusList.value.filter(o => o.status?.toLowerCase() === 'online').length);
const countOffline = computed(() => onusList.value.filter(o => o.status?.toLowerCase() !== 'online').length);

const isHSGQ = computed(() => {
  const type = (activeOlt.value?.tipe_olt || '').toLowerCase();
  const name = (activeOlt.value?.nama_olt || '').toLowerCase();
  return type.includes('hsgq') || name.includes('hsgq');
});

const availableBoards = computed(() => {
  if (isHSGQ.value) {
    return [1];
  }
  return [1, 2];
});

const availablePons = computed(() => {
  if (isHSGQ.value) {
    return [1, 2, 3, 4];
  }
  return Array.from({ length: 16 }, (_, i) => i + 1);
});

// Parse DMI telemetry (Suhu, Voltase, Arus Bias) from description if HSGQ format
const parsedDiagnostics = computed(() => {
  if (!selectedOnuDetail.value?.description) return null;
  const desc = selectedOnuDetail.value.description;
  const tempMatch = desc.match(/Suhu:\s*([^\s|]+)/i);
  const voltMatch = desc.match(/Voltase:\s*([^\s|]+)/i);
  const biasMatch = desc.match(/Arus Bias:\s*([^\s|]+)/i);
  if (tempMatch || voltMatch || biasMatch) {
    return {
      temperature: tempMatch ? tempMatch[1] : null,
      voltage: voltMatch ? voltMatch[1] : null,
      biasCurrent: biasMatch ? biasMatch[1] : null,
    };
  }
  return null;
});

// --- DATA & CONFIGURATION ---
const oltTypes = ['HSGQ', 'ZTE', 'Huawei', 'Fiberhome', 'Lainnya'];

const rules = {
  required: (value: any) => !!value || 'Field ini wajib diisi.',
  ip: (value: string) => /^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$/.test(value) || 'Format IP tidak valid.',
};

// --- LIFECYCLE HOOKS ---
onMounted(() => {
  fetchOLTs();
  fetchMikrotiks();
});

// --- HELPER FUNCTIONS ---
function getMikrotikName(id: number | null | undefined): string {
  if (!id) return '-';
  const server = mikrotikList.value.find(m => m.id === id);
  return server ? server.name : 'Unknown';
}

// Clean and normalize dBm string to avoid duplicate unit issues
function formatDbm(val: string | null | undefined): string {
  if (!val || val === '-' || val === 'N/A') return '-';
  const clean = val.replace(/dBm/gi, '').trim();
  if (!clean || clean === '-') return '-';
  return `${clean} dBm`;
}

function parsePower(val: string | null | undefined): number | null {
  if (!val || val === '-' || val === 'N/A') return null;
  const clean = val.replace(/dBm/gi, '').trim();
  const num = parseFloat(clean);
  return isNaN(num) ? null : num;
}

function getRxColor(powerStr: string | null | undefined): string {
  const p = parsePower(powerStr);
  if (p === null) return 'grey';
  if (p >= -22.5 && p <= -10) return 'success';
  if (p < -22.5 && p >= -26.5) return 'warning';
  if (p < -26.5 || p > -8) return 'error';
  return 'info';
}

function getRxQualityLabel(powerStr: string | null | undefined): string {
  const p = parsePower(powerStr);
  if (p === null) return 'Tidak Terdeteksi';
  if (p >= -22.5 && p <= -10) return 'Optimal';
  if (p < -22.5 && p >= -25.0) return 'Standar Baik';
  if (p < -25.0 && p >= -27.0) return 'Peringatan Redaman';
  if (p < -27.0) return 'Kritis (< -27 dBm)';
  if (p > -8) return 'Sinyal Terlalu Kuat';
  return 'Normal';
}

function getRxBannerClass(powerStr: string | null | undefined): string {
  const p = parsePower(powerStr);
  if (p === null) return 'rx-banner-neutral';
  if (p >= -22.5 && p <= -10) return 'rx-banner-good';
  if (p < -22.5 && p >= -26.5) return 'rx-banner-warning';
  if (p < -26.5 || p > -8) return 'rx-banner-danger';
  return 'rx-banner-neutral';
}

function getStatusBadgeClass(status: string | undefined): string {
  const s = (status || '').toLowerCase();
  if (s === 'online') return 'status-pill status-pill-online';
  if (s === 'offline') return 'status-pill status-pill-offline';
  return 'status-pill status-pill-warning';
}

function copyText(text: string) {
  if (!text) return;
  navigator.clipboard.writeText(text);
  showSnackbar('Serial Number / MAC disalin: ' + text, 'success');
}

// --- PON EXPLORER FUNCTIONS ---
async function openExplorer(olt: OLT) {
  activeOlt.value = olt;
  dialogExplorer.value = true;
  explorerTab.value = 'pon';
  selectedBoard.value = 1;
  selectedPon.value = 1;
  await loadPonONUs();
  loadUplinks();
}

async function loadPonONUs() {
  if (!activeOlt.value) return;
  loadingPonData.value = true;
  try {
    const response = await apiClient.get(`/olt/${activeOlt.value.id}/board/${selectedBoard.value}/pon/${selectedPon.value}/onus`);
    onusList.value = response.data?.data || [];
  } catch (error: any) {
    const errorMsg = error.response?.data?.message || "Gagal memuat data ONU dari OLT";
    showSnackbar(errorMsg, 'error');
    onusList.value = [];
  } finally {
    loadingPonData.value = false;
  }
}

async function loadUplinks() {
  if (!activeOlt.value) return;
  loadingUplinks.value = true;
  try {
    const response = await apiClient.get(`/olt/${activeOlt.value.id}/uplinks`);
    uplinksData.value = response.data?.data || null;
  } catch (error: any) {
    console.error("Failed to load OLT uplinks", error);
    uplinksData.value = null;
  } finally {
    loadingUplinks.value = false;
  }
}

async function refreshPonData() {
  if (!activeOlt.value) return;
  loadingPonData.value = true;
  try {
    await apiClient.delete(`/olt/${activeOlt.value.id}/board/${selectedBoard.value}/pon/${selectedPon.value}/cache`);
    await loadPonONUs();
    showSnackbar('Cache dibersihkan & data PON diperbarui dari OLT', 'success');
  } catch (error: any) {
    showSnackbar(error.response?.data?.message || 'Gagal merefresh data PON', 'error');
  } finally {
    loadingPonData.value = false;
  }
}

async function openEmptyIdsDialog() {
  if (!activeOlt.value) return;
  loadingEmptyIds.value = true;
  try {
    const response = await apiClient.get(`/olt/${activeOlt.value.id}/board/${selectedBoard.value}/pon/${selectedPon.value}/empty-onu-ids`);
    emptyOnuIds.value = response.data?.data || [];
    dialogEmptyIds.value = true;
  } catch (error: any) {
    showSnackbar(error.response?.data?.message || 'Gagal mengambil daftar slot kosong', 'error');
  } finally {
    loadingEmptyIds.value = false;
  }
}

async function openOnuDetail(onu: ZTEONUInfo) {
  if (!activeOlt.value) return;
  selectedOnuDetail.value = null;
  dialogOnuDetail.value = true;
  loadingOnuDetail.value = true;
  try {
    const response = await apiClient.get(`/olt/${activeOlt.value.id}/board/${selectedBoard.value}/pon/${selectedPon.value}/onu/${onu.onu_id}`);
    selectedOnuDetail.value = response.data?.data || null;
    // Reflect immediately in table row if telemetry is present
    if (selectedOnuDetail.value?.rx_power) {
      onu.rx_power = selectedOnuDetail.value.rx_power;
    }
  } catch (error: any) {
    showSnackbar(error.response?.data?.message || 'Gagal mengambil detail telemetri ONU', 'error');
    dialogOnuDetail.value = false;
  } finally {
    loadingOnuDetail.value = false;
  }
}

// --- API CRUD FUNCTIONS ---
async function fetchOLTs() {
  loading.value = true;
  try {
    const response = await apiClient.get('/olt');
    olts.value = Array.isArray(response.data) ? response.data : (response.data.data || []);
  } catch (error) {
    showSnackbar("Gagal memuat data OLT", "error");
  } finally {
    loading.value = false;
  }
}

async function fetchMikrotiks() {
  loadingMikrotiks.value = true;
  try {
    const response = await apiClient.get('/mikrotik_servers');
    mikrotikList.value = Array.isArray(response.data) ? response.data : (response.data.data || []);
  } catch (error) {
    console.error("Failed to load Mikrotik servers", error);
  } finally {
    loadingMikrotiks.value = false;
  }
}

function openDialog(item?: OLT) {
  editedItem.value = item ? { ...item, password: '' } : { tipe_olt: 'HSGQ' };
  dialog.value = true;
}

function closeDialog() {
  dialog.value = false;
  editedItem.value = {};
}

async function saveOLT() {
  saving.value = true;
  const payload = { ...editedItem.value };
  
  if (isEditMode.value && !payload.password) {
    delete payload.password;
  }

  try {
    if (isEditMode.value) {
      await apiClient.patch(`/olt/${payload.id}`, payload);
    } else {
      await apiClient.post('/olt', payload);
    }
    fetchOLTs();
    closeDialog();
    showSnackbar(`OLT berhasil ${isEditMode.value ? 'diperbarui' : 'ditambahkan'}`, 'success');
  } catch (error: any) {
    const errorMsg = error.response?.data?.detail?.[0]?.msg || "Gagal menyimpan data OLT";
    showSnackbar(errorMsg, 'error');
  } finally {
    saving.value = false;
  }
}

function openDeleteDialog(item: OLT) {
  itemToDelete.value = item;
  dialogDelete.value = true;
}

function closeDeleteDialog() {
  dialogDelete.value = false;
  itemToDelete.value = null;
}

async function confirmDelete() {
  if (!itemToDelete.value) return;
  deleting.value = true;
  try {
    await apiClient.delete(`/olt/${itemToDelete.value.id}`);
    fetchOLTs();
    showSnackbar('OLT berhasil dihapus', 'success');
  } catch (error) {
    showSnackbar('Gagal menghapus OLT', 'error');
  } finally {
    deleting.value = false;
    closeDeleteDialog();
  }
}

async function testConnection(item: OLT) {
  testingConnectionId.value = item.id;
  try {
    const response = await apiClient.post(`/olt/${item.id}/test-connection`);
    showSnackbar(response.data.message || 'Koneksi berhasil!', 'success');
    connectionStatus.value[item.id] = true;
    setTimeout(() => {
      connectionStatus.value[item.id] = false;
    }, 4000);
  } catch (error: any) {
    const message = error.response?.data?.message || "Koneksi gagal, terjadi error.";
    showSnackbar(message, 'error');
  } finally {
    testingConnectionId.value = null;
  }
}

function showSnackbar(text: string, color: 'success' | 'error' | 'info') {
  snackbar.value.text = text;
  snackbar.value.color = color;
  snackbar.value.show = true;
}
</script>

<style scoped>
/* ===== PURE LIGHT HEADER ===== */
.header-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  padding: 18px 24px;
}

.header-avatar-box {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 4px 10px rgba(15, 23, 42, 0.15);
}

.header-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: #0f172a;
  letter-spacing: -0.01em;
}

.header-subtitle {
  font-size: 0.8125rem;
  color: #64748b;
}

/* ===== SERVER CARD GRID ===== */
.server-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
  position: relative;
  overflow: hidden;
}

.server-card:hover {
  transform: translateY(-2px);
  border-color: #cbd5e1;
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.06);
}

.card-icon-box {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.dialog-icon-box {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.dialog-icon-box.box-rose {
  background: #fff1f2;
  border-color: #fecdd3;
}

.server-title {
  font-size: 1.05rem;
  font-weight: 700;
  color: #0f172a;
  letter-spacing: -0.01em;
}

.btn-explorer {
  background-color: #0f172a !important;
  color: #ffffff !important;
  transition: background-color 0.2s ease;
}

.btn-explorer:hover {
  background-color: #1e293b !important;
}

/* ===== SUMMARY STATUS BADGES ===== */
.counter-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  border-radius: 8px;
  font-size: 0.8125rem;
  font-weight: 600;
  line-height: 1;
}

.counter-badge .dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
}

.counter-badge .value {
  font-weight: 700;
}

.counter-total {
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  color: #334155;
}

.counter-online {
  background: #ecfdf5;
  border: 1px solid #a7f3d0;
  color: #047857;
}
.counter-online .dot {
  background: #10b981;
}

.counter-offline {
  background: #fff1f2;
  border: 1px solid #fecdd3;
  color: #be123c;
}
.counter-offline .dot {
  background: #f43f5e;
}

/* ===== TABLE & STATUS PILLS ===== */
.onu-table thead tr th {
  background: #f8fafc !important;
  color: #475569 !important;
  border-bottom: 1px solid #e2e8f0 !important;
  font-size: 0.75rem !important;
  letter-spacing: 0.03em;
  text-transform: uppercase;
}

.onu-table tbody tr td {
  border-bottom: 1px solid #f1f5f9 !important;
}

.onu-id-badge {
  display: inline-block;
  padding: 2px 8px;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-weight: 700;
  color: #334155;
  font-size: 0.75rem;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 0.75rem;
  font-weight: 700;
  padding: 3px 10px;
  border-radius: 9999px;
  line-height: 1;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  display: inline-block;
}

.status-pill-online {
  background: #ecfdf5;
  color: #047857;
  border: 1px solid #a7f3d0;
}
.status-pill-online .status-dot {
  background: #10b981;
}

.status-pill-offline {
  background: #fff1f2;
  color: #be123c;
  border: 1px solid #fecdd3;
}
.status-pill-offline .status-dot {
  background: #f43f5e;
}

.status-pill-warning {
  background: #fffbeb;
  color: #b45309;
  border: 1px solid #fde68a;
}
.status-pill-warning .status-dot {
  background: #f59e0b;
}

/* ===== RX POWER HIGHLIGHT BANNER ===== */
.rx-signal-card {
  transition: all 0.2s ease;
}

.rx-banner-neutral {
  background: #f8fafc;
  border-color: #e2e8f0 !important;
  color: #334155;
}

.rx-banner-good {
  background: #ecfdf5;
  border-color: #a7f3d0 !important;
  color: #047857;
}

.rx-banner-warning {
  background: #fffbeb;
  border-color: #fde68a !important;
  color: #b45309;
}

.rx-banner-danger {
  background: #fff1f2;
  border-color: #fecdd3 !important;
  color: #be123c;
}

/* ===== DMI TRANSCEIVER CARDS ===== */
.dmi-item {
  transition: background 0.2s ease;
}

.empty-id-pill {
  display: inline-block;
  padding: 4px 10px;
  background: #f8fafc;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 0.8125rem;
  font-weight: 700;
  color: #1e293b;
}

.font-mono {
  font-family: 'SF Mono', 'Roboto Mono', ui-monospace, monospace;
}

.letter-spacing-wide {
  letter-spacing: 0.05em;
}
</style>
