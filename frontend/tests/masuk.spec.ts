import { expect, test } from "@playwright/test";

// Halaman login portal (/masuk) memakai tema SSO Pemda DIY. Tes ini memisahkan
// perilaku backend lewat emulasi: klik tombol harus SELALU memberi hasil (pesan
// gagal) bila backend mati, dan harus mengikuti redirect saat backend menjawab
// 302 ke issuer SSO. Tidak ada kredensial atau issuer sungguhan yang dipakai.
const SSO_ISSUER = "https://sso.example.invalid/realms/jsp/protocol/openid-connect/auth";

test.describe("halaman masuk SSO", () => {
  test("header portal menuju halaman /masuk", async ({ page }, testInfo) => {
    await page.route("**/auth/me", (route) => route.fulfill({ status: 401, body: "" }));
    await page.goto("/");

    if (testInfo.project.name === "mobile") {
      // Tautan auth berada di sidebar drawer; buka menu terlebih dahulu.
      await page.getByRole("button", { name: "Menu", exact: true }).click();
      await expect(page.locator("#sidebar-portal")).toBeVisible();
    }
    const jsp = page.getByRole("link", { name: "Masuk JSP" });
    await expect(jsp).toHaveAttribute("href", "/masuk?realm=jsp");
    const satriya = page.getByRole("link", { name: "Masuk Satriya" });
    await expect(satriya).toHaveAttribute("href", "/masuk?realm=satriya");

    await jsp.click();
    await expect(page).toHaveURL(/\/masuk\?realm=jsp$/);
    await expect(page.getByRole("heading", { level: 1, name: "Single Sign On" })).toBeVisible();
  });

  test("tombol masuk memberi pesan bila backend tidak dapat dihubungi", async ({ page }) => {
    await page.route("**/auth/me", (route) => route.abort("connectionrefused"));
    await page.route("**/auth/dev/status", (route) => route.abort("connectionrefused"));
    await page.route("**/auth/login/**", (route) => route.abort("connectionrefused"));

    await page.goto("/masuk?realm=satriya");
    const tombol = page.getByRole("button", { name: /Masuk dengan Satriya/ });
    await expect(tombol).toBeEnabled();
    await tombol.click();

    await expect(page.locator(".sso-pesan-galat")).toContainText("tidak dapat dihubungi");
    // Tetap di halaman login sehingga pengguna dapat mencoba lagi.
    await expect(page).toHaveURL(/\/masuk\?realm=satriya$/);
  });

  test("tombol masuk menampilkan error backend tanpa berpindah halaman kosong", async ({ page }) => {
    await page.route("**/auth/me", (route) =>
      route.fulfill({ status: 503, contentType: "application/json", body: JSON.stringify({ error: "konfigurasi SSO JSP/Satriya belum tersedia" }) }),
    );
    await page.route("**/auth/dev/status", (route) => route.fulfill({ status: 404, body: "" }));
    await page.route("**/auth/login/jsp**", (route) =>
      route.fulfill({
        status: 503,
        contentType: "application/json",
        body: JSON.stringify({ error: "konfigurasi SSO JSP/Satriya belum tersedia" }),
      }),
    );

    await page.goto("/masuk");
    await expect(page.getByRole("status")).toContainText("SSO Jogjaprov belum aktif");
    await page.getByRole("button", { name: /Masuk dengan JSP/ }).click();
    await expect(page.locator(".sso-pesan-galat")).toContainText("konfigurasi SSO JSP/Satriya belum tersedia");
  });

  test("klik tidak pernah diam saat redirect tidak terekspos", async ({ page }) => {
    const loginURLs: string[] = [];
    // Backend membalas teks kosong tanpa pesan JSON (fetch tidak melihat redirect):
    // halaman harus berpindah ke endpoint backend, yang lalu mengalihkan ke issuer.
    await page.route("**/auth/me", (route) => route.fulfill({ status: 401, body: "" }));
    await page.route("**/auth/dev/status", (route) => route.fulfill({ status: 404, body: "" }));
    await page.route("**/auth/login/satriya**", (route) => {
      loginURLs.push(route.request().url());
      return route.fulfill({ status: 200, contentType: "text/html", body: "" });
    });

    await page.goto("/masuk?redirect=%2Fdashboard%2Fprofil");
    await page.getByLabel("Satriya").check();
    await page.getByRole("button", { name: /Masuk dengan Satriya/ }).click();

    await expect.poll(() => loginURLs.length).toBeGreaterThan(1);
    expect(loginURLs[0]).toContain("/auth/login/satriya?redirect=");
    expect(decodeURIComponent(loginURLs[0])).toContain("redirect=/dashboard/profil");
    await expect.poll(() => new URL(page.url()).pathname, { timeout: 10000 }).toContain(
      "/auth/login/satriya",
    );
  });

  test("redirect eksternal diabaikan", async ({ page }) => {
    await page.route("**/auth/me", (route) => route.fulfill({ status: 401, body: "" }));
    await page.route("**/auth/dev/status", (route) => route.fulfill({ status: 404, body: "" }));

    // URLSearchParams menormalkan "//evil..." menjadi path internal "/evil...",
    // sehingga tautan kembali tidak pernah mengarah keluar portal.
    await page.goto("/masuk?redirect=//evil.example.invalid");
    await expect(page.locator(".sso-kembali a")).toHaveAttribute("href", "/");
  });
});
