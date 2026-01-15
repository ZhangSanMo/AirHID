# AirHID - Wireless Remote Control Utility

**AirHID** turns your smartphone into a secure, professional-grade remote keyboard, mouse, and clipboard for your computer. No apps to install on your phone—just scan and control.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.25%2B-cyan.svg)
![Platform](https://img.shields.io/badge/platform-Windows-lightgrey.svg)

## ✨ Key Features

- **🛡️ Secure by Design**: Auto-generated security tokens ensure only *you* can control your PC.
- **🖥️ System Tray App**: Runs quietly in the background with a system tray icon.
- **⚡ Auto-Start**: Easily enable "Run on Login" via the tray menu to have it always ready.
- **🔄 Seamless Reconnection**: 
    - **"Pair Once, Trust Forever"**: The web app remembers your token. 
    - Add to your home screen or bookmarks for one-tap access next time.
- **⌨️ Smart Typing**:
    - Type on your phone's native keyboard (with auto-correct/suggestions) and send text instantly.
    - Support for special keys (Ctrl, Alt, Win, F1-F12) and shortcuts.
    - **Repeat commands**: Execute any command multiple times:
      - English: `5x backspace`, `ctrl+z*3`, `enter*10`
      - 中文口语: `三次ctrl加z`, `回车五次`, `按十下退格`
- **📋 Clipboard Sync**: Instantly paste text from your phone to your PC's clipboard.
- **🖱️ Multi-Touch Trackpad**: 
    - Silky smooth mouse control with sensitivity adjustment.
    - Supports tap-to-click, two-finger scroll, and right-click.

## 🚀 Quick Start

### 1. Run
Download and run `airhid.exe`. It will minimize to the **System Tray** (bottom right of your taskbar).

### 2. Connect
Right-click the AirHID tray icon and select **"Show Connection Info"**.
This will open a page with a QR code in your browser.

### 3. Scan & Control
Use your phone's camera to scan the QR code.
- **Type Mode**: Type text and hit "Send".
- **Clipboard Mode**: Paste long text blocks directly to PC clipboard.
- **Touchpad Mode**: Use screen as a trackpad.

> **Pro Tip:** Enable **"Enable Auto-Start"** in the tray menu so AirHID is always ready when you turn on your PC.

## ⚙️ Configuration

AirHID creates a `config.json` file next to the executable on the first run:

```json
{
  "token": "your-secret-token",  // Security token (keep secret!)
  "host": "0.0.0.0",             // Bind address
  "port": "5000"                 // Server port
}
```

## 🛠️ Build from Source

Requirements: [Go 1.25+](https://go.dev/)

1.  Clone the repository:
    ```bash
    git clone https://github.com/ZhangSanmo/airhid.git
    cd airhid
    ```

2.  (Optional) Install `rsrc` to embed the icon:
    ```bash
    go install github.com/akavel/rsrc@latest
    rsrc -ico internal/icon/icon.ico -o rsrc.syso
    ```

3.  Build (Windows GUI mode):
    ```bash
    go mod tidy
    go build -ldflags="-H windowsgui" -o airhid.exe
    ```

## ⚠️ Important Note
**Administrative Privileges**:
AirHID does **not** run as Administrator by default. This is safer, but it means you cannot control windows that *are* running as Administrator (e.g., Task Manager, Registry Editor) due to Windows security isolation.

If you need to control these windows:
1.  Right-click `airhid.exe`
2.  Select **Run as Administrator**

## 📄 License

MIT License. Free for personal and commercial use.