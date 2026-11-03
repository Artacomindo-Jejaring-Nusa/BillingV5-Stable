<template>
  <div class="gis-map-wrapper position-relative" :class="{ 'gis-fullscreen': isFullscreen }">
    <!-- Top Floating Toolbar / HUD -->
    <div class="gis-hud-toolbar elevation-4 rounded-lg pa-2 bg-white d-flex flex-wrap align-center gap-2">
      <!-- Search Input -->
      <v-text-field
        v-model="searchQuery"
        placeholder="Cari ODP, ODC, atau Pelanggan..."
        prepend-inner-icon="mdi-magnify"
        density="compact"
        variant="solo"
        flat
        hide-details
        clearable
        class="gis-search-input"
        @keyup.enter="searchAndZoom"
      ></v-text-field>

      <v-btn
        size="small"
        color="primary"
        variant="tonal"
        class="text-none"
        @click="searchAndZoom"
      >
        Cari
      </v-btn>

      <v-divider vertical class="mx-1"></v-divider>

      <!-- Drag / Edit Coordinates Toggle -->
      <v-btn
        size="small"
        :color="isEditMode ? 'error' : 'secondary'"
        :variant="isEditMode ? 'flat' : 'tonal'"
        :prepend-icon="isEditMode ? 'mdi-cursor-move' : 'mdi-map-marker-distance'"
        class="text-none font-weight-bold"
        @click="toggleEditMode"
      >
        {{ isEditMode ? 'Mode Geser: AKTIF' : 'Atur / Geser Posisi' }}
      </v-btn>

      <!-- Basemap Switcher -->
      <v-btn-toggle
        v-model="activeBasemap"
        mandatory
        density="compact"
        variant="outlined"
        divided
        class="ms-1"
        @update:model-value="changeBasemap"
      >
        <v-btn value="osm" size="small" class="text-none">
          <v-icon start size="16">mdi-map-outline</v-icon>
          Peta
        </v-btn>
        <v-btn value="satellite" size="small" class="text-none">
          <v-icon start size="16">mdi-satellite-variant</v-icon>
          Satelit
        </v-btn>
      </v-btn-toggle>

      <!-- Layer Filter Menu -->
      <v-menu :close-on-content-click="false" location="bottom end">
        <template v-slot:activator="{ props }">
          <v-btn
            size="small"
            variant="tonal"
            color="primary"
            prepend-icon="mdi-layers-outline"
            v-bind="props"
            class="text-none"
          >
            Filter Layer
          </v-btn>
        </template>
        <v-card min-width="220" rounded="lg" class="pa-2">
          <v-list density="compact">
            <v-list-item>
              <template v-slot:prepend>
                <v-checkbox-btn v-model="visibleLayers.server" color="indigo" density="compact"></v-checkbox-btn>
              </template>
              <v-list-item-title class="d-flex align-center">
                <span class="legend-dot bg-indigo me-2"></span>
                Server ({{ servers.length }})
              </v-list-item-title>
            </v-list-item>
            <v-list-item>
              <template v-slot:prepend>
                <v-checkbox-btn v-model="visibleLayers.olt" color="purple" density="compact"></v-checkbox-btn>
              </template>
              <v-list-item-title class="d-flex align-center">
                <span class="legend-dot bg-purple me-2"></span>
                OLT ({{ olts.length }})
              </v-list-item-title>
            </v-list-item>
            <v-list-item>
              <template v-slot:prepend>
                <v-checkbox-btn v-model="visibleLayers.odc" color="orange" density="compact"></v-checkbox-btn>
              </template>
              <v-list-item-title class="d-flex align-center">
                <span class="legend-dot bg-orange me-2"></span>
                ODC ({{ odcs.length }})
              </v-list-item-title>
            </v-list-item>
            <v-list-item>
              <template v-slot:prepend>
                <v-checkbox-btn v-model="visibleLayers.odp" color="success" density="compact"></v-checkbox-btn>
              </template>
              <v-list-item-title class="d-flex align-center">
                <span class="legend-dot bg-success me-2"></span>
                ODP ({{ odps.length }})
              </v-list-item-title>
            </v-list-item>
            <v-list-item>
              <template v-slot:prepend>
                <v-checkbox-btn v-model="visibleLayers.pelanggan" color="cyan" density="compact"></v-checkbox-btn>
              </template>
              <v-list-item-title class="d-flex align-center">
                <span class="legend-dot bg-cyan me-2"></span>
                Pelanggan ({{ pelanggans.length }})
              </v-list-item-title>
            </v-list-item>
            <v-list-item>
              <template v-slot:prepend>
                <v-checkbox-btn v-model="visibleLayers.lines" color="grey-darken-2" density="compact"></v-checkbox-btn>
              </template>
              <v-list-item-title class="d-flex align-center">
                <v-icon size="16" class="me-2">mdi-vector-polyline</v-icon>
                Jalur Kabel Fiber
              </v-list-item-title>
            </v-list-item>
          </v-list>
        </v-card>
      </v-menu>

      <!-- Center / Reset Zoom -->
      <v-btn icon="mdi-crosshairs-gps" size="small" variant="text" title="Pusatkan Peta" @click="fitAllBounds"></v-btn>

      <!-- Fullscreen Toggle -->
      <v-btn
        :icon="isFullscreen ? 'mdi-fullscreen-exit' : 'mdi-fullscreen'"
        size="small"
        variant="text"
        :title="isFullscreen ? 'Keluar Layar Penuh' : 'Layar Penuh'"
        @click="toggleFullscreen"
      ></v-btn>
    </div>

    <!-- Active Edit Mode Alert Banner -->
    <v-alert
      v-if="isEditMode"
      type="warning"
      variant="elevated"
      density="compact"
      class="gis-edit-alert elevation-4"
    >
      <div class="d-flex align-center justify-space-between w-100">
        <div>
          <strong>Mode Geser Titik Aktif!</strong> Drag icon titik ODP / ODC / Server pada peta untuk memperbarui koordinat Latitude &amp; Longitude secara instan.
        </div>
        <v-btn size="x-small" variant="elevated" color="white" class="ms-3 text-dark" @click="isEditMode = false">
          Selesai
        </v-btn>
      </div>
    </v-alert>

    <!-- Map Container -->
    <div
      ref="mapContainer"
      class="gis-map-canvas"
      :style="{ height: mapHeightStyle }"
    ></div>

    <!-- Bottom Flow Legend -->
    <div class="gis-flow-legend elevation-3 pa-2 rounded-lg bg-white d-none d-md-flex align-center gap-2">
      <span class="text-caption font-weight-bold me-1 text-grey-darken-2">Hierarki:</span>
      <div class="legend-badge bg-indigo-lighten-5 text-indigo">
        <v-icon size="14" class="me-1">mdi-server-network</v-icon> Server
      </div>
      <v-icon size="14" color="grey">mdi-arrow-right</v-icon>
      <div class="legend-badge bg-purple-lighten-5 text-purple">
        <v-icon size="14" class="me-1">mdi-router-wireless</v-icon> OLT
      </div>
      <v-icon size="14" color="grey">mdi-arrow-right</v-icon>
      <div class="legend-badge bg-orange-lighten-5 text-orange-darken-2">
        <v-icon size="14" class="me-1">mdi-inbox-multiple</v-icon> ODC
      </div>
      <v-icon size="14" color="grey">mdi-arrow-right</v-icon>
      <div class="legend-badge bg-green-lighten-5 text-green-darken-2">
        <v-icon size="14" class="me-1">mdi-transmission-tower</v-icon> ODP
      </div>
      <v-icon size="14" color="grey">mdi-arrow-right</v-icon>
      <div class="legend-badge bg-cyan-lighten-5 text-cyan-darken-2">
        <v-icon size="14" class="me-1">mdi-home-outline</v-icon> Pelanggan
      </div>
    </div>

    <!-- Move Confirmation Dialog -->
    <v-dialog v-model="moveConfirmDialog" max-width="420px" persistent>
      <v-card rounded="xl" class="pa-4">
        <v-card-title class="d-flex align-center pb-2">
          <v-icon color="warning" class="me-2">mdi-map-marker-alert</v-icon>
          Simpan Posisi Baru?
        </v-card-title>
        <v-card-text class="pt-2 text-body-2">
          Titik <strong>{{ pendingMove?.name }}</strong> ({{ pendingMove?.type.toUpperCase() }}) dipindahkan ke:
          <div class="pa-3 bg-grey-lighten-4 rounded-lg font-mono text-caption mt-2">
            <div><strong>Lat:</strong> {{ pendingMove?.lat.toFixed(6) }}</div>
            <div><strong>Lng:</strong> {{ pendingMove?.lng.toFixed(6) }}</div>
          </div>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn variant="text" @click="cancelMove">Batal</v-btn>
          <v-btn color="primary" variant="flat" :loading="savingMove" @click="confirmMove">Simpan Koordinat</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, computed, nextTick } from 'vue';
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';
import apiClient from '@/services/api';
import type { KmzParseResult } from '@/utils/kmzParser';

// --- Interfaces ---
export interface ServerNode {
  id: number;
  name: string;
  host_ip: string;
  latitude?: number | null;
  longitude?: number | null;
  is_active: boolean;
}

export interface OltNode {
  id: number;
  nama_olt: string;
  ip_address: string;
  tipe_olt: string;
  mikrotik_server_id?: number | null;
  latitude?: number | null;
  longitude?: number | null;
}

export interface OdcNode {
  id: number;
  kode_odc: string;
  nama_odc: string;
  alamat?: string | null;
  kapasitas_port: number;
  latitude?: number | null;
  longitude?: number | null;
  olt_id?: number | null;
}

export interface OdpNode {
  id: number;
  kode_odp: string;
  alamat?: string | null;
  kapasitas_port: number;
  port_terpakai?: number;
  latitude?: number | null;
  longitude?: number | null;
  olt_id: number;
  odc_id?: number | null;
  parent_odp_id?: number | null;
  olt?: { id: number; nama_olt: string };
  odc?: { id: number; kode_odc: string; nama_odc: string };
}

export interface PelangganNode {
  id: number;
  nama: string;
  alamat: string;
  no_telp: string;
  latitude?: number | null;
  longitude?: number | null;
  odp_id?: number | null;
  kode_odp?: string;
  port_odp?: number | null;
  username_pppoe?: string;
  status: string;
}

// --- Props & Emits ---
const props = withDefaults(
  defineProps<{
    odps?: OdpNode[];
    olts?: OltNode[];
    servers?: ServerNode[];
    odcs?: OdcNode[];
    pelanggans?: PelangganNode[];
    height?: string;
    kmzPreview?: KmzParseResult | null;
  }>(),
  {
    odps: () => [],
    olts: () => [],
    servers: () => [],
    odcs: () => [],
    pelanggans: () => [],
    height: '650px',
    kmzPreview: null,
  }
);

const emit = defineEmits<{
  (e: 'coordinate-updated', payload: { type: string; id: number; lat: number; lng: number }): void;
  (e: 'add-odp-click', payload: { lat: number; lng: number }): void;
  (e: 'edit-odp', odp: OdpNode): void;
}>();

// --- Map References & State ---
const mapContainer = ref<HTMLDivElement | null>(null);
let map: L.Map | null = null;
let osmTileLayer: L.TileLayer | null = null;
let satTileLayer: L.TileLayer | null = null;

// Layer Groups
let serverLayerGroup: L.LayerGroup | null = null;
let oltLayerGroup: L.LayerGroup | null = null;
let odcLayerGroup: L.LayerGroup | null = null;
let odpLayerGroup: L.LayerGroup | null = null;
let pelangganLayerGroup: L.LayerGroup | null = null;
let cableLinesLayerGroup: L.LayerGroup | null = null;
let kmzPreviewLayerGroup: L.LayerGroup | null = null;

const isFullscreen = ref(false);
const isEditMode = ref(false);
const activeBasemap = ref<'osm' | 'satellite'>('osm');
const searchQuery = ref('');

const visibleLayers = ref({
  server: true,
  olt: true,
  odc: true,
  odp: true,
  pelanggan: true,
  lines: true,
});

// Drag & drop confirmation
const moveConfirmDialog = ref(false);
const savingMove = ref(false);
const pendingMove = ref<{
  type: 'server' | 'olt' | 'odc' | 'odp' | 'pelanggan';
  id: number;
  name: string;
  lat: number;
  lng: number;
  originalMarker: L.Marker;
  originalLatLng: L.LatLng;
} | null>(null);

const mapHeightStyle = computed(() => {
  if (isFullscreen.value) return '100vh';
  return props.height;
});

// --- Custom Vector Icons ---
function createSvgIcon(type: 'server' | 'olt' | 'odc' | 'odp' | 'pelanggan', label: string, statusColor = '#10b981') {
  let bgColor = '#10b981';
  let iconSvg = '';

  switch (type) {
    case 'server':
      bgColor = '#4f46e5'; // Indigo
      iconSvg = `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="2" width="20" height="8" rx="2" ry="2"></rect><rect x="2" y="14" width="20" height="8" rx="2" ry="2"></rect><line x1="6" y1="6" x2="6.01" y2="6"></line><line x1="6" y1="18" x2="6.01" y2="18"></line></svg>`;
      break;
    case 'olt':
      bgColor = '#9333ea'; // Purple
      iconSvg = `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M4.93 19.07a10 10 0 0 1 0-14.14"></path><path d="M7.76 16.24a6 6 0 0 1 0-8.48"></path><circle cx="12" cy="12" r="2"></circle><path d="M16.24 7.76a6 6 0 0 1 0 8.48"></path><path d="M19.07 4.93a10 10 0 0 1 0 14.14"></path></svg>`;
      break;
    case 'odc':
      bgColor = '#ea580c'; // Orange
      iconSvg = `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 12 16 12 14 15 10 15 8 12 2 12"></polyline><path d="M5.45 5.11L2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z"></path></svg>`;
      break;
    case 'odp':
      bgColor = statusColor; // Green, Orange, or Red
      iconSvg = `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polygon points="12 2 2 7 12 12 22 7 12 2"></polygon><polyline points="2 17 12 22 22 17"></polyline><polyline points="2 12 12 17 22 12"></polyline></svg>`;
      break;
    case 'pelanggan':
      bgColor = '#06b6d4'; // Cyan
      iconSvg = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path><polyline points="9 22 9 12 15 12 15 22"></polyline></svg>`;
      break;
  }

  const html = `
    <div class="gis-marker-container">
      <div class="gis-marker-bubble" style="background-color: ${bgColor}; box-shadow: 0 4px 10px ${bgColor}66;">
        ${iconSvg}
      </div>
      <div class="gis-marker-label">${label}</div>
    </div>
  `;

  return L.divIcon({
    className: 'gis-custom-div-icon',
    html,
    iconSize: [36, 48],
    iconAnchor: [18, 20],
    popupAnchor: [0, -22],
  });
}

// --- Map Initialization ---
async function initMap() {
  await nextTick();
  if (!mapContainer.value) return;

  // Initialize Map with smooth zoom
  map = L.map(mapContainer.value, {
    zoomControl: false,
    minZoom: 4,
    maxZoom: 21,
  }).setView([-6.2383, 106.9756], 13);

  // Zoom control in bottom right
  L.control.zoom({ position: 'bottomright' }).addTo(map);

  // Basemap Tile Layers
  osmTileLayer = L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; OpenStreetMap',
    maxZoom: 19,
  });

  satTileLayer = L.tileLayer(
    'https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}',
    {
      attribution: 'Esri World Imagery',
      maxZoom: 19,
    }
  );

  // Add default layer
  osmTileLayer.addTo(map);

  // Layer Groups
  cableLinesLayerGroup = L.layerGroup().addTo(map);
  serverLayerGroup = L.layerGroup().addTo(map);
  oltLayerGroup = L.layerGroup().addTo(map);
  odcLayerGroup = L.layerGroup().addTo(map);
  odpLayerGroup = L.layerGroup().addTo(map);
  pelangganLayerGroup = L.layerGroup().addTo(map);
  kmzPreviewLayerGroup = L.layerGroup().addTo(map);

  // Right-click or Click to add ODP
  map.on('contextmenu', (e: L.LeafletMouseEvent) => {
    emit('add-odp-click', { lat: e.latlng.lat, lng: e.latlng.lng });
  });

  renderAll();
}

// --- Basemap Switch ---
function changeBasemap(val: 'osm' | 'satellite') {
  if (!map) return;
  if (val === 'satellite') {
    if (osmTileLayer) map.removeLayer(osmTileLayer);
    if (satTileLayer) satTileLayer.addTo(map);
  } else {
    if (satTileLayer) map.removeLayer(satTileLayer);
    if (osmTileLayer) osmTileLayer.addTo(map);
  }
}

// --- Render All Nodes & Cables ---
function renderAll() {
  if (!map) return;

  // Clear existing layers
  serverLayerGroup?.clearLayers();
  oltLayerGroup?.clearLayers();
  odcLayerGroup?.clearLayers();
  odpLayerGroup?.clearLayers();
  pelangganLayerGroup?.clearLayers();
  cableLinesLayerGroup?.clearLayers();

  const allBounds: L.LatLng[] = [];

  // 1. Render Servers
  const serverMap = new Map<number, ServerNode>();
  props.servers.forEach((server) => {
    serverMap.set(server.id, server);
    if (server.latitude && server.longitude && visibleLayers.value.server) {
      const latlng = L.latLng(server.latitude, server.longitude);
      allBounds.push(latlng);

      const marker = L.marker(latlng, {
        icon: createSvgIcon('server', server.name),
        draggable: isEditMode.value,
      });

      marker.bindPopup(`
        <div class="gis-popup-content">
          <div class="gis-popup-header bg-indigo text-white pa-2 font-weight-bold">
            <i class="mdi mdi-server-network"></i> MikroTik Server
          </div>
          <div class="pa-3 text-body-2">
            <div class="font-weight-bold text-subtitle-2 mb-1">${server.name}</div>
            <div><strong>Host IP:</strong> ${server.host_ip}</div>
            <div><strong>Status:</strong> ${server.is_active ? 'Online' : 'Offline'}</div>
            <div class="text-caption text-grey mt-2 font-mono">Lat: ${server.latitude.toFixed(6)}, Lng: ${server.longitude.toFixed(6)}</div>
          </div>
        </div>
      `);

      setupDragEvents(marker, 'server', server.id, server.name);
      serverLayerGroup?.addLayer(marker);
    }
  });

  // 2. Render OLTs
  const oltMap = new Map<number, OltNode>();
  props.olts.forEach((olt) => {
    oltMap.set(olt.id, olt);
    if (olt.latitude && olt.longitude && visibleLayers.value.olt) {
      const latlng = L.latLng(olt.latitude, olt.longitude);
      allBounds.push(latlng);

      const marker = L.marker(latlng, {
        icon: createSvgIcon('olt', olt.nama_olt),
        draggable: isEditMode.value,
      });

      marker.bindPopup(`
        <div class="gis-popup-content">
          <div class="gis-popup-header bg-purple text-white pa-2 font-weight-bold">
            <i class="mdi mdi-router-wireless"></i> OLT Node
          </div>
          <div class="pa-3 text-body-2">
            <div class="font-weight-bold text-subtitle-2 mb-1">${olt.nama_olt}</div>
            <div><strong>IP OLT:</strong> ${olt.ip_address}</div>
            <div><strong>Tipe:</strong> ${olt.tipe_olt}</div>
            <div class="text-caption text-grey mt-2 font-mono">Lat: ${olt.latitude.toFixed(6)}, Lng: ${olt.longitude.toFixed(6)}</div>
          </div>
        </div>
      `);

      setupDragEvents(marker, 'olt', olt.id, olt.nama_olt);
      oltLayerGroup?.addLayer(marker);

      // Line: Server -> OLT
      if (olt.mikrotik_server_id && serverMap.has(olt.mikrotik_server_id) && visibleLayers.value.lines) {
        const srv = serverMap.get(olt.mikrotik_server_id)!;
        if (srv.latitude && srv.longitude) {
          const polyline = L.polyline(
            [
              [srv.latitude, srv.longitude],
              [olt.latitude, olt.longitude],
            ],
            {
              color: '#4f46e5',
              weight: 3.5,
              opacity: 0.85,
              dashArray: '6, 6',
            }
          ).bindTooltip(`Feeder: ${srv.name} → ${olt.nama_olt}`, { sticky: true });
          cableLinesLayerGroup?.addLayer(polyline);
        }
      }
    }
  });

  // 3. Render ODCs
  const odcMap = new Map<number, OdcNode>();
  props.odcs.forEach((odc) => {
    odcMap.set(odc.id, odc);
    if (odc.latitude && odc.longitude && visibleLayers.value.odc) {
      const latlng = L.latLng(odc.latitude, odc.longitude);
      allBounds.push(latlng);

      const marker = L.marker(latlng, {
        icon: createSvgIcon('odc', odc.nama_odc),
        draggable: isEditMode.value,
      });

      marker.bindPopup(`
        <div class="gis-popup-content">
          <div class="gis-popup-header bg-orange-darken-2 text-white pa-2 font-weight-bold">
            <i class="mdi mdi-inbox-multiple"></i> ODC (Kabinet Fiber)
          </div>
          <div class="pa-3 text-body-2">
            <div class="font-weight-bold text-subtitle-2 mb-1">${odc.nama_odc} (${odc.kode_odc})</div>
            <div><strong>Alamat:</strong> ${odc.alamat || '-'}</div>
            <div><strong>Kapasitas:</strong> ${odc.kapasitas_port} Port</div>
            <div class="text-caption text-grey mt-2 font-mono">Lat: ${odc.latitude.toFixed(6)}, Lng: ${odc.longitude.toFixed(6)}</div>
          </div>
        </div>
      `);

      setupDragEvents(marker, 'odc', odc.id, odc.nama_odc);
      odcLayerGroup?.addLayer(marker);

      // Line: OLT -> ODC
      if (odc.olt_id && oltMap.has(odc.olt_id) && visibleLayers.value.lines) {
        const parentOlt = oltMap.get(odc.olt_id)!;
        if (parentOlt.latitude && parentOlt.longitude) {
          const polyline = L.polyline(
            [
              [parentOlt.latitude, parentOlt.longitude],
              [odc.latitude, odc.longitude],
            ],
            {
              color: '#ea580c',
              weight: 3,
              opacity: 0.85,
            }
          ).bindTooltip(`Main Line: ${parentOlt.nama_olt} → ${odc.nama_odc}`, { sticky: true });
          cableLinesLayerGroup?.addLayer(polyline);
        }
      }
    }
  });

  // 4. Render ODPs
  const odpMap = new Map<number, OdpNode>();
  props.odps.forEach((odp) => {
    odpMap.set(odp.id, odp);
    if (odp.latitude && odp.longitude && visibleLayers.value.odp) {
      const latlng = L.latLng(odp.latitude, odp.longitude);
      allBounds.push(latlng);

      // Color by usage
      const used = odp.port_terpakai || 0;
      const total = odp.kapasitas_port || 8;
      const ratio = total > 0 ? used / total : 0;
      let statusColor = '#10b981'; // Green
      if (ratio >= 1) statusColor = '#ef4444'; // Red
      else if (ratio >= 0.75) statusColor = '#f59e0b'; // Amber

      const marker = L.marker(latlng, {
        icon: createSvgIcon('odp', odp.kode_odp, statusColor),
        draggable: isEditMode.value,
      });

      marker.bindPopup(`
        <div class="gis-popup-content">
          <div class="gis-popup-header" style="background-color: ${statusColor}; color: white; padding: 8px; font-weight: bold;">
            <i class="mdi mdi-transmission-tower"></i> ODP: ${odp.kode_odp}
          </div>
          <div class="pa-3 text-body-2">
            <div><strong>Lokasi:</strong> ${odp.alamat || '-'}</div>
            <div class="mt-2 mb-1">
              <strong>Kapasitas Port:</strong> ${used} / ${total}
              <div class="progress-bar-container" style="background: #e2e8f0; height: 8px; border-radius: 4px; overflow: hidden; margin-top: 4px;">
                <div style="background: ${statusColor}; width: ${(ratio * 100).toFixed(0)}%; height: 100%;"></div>
              </div>
            </div>
            ${odp.olt ? `<div class="mt-1"><strong>OLT:</strong> ${odp.olt.nama_olt}</div>` : ''}
            ${odp.odc ? `<div class="mt-1"><strong>ODC:</strong> ${odp.odc.nama_odc}</div>` : ''}
            <div class="text-caption text-grey mt-2 font-mono">Lat: ${odp.latitude.toFixed(6)}, Lng: ${odp.longitude.toFixed(6)}</div>
          </div>
        </div>
      `);

      setupDragEvents(marker, 'odp', odp.id, odp.kode_odp);
      odpLayerGroup?.addLayer(marker);

      // Line: ODC -> ODP or OLT -> ODP
      if (visibleLayers.value.lines) {
        if (odp.odc_id && odcMap.has(odp.odc_id)) {
          const parentOdc = odcMap.get(odp.odc_id)!;
          if (parentOdc.latitude && parentOdc.longitude) {
            const polyline = L.polyline(
              [
                [parentOdc.latitude, parentOdc.longitude],
                [odp.latitude, odp.longitude],
              ],
              {
                color: '#10b981',
                weight: 2.5,
                opacity: 0.8,
              }
            ).bindTooltip(`Distribusi: ${parentOdc.nama_odc} → ${odp.kode_odp}`, { sticky: true });
            cableLinesLayerGroup?.addLayer(polyline);
          }
        } else if (odp.olt_id && oltMap.has(odp.olt_id)) {
          const parentOlt = oltMap.get(odp.olt_id)!;
          if (parentOlt.latitude && parentOlt.longitude) {
            const polyline = L.polyline(
              [
                [parentOlt.latitude, parentOlt.longitude],
                [odp.latitude, odp.longitude],
              ],
              {
                color: '#10b981',
                weight: 2.5,
                opacity: 0.8,
              }
            ).bindTooltip(`Distribusi: ${parentOlt.nama_olt} → ${odp.kode_odp}`, { sticky: true });
            cableLinesLayerGroup?.addLayer(polyline);
          }
        }
      }
    }
  });

  // 5. Render Pelanggans
  props.pelanggans.forEach((pelanggan) => {
    if (pelanggan.latitude && pelanggan.longitude && visibleLayers.value.pelanggan) {
      const latlng = L.latLng(pelanggan.latitude, pelanggan.longitude);
      allBounds.push(latlng);

      const marker = L.marker(latlng, {
        icon: createSvgIcon('pelanggan', pelanggan.nama),
        draggable: isEditMode.value,
      });

      marker.bindPopup(`
        <div class="gis-popup-content">
          <div class="gis-popup-header bg-cyan-darken-2 text-white pa-2 font-weight-bold">
            <i class="mdi mdi-home-account"></i> Pelanggan: ${pelanggan.nama}
          </div>
          <div class="pa-3 text-body-2">
            <div><strong>Username PPPoE:</strong> ${pelanggan.username_pppoe || '-'}</div>
            <div><strong>Alamat:</strong> ${pelanggan.alamat}</div>
            <div><strong>Status:</strong> ${pelanggan.status}</div>
            <div><strong>ODP:</strong> ${pelanggan.kode_odp || '-'} (Port: ${pelanggan.port_odp || '-'})</div>
            <div class="text-caption text-grey mt-2 font-mono">Lat: ${pelanggan.latitude.toFixed(6)}, Lng: ${pelanggan.longitude.toFixed(6)}</div>
          </div>
        </div>
      `);

      setupDragEvents(marker, 'pelanggan', pelanggan.id, pelanggan.nama);
      pelangganLayerGroup?.addLayer(marker);

      // Line: ODP -> Pelanggan (Drop Cable)
      if (pelanggan.odp_id && odpMap.has(pelanggan.odp_id) && visibleLayers.value.lines) {
        const parentOdp = odpMap.get(pelanggan.odp_id)!;
        if (parentOdp.latitude && parentOdp.longitude) {
          const polyline = L.polyline(
            [
              [parentOdp.latitude, parentOdp.longitude],
              [pelanggan.latitude, pelanggan.longitude],
            ],
            {
              color: '#06b6d4',
              weight: 1.5,
              opacity: 0.7,
              dashArray: '3, 4',
            }
          ).bindTooltip(`Drop Cable: ${parentOdp.kode_odp} → ${pelanggan.nama}`, { sticky: true });
          cableLinesLayerGroup?.addLayer(polyline);
        }
      }
    }
  });

  // 6. Render KMZ Preview (if active)
  renderKmzPreview();

  // Auto-fit bounds if we have points and no specific zoom is locked
  if (allBounds.length > 0) {
    const bounds = L.latLngBounds(allBounds);
    map.fitBounds(bounds, { padding: [40, 40], maxZoom: 16 });
  }
}

// --- Render KMZ Preview Layer ---
function renderKmzPreview() {
  kmzPreviewLayerGroup?.clearLayers();
  if (!props.kmzPreview || !map) return;

  const { nodes, lines } = props.kmzPreview;

  // Render KMZ LineStrings
  lines.forEach((line) => {
    if (line.coordinates.length >= 2) {
      const poly = L.polyline(line.coordinates, {
        color: '#dc2626',
        weight: 3,
        dashArray: '4, 4',
        opacity: 0.9,
      }).bindTooltip(`KMZ Line: ${line.name}`, { sticky: true });
      kmzPreviewLayerGroup?.addLayer(poly);
    }
  });

  // Render KMZ Placemark Points
  nodes.forEach((node) => {
    const latlng = L.latLng(node.lat, node.lng);
    const marker = L.circleMarker(latlng, {
      radius: 7,
      fillColor: '#ef4444',
      fillOpacity: 0.9,
      color: '#ffffff',
      weight: 2,
    }).bindPopup(`
      <div class="pa-2">
        <div class="font-weight-bold text-red">KMZ Placemark</div>
        <div><strong>Nama:</strong> ${node.name}</div>
        <div><strong>Tipe:</strong> ${node.type.toUpperCase()}</div>
        <div class="text-caption">${node.description || ''}</div>
      </div>
    `);
    kmzPreviewLayerGroup?.addLayer(marker);
  });
}

// --- Drag and Drop Handling ---
function setupDragEvents(
  marker: L.Marker,
  type: 'server' | 'olt' | 'odc' | 'odp' | 'pelanggan',
  id: number,
  name: string
) {
  marker.on('dragstart', (e) => {
    const target = e.target as L.Marker;
    (target as any)._origLatLng = target.getLatLng();
  });

  marker.on('dragend', (e) => {
    const target = e.target as L.Marker;
    const newPos = target.getLatLng();
    const origPos = (target as any)._origLatLng || newPos;

    pendingMove.value = {
      type,
      id,
      name,
      lat: newPos.lat,
      lng: newPos.lng,
      originalMarker: target,
      originalLatLng: origPos,
    };

    moveConfirmDialog.value = true;
  });
}

async function confirmMove() {
  if (!pendingMove.value) return;

  savingMove.value = true;
  const { type, id, lat, lng } = pendingMove.value;

  try {
    // Call generic coordinate update endpoint
    await apiClient.patch('/network/coordinate', {
      type,
      id,
      latitude: lat,
      longitude: lng,
    });

    emit('coordinate-updated', { type, id, lat, lng });
    moveConfirmDialog.value = false;
    pendingMove.value = null;
  } catch (err: any) {
    console.error('Gagal memperbarui koordinat:', err);
    cancelMove();
  } finally {
    savingMove.value = false;
  }
}

function cancelMove() {
  if (pendingMove.value) {
    pendingMove.value.originalMarker.setLatLng(pendingMove.value.originalLatLng);
  }
  moveConfirmDialog.value = false;
  pendingMove.value = null;
}

// --- Toolbar Helpers ---
function toggleEditMode() {
  isEditMode.value = !isEditMode.value;
  renderAll();
}

function toggleFullscreen() {
  isFullscreen.value = !isFullscreen.value;
  setTimeout(() => {
    map?.invalidateSize();
  }, 200);
}

function fitAllBounds() {
  if (!map) return;
  const allPoints: L.LatLng[] = [];

  [...props.odps, ...props.olts, ...props.servers, ...props.odcs, ...props.pelanggans].forEach((node: any) => {
    if (node.latitude && node.longitude) {
      allPoints.push(L.latLng(node.latitude, node.longitude));
    }
  });

  if (allPoints.length > 0) {
    map.fitBounds(L.latLngBounds(allPoints), { padding: [40, 40], maxZoom: 16 });
  }
}

function searchAndZoom() {
  if (!searchQuery.value || !map) return;
  const q = searchQuery.value.toLowerCase().trim();

  // Search across ODPs, ODCs, Pelanggan
  const foundOdp = props.odps.find(
    (o) => o.kode_odp.toLowerCase().includes(q) || (o.alamat && o.alamat.toLowerCase().includes(q))
  );
  if (foundOdp && foundOdp.latitude && foundOdp.longitude) {
    map.flyTo([foundOdp.latitude, foundOdp.longitude], 18, { duration: 1.5 });
    return;
  }

  const foundPelanggan = props.pelanggans.find(
    (p) => p.nama.toLowerCase().includes(q) || (p.username_pppoe && p.username_pppoe.toLowerCase().includes(q))
  );
  if (foundPelanggan && foundPelanggan.latitude && foundPelanggan.longitude) {
    map.flyTo([foundPelanggan.latitude, foundPelanggan.longitude], 18, { duration: 1.5 });
    return;
  }

  const foundOdc = props.odcs.find(
    (o) => o.kode_odc.toLowerCase().includes(q) || o.nama_odc.toLowerCase().includes(q)
  );
  if (foundOdc && foundOdc.latitude && foundOdc.longitude) {
    map.flyTo([foundOdc.latitude, foundOdc.longitude], 18, { duration: 1.5 });
  }
}

// --- Watchers ---
watch(
  () => [props.odps, props.olts, props.servers, props.odcs, props.pelanggans],
  () => {
    renderAll();
  },
  { deep: true }
);

watch(
  () => props.kmzPreview,
  () => {
    renderKmzPreview();
  },
  { deep: true }
);

watch(
  visibleLayers,
  () => {
    renderAll();
  },
  { deep: true }
);

// --- Lifecycle Hooks ---
onMounted(() => {
  initMap();
});

onUnmounted(() => {
  map?.remove();
});
</script>

<style>
/* Leaflet Global Marker Styling */
.gis-custom-div-icon {
  background: transparent !important;
  border: none !important;
}

.gis-marker-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  position: relative;
}

.gis-marker-bubble {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: 2px solid white;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: transform 0.2s ease;
}

.gis-marker-bubble:hover {
  transform: scale(1.15);
}

.gis-marker-label {
  margin-top: 3px;
  font-size: 10px;
  font-weight: 700;
  color: #1e293b;
  background: rgba(255, 255, 255, 0.92);
  padding: 1px 5px;
  border-radius: 4px;
  white-space: nowrap;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(0, 0, 0, 0.05);
}

.gis-popup-content {
  min-width: 200px;
  border-radius: 8px;
  overflow: hidden;
}

.gis-popup-header {
  font-size: 12px;
}
</style>

<style scoped>
.gis-map-wrapper {
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid #e2e8f0;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.gis-fullscreen {
  position: fixed !important;
  top: 0 !important;
  left: 0 !important;
  width: 100vw !important;
  height: 100vh !important;
  z-index: 99999 !important;
  border-radius: 0 !important;
}

.gis-map-canvas {
  width: 100%;
}

.gis-hud-toolbar {
  position: absolute;
  top: 16px;
  left: 16px;
  right: 16px;
  z-index: 1000;
  border: 1px solid rgba(0, 0, 0, 0.08);
}

.gis-search-input {
  max-width: 260px;
}

.gis-edit-alert {
  position: absolute;
  top: 76px;
  left: 16px;
  right: 16px;
  z-index: 1000;
  border-radius: 8px;
}

.gis-flow-legend {
  position: absolute;
  bottom: 20px;
  left: 16px;
  z-index: 1000;
  border: 1px solid rgba(0, 0, 0, 0.08);
}

.legend-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 6px;
  display: flex;
  align-items: center;
}

.legend-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  display: inline-block;
}

.font-mono {
  font-family: monospace;
}
</style>
