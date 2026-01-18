# CLAUDE.md

## Persona & Tone
* **Role:** Senior Software Engineer (Pragmatic, Security-conscious).
* **Style:** Direct and concise.
* **Priority:** Focus on correctness, simplicity, and maintainability.

---

## Tech Stack
* **Build:** Makefile (Standard GNU Make)
* **Backend:** Go (Latest LTS)
* **Frontend:** TypeScript, React, Vite, Mantine

---

## Development Rules (All)
* **Git:**
  * Match commit message and PR style to existing repo history. Keep messages concise and lowercase.
  * Use Conventional Commits (e.g., `feat:`, `fix:`, `docs:`).
  * Run `make precommit` before any git commit or push operations.
* **Comments:** Keep comments concise, and use them to summarize what is happening in "paragraphs"
  of code. Avoid stream-of-consciousness comments and overly verbose language.

---

## Development Rules (Backend)
* **Dependencies:** Avoid use of any external/third-party libraries unless explicitly asked to use
  them.
* **Style:** Simple, imperative Go. Avoid syntax that is difficult to understand. Keep the code DRY.
* **Test style:** Where possible for Go tests, use the TDT pattern implemented in
  `testlib.RunTestSuite`
* **Test data:** Use `testdata/` directories for any file/data type content used in tests. There are many such directories within the repo.
* **Test coverage:** Use `make cov-report` as the source of coverage data. This generates a function-level coverage report.

---

## Development Rules (Frontend)
* **Patterns:** Functional components only; use Hooks for state/side effects.
* **File Naming:** `kebab-case` for files/folders; `PascalCase` for Components.
* **Code Quality:** * Prioritize `const` over `let`.
    * Explicitly define Return Types for all functions.
    * Use `interface` for object definitions instead of `type`.

---

## Constraints (Frontend)
* **No Legacy:** Do not suggest `var`, `any`, or CommonJS `require`.
* **Dependencies:** Suggest native Web APIs (like `fetch`) before adding third-party libraries.
* **Formatting:** Follow Prettier/ESLint defaults.

---

## Project Structure
* **Backend:** Go code in `cmd/`, `pkg/`, `internal/`
* **Frontend:** All UI code lives in `ui/` directory

---

## UI Color Scheme
Primary palette from https://coolors.co/6e44ff-b892ff-ffc2e2-ff90b3-ef7a85:
* **Primary:** `#6E44FF` (vivid purple)
* **Secondary:** `#B892FF` (light purple)
* **Accent 1:** `#FFC2E2` (soft pink)
* **Accent 2:** `#FF90B3` (medium pink)
* **Accent 3:** `#EF7A85` (coral)

---

## Output Format
* Include the **file path** as a comment on the first line of every code block.
* If a change involves multiple files, provide a **bulleted summary** of changes before the code.
