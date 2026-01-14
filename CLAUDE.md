# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Quiz Sprint TMA is a Telegram Mini App built with Vue 3, TypeScript, and Vite. The main application lives in the `tma/` subdirectory.

## Commands

All commands run from the `tma/` directory using pnpm:

```bash
# Development
pnpm dev              # Start dev server (port 5173)
pnpm preview          # Preview production build (port 4173)

# Building
pnpm build            # Type-check + build for production
pnpm build-only       # Build only (skip type-check)

# Type Checking
pnpm type-check       # Run vue-tsc

# Linting
pnpm lint             # Run all linters (oxlint + eslint)
pnpm format           # Format with Prettier

# Testing
pnpm test:unit                              # Run Vitest unit tests
pnpm test:e2e                               # Run Playwright E2E tests
pnpm test:e2e --project=chromium            # Run E2E on specific browser
npx playwright install                       # Install browser drivers (first run)
```

## Architecture

### Monorepo Structure
- `tma/` - Main Vue 3 application
- `infrastructure/` - VPS server configurations and nginx setup
- `dev-tunnel/` - SSH tunnel scripts for HTTPS development
- `.github/workflows/` - CI/CD pipelines

### TMA Application (`tma/src/`)
- `main.ts` - Vue app initialization
- `App.vue` - Root component
- `router/` - Vue Router configuration
- `views/` - Page components
- `__tests__/` - Vitest unit tests

### Build Once, Deploy Many
The CI/CD uses a two-stage workflow:
1. `build.yml` - Runs quality checks (type-check, lint), builds, and uploads artifact
2. `deploy.yml` - Downloads artifact and deploys to staging or production

### Environments
- Development: `dev.quiz-sprint-tma.online` (via SSH tunnel to VPS)
- Staging: `staging.quiz-sprint-tma.online`
- Production: `quiz-sprint-tma.online`

## Tech Stack

- Vue 3.5 with Composition API (`<script setup>`)
- TypeScript 5.9
- Vite (dev server and bundler)
- Vue Router 4
- Vitest + Vue Test Utils (unit testing)
- Playwright (E2E testing)
- ESLint + Oxlint + Prettier (code quality)
- pnpm 9 (package manager)
- Node.js ^20.19.0 || >=22.12.0

## Code Style

- No semicolons
- Single quotes
- 100 character line width
- Path alias: `@` maps to `./src`

## Workflow Requirements (from AGENTS.md)

Before completing a session:
1. Track issues with `bd` (beads) tool
2. Always push to remote: `git push` is required
