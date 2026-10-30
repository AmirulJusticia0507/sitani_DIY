import {
  getRealmLabel as getRealmLabelFromRealm,
  getRoleAccess,
  getRoleLabel as getRoleLabelFromRole,
  type Role,
} from "./role-access";

export type DashboardPrincipal = {
  user_id?: number;
  email: string;
  name?: string;
  username?: string;
  role: string;
  realm?: string;
};

export type DashboardSession =
  | { status: "memuat" }
  | { status: "tamu" }
  | { status: "masuk"; principal: DashboardPrincipal };

export const apiURL = process.env.NEXT_PUBLIC_BACKEND_URL ?? "http://localhost:8080";

// Endpoint sesi backend memakai principal JSON dari paket auth
// (`/auth/me`), bukan nama/role permukaan frontend. Ambil sekali per halaman
// dashboard; cache HTTP dinonaktifkan agar status keluar/masuk selalu segar.
export async function fetchSession(): Promise<DashboardSession> {
  try {
    const res = await fetch(`${apiURL}/auth/me`, { credentials: "include", cache: "no-store" });
    if (!res.ok) return { status: "tamu" };
    const data = (await res.json()) as DashboardPrincipal;
    if (!data || typeof data.role !== "string") return { status: "tamu" };
    return { status: "masuk", principal: data };
  } catch {
    return { status: "tamu" };
  }
}

export function getRoleAccessFor(principal: DashboardPrincipal) {
  return getRoleAccess(principal.role as Role);
}

export function getRoleLabel(principal: DashboardPrincipal) {
  return getRoleLabelFromRole(principal.role as Role);
}

export function getRealmLabel(principal: DashboardPrincipal) {
  return getRealmLabelFromRealm(principal.realm);
}

// Login lokal (backend `internal/auth/devlogin`) hanya tersedia bila backend
// mengaktifkannya lewat SITANI_DEV_LOGIN=true di luar produksi. Respons berupa
// 200 saat aktif dan 404 saat tidak, sehingga antarmuka tidak pernah
// menawarkan formulir yang akan ditolak.
export async function devLoginTersedia(): Promise<boolean> {
  try {
    const res = await fetch(`${apiURL}/auth/dev/status`, { cache: "no-store" });
    if (!res.ok) return false;
    const data = (await res.json()) as { aktif?: boolean };
    return data.aktif === true;
  } catch {
    return false;
  }
}

// akhiriSesi mengakhiri sesi lokal dan sesi SSO. Backend mencabut refresh
// token lalu mengembalikan `logout_url` end_session_endpoint penyedia
// identitas; hanya URL https/http milik penyedia yang diikuti agar respons
// backend yang tidak terduga tidak dapat mengalihkan pengguna ke situs lain.
export async function akhiriSesi(): Promise<string | null> {
  const [, sso] = await Promise.all([
    fetch(`${apiURL}/auth/dev/logout`, { method: "POST", credentials: "include" }).catch(() => null),
    fetch(`${apiURL}/auth/logout`, { method: "POST", credentials: "include" }).catch(() => null),
  ]);
  if (!sso || !sso.ok) return null;
  try {
    const data = (await sso.json()) as { logout_url?: unknown };
    if (typeof data.logout_url !== "string" || data.logout_url === "") return null;
    const tujuan = new URL(data.logout_url);
    if (tujuan.protocol !== "https:" && tujuan.protocol !== "http:") return null;
    return tujuan.toString();
  } catch {
    return null;
  }
}
