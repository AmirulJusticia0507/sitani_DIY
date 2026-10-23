import { expect, test, type Page } from "@playwright/test";

// Dashboard memakai kerangka sendiri (header area kerja, sidebar, footer
// ringkas) yang terpisah dari header/footer portal publik. Sesi dimock agar
// tes tidak bergantung pada backend SSO yang belum aktif di lingkungan uji.
const SESI_ADMIN = {
  user_id: 12,
  email: "admin@diy.go.id",
  name: "Admin DPKP",
  username: "admin.dpkp",
  role: "admin",
  realm: "satriya",
};

// Pada layar kecil sidebar tersembunyi sampai menu dibuka; buka dulu agar
// navigasi dapat diperiksa. Pada layar lebar fungsi ini tidak melakukan apa pun.
async function bukaSidebarJikaTersembunyi(page: Page) {
  const sidebar = page.locator("#sidebar-dashboard");
  if (await sidebar.isVisible()) return;
  await page.getByRole("button", { name: "Buka menu dashboard" }).click();
  await expect(sidebar).toBeVisible();
}

test.describe("dashboard tanpa sesi", () => {
  test("menampilkan ajakan masuk dan kerangka dashboard", async ({ page }) => {
    await page.route("**/auth/me", (route) => route.fulfill({ status: 401, body: "" }));
    await page.goto("/dashboard");

    await expect(page).toHaveTitle("Dashboard — SiTani DIY");
    await expect(page.getByRole("heading", { level: 1, name: "Dashboard SiTani DIY" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Masuk untuk membuka area kerja" })).toBeVisible();

    // Kerangka dashboard menggantikan header/footer portal publik.
    await expect(page.locator(".sitani-header")).toHaveCount(0);
    await expect(page.locator(".dashboard-header")).toHaveCount(1);
    await expect(page.locator(".dashboard-footer")).toHaveCount(1);
    await expect(page.getByRole("contentinfo").getByRole("link", { name: "Bantuan" })).toHaveAttribute(
      "href",
      "/bantuan",
    );
  });
});

test.describe("dashboard dengan sesi admin", () => {
  test.beforeEach(async ({ page }) => {
    await page.route("**/auth/me", (route) =>
      route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify(SESI_ADMIN) }),
    );
  });

  test("menampilkan seluruh modul dan identitas sesi", async ({ page }) => {
    await page.goto("/dashboard");
    await expect(page.getByRole("heading", { name: "Admin DPKP" })).toBeVisible();
    await bukaSidebarJikaTersembunyi(page);

    const nav = page.getByRole("navigation", { name: "Modul dashboard" });
    await expect(nav.getByRole("link")).toHaveCount(11);
    await expect(nav.getByRole("link", { name: "Ringkasan" })).toHaveAttribute("aria-current", "page");
    await expect(page.locator(".dashboard-tile").first()).toBeVisible();
    await expect(page.getByText("Satriya (ASN)", { exact: true })).toBeVisible();

    await page.goto("/dashboard/profil");
    await expect(page.getByRole("heading", { level: 1, name: "Profil & Akses" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Modul yang dapat dibuka" })).toBeVisible();
    await bukaSidebarJikaTersembunyi(page);
    await expect(
      page.getByRole("navigation", { name: "Modul dashboard" }).getByRole("link", { name: "Profil & Akses" }),
    ).toHaveAttribute("aria-current", "page");
  });
});

test.describe("dashboard dengan peran terbatas", () => {
  test("pelanggan hanya melihat modul yang diizinkan", async ({ page }) => {
    await page.route("**/auth/me", (route) =>
      route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ ...SESI_ADMIN, name: "Siti Pelanggan", role: "pelanggan", realm: "jsp" }),
      }),
    );
    await page.goto("/dashboard");
    await expect(page.getByRole("heading", { name: "Siti Pelanggan" })).toBeVisible();
    await bukaSidebarJikaTersembunyi(page);

    const nav = page.getByRole("navigation", { name: "Modul dashboard" });
    await expect(nav.getByRole("link")).toHaveCount(5);
    await expect(nav.getByRole("link", { name: "Repository Data" })).toHaveCount(0);
    await expect(nav.getByRole("link", { name: "Penjualan Benih" })).toHaveCount(0);
  });

  test("memperingatkan peran yang belum dipetakan", async ({ page }) => {
    await page.route("**/auth/me", (route) =>
      route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ ...SESI_ADMIN, role: "operator_baru" }),
      }),
    );
    await page.goto("/dashboard");
    await expect(page.locator(".dashboard-callout")).toBeVisible();
    await bukaSidebarJikaTersembunyi(page);

    await expect(
      page.getByRole("navigation", { name: "Modul dashboard" }).getByRole("link"),
    ).toHaveCount(2);
  });
});
