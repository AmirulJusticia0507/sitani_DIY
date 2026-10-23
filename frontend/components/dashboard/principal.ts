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
