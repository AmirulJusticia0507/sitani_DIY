// Diagnostik: apakah request /v1/permohonan/aktif dan /auth/me membawa cookie
// sesi setelah dev-login, dan apakah galat 401 datang dari request sebelum login.
export default async function run(page) {
  const catatan = [];
  page.on("request", (r) => {
    const url = r.url();
    if (!url.startsWith("http://localhost:8090")) return;
    catatan.push({
      url: url.replace("http://localhost:8090", ""),
      method: r.method(),
      cookie: (r.headers()["cookie"] || "").includes("sitani_session") ? "ada" : "tidak",
    });
  });
  const respons = [];
  page.on("response", (r) => {
    const url = r.url();
    if (!url.startsWith("http://localhost:8090")) return;
    respons.push({ url: url.replace("http://localhost:8090", ""), status: r.status() });
  });

  await page.goto("http://localhost:3002/certification");
  await page.waitForTimeout(2000);
  const sebelumLogin = { catatan: [...catatan], respons: [...respons] };

  const login = await page.request.post("http://localhost:8090/auth/dev/login", {
    data: { email: "uji@contoh.id", password: "rahasia-uji-123" },
    headers: { Origin: "http://localhost:3002" },
  });
  const cookies = await page.context().cookies("http://localhost:8090");

  catatan.length = 0;
  respons.length = 0;
  await page.reload();
  await page.waitForSelector("#form-jenis-benih");
  await page.waitForTimeout(2000);

  return {
    loginStatus: login.status(),
    cookieSesi: cookies.map((c) => c.name),
    sebelumLogin,
    sesudahLogin: { catatan: [...catatan], respons: [...respons] },
  };
}
