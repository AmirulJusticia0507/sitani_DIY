// Peta peran dan modul dashboard. Sumber kebenaran otorisasi tetap backend
// (`backend/internal/httpapi` dan `backend/internal/service`); peta ini hanya
// mengatur modul mana yang tampil di navigasi dan panel kerja.
export type Role =
  | "pelanggan"
  | "admin"
  | "superadmin"
  | "pengawas"
  | "verifikator"
  | "inspektor"
  | "pbt"
  | "kasi_pmbtp"
  | "analis"
  | "penandatangan"
  | "legalisasi"
  | "percetakan";

export type DashboardModule =
  | "ringkasan"
  | "repository-data"
  | "hortikultura"
  | "sertifikasi"
  | "laboratorium"
  | "penerbitan"
  | "registrasi-benih"
  | "penjualan-benih"
  | "laporan"
  | "dokumen"
  | "profil";

const SEMUA_MODUL: DashboardModule[] = [
  "ringkasan",
  "repository-data",
  "hortikultura",
  "sertifikasi",
  "laboratorium",
  "penerbitan",
  "registrasi-benih",
  "penjualan-benih",
  "laporan",
  "dokumen",
  "profil",
];

const MODUL: Record<Role, DashboardModule[]> = {
  superadmin: SEMUA_MODUL,
  admin: SEMUA_MODUL,
  pelanggan: ["ringkasan", "sertifikasi", "hortikultura", "dokumen", "profil"],
  pengawas: ["ringkasan", "sertifikasi", "hortikultura", "registrasi-benih", "penjualan-benih", "laporan", "dokumen", "profil"],
  verifikator: ["ringkasan", "hortikultura", "dokumen", "profil"],
  inspektor: ["ringkasan", "hortikultura", "dokumen", "profil"],
  pbt: ["ringkasan", "hortikultura", "dokumen", "profil"],
  kasi_pmbtp: ["ringkasan", "hortikultura", "dokumen", "profil"],
  analis: ["ringkasan", "repository-data", "laboratorium", "dokumen", "profil"],
  penandatangan: ["ringkasan", "penerbitan", "dokumen", "profil"],
  legalisasi: ["ringkasan", "penerbitan", "dokumen", "profil"],
  percetakan: ["ringkasan", "penerbitan", "dokumen", "profil"],
};

const LABEL_PERAN: Record<Role, string> = {
  pelanggan: "Pelanggan",
  admin: "Administrator",
  superadmin: "Superadmin",
  pengawas: "Pengawas / PBT",
  verifikator: "Petugas Verifikator",
  inspektor: "Petugas Inspektor",
  pbt: "PBT",
  kasi_pmbtp: "Kasi PMBTP",
  analis: "Analis Laboratorium",
  penandatangan: "Penandatangan TTE",
  legalisasi: "Petugas Legalisasi",
  percetakan: "Percetakan",
};

export function isRole(value: string): value is Role {
  return Object.prototype.hasOwnProperty.call(MODUL, value);
}

export function getRoleLabel(role?: string): string {
  if (role && isRole(role)) return LABEL_PERAN[role];
  return role ? role.replaceAll("_", " ") : "Tanpa peran";
}

export function getRealmLabel(realm?: string): string {
  return realm === "satriya" ? "Satriya (ASN)" : "JSP (Masyarakat)";
}

// Peran yang belum dipetakan tidak diberi akses modul apa pun selain ringkasan
// dan profil agar antarmuka tidak menampilkan menu yang akan ditolak backend.
export function getRoleAccess(role?: string): Set<DashboardModule> {
  const daftar: DashboardModule[] = role && isRole(role) ? MODUL[role] : ["ringkasan", "profil"];
  return new Set<DashboardModule>(daftar);
}

export type DashboardNavItem = {
  key: DashboardModule;
  label: string;
  href: string;
  deskripsi: string;
};

const SEMUA_NAV: Record<DashboardModule, DashboardNavItem> = {
  ringkasan: { key: "ringkasan", label: "Ringkasan", href: "/dashboard", deskripsi: "Ikhtisar pekerjaan dan status layanan." },
  "repository-data": { key: "repository-data", label: "Repository Data", href: "/dashboard/repository-data", deskripsi: "Impor dan pantau dataset pertanian." },
  hortikultura: { key: "hortikultura", label: "Hortikultura", href: "/dashboard/horticulture", deskripsi: "Permohonan, verifikasi, inspeksi, dan penilaian." },
  sertifikasi: { key: "sertifikasi", label: "Sertifikasi Benih", href: "/dashboard/certification", deskripsi: "Alur sertifikasi dan penerbitan label BP3MBTP." },
  laboratorium: { key: "laboratorium", label: "Uji Mutu Lab", href: "/dashboard/certification", deskripsi: "Kodifikasi sampel, hasil uji, dan laporan mutu." },
  penerbitan: { key: "penerbitan", label: "Penerbitan Dokumen", href: "/dashboard/certification", deskripsi: "TTE sertifikat, legalisasi, cetak, dan penyerahan." },
  "registrasi-benih": { key: "registrasi-benih", label: "Registrasi & Stok Benih", href: "/dashboard/benih", deskripsi: "Master lot, persediaan, dan mutasi stok." },
  "penjualan-benih": { key: "penjualan-benih", label: "Penjualan Benih", href: "/dashboard/benih", deskripsi: "Transaksi penjualan dan pembatalan." },
  laporan: { key: "laporan", label: "Laporan", href: "/dashboard/benih", deskripsi: "Registrasi, stok, kedaluwarsa, dan penjualan." },
  dokumen: { key: "dokumen", label: "Dokumen", href: "/dashboard/dokumen", deskripsi: "Lampiran privat sesuai hak akses." },
  profil: { key: "profil", label: "Profil & Akses", href: "/dashboard/profil", deskripsi: "Identitas sesi dan modul yang tersedia." },
};

const URUTAN_NAV: DashboardModule[] = [
  "ringkasan",
  "repository-data",
  "hortikultura",
  "sertifikasi",
  "laboratorium",
  "penerbitan",
  "registrasi-benih",
  "penjualan-benih",
  "laporan",
  "dokumen",
  "profil",
];

export function getDashboardNav(role?: string): DashboardNavItem[] {
  const akses = getRoleAccess(role);
  return URUTAN_NAV.filter((key) => akses.has(key)).map((key) => SEMUA_NAV[key]);
}
