import { expect, test } from "@playwright/test";

test("menampilkan identitas dan modul utama", async ({ page }, testInfo) => {
  await page.goto("/");

  await expect(page).toHaveTitle("SiTani DIY");
  await expect(page.getByRole("heading", { level: 1 })).toContainText(
    "Portal Pertanian dan Ketahanan Pangan DIY",
  );
  await expect(page.getByRole("heading", { name: "Repository Data Lokal" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Registrasi Hortikultura" })).toBeVisible();

  await page.screenshot({
    path: `../.ui-check/${testInfo.project.name}.png`,
    fullPage: true,
  });
});
