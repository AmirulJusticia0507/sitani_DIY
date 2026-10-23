"use client";

import { useEffect, useState } from "react";

type Principal = {
  email: string;
  name: string;
  role: string;
  realm: string;
};

const apiURL = process.env.NEXT_PUBLIC_BACKEND_URL ?? "http://localhost:8080";

export function AuthPanel() {
  const [principal, setPrincipal] = useState<Principal | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch(`${apiURL}/auth/me`, { credentials: "include" })
      .then((response) => (response.ok ? response.json() : null))
      .then((data) => setPrincipal(data))
      .catch(() => setPrincipal(null))
      .finally(() => setLoading(false));
  }, []);

  async function logout() {
    await fetch(`${apiURL}/auth/logout`, { method: "POST", credentials: "include" });
    setPrincipal(null);
  }

  if (loading) {
    return <p className="auth-note">Memeriksa sesi SSO…</p>;
  }

  if (principal) {
    return (
      <div className="auth-panel">
        <div>
          <strong>{principal.name || principal.email}</strong>
          <span>{principal.role} · realm {principal.realm}</span>
        </div>
        <button type="button" onClick={logout}>Keluar</button>
      </div>
    );
  }

  const redirect = encodeURIComponent("/");
  return (
    <div className="auth-panel">
      <div>
        <strong>Masuk melalui SSO Jogjaprov</strong>
        <span>Gunakan JSP untuk masyarakat atau Satriya untuk ASN.</span>
      </div>
      <div className="auth-actions">
        <a href={`${apiURL}/auth/login/jsp?redirect=${redirect}`}>Masuk dengan JSP</a>
        <a href={`${apiURL}/auth/login/satriya?redirect=${redirect}`}>Masuk dengan Satriya</a>
      </div>
    </div>
  );
}
