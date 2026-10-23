import { expect, test } from "@playwright/test";

test("menampilkan identitas dan modul utama", async ({ page }, testInfo) => {
  await page.goto("/");

  await expect(page).toHaveTitle("SiTani DIY");
  await expect(page.getByRole("heading", { level: 1 })).toContainText(
    "Portal Pertanian dan Ketahanan Pangan DIY",
  );
  await expect(page.getByRole("heading", { name: "Repository Data Lokal" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Registrasi Hortikultura" })).toBeVisible();
  await expect(page.getByRole("link", { name: "Masuk dengan JSP" })).toHaveAttribute(
    "href",
    "http://localhost:8080/auth/login/jsp?redirect=%2F",
  );
  await expect(page.getByRole("link", { name: "Masuk dengan Satriya" })).toHaveAttribute(
    "href",
    "http://localhost:8080/auth/login/satriya?redirect=%2F",
  );

  await page.screenshot({
    path: `../.ui-check/${testInfo.project.name}.png`,
    fullPage: true,
  });
});
