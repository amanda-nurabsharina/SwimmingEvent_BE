const express = require("express");
const cors = require("cors");
const qrcode = require("qrcode");
const path = require("path");
const fs = require("fs");
const pino = require("pino");

const {
  default: makeWASocket,
  useMultiFileAuthState,
  DisconnectReason,
  fetchLatestBaileysVersion,
  makeCacheableSignalKeyStore,
  Browsers,
} = require("@whiskeysockets/baileys");

const app = express();
const PORT = process.env.PORT || 5001;

app.use(cors());
app.use(express.json());

// Distinguish Local vs Server environment
const IS_PRODUCTION =
  process.env.NODE_ENV === "production" ||
  process.env.APP_ENV === "production" ||
  process.env.IS_SERVER === "true";

// Use distinct session folder for Local vs Server to prevent session collision
const SESSION_NAME =
  process.env.WA_SESSION_NAME ||
  (IS_PRODUCTION ? "auth_info_baileys_server" : "auth_info_baileys_local");

const AUTH_DIR = path.join(__dirname, SESSION_NAME);
const logger = pino({ level: process.env.DEBUG_WA ? "info" : "silent" });

let sock = null;
let currentQR = null;
let currentQRDataUrl = null;
let isConnected = false;
let connectedPhone = null;
let isConnecting = false;
let pairingCode = null;

// Helper: Format phone to WhatsApp international 62 format
function formatWhatsAppNumber(phone) {
  if (!phone) return "";
  let clean = phone.replace(/\D/g, "");
  if (clean.startsWith("0")) {
    clean = "62" + clean.slice(1);
  } else if (clean.startsWith("8")) {
    clean = "62" + clean;
  }
  return clean;
}

// Start Baileys connection
async function startWhatsAppSocket() {
  if (isConnecting) return;
  isConnecting = true;

  try {
    console.log(`[WA Gateway] Initializing socket (${IS_PRODUCTION ? "SERVER" : "LOCAL"}). Session: ${SESSION_NAME}`);
    const { state, saveCreds } = await useMultiFileAuthState(AUTH_DIR);

    let version = [2, 3000, 1015901307];
    try {
      const v = await fetchLatestBaileysVersion();
      if (v?.version) version = v.version;
    } catch {
      // Fallback version if offline or github rate-limited
    }

    sock = makeWASocket({
      version,
      logger,
      printQRInTerminal: false,
      auth: {
        creds: state.creds,
        keys: makeCacheableSignalKeyStore(state.keys, logger),
      },
      browser: Browsers.ubuntu("Chrome"),
      generateHighQualityLinkPreview: true,
      syncFullHistory: false,
      connectTimeoutMs: 60000,
      defaultQueryTimeoutMs: 60000,
      keepAliveIntervalMs: 25000,
    });

    sock.ev.on("creds.update", saveCreds);

    sock.ev.on("connection.update", async (update) => {
      const { connection, lastDisconnect, qr } = update;

      if (qr) {
        currentQR = qr;
        try {
          currentQRDataUrl = await qrcode.toDataURL(qr, { margin: 2, scale: 6 });
          console.log("[WA Gateway] New QR code generated successfully");
        } catch (err) {
          console.error("[WA Gateway] Failed to generate QR data URL:", err);
        }
      }

      if (connection === "close") {
        isConnected = false;
        connectedPhone = null;
        isConnecting = false;
        currentQR = null;
        currentQRDataUrl = null;

        const statusCode = lastDisconnect?.error?.output?.statusCode;
        console.log(`[WA Gateway] Connection closed (${statusCode || "unknown"}). Reason: ${lastDisconnect?.error?.message || "none"}`);

        if (statusCode === DisconnectReason.loggedOut || statusCode === 401) {
          console.log(`[WA Gateway] Logged out / session expired. Resetting session directory ${SESSION_NAME}...`);
          try {
            fs.rmSync(AUTH_DIR, { recursive: true, force: true });
          } catch (e) {}
          // Immediately restart socket so a fresh QR code is available for scanning!
          setTimeout(() => startWhatsAppSocket(), 2000);
        } else {
          // Reconnect for other reasons (timeout, network blink, restartRequired)
          const delay = statusCode === DisconnectReason.restartRequired ? 1000 : 3000;
          setTimeout(() => startWhatsAppSocket(), delay);
        }
      } else if (connection === "open") {
        isConnected = true;
        isConnecting = false;
        currentQR = null;
        currentQRDataUrl = null;
        pairingCode = null;

        const userJid = sock.user?.id || "";
        const rawPhone = userJid.split(":")[0] || userJid.split("@")[0];
        connectedPhone = rawPhone;

        console.log(`[WA Gateway] Connected successfully to WhatsApp! Admin Phone: +${connectedPhone}`);
      }
    });
  } catch (error) {
    console.error("[WA Gateway] Error initializing socket:", error);
    isConnecting = false;
    setTimeout(() => startWhatsAppSocket(), 5000);
  }
}

// --------------------------------------------------------------------------
// REST API ENDPOINTS
// --------------------------------------------------------------------------

// 1. GET STATUS & QR
app.get("/api/wa/status", (req, res) => {
  // If not connected, not currently connecting, and no QR exists, wake up socket
  if (!isConnected && !isConnecting && !currentQRDataUrl) {
    startWhatsAppSocket();
  }

  res.json({
    success: true,
    isConnected,
    phoneNumber: connectedPhone,
    qrCodeUrl: currentQRDataUrl,
    pairingCode,
    status: isConnected ? "connected" : currentQRDataUrl ? "scan_needed" : isConnecting ? "connecting" : "disconnected",
    environment: IS_PRODUCTION ? "production" : "development",
    sessionDir: SESSION_NAME,
  });
});

// 1b. POST RESTART / FORCE REFRESH QR
app.post("/api/wa/restart", async (req, res) => {
  try {
    if (sock) {
      try {
        sock.end(new Error("Manual restart requested"));
      } catch (e) {}
    }
    isConnected = false;
    connectedPhone = null;
    isConnecting = false;
    currentQR = null;
    currentQRDataUrl = null;
    pairingCode = null;

    if (req.body?.clearSession) {
      try {
        fs.rmSync(AUTH_DIR, { recursive: true, force: true });
      } catch (e) {}
    }

    setTimeout(() => startWhatsAppSocket(), 1000);
    return res.json({
      success: true,
      message: "Gateway restarting, generating fresh QR code...",
    });
  } catch (err) {
    return res.status(500).json({ success: false, message: err.message });
  }
});

// 2. POST PAIRING CODE (Using phone number from Settings)
app.post("/api/wa/pair", async (req, res) => {
  const { phoneNumber } = req.body;
  if (!phoneNumber) {
    return res.status(400).json({ success: false, message: "Nomor telepon WhatsApp wajib disertakan" });
  }

  const cleanPhone = formatWhatsAppNumber(phoneNumber);

  if (isConnected) {
    return res.json({ success: true, message: "Sudah terhubung", isConnected: true, phoneNumber: connectedPhone });
  }

  if (!sock) {
    await startWhatsAppSocket();
  }

  try {
    if (!sock.authState.creds.registered) {
      const code = await sock.requestPairingCode(cleanPhone);
      pairingCode = code;
      return res.json({
        success: true,
        pairingCode: code,
        message: `Kode pairing WhatsApp berhasil dibuat untuk ${cleanPhone}`,
      });
    } else {
      return res.json({ success: false, message: "Perangkat sudah terdaftar, silakan tunggu koneksi" });
    }
  } catch (error) {
    console.error("[WA Gateway] Pairing code error:", error);
    return res.status(500).json({ success: false, message: error.message || "Gagal meminta kode pairing" });
  }
});

// 3. POST SEND SINGLE MESSAGE
app.post("/api/wa/send", async (req, res) => {
  const { to, text } = req.body;

  if (!isConnected || !sock) {
    return res.status(503).json({
      success: false,
      message: "WhatsApp Gateway belum terhubung. Silakan scan QR code terlebih dahulu di Admin CMS.",
    });
  }

  if (!to || !text) {
    return res.status(400).json({ success: false, message: "Nomor tujuan ('to') dan pesan ('text') wajib diisi" });
  }

  const cleanPhone = formatWhatsAppNumber(to);
  const jid = `${cleanPhone}@s.whatsapp.net`;

  try {
    const result = await sock.sendMessage(jid, { text });
    return res.json({
      success: true,
      message: `Pesan berhasil dikirim ke ${cleanPhone}`,
      messageId: result?.key?.id,
    });
  } catch (error) {
    console.error(`[WA Gateway] Error sending message to ${cleanPhone}:`, error);
    return res.status(500).json({
      success: false,
      message: `Gagal mengirim pesan: ${error.message || "Unknown error"}`,
    });
  }
});

// 4. POST BROADCAST (Batch send with anti-ban delay)
app.post("/api/wa/broadcast", async (req, res) => {
  const { recipients, delayMs = 2000 } = req.body;

  if (!isConnected || !sock) {
    return res.status(503).json({
      success: false,
      message: "WhatsApp Gateway belum terhubung. Silakan scan QR code terlebih dahulu.",
    });
  }

  if (!Array.isArray(recipients) || recipients.length === 0) {
    return res.status(400).json({ success: false, message: "Daftar penerima ('recipients') tidak boleh kosong" });
  }

  // Execute in background or synchronously stream
  const results = [];

  for (let i = 0; i < recipients.length; i++) {
    const item = recipients[i];
    const cleanPhone = formatWhatsAppNumber(item.phone);

    if (!cleanPhone) {
      results.push({ id: item.id, phone: item.phone, success: false, error: "Nomor tidak valid" });
      continue;
    }

    const jid = `${cleanPhone}@s.whatsapp.net`;

    try {
      await sock.sendMessage(jid, { text: item.text });
      results.push({ id: item.id, phone: cleanPhone, success: true });
    } catch (err) {
      console.error(`[WA Gateway] Error broadcasting to ${cleanPhone}:`, err);
      results.push({ id: item.id, phone: cleanPhone, success: false, error: err.message });
    }

    // Anti-ban delay between messages
    if (i < recipients.length - 1) {
      await new Promise((resolve) => setTimeout(resolve, delayMs));
    }
  }

  const successCount = results.filter((r) => r.success).length;
  const failCount = results.length - successCount;

  return res.json({
    success: true,
    total: results.length,
    sent: successCount,
    failed: failCount,
    results,
  });
});

// 5. POST LOGOUT / DISCONNECT
app.post("/api/wa/logout", async (req, res) => {
  try {
    if (sock) {
      try {
        await sock.logout();
      } catch (e) {}
    }

    isConnected = false;
    connectedPhone = null;
    currentQR = null;
    currentQRDataUrl = null;

    try {
      fs.rmSync(AUTH_DIR, { recursive: true, force: true });
    } catch (e) {}

    setTimeout(() => startWhatsAppSocket(), 2000);

    return res.json({ success: true, message: "Koneksi WhatsApp berhasil diputus dan sesi dibersihkan" });
  } catch (error) {
    return res.status(500).json({ success: false, message: error.message });
  }
});

// Start Express Server & Baileys
app.listen(PORT, () => {
  console.log(`[WA Gateway] Server listening on http://localhost:${PORT}`);
  startWhatsAppSocket();
});
