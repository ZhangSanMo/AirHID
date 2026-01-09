# AirHID

AirHID is a professional-grade wireless input utility that allows you to control your computer's keyboard and clipboard from any mobile device via a local web interface. It acts as a virtual HID (Human Interface Device) over the network.

## Features

- **Wireless Control:** Turn your smartphone into a remote input device.
- **QR Code Pairing:** Instant connection by scanning a QR code at startup.
- **Real-time Typing:** Simulate keystrokes on the host machine (bypassing VM/RDP clipboard restrictions).
- **Clipboard Sync:** Instantly sync text from your phone to your computer's clipboard.
- **Shortcuts & Keys:** dedicated buttons for control keys (Enter, Tab, Esc, etc.).

## Installation

### Prerequisites
- [Go](https://go.dev/dl/) 1.20+

### Build from Source

```bash
git clone https://github.com/yourusername/airhid.git
cd airhid
go mod tidy
go build -o airhid.exe
```

## Usage

1. Run the application:
   ```bash
   ./airhid.exe
   ```
   *Note: Run as Administrator on Windows if you need to control elevated applications (like Task Manager).*

2. Scan the displayed **QR Code** with your phone to open the control interface.

3. Use the web interface to:
   - **Type:** Send text character-by-character.
   - **Copy:** Sync text to the PC clipboard.
   - **Control:** Use function keys remotely.

## Acknowledgments

*   Original inspiration and core logic from [ychisbest/easytype](https://github.com/ychisbest/easytype).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.