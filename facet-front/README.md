# Facet — Frontend

Feature flag management platform frontend built with React 19, Material UI 7, TypeScript, and Vite 8.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Framework | React 19 |
| Language | TypeScript 7 |
| UI Library | Material UI 7 (MUI) |
| State Management | Zustand 5 |
| Routing | React Router v7 |
| HTTP Client | Axios |
| Build Tool | Vite 8 |
| Testing | Vitest, jsdom, Testing Library |
| Linting | ESLint (flat config) + Prettier |
| Deployment | Docker + Nginx (multi-stage build) |

## Project Structure

```
facet-front/
├── src/
│   ├── api/              # API client modules (one per resource)
│   ├── components/       # Shared UI components
│   ├── layouts/          # Page layouts (MainLayout)
│   ├── router/           # React Router config + guards
│   ├── stores/           # Zustand state stores
│   ├── types/            # TypeScript type definitions
│   ├── utils/            # Utilities (request helper, formatters)
│   ├── views/            # Page-level components
│   ├── App.tsx           # Root component
│   └── main.tsx          # Entry point
├── website/              # Marketing landing page (static HTML)
├── nginx.conf            # Nginx config for production
├── Dockerfile            # Multi-stage Docker build
├── eslint.config.js      # ESLint flat configuration
├── vite.config.js        # Vite configuration
└── package.json
```

## Development

```bash
# Install dependencies
npm ci

# Start dev server with HMR (proxies /api/v1 to localhost:8080)
npm run dev

# Run tests in watch mode
npm test

# Lint
npm run lint

# Format
npm run format
```

The Vite dev server proxies `/api/v1/*` requests to `http://localhost:8080` so you can run the backend locally alongside.

## Build & Deploy

```bash
# Build for production
npm run build              # outputs to dist/

# Preview production build
npm run preview            # serves dist/ locally
```

Docker:

```bash
docker build -t facet-front .
docker run -p 80:80 facet-front
```

Or run as part of the full stack from the project root:

```bash
docker compose up
```

## Testing

Tests use Vitest with jsdom environment. All test files are co-located with source:

```
src/api/auth.test.js        # API module tests
src/views/Login.test.jsx    # View component tests
src/stores/authStore.test.js  # Store tests
src/utils/request.test.js    # Utility tests
```

Run all tests:

```bash
npm run test
```

Or with coverage:

```bash
npx vitest run --coverage
```

## Conventions

- TypeScript throughout (`.tsx` / `.ts`) including tests
- API modules use default export with Axios instance
- Views use named exports with lazy loading
- State stores use Zustand with `create()`
- All API responses wrapped in `{code, message, data}` format
- Path alias `@` maps to `src/`
- i18n with Zustand store (`localeStore`) and localStorage persistence; switch between Chinese and English in-app

## API Reference

See the [backend API documentation](../facet-backend/API_DOC.md) for available endpoints.
