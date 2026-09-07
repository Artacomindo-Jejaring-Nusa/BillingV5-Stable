<template>
  <v-container fluid class="pa-4 pa-md-6">
    <!-- Header Section with Gradient Background -->
    <div class="header-card mb-4 mb-md-6">
      <div class="header-section">
        <div class="header-content">
          <div class="d-flex align-center flex-wrap gap-4">
            <div class="d-flex align-center flex-grow-1">
              <v-avatar class="me-4 elevation-4 bg-white/10" color="transparent" size="64">
                <v-icon color="white" size="32">mdi-router-wireless</v-icon>
              </v-avatar>
              <div>
                <h1 class="text-h4 font-weight-bold text-white mb-1">OLT Management</h1>
                <p class="header-subtitle mb-0">
                  Manage and monitor your OLT devices
                </p>
              </div>
            </div>
            <v-btn
              color="white"
              variant="elevated"
              size="large"
              elevation="4"
              @click="openDialog()"
              prepend-icon="mdi-plus"
              class="text-none font-weight-bold px-6 rounded-lg add-btn"
            >
              Tambah OLT
            </v-btn>
          </div>
        </div>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="d-flex justify-center align-center py-12">
      <v-progress-circular indeterminate color="primary" size="64"></v-progress-circular>
    </div>

    <!-- Empty State -->
    <div v-else-if="olts.length === 0" class="text-center py-12">
      <v-icon size="64" color="grey-lighten-1" class="mb-4">mdi-server-network-off</v-icon>
      <h3 class="text-h6 text-grey-darken-1 font-weight-regular">Belum ada OLT yang terdaftar</h3>
      <v-btn
        color="primary"
        variant="text"
        class="mt-2 text-none"
        @click="openDialog()"
      >
        Tambah OLT Baru
      </v-btn>
    </div>

    <!-- Server Cards Grid -->
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
          border
        >
          <!-- Status Line Indicator -->
          <div class="status-line"></div>

          <v-card-text class="pt-5 pb-2 flex-grow-1">
            <div class="d-flex justify-space-between align-start mb-4">
              <!-- Icon Container -->
              <div class="icon-box elevation-2">
                <v-icon color="primary" size="28">mdi-router-wireless</v-icon>
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
                    <v-icon>mdi-dots-vertical</v-icon>
                  </v-btn>
                </template>
                <v-list density="compact" elevation="3" rounded="lg" class="py-2">
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

            <!-- Server Info -->
            <div class="mb-4">
              <h3 class="text-h6 font-weight-bold text-grey-darken-3 mb-1 text-truncate">
                {{ item.nama_olt }}
              </h3>
              <div class="d-flex align-center">
                <v-icon size="16" color="primary" class="me-2">mdi-ip-network</v-icon>
                <span class="text-body-2 font-weight-medium text-grey-darken-1 font-mono">
                  {{ item.ip_address }}
                </span>
              </div>
            </div>

            <v-divider class="mb-4 border-opacity-75"></v-divider>

            <!-- Details Grid -->
            <v-row dense class="mb-0 g-3">
              <v-col cols="6">
                <div class="detail-item">
                  <div class="text-caption text-grey mb-1">Tipe OLT</div>
                  <div class="d-flex align-center">
                    <v-chip
                      size="x-small"
                      color="blue-grey"
                      variant="flat"
                      class="font-weight-bold"
                    >
                      {{ item.tipe_olt }}
                    </v-chip>
                  </div>
                </div>
              </v-col>
              
              <v-col cols="6">
                <div class="detail-item text-right">
                  <div class="text-caption text-grey mb-1">Username</div>
                  <div class="font-weight-medium text-body-2 text-grey-darken-2 text-truncate">
                    {{ item.username || '-' }}
                  </div>
                </div>
              </v-col>

              <v-col cols="12" class="mt-2">
                 <div class="detail-item bg-grey-lighten-4 pa-2 rounded-lg">
                    <div class="d-flex align-center justify-space-between">
                       <span class="text-caption text-grey-darken-1">Mikrotik Server:</span>
                       <span class="text-caption font-weight-bold text-primary text-truncate ms-2">
                          {{ getMikrotikName(item.mikrotik_server_id) }}
                       </span>
                    </div>
                 </div>
              </v-col>
            </v-row>
          </v-card-text>

          <v-card-actions class="px-4 pb-4 pt-0 gap-2 flex-column flex-sm-row">
            <v-btn
              height="40"
              rounded="lg"
              color="primary"
              variant="flat"
              class="flex-grow-1 font-weight-bold text-capitalize"
              elevation="1"
              @click="openExplorer(item)"
              prepend-icon="mdi-lan-connect"
            >
              PON Explorer
            </v-btn>
            <v-btn
              height="40"
              rounded="lg"
              :loading="testingConnectionId === item.id"
              @click="testConnection(item)"
              :color="connectionStatus[item.id] ? 'success' : 'grey-darken-1'"
              :variant="connectionStatus[item.id] ? 'flat' : 'tonal'"
              class="font-weight-bold text-capitalize"
              elevation="0"
            >
              <v-icon :icon="connectionStatus[item.id] ? 'mdi-check-circle' : 'mdi-connection'" class="me-1"></v-icon>
              {{ connectionStatus[item.id] ? 'Terhubung' : 'Test' }}
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>

    <!-- Dialog Edit/Add -->
    <v-dialog v-model="dialog" max-width="500px" persistent transition="dialog-bottom-transition">
      <v-card rounded="xl" class="overflow-visible">
        <div class="dialog-header-accent"></div>
        <v-card-title class="text-h5 font-weight-bold pt-6 px-6 pb-2">
          {{ formTitle }}
        </v-card-title>
        
        <v-card-text class="pt-4 px-6 pb-6">
          <v-form @submit.prevent="saveOLT" ref="form">
            <v-row dense>
               <v-col cols="12">
                   <p class="text-caption text-grey-darken-1 font-weight-bold mb-2 text-uppercase">Informasi Utama</p>
               </v-col>
              <v-col cols="12">
                <v-text-field 
                  v-model="editedItem.nama_olt" 
                  label="Nama OLT" 
                  placeholder="Contoh: OLT-Pusat"
                  variant="outlined" 
                  density="comfortable"
                  color="primary"
                  bg-color="grey-lighten-5"
                  prepend-inner-icon="mdi-format-title"
                  :rules="[rules.required]"
                  class="mb-1"
                ></v-text-field>
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field 
                  v-model="editedItem.ip_address" 
                  label="IP Address" 
                  placeholder="192.168.1.1"
                  variant="outlined" 
                  density="comfortable"
                  color="primary"
                  bg-color="grey-lighten-5"
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
                  bg-color="grey-lighten-5"
                  prepend-inner-icon="mdi-server"
                  :rules="[rules.required]"
                ></v-select>
              </v-col>

               <v-col cols="12" class="mt-2">
                   <p class="text-caption text-grey-darken-1 font-weight-bold mb-2 text-uppercase">Kredensial & Integrasi</p>
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
                  bg-color="grey-lighten-5"
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
                  variant="outlined" 
                  density="comfortable"
                  color="primary"
                  bg-color="grey-lighten-5"
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
                  bg-color="grey-lighten-5"
                  prepend-inner-icon="mdi-lock"
                  :placeholder="isEditMode ? '••••••••' : ''"
                  :hint="isEditMode ? 'Biarkan kosong jika tidak ingin mengubah' : ''"
                  persistent-hint
                  :rules="isEditMode ? [] : [rules.required]"
                ></v-text-field>
              </v-col>
            </v-row>

            <div class="d-flex gap-3 mt-6">
               <v-btn 
                  variant="tonal" 
                  color="grey" 
                  size="large" 
                  class="flex-grow-1"
                  @click="closeDialog"
               >
                  Batal
               </v-btn>
               <v-btn 
                  color="primary" 
                  type="submit" 
                  size="large" 
                  elevation="2"
                  class="flex-grow-1"
                  :loading="saving"
               >
                  {{ isEditMode ? 'Simpan Perubahan' : 'Tambah OLT' }}
               </v-btn>
            </div>
          </v-form>
        </v-card-text>
      </v-card>
    </v-dialog>

    <!-- Dialog PON Explorer & Telemetry Live -->
    <v-dialog v-model="dialogExplorer" max-width="1100px" scrollable transition="dialog-bottom-transition">
      <v-card rounded="xl" class="overflow-hidden">
        <!-- Header -->
        <div class="pa-4 pa-md-5 bg-gradient-to-r from-indigo-600 to-purple-600 text-white d-flex align-center justify-space-between" style="background: linear-gradient(135deg, #4F46E5 0%, #7C3AED 100%);">
          <div class="d-flex align-center">
            <v-avatar color="white" variant="tonal" size="48" class="me-3">
              <v-icon size="28" color="white">mdi-lan-connect</v-icon>
            </v-avatar>
            <div>
              <div class="d-flex align-center gap-2">
                <h2 class="text-h6 font-weight-bold text-white mb-0">{{ activeOlt?.nama_olt }}</h2>
                <v-chip size="x-small" color="white" variant="outlined" class="font-weight-bold">
                  {{ activeOlt?.tipe_olt }}
                </v-chip>
              </div>
              <p class="text-caption text-white opacity-80 mb-0 font-mono">
                IP: {{ activeOlt?.ip_address }} | ZTE C320 SNMP Gateway
              </p>
            </div>
          </div>
          <v-btn icon variant="text" color="white" @click="dialogExplorer = false">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </div>

        <!-- Navigation Tabs -->
        <v-tabs v-model="explorerTab" color="primary" density="comfortable" class="border-b">
          <v-tab value="pon" prepend-icon="mdi-access-point-network" class="text-none font-weight-bold">
            PON & Telemetri ONU
          </v-tab>
          <v-tab value="uplinks" prepend-icon="mdi-server-network" class="text-none font-weight-bold">
            Kartu & Uplink Port
          </v-tab>
        </v-tabs>

        <v-card-text class="pa-4 pa-md-6 bg-grey-lighten-5">
          <!-- TAB 1: PON & ONUs -->
          <div v-if="explorerTab === 'pon'">
            <!-- Filter Bar -->
            <v-card class="mb-4 pa-3 rounded-lg" elevation="0" border>
              <div class="d-flex align-center flex-wrap gap-3">
                <div style="min-width: 140px;" class="flex-grow-1 flex-sm-grow-0">
                  <v-select
                    v-model="selectedBoard"
                    :items="[1, 2]"
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
                    :items="Array.from({ length: 16 }, (_, i) => i + 1)"
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
                    class="text-none font-weight-bold"
                    prepend-icon="mdi-numeric-positive-1"
                    @click="openEmptyIdsDialog"
                    :loading="loadingEmptyIds"
                  >
                    Cek Slot Kosong
                  </v-btn>
                  <v-btn
                    variant="tonal"
                    color="grey-darken-2"
                    density="comfortable"
                    class="text-none font-weight-bold"
                    prepend-icon="mdi-refresh"
                    @click="refreshPonData"
                    :loading="loadingPonData"
                  >
                    Refresh
                  </v-btn>
                </div>
              </div>
            </v-card>

            <!-- Loading Indicator -->
            <v-progress-linear
              v-if="loadingPonData"
              indeterminate
              color="primary"
              class="mb-4 rounded-pill"
            ></v-progress-linear>

            <!-- Summary Chips -->
            <div class="d-flex align-center flex-wrap gap-2 mb-3">
              <v-chip size="small" variant="flat" color="blue-lighten-5" class="text-primary font-weight-bold">
                Total ONU: {{ onusList.length }}
              </v-chip>
              <v-chip size="small" variant="flat" color="green-lighten-5" class="text-success font-weight-bold">
                Online: {{ countOnline }}
              </v-chip>
              <v-chip size="small" variant="flat" color="red-lighten-5" class="text-error font-weight-bold">
                Offline: {{ countOffline }}
              </v-chip>
            </div>

            <!-- ONUs Data Table -->
            <v-card rounded="lg" elevation="0" border class="overflow-hidden">
              <v-table density="comfortable" hover>
                <thead>
                  <tr class="bg-grey-lighten-4">
                    <th class="font-weight-bold text-caption">ONU ID</th>
                    <th class="font-weight-bold text-caption">NAMA / DESKRIPSI</th>
                    <th class="font-weight-bold text-caption">TIPE MODEM</th>
                    <th class="font-weight-bold text-caption">SERIAL NUMBER</th>
                    <th class="font-weight-bold text-caption text-center">RX POWER (SIGNAL)</th>
                    <th class="font-weight-bold text-caption text-center">STATUS</th>
                    <th class="font-weight-bold text-caption text-center">DETAIL</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="!loadingPonData && onusList.length === 0">
                    <td colspan="7" class="text-center py-8 text-grey">
                      <v-icon size="40" color="grey-lighten-1" class="mb-2">mdi-router-wireless-off</v-icon>
                      <div class="text-body-2">Tidak ada ONU terdaftar pada Board {{ selectedBoard }} PON {{ selectedPon }}</div>
                    </td>
                  </tr>
                  <tr v-for="onu in onusList" :key="onu.onu_id">
                    <td>
                      <v-chip size="x-small" variant="flat" color="grey-lighten-3" class="font-mono font-weight-bold">
                        #{{ onu.onu_id }}
                      </v-chip>
                    </td>
                    <td>
                      <div class="font-weight-medium text-body-2">{{ onu.name || '-' }}</div>
                    </td>
                    <td>
                      <v-chip v-if="onu.onu_type" size="x-small" variant="tonal" color="indigo" class="font-weight-bold">
                        {{ onu.onu_type }}
                      </v-chip>
                      <span v-else class="text-caption text-grey">-</span>
                    </td>
                    <td>
                      <div class="d-flex align-center font-mono text-caption">
                        <span class="font-weight-bold text-grey-darken-3 me-1">{{ onu.serial_number }}</span>
                        <v-btn
                          icon
                          size="x-small"
                          variant="text"
                          density="compact"
                          color="grey"
                          @click="copyText(onu.serial_number)"
                          title="Salin Serial Number"
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
                        {{ onu.rx_power }} dBm
                      </v-chip>
                      <span v-else class="text-caption text-grey">-</span>
                    </td>
                    <td class="text-center">
                      <v-chip
                        size="x-small"
                        :color="onu.status?.toLowerCase() === 'online' ? 'success' : 'error'"
                        variant="flat"
                        class="font-weight-bold"
                      >
                        <v-icon start size="10">mdi-circle</v-icon>
                        {{ onu.status || 'Offline' }}
                      </v-chip>
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
                <v-card rounded="lg" elevation="0" border class="pa-4 h-100">
                  <div class="d-flex align-center mb-3">
                    <v-icon color="primary" class="me-2">mdi-credit-card-chip-outline</v-icon>
                    <h3 class="text-subtitle-1 font-weight-bold text-grey-darken-3 mb-0">Kartu Terpasang (Line Cards)</h3>
                  </div>
                  <v-list density="compact" class="pa-0">
                    <v-list-item
                      v-for="(card, i) in uplinksData.cards"
                      :key="i"
                      class="px-0 py-2 border-b"
                    >
                      <template v-slot:prepend>
                        <v-chip size="x-small" color="primary" variant="flat" class="font-weight-bold me-2">
                          Slot {{ card.slot }}
                        </v-chip>
                      </template>
                      <v-list-item-title class="font-weight-medium text-body-2">{{ card.type }}</v-list-item-title>
                      <v-list-item-subtitle class="text-caption">Role: {{ card.role }}</v-list-item-subtitle>
                    </v-list-item>
                  </v-list>
                </v-card>
              </v-col>

              <!-- Uplink Ports -->
              <v-col cols="12" md="7">
                <v-card rounded="lg" elevation="0" border class="pa-4 h-100">
                  <div class="d-flex align-center mb-3">
                    <v-icon color="primary" class="me-2">mdi-ethernet</v-icon>
                    <h3 class="text-subtitle-1 font-weight-bold text-grey-darken-3 mb-0">Port Uplink Ethernet</h3>
                  </div>
                  <v-table density="compact">
                    <thead>
                      <tr class="bg-grey-lighten-4">
                        <th class="text-caption font-weight-bold">PORT</th>
                        <th class="text-caption font-weight-bold">KIND</th>
                        <th class="text-caption font-weight-bold text-center">ADMIN</th>
                        <th class="text-caption font-weight-bold text-center">LINK STATUS</th>
                        <th class="text-caption font-weight-bold text-right">SPEED</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="(port, i) in uplinksData.ports" :key="i">
                        <td class="font-mono font-weight-bold text-caption text-primary">{{ port.name }}</td>
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
                        <td class="text-right font-mono text-caption">{{ port.speed_mbps }} Mbps</td>
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

    <!-- Dialog Detail Telemetri ONU -->
    <v-dialog v-model="dialogOnuDetail" max-width="560px" transition="dialog-bottom-transition">
      <v-card rounded="xl" class="overflow-hidden">
        <div class="pa-4 bg-primary text-white d-flex align-center justify-space-between">
          <div class="d-flex align-center">
            <v-avatar color="white" variant="tonal" size="40" class="me-3">
              <v-icon size="24" color="white">mdi-router-wireless</v-icon>
            </v-avatar>
            <div>
              <h3 class="text-subtitle-1 font-weight-bold mb-0 text-white">{{ selectedOnuDetail?.name || 'Detail Telemetri ONU' }}</h3>
              <p class="text-caption text-white opacity-80 mb-0 font-mono">SN: {{ selectedOnuDetail?.serial_number }}</p>
            </div>
          </div>
          <v-btn icon size="small" variant="text" color="white" @click="dialogOnuDetail = false">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </div>

        <v-card-text class="pa-5">
          <div v-if="loadingOnuDetail" class="text-center py-8">
            <v-progress-circular indeterminate color="primary" size="48"></v-progress-circular>
            <p class="text-caption text-grey mt-2">Mengambil data telemetri real-time dari OLT...</p>
          </div>

          <div v-else-if="selectedOnuDetail">
            <!-- Signal Highlight Banner -->
            <div class="pa-4 rounded-lg mb-4 text-center" :class="getRxBgClass(selectedOnuDetail.rx_power)">
              <div class="text-caption text-grey-darken-2 font-weight-bold text-uppercase mb-1">Optical Rx Signal</div>
              <div class="text-h4 font-weight-bold font-mono" :class="getRxTextClass(selectedOnuDetail.rx_power)">
                {{ selectedOnuDetail.rx_power || '-' }} dBm
              </div>
              <div class="text-caption mt-1 font-weight-medium text-grey-darken-1">
                Tx Power: {{ selectedOnuDetail.tx_power ? selectedOnuDetail.tx_power + ' dBm' : '-' }}
              </div>
            </div>

            <!-- Detail Grid -->
            <v-row dense class="g-2">
              <v-col cols="6">
                <div class="pa-3 bg-grey-lighten-4 rounded-lg">
                  <div class="text-caption text-grey">Jarak Kabel Optik (FO)</div>
                  <div class="text-subtitle-2 font-weight-bold font-mono text-grey-darken-3">
                    {{ selectedOnuDetail.gpon_optical_distance ? selectedOnuDetail.gpon_optical_distance + ' Meter' : '-' }}
                  </div>
                </div>
              </v-col>
              <v-col cols="6">
                <div class="pa-3 bg-grey-lighten-4 rounded-lg">
                  <div class="text-caption text-grey">Status Modem</div>
                  <div class="text-subtitle-2 font-weight-bold" :class="selectedOnuDetail.status?.toLowerCase() === 'online' ? 'text-success' : 'text-error'">
                    {{ selectedOnuDetail.status || '-' }}
                  </div>
                </div>
              </v-col>
              <v-col cols="6">
                <div class="pa-3 bg-grey-lighten-4 rounded-lg">
                  <div class="text-caption text-grey">Tipe Perangkat</div>
                  <div class="text-subtitle-2 font-weight-bold text-grey-darken-3">
                    {{ selectedOnuDetail.onu_type || '-' }}
                  </div>
                </div>
              </v-col>
              <v-col cols="6">
                <div class="pa-3 bg-grey-lighten-4 rounded-lg">
                  <div class="text-caption text-grey">IP Management ONU</div>
                  <div class="text-subtitle-2 font-weight-bold font-mono text-grey-darken-3">
                    {{ selectedOnuDetail.ip_address || '-' }}
                  </div>
                </div>
              </v-col>
              <v-col cols="12">
                <div class="pa-3 bg-grey-lighten-4 rounded-lg">
                  <div class="text-caption text-grey">Uptime Terakhir</div>
                  <div class="text-body-2 font-weight-medium text-grey-darken-3 font-mono">
                    {{ selectedOnuDetail.uptime || '-' }}
                  </div>
                </div>
              </v-col>
              <v-col cols="12">
                <div class="pa-3 bg-grey-lighten-4 rounded-lg">
                  <div class="text-caption text-grey">Alasan Offline / Diagnosa</div>
                  <div class="text-body-2 font-weight-bold" :class="selectedOnuDetail.offline_reason?.toLowerCase().includes('power') ? 'text-warning' : 'text-grey-darken-3'">
                    {{ selectedOnuDetail.offline_reason || 'Normal / Tidak Ada Error' }}
                  </div>
                </div>
              </v-col>
            </v-row>
          </div>
        </v-card-text>
      </v-card>
    </v-dialog>

    <!-- Dialog Slot Kosong -->
    <v-dialog v-model="dialogEmptyIds" max-width="450px">
      <v-card rounded="xl" class="pa-4">
        <v-card-title class="text-h6 font-weight-bold d-flex align-center">
          <v-icon color="primary" class="me-2">mdi-numeric-positive-1</v-icon>
          ONU ID Kosong (Bebas)
        </v-card-title>
        <v-card-text class="pt-2">
          <p class="text-body-2 text-grey-darken-1 mb-3">
            Daftar nomor ONU ID yang belum terpakai di <strong>Board {{ selectedBoard }} PON {{ selectedPon }}</strong>:
          </p>
          <div v-if="emptyOnuIds.length > 0" class="d-flex flex-wrap gap-2 max-h-48 overflow-y-auto pa-2 bg-grey-lighten-4 rounded-lg">
            <v-chip
              v-for="id in emptyOnuIds"
              :key="id"
              size="small"
              color="primary"
              variant="flat"
              class="font-weight-bold font-mono"
            >
              #{{ id }}
            </v-chip>
          </div>
          <div v-else class="text-center py-4 text-grey">
            Semua slot ONU ID di port ini sudah penuh.
          </div>
        </v-card-text>
        <v-card-actions class="pt-0">
          <v-spacer></v-spacer>
          <v-btn color="primary" variant="flat" rounded="lg" @click="dialogEmptyIds = false">Tutup</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-snackbar
      v-model="snackbar.show"
      :color="snackbar.color"
      :timeout="4000"
      location="top right"
      variant="elevated"
      elevation="8"
      rounded="lg"
    >
      <div class="d-flex align-center">
        <v-icon
          :icon="snackbar.color === 'success' ? 'mdi-check-circle' : 'mdi-alert-circle'"
          class="me-3"
          size="24"
        ></v-icon>
        <div class="font-weight-medium text-body-1">{{ snackbar.text }}</div>
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
const connectionStatus = ref<Record<number, boolean>>({}); // Track success status per item

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
const formTitle = computed(() => isEditMode.value ? 'Edit OLT' : 'Tambah OLT Baru');
const countOnline = computed(() => onusList.value.filter(o => o.status?.toLowerCase() === 'online').length);
const countOffline = computed(() => onusList.value.filter(o => o.status?.toLowerCase() !== 'online').length);

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

function getRxColor(powerStr: string): string {
  if (!powerStr) return 'grey';
  const p = parseFloat(powerStr);
  if (isNaN(p)) return 'grey';
  if (p >= -22.5 && p <= -10) return 'success';
  if (p < -22.5 && p >= -26.5) return 'warning';
  if (p < -26.5 || p > -8) return 'error';
  return 'info';
}

function getRxBgClass(powerStr: string): string {
  if (!powerStr) return 'bg-grey-lighten-4';
  const p = parseFloat(powerStr);
  if (isNaN(p)) return 'bg-grey-lighten-4';
  if (p >= -22.5 && p <= -10) return 'bg-green-50 border border-green-200';
  if (p < -22.5 && p >= -26.5) return 'bg-amber-50 border border-amber-200';
  if (p < -26.5 || p > -8) return 'bg-red-50 border border-red-200';
  return 'bg-blue-50 border border-blue-200';
}

function getRxTextClass(powerStr: string): string {
  if (!powerStr) return 'text-grey-darken-3';
  const p = parseFloat(powerStr);
  if (isNaN(p)) return 'text-grey-darken-3';
  if (p >= -22.5 && p <= -10) return 'text-green-700';
  if (p < -22.5 && p >= -26.5) return 'text-amber-700';
  if (p < -26.5 || p > -8) return 'text-red-700';
  return 'text-blue-700';
}

function copyText(text: string) {
  if (!text) return;
  navigator.clipboard.writeText(text);
  showSnackbar('Serial Number disalin: ' + text, 'success');
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
  } catch (error: any) {
    showSnackbar(error.response?.data?.message || 'Gagal mengambil detail telemetri ONU', 'error');
    dialogOnuDetail.value = false;
  } finally {
    loadingOnuDetail.value = false;
  }
}

// --- API FUNCTIONS ---
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
  editedItem.value = item ? { ...item, password: '' } : { tipe_olt: 'ZTE' };
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
    
    // Set success status for visuals
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
/* Main Header Styling */
.header-card {
  border-radius: 24px;
  overflow: hidden;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.08);
  background: white;
  position: relative;
  z-index: 1;
}

.header-section {
  background: linear-gradient(135deg, #4F46E5 0%, #7C3AED 100%);
  position: relative;
  overflow: hidden;
}

/* Background Pattern overlay */
.header-section::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-image: 
    radial-gradient(circle at 20% 150%, rgba(255, 255, 255, 0.1) 0%, transparent 50%),
    radial-gradient(circle at 80% -50%, rgba(255, 255, 255, 0.15) 0%, transparent 50%);
  z-index: 1;
}

.header-content {
  position: relative;
  padding: 40px 32px;
  z-index: 2;
}

.header-subtitle {
  color: rgba(255, 255, 255, 0.85) !important;
  font-size: 1.1rem;
  letter-spacing: 0.01em;
}

.add-btn {
  color: #4F46E5 !important;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.add-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.2);
}

/* Server Card Styling */
.server-card {
  border-radius: 20px;
  background: white;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid rgba(0, 0, 0, 0.05);
  position: relative;
  overflow: hidden;
}

.server-card:hover {
  transform: translateY(-8px);
  box-shadow: 0 20px 40px rgba(79, 70, 229, 0.1);
  border-color: rgba(79, 70, 229, 0.2);
}

.status-line {
  height: 4px;
  background: linear-gradient(90deg, #4F46E5, #7C3AED);
  width: 100%;
}

.icon-box {
  width: 50px;
  height: 50px;
  border-radius: 14px;
  background: linear-gradient(135deg, #EEF2FF 0%, #E0E7FF 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
}

.server-card:hover .icon-box {
  background: linear-gradient(135deg, #4F46E5 0%, #7C3AED 100%);
}

.server-card:hover .icon-box .v-icon {
  color: white !important;
}

.font-mono {
  font-family: 'SF Mono', 'Roboto Mono', monospace;
  letter-spacing: -0.5px;
}

.detail-item {
  transition: all 0.2s ease;
}

/* Dialog Styling */
.dialog-header-accent {
  height: 8px;
  background: linear-gradient(90deg, #4F46E5, #7C3AED);
  width: 100%;
}
</style>