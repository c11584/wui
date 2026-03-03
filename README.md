# WUI - Next-Generation Proxy Management Panel

A modern, high-performance proxy management panel that surpasses x-ui.

## Features

- 🔥 **Tunnel Orchestration**: Inbound protocol + port → Outbound protocol + port (with UDP support)
- ⚡ **One-Click Install**: Automated installation with configurable parameters
- 🎨 **Modern UI**: React 18 + TypeScript + Vite + shadcn/ui
- 🚀 **High Performance**: WebSocket real-time communication, < 200KB bundle size
- 🔐 **Enterprise Security**: JWT authentication + 2FA support
- 🎭 **Dynamic Theming**: Light/Dark mode + custom themes
- 📊 **Real-time Monitoring**: Traffic statistics + connection tracking
- 🌐 **Multi-Protocol**: VMess, VLESS, Trojan, Shadowsocks, Hysteria, SOCKS5, HTTP

## Quick Start

### One-Click Installation

```bash
curl -fsSL https://your-domain.com/install.sh | bash
```

Or with custom parameters:

```bash
curl -fsSL https://your-domain.com/install.sh | bash -s -- --port 32451 --username admin --password admin
```

Default credentials:
- **Port**: 32451
- **Username**: admin
- **Password**: admin

## Development

### Prerequisites

- Node.js 18+
- Go 1.21+
- pnpm 8+

### Setup

```bash
# Install dependencies
pnpm install

# Start development servers
pnpm dev
```

## Tech Stack

- **Frontend**: React 18, TypeScript, Vite, shadcn/ui, Tailwind CSS
- **Backend**: Go, Gin, GORM, SQLite
- **Proxy Core**: Xray
- **Build**: pnpm, Monorepo

## License

MIT
