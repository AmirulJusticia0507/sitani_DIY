// Tautan panduan resmi SSO Pemda DIY. Dipakai hanya untuk mengarahkan pengguna
// pengguna ke kanal bantuan; bukan alat autentikasi dan tidak mengirim data.
export const BANTUAN_SSO_URL = "https://wiki.jogjaprov.go.id/diskominfo/panduan/panduan-2fa";

// Realm SSO portal. Nilai ini sekadar kunci realm di backend
// (`/auth/login/{realm}`), bukan kredensial atau URL issuer.
export type RealmSSO = "jsp" | "satriya";

const REALM_LABEL: Record<RealmSSO, string> = {
  jsp: "JSP",
  satriya: "Satriya",
};

export function labelRealm(realm: RealmSSO): string {
  return REALM_LABEL[realm];
}

// redirect hanya menerima path internal agar tautan tidak bisa dipakai untuk
// open redirect. Hal yang sama divalidasi ulang oleh backend (safeRedirect).
export function pathAman(redirect: string | undefined): string {
  if (!redirect || !redirect.startsWith("/") || redirect.startsWith("//")) return "/";
  return redirect;
}

// Halaman login SiTani. Semua tombol "Masuk JSP/Satriya" mengarah ke sini agar
// pengguna melihat pilihan realm dan pesan yang jelas bila SSO belum aktif,
// lalu melanjutkan ke backend hanya saat menekan tombol di halaman login.
export function masukHref(redirect?: string, realm?: RealmSSO): string {
  const params = new URLSearchParams();
  const tujuan = pathAman(redirect);
  if (tujuan !== "/") params.set("redirect", tujuan);
  if (realm) params.set("realm", realm);
  const query = params.toString();
  return query ? `/masuk?${query}` : "/masuk";
}

// URL backend yang memulai alur OIDC. Backend membuat state + PKCE lalu
// mengalihkan browser ke issuer SSO resmi; frontend tidak pernah menebak
// endpoint SSO maupun token.
export function ssoHref(apiURL: string, realm: RealmSSO, redirect?: string): string {
  const params = new URLSearchParams({ redirect: pathAman(redirect) });
  return `${apiURL}/auth/login/${realm}?${params.toString()}`;
}
