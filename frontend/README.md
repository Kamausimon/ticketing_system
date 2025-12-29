# Ticketing System Frontend

Next.js frontend for the ticketing system platform.

## Tech Stack

- **Next.js 15** with App Router
- **TypeScript** (strict mode)
- **Tailwind CSS** for styling
- **React 18**

## Getting Started

1. Install dependencies:
```bash
npm install
```

2. Configure environment variables:
```bash
cp .env.local .env.local
# Edit .env.local with your API URL and IntaSend keys
```

3. Run the development server:
```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

## Project Structure

```
src/
├── app/              # Next.js App Router pages
│   ├── layout.tsx    # Root layout
│   ├── page.tsx      # Home page
│   └── globals.css   # Global styles
├── components/       # React components
├── lib/             # Utilities and API client
│   └── api-client.ts # API communication
└── types/           # TypeScript type definitions
    └── index.ts     # Shared types
```

## API Integration

The frontend connects to the Go backend at `http://localhost:8080` (configurable via `NEXT_PUBLIC_API_URL`).

## Building for Production

```bash
npm run build
npm start
```

## Environment Variables

- `NEXT_PUBLIC_API_URL` - Backend API URL
- `NEXT_PUBLIC_INTASEND_PUBLISHABLE_KEY` - IntaSend public key for payments
