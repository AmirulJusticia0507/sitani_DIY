// Worker router SiTani DIY.
//
// Satu domain publik, dua origin:
//   /v1/*, /auth/*, /health  -> backend Go
//   sisanya                  -> frontend Next.js di Vercel
//
// Tujuannya menjaga frontend dan backend tetap satu situs sehingga cookie sesi
// SameSite=Lax tetap terkirim. Worker ini belum pernah di-deploy dari repositori
// ini; lihat docs/deployment-cloudflare-vercel.md.

const BACKEND_PREFIXES = ["/v1/", "/auth/", "/health"];

function keBackend(pathname) {
  if (pathname === "/health") return true;
  return BACKEND_PREFIXES.some((prefix) =>
    prefix.endsWith("/") ? pathname.startsWith(prefix) : pathname === prefix,
  );
}

export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    const origin = keBackend(url.pathname) ? env.BACKEND_ORIGIN : env.FRONTEND_ORIGIN;

    if (!origin) {
      // Jangan meneruskan ke origin yang tidak diketahui: lebih baik gagal jelas
      // daripada mengirim trafik API ke frontend.
      return new Response("Origin tujuan belum dikonfigurasi pada Worker.", {
        status: 503,
        headers: { "content-type": "text/plain; charset=utf-8" },
      });
    }

    const target = new URL(url.pathname + url.search, origin);
    const upstream = new Request(target, request);

    // Redirect harus diteruskan ke peramban, bukan diikuti Worker. Alur
    // /auth/login mengalihkan pengguna ke halaman SSO, dan Location itu harus
    // sampai apa adanya.
    const response = await fetch(upstream, { redirect: "manual" });

    // Set-Cookie ganda (access token, refresh token, id token) tetap utuh karena
    // respons dikembalikan tanpa dibungkus ulang.
    return response;
  },
};
