# AirHID - Wireless Remote Control Utility

**AirHID** turns your smartphone into a secure, professional-grade remote keyboard, mouse, and clipboard for your computer. No apps to install on your phone—just scan and control.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.20%2B-cyan.svg)
![Platform](https://img.shields.io/badge/platform-Windows-lightgrey.svg)

## ✨ Key Features

- **🛡️ Secure by Design**: Auto-generated security tokens ensure only *you* can control your PC.
- **⚡ Instant Connection**: Zero-config startup. Just run the exe and scan the QR code.
- **🔄 Seamless Reconnection**: 
    - **"Pair Once, Trust Forever"**: The web app remembers your token. 
    - Add to your home screen or bookmarks for one-tap access next time.
- **⌨️ Smart Typing**: 
    - Type on your phone's native keyboard (with auto-correct/suggestions) and send text instantly.
    - Support for special keys (Ctrl, Alt, Win, F1-F12) and shortcuts.
- **📋 Clipboard Sync**: Instantly paste text from your phone to your PC's clipboard.
- **🖱️ Multi-Touch Trackpad**: 
    - Silky smooth mouse control with sensitivity adjustment.
    - Supports tap-to-click, two-finger scroll, and right-click.

## 🚀 Quick Start

### 1. Run
Download and run `airhid.exe` on your Windows PC.

```text
AirHID Running (Secure Mode)
Listening on: 0.0.0.0:5000
Connect URL:  http://192.168.1.5:5000/?token=abc123...
[QR Code Here]
```

### 2. Scan
Use your phone's camera to scan the QR code.

### 3. Control
- **Type Mode**: Type text and hit "Send".
- **Clipboard Mode**: Paste long text blocks directly to PC clipboard.
- **Touchpad Mode**: Use screen as a trackpad.

> **Pro Tip:** Add the webpage to your phone's Home Screen. Next time you launch AirHID on PC, just tap the icon on your phone to reconnect instantly!

## ⚙️ Configuration

AirHID creates a `config.json` file on the first run. You can customize it:

```json
{
  "token": "your-secret-token",  // Security token (keep secret!)
  "host": "0.0.0.0",             // Bind address (e.g., "127.0.0.1" for local only)
  "port": "5000"                 // Server port
}
```

## 🛠️ Build from Source

Requirements: [Go 1.20+](https://go.dev/)

```bash
git clone https://github.com/ZhangSanmo/airhid.git
cd airhid
go mod tidy
go build -o airhid.exe main.go
```

## ⚠️ Note
Run as **Administrator** if you need to simulate input into elevated windows (like Task Manager or some full-screen games).

## 📄 License

MIT License. Free for personal and commercial use.
