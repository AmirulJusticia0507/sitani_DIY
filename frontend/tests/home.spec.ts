import { expect, test } from "@playwright/test";

test("menampilkan identitas dan modul utama", async ({ page }, testInfo) => {
  await page.goto("/");

  await expect(page).toHaveTitle("SiTani DIY — Portal Pertanian dan Ketahanan Pangan DIY");
  await expect(page.getByRole("heading", { level: 1 })).toContainText(
    "Portal Pertanian dan Ketahanan Pangan DIY",
  );
  await expect(page.getByRole("heading", { name: "Repository Data Lokal" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Registrasi Hortikultura" })).toBeVisible();
  // Modul layanan berada di area kerja; kartu mengarah ke login dengan tujuan
  // modulnya, bukan langsung membuka halaman modul.
  await expect(
    page.getByRole("link", { name: /Registrasi Hortikultura/ }).first(),
  ).toHaveAttribute("href", "/masuk?redirect=%2Fdashboard%2Fhorticulture");
  // Tautan masuk kini menuju halaman portal /masuk (tema SSO Pemda DIY), bukan
  // langsung ke endpoint backend; pengalihan ke issuer dilakukan di sana.
  await expect(page.getByRole("link", { name: "Masuk JSP" })).toHaveAttribute(
    "href",
    "/masuk?redirect=%2Fdashboard&realm=jsp",
  );
  await expect(page.getByRole("link", { name: "Masuk Satriya" })).toHaveAttribute(
    "href",
    "/masuk?redirect=%2Fdashboard&realm=satriya",
  );
  await expect(page.locator('link[rel="icon"]')).toHaveAttribute("href", "/favicon.svg");
  // Navigasi utama kini berada di sidebar, bukan header horizontal.
  const navigasi = page.getByRole("navigation", { name: "Navigasi utama" });
  if (testInfo.project.name === "mobile") {
    // Di layar sempit sidebar berupa drawer; buka lewat tombol Menu.
    await page.getByRole("button", { name: "Menu", exact: true }).click();
    await expect(page.locator("#sidebar-portal")).toBeVisible();
  }
  // Sidebar publik hanya memuat beranda dan pintu masuk area kerja; modul
  // layanan dibuka dari sidebar dashboard setelah login.
  await expect(navigasi.getByRole("link")).toHaveCount(2);
  await expect(navigasi.getByRole("link", { name: /Masuk area kerja/ })).toHaveAttribute(
    "href",
    "/masuk",
  );
  await expect(page.getByRole("contentinfo").getByRole("link", { name: "Tentang" })).toHaveAttribute("href", "/tentang");
  await expect(page.getByRole("contentinfo").getByRole("link", { name: "Kebijakan Privasi" })).toHaveAttribute("href", "/kebijakan-privasi");
  await expect(page.getByRole("contentinfo").getByRole("link", { name: "Bantuan" })).toHaveAttribute("href", "/bantuan");

  // Halaman publik tetap memakai kerangka portal (sidebar), bukan kerangka dashboard.
  await expect(page.locator(".dashboard-header")).toHaveCount(0);
  await expect(page.locator("#sidebar-portal")).toHaveCount(1);

  const manifest = await page.request.get("/manifest.webmanifest");
  expect(manifest.ok()).toBeTruthy();
  const manifestBody = await manifest.json();
  expect(manifestBody.name).toContain("SiTani DIY");
  expect(manifestBody.display).toBe("standalone");
  expect(manifestBody.icons).toHaveLength(3);

  await page.screenshot({
    path: `../.ui-check/${testInfo.project.name}.png`,
    fullPage: true,
  });
});

test("halaman informasi portal tersedia", async ({ page }) => {
  for (const [path, heading] of [
    ["/tentang", "Tentang SiTani DIY"],
    ["/kebijakan-privasi", "Kebijakan Privasi"],
    ["/bantuan", "Bantuan SiTani DIY"],
  ]) {
    await page.goto(path);
    await expect(page.getByRole("heading", { level: 1, name: heading })).toBeVisible();
  }
});

test("menampilkan modul registrasi dan penjualan benih", async ({ page }) => {
  await page.goto("/dashboard/benih");

  await expect(page.getByRole("heading", { level: 1 })).toContainText(
    "Registrasi & Penjualan Benih",
  );
  await expect(page.getByRole("heading", { name: "Registrasi Benih", exact: true })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Penjualan Benih", exact: true })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Laporan", exact: true })).toBeVisible();
  await expect(page.getByRole("button", { name: "Daftarkan Lot" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Catat Penjualan" })).toBeVisible();
  // Form registrasi memilih permohonan, tidak meminta ID permohonan internal.
  await expect(page.getByLabel("Permohonan")).toBeVisible();
  await expect(page.getByLabel("ID permohonan")).toHaveCount(0);
});
