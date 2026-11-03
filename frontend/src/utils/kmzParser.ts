import JSZip from 'jszip';

export interface KmzNode {
  id: string;
  name: string;
  type: 'server' | 'olt' | 'odc' | 'odp' | 'pelanggan';
  lat: number;
  lng: number;
  description?: string;
  kapasitas_port?: number;
  port_terpakai?: number;
  parent_name?: string;
  selected?: boolean;
}

export interface KmzLine {
  id: string;
  name: string;
  type: 'server_olt' | 'olt_odc' | 'odc_odp' | 'odp_pelanggan' | 'cable';
  coordinates: Array<[number, number]>; // [lat, lng]
  color?: string;
  description?: string;
}

export interface KmzParseResult {
  nodes: KmzNode[];
  lines: KmzLine[];
  fileName: string;
  summary: {
    servers: number;
    olts: number;
    odcs: number;
    odps: number;
    pelanggans: number;
    lines: number;
  };
}

/**
 * Mendeteksi tipe perangkat/titik berdasarkan nama atau deskripsi KML
 */
export function guessNodeType(name: string, desc: string = ''): 'server' | 'olt' | 'odc' | 'odp' | 'pelanggan' {
  const text = (name + ' ' + desc).toLowerCase();

  if (text.includes('server') || text.includes('core') || text.includes('noc') || text.includes('pop')) {
    return 'server';
  }
  if (text.includes('olt')) {
    return 'olt';
  }
  if (text.includes('odc') || text.includes('fdt') || text.includes('feeder')) {
    return 'odc';
  }
  if (text.includes('odp') || text.includes('tiang') || text.includes('fat') || text.includes('distribusi')) {
    return 'odp';
  }
  if (text.includes('pelanggan') || text.includes('cust') || text.includes('home') || text.includes('rumah') || text.includes('user') || text.includes('ont')) {
    return 'pelanggan';
  }

  // Default ke ODP jika tidak ada pola lain
  return 'odp';
}

/**
 * Parsing file KMZ (zip) atau KML (xml)
 */
export async function parseKmzOrKml(file: File): Promise<KmzParseResult> {
  let kmlText = '';
  const isKmz = file.name.toLowerCase().endsWith('.kmz');

  if (isKmz) {
    const zip = await JSZip.loadAsync(file);
    // Cari file dengan ekstensi .kml di dalam arsip ZIP
    const kmlEntry = Object.values(zip.files).find(
      (f) => !f.dir && f.name.toLowerCase().endsWith('.kml')
    );
    if (!kmlEntry) {
      throw new Error('File doc.kml tidak ditemukan di dalam arsip KMZ');
    }
    kmlText = await kmlEntry.async('text');
  } else {
    kmlText = await file.text();
  }

  const parser = new DOMParser();
  const xmlDoc = parser.parseFromString(kmlText, 'text/xml');

  // Cek apakah ada error parsing XML
  const parseError = xmlDoc.querySelector('parsererror');
  if (parseError) {
    throw new Error('Gagal membaca format XML KML: ' + parseError.textContent);
  }

  const placemarks = xmlDoc.querySelectorAll('Placemark');
  const nodes: KmzNode[] = [];
  const lines: KmzLine[] = [];

  let idCounter = 1;

  placemarks.forEach((pm) => {
    const name = pm.querySelector('name')?.textContent?.trim() || `Titik #${idCounter}`;
    const desc = pm.querySelector('description')?.textContent?.trim() || '';

    // 1. Periksa apakah Placemark berupa Point (Titik Koordinat)
    const pointCoord = pm.querySelector('Point coordinates')?.textContent?.trim();
    if (pointCoord) {
      // Format KML: longitude,latitude,altitude
      const parts = pointCoord.split(',').map((s) => parseFloat(s.trim()));
      if (parts.length >= 2 && !isNaN(parts[0]) && !isNaN(parts[1])) {
        const lng = parts[0];
        const lat = parts[1];

        // Validasi koordinat bumi
        if (lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180) {
          const type = guessNodeType(name, desc);
          nodes.push({
            id: `kmz_node_${idCounter++}`,
            name,
            type,
            lat,
            lng,
            description: desc,
            kapasitas_port: type === 'odc' ? 144 : type === 'odp' ? 8 : undefined,
            port_terpakai: 0,
            selected: true,
          });
        }
      }
    }

    // 2. Periksa apakah Placemark berupa LineString (Jalur Kabel)
    const lineCoord = pm.querySelector('LineString coordinates')?.textContent?.trim();
    if (lineCoord) {
      const coordPairs = lineCoord.split(/\s+/).filter(Boolean);
      const points: Array<[number, number]> = [];

      coordPairs.forEach((pair) => {
        const parts = pair.split(',').map((s) => parseFloat(s.trim()));
        if (parts.length >= 2 && !isNaN(parts[0]) && !isNaN(parts[1])) {
          const lng = parts[0];
          const lat = parts[1];
          if (lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180) {
            points.push([lat, lng]);
          }
        }
      });

      if (points.length >= 2) {
        lines.push({
          id: `kmz_line_${idCounter++}`,
          name,
          type: 'cable',
          coordinates: points,
          description: desc,
        });
      }
    }
  });

  const summary = {
    servers: nodes.filter((n) => n.type === 'server').length,
    olts: nodes.filter((n) => n.type === 'olt').length,
    odcs: nodes.filter((n) => n.type === 'odc').length,
    odps: nodes.filter((n) => n.type === 'odp').length,
    pelanggans: nodes.filter((n) => n.type === 'pelanggan').length,
    lines: lines.length,
  };

  return {
    nodes,
    lines,
    fileName: file.name,
    summary,
  };
}
