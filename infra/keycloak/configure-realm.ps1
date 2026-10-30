<#
.SYNOPSIS
    Membuat realm SSO lokal SiTani DIY di Keycloak development secara idempoten.

.DESCRIPTION
    Membuat (atau memperbarui) di Keycloak lokal:
      - realm `sitani` (HTTP diizinkan karena hanya untuk pengembangan)
      - client `sitani-portal` (confidential, PKCE S256 + client secret)
      - 5 akun uji @sitani.local dengan peran berbeda

    Skrip ini TIDAK menyentuh realm Jogjaprov atau Keycloak proyek lain. Realm
    produksi (JSP/Satriya) tetap milik pengelola SSO Jogjaprov.

    Kredensial dibaca dari parameter/environment, tidak pernah ditulis ke disk:
      KEYCLOAK_ADMIN_PASSWORD  (wajib) password administrator Keycloak lokal
      KEYCLOAK_ADMIN           (default admin)

.PARAMETER KeycloakUrl
    Base URL Keycloak lokal, default http://localhost:8085

.PARAMETER ClientSecret
    Client secret untuk `sitani-portal`. Bila kosong, dihasilkan acak sekali
    lalu dicetak agar dapat disalin ke .env backend.

.PARAMETER SeedPassword
    Password akun uji; default dari SEED_DEV_PASSWORD. Minimal 12 karakter.

.PARAMETER IssuerHosts
    Nama host pada redirect_uri, default localhost dan host.docker.internal.

.PARAMETER CallbackPorts
    Port backend pada redirect_uri, default 8080 (Compose) dan 8081 (dev host).

.EXAMPLE
    $env:KEYCLOAK_ADMIN_PASSWORD = 'password-admin-lokal'
    $env:SEED_DEV_PASSWORD = 'password-dev-lokal-123'
    ./infra/keycloak/configure-realm.ps1
#>
[CmdletBinding()]
param(
    [string]$KeycloakUrl = 'http://localhost:8085',
    [string]$Realm = 'sitani',
    [string]$ClientId = 'sitani-portal',
    [string]$ClientSecret = '',
    [string]$SeedPassword = '',
    # Nama host issuer dapat berbeda antara proses host (localhost) dan
    # container (host.docker.internal); keduanya didaftarkan.
    [string[]]$IssuerHosts = @('localhost', 'host.docker.internal'),
    [string[]]$CallbackPorts = @('8080', '8081')
)

$ErrorActionPreference = 'Stop'

$adminUser = if ($env:KEYCLOAK_ADMIN) { $env:KEYCLOAK_ADMIN } else { 'admin' }
$adminPassword = $env:KEYCLOAK_ADMIN_PASSWORD
if (-not $adminPassword) {
    throw 'KEYCLOAK_ADMIN_PASSWORD belum diisi. Ekspor environment atau jalankan dengan -KeycloakAdminPassword lewat variabel environment.'
}
if (-not $SeedPassword) { $SeedPassword = $env:SEED_DEV_PASSWORD }
if ($SeedPassword.Length -lt 12) {
    throw 'SEED_DEV_PASSWORD wajib diisi minimal 12 karakter (dipakai untuk akun uji).'
}

# Client secret acak bila tidak diberikan. Base64URL tanpa padding.
# RandomNumberGenerator.Create() dipakai agar kompatibel dengan PowerShell 5.1.
if (-not $ClientSecret) {
    $bytes = New-Object 'byte[]' 32
    $rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
    try { $rng.GetBytes($bytes) } finally { $rng.Dispose() }
    $ClientSecret = [Convert]::ToBase64String($bytes).Replace('+', '-').Replace('/', '_').TrimEnd('=')
}

$base = $KeycloakUrl.TrimEnd('/')

# Daftar redirect_uri dibentuk dari semua kombinasi host/port agar issuer boleh
# diakses dari host maupun container docker.
$RedirectUris = @(
    foreach ($h in $IssuerHosts) {
        foreach ($p in $CallbackPorts) {
            "http://${h}:${p}/auth/callback/jsp"
        }
    }
)
$adminBase = "$base/admin/realms"

function Invoke-Kc {
    param(
        [Parameter(Mandatory)][string]$Method,
        [Parameter(Mandatory)][string]$Uri,
        $Body,
        [string[]]$OkCodes = @(200, 201, 204)
    )
    $params = @{
        Method      = $Method
        Uri         = $Uri
        Headers     = @{ Authorization = "Bearer $script:token" }
        ContentType = 'application/json'
        ErrorAction = 'Stop'
    }
    if ($null -ne $Body) { $params.Body = ($Body | ConvertTo-Json -Depth 12 -Compress) }
    try {
        return Invoke-RestMethod @params
    } catch {
        $status = [int]$_.Exception.Response.StatusCode
        if ($OkCodes -contains $status) { return $null }
        # Body kesalahan Keycloak membantu diagnosis; tidak berisi kredensial.
        $detail = ''
        if ($_.ErrorDetails -and $_.ErrorDetails.Message) { $detail = $_.ErrorDetails.Message }
        throw "Keycloak $Method $Uri gagal (HTTP $status): $detail"
    }
}

Write-Host '1/5 Autentikasi administrator Keycloak...' -ForegroundColor Cyan
$tokenBody = @{
    grant_type = 'password'
    client_id  = 'admin-cli'
    username   = $adminUser
    password   = $adminPassword
}
# Token dikirim lewat body (bukan URL) agar tidak tercatat di log akses.
$tokenResponse = Invoke-RestMethod -Method Post -Uri "$base/realms/master/protocol/openid-connect/token" `
    -ContentType 'application/x-www-form-urlencoded' -Body $tokenBody -ErrorAction Stop
$script:token = $tokenResponse.access_token
if (-not $script:token) { throw 'Token administrator tidak diterima Keycloak.' }

Write-Host "2/5 Pastikan realm '$Realm' ada..." -ForegroundColor Cyan
$realmBody = @{
    realm                  = $Realm
    displayName            = 'SiTani DIY (lokal)'
    enabled                = $true
    # Hanya untuk pengembangan: Keycloak lokal berjalan di HTTP.
    sslRequired            = 'none'
    registrationAllowed    = $false
    loginWithEmailAllowed  = $true
    duplicateEmailsAllowed = $false
    resetPasswordAllowed   = $true
    editUsernameAllowed    = $false
    accessTokenLifespan    = 900
    ssoSessionIdleTimeout  = 28800
    ssoSessionMaxLifespan  = 28800
}
$null = Invoke-Kc -Method Post -Uri $adminBase -Body $realmBody -OkCodes @(201, 409)

Write-Host "3/5 Pastikan client '$ClientId' ada..." -ForegroundColor Cyan
$existingClients = Invoke-Kc -Method Get -Uri ("{0}/{1}/clients?clientId={2}" -f $adminBase, $Realm, $ClientId)
$clientBody = @{
    clientId                  = $ClientId
    name                      = 'Portal SiTani DIY (lokal)'
    enabled                   = $true
    protocol                  = 'openid-connect'
    publicClient              = $false
    clientAuthenticatorType   = 'client-secret'
    secret                    = $ClientSecret
    standardFlowEnabled       = $true
    implicitFlowEnabled       = $false
    directAccessGrantsEnabled = $false
    serviceAccountsEnabled    = $false
    rootUrl                   = 'http://localhost:3000'
    baseUrl                   = '/'
    redirectUris              = $RedirectUris
    webOrigins                = @('http://localhost:3000', 'http://localhost:3001', 'http://localhost:3002')
    attributes                = @{
        'pkce.code.challenge.method' = 'S256'
        'post.logout.redirect.uris'  = 'http://localhost:3000/*##http://localhost:3001/*##http://localhost:3002/*'
        'access.token.lifespan'      = '900'
    }
}
if ($existingClients -and $existingClients.Count -gt 0) {
    $clientUuid = $existingClients[0].id
    # Body PUT tidak menyertakan `secret` yang sudah ada; tanpa langkah ini
    # secret baru tidak tersimpan sehingga pertukaran token selalu gagal.
    $clientBody.secret = $ClientSecret
    $null = Invoke-Kc -Method Put -Uri ("{0}/{1}/clients/{2}" -f $adminBase, $Realm, $clientUuid) -Body $clientBody
    Write-Host "   client sudah ada, diperbarui (id $clientUuid)." -ForegroundColor DarkGray
} else {
    $null = Invoke-Kc -Method Post -Uri ("{0}/{1}/clients" -f $adminBase, $Realm) -Body $clientBody -OkCodes @(201)
    $clientUuid = (Invoke-Kc -Method Get -Uri ("{0}/{1}/clients?clientId={2}" -f $adminBase, $Realm, $ClientId))[0].id
    Write-Host "   client dibuat (id $clientUuid)." -ForegroundColor DarkGray
}

Write-Host '4/5 Pastikan akun uji tersedia...' -ForegroundColor Cyan
# Email memakai domain lokal sitani.local agar tidak menyerupai alamat dinas.
# Peran bisnis (admin, pengawas, ...) TIDAK diambil dari claim token: backend
# memetakan identitas SSO ke tabel users lokal berdasarkan email.
$akunUji = @(
    @{ username = 'admin.dev';        email = 'admin@sitani.local';        nama = 'Admin Development' },
    @{ username = 'pengawas.dev';     email = 'pengawas@sitani.local';     nama = 'Pengawas Development' },
    @{ username = 'analis.dev';       email = 'analis@sitani.local';       nama = 'Analis Development' },
    @{ username = 'penandatangan.dev'; email = 'penandatangan@sitani.local'; nama = 'Penandatangan Development' },
    @{ username = 'pelanggan.dev';    email = 'pelanggan@sitani.local';    nama = 'Pelanggan Development' }
)
foreach ($akun in $akunUji) {
    $ada = Invoke-Kc -Method Get -Uri ("{0}/{1}/users?exact=true&email={2}" -f $adminBase, $Realm, [uri]::EscapeDataString($akun.email))
    $data = @{
        username      = $akun.username
        email         = $akun.email
        firstName     = ($akun.nama -split ' ')[0]
        lastName      = ($akun.nama -split ' ')[-1]
        emailVerified = $true
        enabled       = $true
        attributes    = @{ unit_kerja = @('Dev') }
    }
    if ($ada -and $ada.Count -gt 0) {
        $userId = $ada[0].id
        $null = Invoke-Kc -Method Put -Uri ("{0}/{1}/users/{2}" -f $adminBase, $Realm, $userId) -Body $data
    } else {
        $null = Invoke-Kc -Method Post -Uri ("{0}/{1}/users" -f $adminBase, $Realm) -Body $data -OkCodes @(201)
        $userId = (Invoke-Kc -Method Get -Uri ("{0}/{1}/users?exact=true&email={2}" -f $adminBase, $Realm, [uri]::EscapeDataString($akun.email)))[0].id
    }
    # Password selalu diset ulang agar skrip idempoten; temporary=false agar
    # tidak wajib ganti password saat login UAT.
    $null = Invoke-Kc -Method Put -Uri ("{0}/{1}/users/{2}/reset-password" -f $adminBase, $Realm, $userId) `
        -Body @{ type = 'password'; value = $SeedPassword; temporary = $false }
    Write-Host "   akun siap: $($akun.email)" -ForegroundColor DarkGray
}

Write-Host '5/5 Verifikasi openid-configuration dan secret client...' -ForegroundColor Cyan
$tersimpan = Invoke-Kc -Method Get -Uri ("{0}/{1}/clients/{2}/client-secret" -f $adminBase, $Realm, $clientUuid)
if (-not $tersimpan -or $tersimpan.value -ne $ClientSecret) {
    throw 'Secret client di Keycloak tidak sama dengan yang dipakai skrip. Hentikan sebelum konfigurasi .env disalin.'
}
Write-Host '   secret client cocok.' -ForegroundColor DarkGray
$discovery = Invoke-RestMethod -Method Get -Uri "$base/realms/$Realm/.well-known/openid-configuration" -ErrorAction Stop
$issuerCotak = 'http://localhost:8085/realms/sitani'
Write-Host "   issuer = $($discovery.issuer)"

Write-Host ''
Write-Host 'Realm siap. Tambahkan ke .env backend (jangan di-commit):' -ForegroundColor Green
Write-Host "  SSO_JSP_ISSUER=$($discovery.issuer)"
Write-Host "  SSO_JSP_CLIENT_ID=$ClientId"
Write-Host "  SSO_JSP_CLIENT_SECRET=<lihat di bawah>"
Write-Host ''
Write-Host 'Client secret sitani-portal:' -ForegroundColor Yellow
Write-Host "  $ClientSecret"
Write-Host ''
Write-Host 'Akun uji (password dari SEED_DEV_PASSWORD):' -ForegroundColor Green
$akunUji | ForEach-Object { Write-Host "  $($_.email)" }
Write-Host ''
Write-Host 'Catatan: realm ini HANYA untuk development/UAT. Jangan dipakai sebagai identitas produksi.'
