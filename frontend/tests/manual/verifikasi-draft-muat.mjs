// Pemuatan ulang halaman saat draft tersimpan: isian terisi otomatis dari
// /v1/permohonan/aktif dan tombol Edit Draft tersedia tanpa menekan apa pun.
export default async function run(page, ui) {
  await page.request.post("http://localhost:8090/auth/dev/login", {
    data: { email: "uji@contoh.id", password: "rahasia-uji-123" },
    headers: { Origin: "http://localhost:3002" },
  });
  await page.goto("http://localhost:3002/certification");
  await page.waitForSelector("#form-jenis-benih");

  // Hapus permohonan lama lewat SQL sudah dilakukan seeder; buat draft baru.
  await page.selectOption("#form-jenis-benih", "Cabai");
  await page.fill("#varietas", "Lado");
  await page.fill("#luas-lahan", "0.25");
  await page.fill("#lokasi-kebun", "Kulon Progo");
  await page.getByRole("button", { name: "Simpan Draft", exact: true }).click();
  await page.waitForSelector("text=Draft tersimpan");

  await page.reload();
  await page.waitForSelector("text=DRAFT (BELUM DIAJUKAN)");
  return {
    jenisTersimpan: await page.inputValue("#form-jenis-benih"),
    varietasTersimpan: await page.inputValue("#varietas"),
    luasTersimpan: await page.inputValue("#luas-lahan"),
    lokasiTersimpan: await page.inputValue("#lokasi-kebun"),
    editDraftTersedia: await page.getByRole("button", { name: "Edit Draft", exact: true }).isVisible(),
    analisTerkunci: await page.locator(".stepper-mt-tombol").nth(1).getAttribute("aria-disabled"),
  };
}
