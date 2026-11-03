// Verifikasi alur draft di halaman /certification.
//
//  1. Tombol Simpan Draft dan Ajukan Permohonan tersedia; stepper Analis Lab /
//     Penandatangan terkunci sebelum permohonan diajukan.
//  2. Simpan Draft menyimpan walau isian belum lengkap, memunculkan Edit Draft,
//     dan belum memberi nomor SPB.
//  3. Edit Draft mempertahankan isian tersimpan.
//  4. Ajukan Permohonan memberi nomor SPB dan membuka langkah berikutnya.
//  5. Stepper dapat berpindah ke Analis Lab, dan form pelanggan disembunyikan.
export default async function run(page, ui) {
  const hasil = { langkah: [] };

  // Masuk lewat dev-login agar permintaan dikenali sebagai sesi pelanggan.
  await page.request.post("http://localhost:8090/auth/dev/login", {
    data: { email: "uji@contoh.id", password: "rahasia-uji-123" },
    headers: { Origin: "http://localhost:3002" },
  });
  await page.goto("http://localhost:3002/certification");

  const tombolStepper = () => page.locator(".stepper-mt-tombol");
  const analis = tombolStepper().nth(1);
  const tombol = (nama) => page.getByRole("button", { name: nama, exact: true });

  // 1. Keadaan awal.
  await page.waitForSelector("#form-jenis-benih");
  hasil.simpanDraftTersedia = await tombol("Simpan Draft").isVisible();
  hasil.analisTerkunciAwal = await analis.getAttribute("aria-disabled");
  hasil.nextTerkunciAwal = await tombol("Next").isDisabled();
  hasil.catatanKunci = await page.getByRole("note").innerText();

  // 2. Simpan draft dengan isian belum lengkap (hanya jenis benih).
  await page.selectOption("#form-jenis-benih", "Tomat");
  await tombol("Simpan Draft").click();
  await page.waitForSelector("text=Draft tersimpan");
  hasil.badgeSetelahDraft = await page.locator(".badge-warning").first().innerText();
  hasil.editDraftTersedia = await tombol("Edit Draft").isVisible();
  hasil.analisTerkunciSetelahDraft = await analis.getAttribute("aria-disabled");
  hasil.formBenihNilai = await page.inputValue("#form-jenis-benih");

  // 3. Edit Draft: isian tersimpan tetap ada dan dapat dilengkapi.
  await tombol("Edit Draft").click();
  await page.waitForSelector("text=Mode edit draft");
  hasil.isianDraftSetelahEdit = await page.inputValue("#form-jenis-benih");
  await page.fill("#varietas", "Rama");
  await page.fill("#luas-lahan", "0.5");
  await page.fill("#lokasi-kebun", "Sleman");

  // 4. Ajukan.
  await tombol("Ajukan Permohonan").click();
  await page.waitForSelector("text=Permohonan diajukan dengan nomor");
  const pesan = (await page.locator(".alert-success").last().innerText()).trim();
  hasil.pesanAjukan = pesan;
  hasil.nomor = pesan.match(/SPB-\d{8}-\d{4,}/)?.[0] ?? null;
  hasil.badgeSetelahAjukan = await page.locator(".badge-success").first().innerText();
  hasil.analisTerkunciSetelahAjukan = await analis.getAttribute("aria-disabled");
  hasil.nextAktifSetelahAjukan = await tombol("Next").isEnabled();
  hasil.isianDikunci = await page.locator("#form-jenis-benih").isDisabled();
  hasil.catatanKunciHilang = (await page.getByRole("note").count()) === 0;

  // 5. Stepper berpindah ke Analis Lab; form hanya untuk pelanggan.
  await analis.click();
  await page.waitForTimeout(300);
  hasil.langkahAktif = await page.locator(".stepper-mt-tombol.is-aktif .stepper-mt-label-judul").innerText();
  hasil.formPelangganDisembunyikan = (await page.locator("#form-jenis-benih").count()) === 0;
  hasil.teksFase23BukanFormPelanggan = await page.getByText("bukan form pelanggan", { exact: false }).count();

  return hasil;
}
