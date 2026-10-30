import { expect, test } from "@playwright/test";

// Login lokal development (/dashboard/masuk) hanya muncul bila backend
// mengaktifkannya. Tes memock /auth/dev/status dan /auth/dev/login agar tidak
// bergantung pada backend berjalan maupun database berisi akun seed.
test.describe("login lokal development", () => {
  test("formulir tampil bila backend mengaktifkan login lokal", async ({ page }) => {
    let payload: { email?: string; password?: string } | null = null;
    await page.route("**/auth/dev/status", (route) =>
      route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ aktif: true }) }),
    );
    await page.route("**/auth/dev/login", async (route) => {
      payload = route.request().postDataJSON() as { email?: string; password?: string };
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ role: "admin" }) });
    });
    await page.route("**/auth/me", (route) => route.fulfill({ status: 401, body: "" }));

    await page.goto("/dashboard/masuk");
    const formulir = page.locator(".masuk-lokal-form");
    await expect(formulir).toBeVisible();
    await expect(page.getByText("Lingkungan development")).toBeVisible();

    await page.getByLabel("Email").fill("admin@sitani.local");
    await page.getByLabel("Password").fill("rahasia-uji");
    await formulir.getByRole("button", { name: "Masuk", exact: true }).click();

    await expect.poll(() => payload).toEqual({ email: "admin@sitani.local", password: "rahasia-uji" });
    // Setelah berhasil, halaman berpindah ke dashboard.
    await page.waitForURL("**/dashboard");
  });

  test("formulir tidak ditawarkan bila login lokal nonaktif", async ({ page }) => {
    await page.route("**/auth/dev/status", (route) =>
      route.fulfill({ status: 404, body: "" }),
    );
    await page.route("**/auth/me", (route) => route.fulfill({ status: 503, body: "" }));

    await page.goto("/dashboard/masuk");
    await expect(page.getByRole("heading", { name: "Login lokal tidak tersedia" })).toBeVisible();
    await expect(page.locator(".masuk-lokal-form")).toHaveCount(0);
    await expect(page.getByRole("main").getByRole("link", { name: "Masuk JSP" })).toBeVisible();

    // Dashboard tidak menautkan halaman login lokal.
    await page.goto("/dashboard");
    await expect(page.getByRole("link", { name: "Masuk lokal" })).toHaveCount(0);
  });

  test("pesan gagal login ditampilkan tanpa berpindah halaman", async ({ page }) => {
    await page.route("**/auth/dev/status", (route) =>
      route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ aktif: true }) }),
    );
    await page.route("**/auth/dev/login", (route) =>
      route.fulfill({
        status: 401,
        contentType: "application/json",
        body: JSON.stringify({ error: "email atau password salah" }),
      }),
    );
    await page.route("**/auth/me", (route) => route.fulfill({ status: 401, body: "" }));

    await page.goto("/dashboard/masuk");
    await page.getByLabel("Email").fill("salah@sitani.local");
    await page.getByLabel("Password").fill("salah");
    await page.locator(".masuk-lokal-form").getByRole("button", { name: "Masuk", exact: true }).click();

    await expect(page.locator(".masuk-lokal-pesan")).toHaveText("email atau password salah");
    await expect(page).toHaveURL(/\/dashboard\/masuk$/);
  });
});
