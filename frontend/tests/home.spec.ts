import { expect, test } from "@playwright/test";

test("menampilkan identitas dan modul utama", async ({ page }, testInfo) => {
  await page.goto("/");

  await expect(page).toHaveTitle("SiTani DIY — Portal Pertanian dan Ketahanan Pangan DIY");
  await expect(page.getByRole("heading", { level: 1 })).toContainText(
    "Portal Pertanian dan Ketahanan Pangan DIY",
  );
  await expect(page.getByRole("heading", { name: "Repository Data Lokal" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Registrasi Hortikultura" })).toBeVisible();
  await expect(page.getByRole("link", { name: "Masuk JSP" })).toHaveAttribute(
    "href",
    "http://localhost:8080/auth/login/jsp?redirect=%2F",
  );
  await expect(page.getByRole("link", { name: "Masuk Satriya" })).toHaveAttribute(
    "href",
    "http://localhost:8080/auth/login/satriya?redirect=%2F",
  );
  await expect(page.locator('link[rel="icon"]')).toHaveAttribute("href", "/favicon.svg");
  await expect(page.getByRole("navigation").getByRole("link")).toHaveCount(5);

  await page.screenshot({
    path: `../.ui-check/${testInfo.project.name}.png`,
    fullPage: true,
  });
});

test("menampilkan modul registrasi dan penjualan benih", async ({ page }) => {
  await page.goto("/benih");

  await expect(page.getByRole("heading", { level: 1 })).toContainText(
    "Registrasi & Penjualan Benih",
  );
  await expect(page.getByRole("heading", { name: "Registrasi Benih", exact: true })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Penjualan Benih", exact: true })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Laporan", exact: true })).toBeVisible();
  await expect(page.getByRole("button", { name: "Daftarkan Lot" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Catat Penjualan" })).toBeVisible();
});
