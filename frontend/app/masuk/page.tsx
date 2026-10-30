import type { Metadata } from "next";

import { MasukKlien } from "../../components/auth/masuk-klien";
import type { RealmSSO } from "../../components/auth/masuk-url";

export const metadata: Metadata = {
  title: { absolute: "Masuk — SiTani DIY" },
  robots: { index: false, follow: false },
};

type Props = {
  searchParams: Promise<{ redirect?: string | string[]; realm?: string | string[] }>;
};

// Halaman login portal dengan tema SSO Pemda DIY. Satu-satunya tautan SSO yang
// dibuka dari header/panel adalah halaman ini; pengalihan nyata ke issuer
// dilakukan backend setelah tombol ditekan sehingga pengguna selalu melihat
// hasilnya.
export default async function MasukPage({ searchParams }: Props) {
  const params = await searchParams;
  const redirect = typeof params.redirect === "string" ? params.redirect : undefined;
  const realmParam = typeof params.realm === "string" ? params.realm : undefined;
  const realmAwal: RealmSSO | undefined =
    realmParam === "jsp" || realmParam === "satriya" ? realmParam : undefined;

  return (
    <main className="sso-halaman">
      <MasukKlien redirect={redirect} realmAwal={realmAwal} />
    </main>
  );
}
