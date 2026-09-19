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
} = require("@whiskeysockets/baileys");

const app = express();
const PORT = process.env.PORT || 5001;

app.use(cors());
app.use(express.json());

const AUTH_DIR = path.join(__dirname, "auth_info_baileys");
const logger = pino({ level: "silent" });

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
    const { state, saveCreds } = await useMultiFileAuthState(AUTH_DIR);
    const { version, isLatest } = await fetchLatestBaileysVersion();

    sock = makeWASocket({
      version,
      logger,
      printQRInTerminal: false,
      auth: {
        creds: state.creds,
        keys: makeCacheableSignalKeyStore(state.keys, logger),
      },
      browser: ["Akuatik Indonesia Admin", "Chrome", "1.0.0"],
      generateHighQualityLinkPreview: true,
      syncFullHistory: false,
    });

    sock.ev.on("creds.update", saveCreds);

    sock.ev.on("connection.update", async (update) => {
      const { connection, lastDisconnect, qr } = update;

      if (qr) {
        currentQR = qr;
        try {
          currentQRDataUrl = await qrcode.toDataURL(qr, { margin: 2, scale: 6 });
        } catch (err) {
          console.error("Failed to generate QR data URL:", err);
        }
      }

      if (connection === "close") {
        isConnected = false;
        connectedPhone = null;
        isConnecting = false;
        currentQR = null;
        currentQRDataUrl = null;

        const statusCode = lastDisconnect?.error?.output?.statusCode;
        const shouldReconnect = statusCode !== DisconnectReason.loggedOut;

        console.log(`[WA Gateway] Connection closed (${statusCode}). Reconnecting: ${shouldReconnect}`);

        if (statusCode === DisconnectReason.loggedOut) {
          // Clear auth credentials if logged out
          try {
            fs.rmSync(AUTH_DIR, { recursive: true, force: true });
          } catch (e) {}
        }

        if (shouldReconnect) {
          setTimeout(() => startWhatsAppSocket(), 3000);
        }
      } else if (connection === "open") {
        isConnected = true;
        isConnecting = false;
        currentQR = null;
        currentQRDataUrl = null;
        pairingCode = null;

        // Extract phone number from JID (e.g. 628123456789:1@s.whatsapp.net)
        const userJid = sock.user?.id || "";
        const rawPhone = userJid.split(":")[0] || userJid.split("@")[0];
        connectedPhone = rawPhone;

        console.log(`[WA Gateway] Connected successfully to WhatsApp! Admin Phone: ${connectedPhone}`);
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
  res.json({
    success: true,
    isConnected,
    phoneNumber: connectedPhone,
    qrCodeUrl: currentQRDataUrl,
    pairingCode,
    status: isConnected ? "connected" : currentQRDataUrl ? "scan_needed" : isConnecting ? "connecting" : "disconnected",
  });
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
