# GEMINI.md

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
* **Git:** Use Conventional Commits (e.g., `feat:`, `fix:`, `docs:`).
* **Comments:** Keep comments concise, and use them to summarize what is happening in "paragraphs"
  of code. Avoid stream-of-consciousness comments and overly verbose language.

---

## Development Rules (Backend)
* **Style:** Simple, imperative Go. Avoid syntax that is difficult to understand. Keep the code DRY.
* **Dependencies:** Avoid use of any external dependencies other than the AWS SDK for Go.
* **Test style:** Where possible for Go tests, use the TDT pattern implemented in
  `testlib.RunTestSuite`
* **Test data:** Use `testdata/` directories for any file/data type content used in tests. There are many such directories within the repo.
* **Yams server:** When starting a one-off yams server for testing, use a non-default port (e.g., `-a :9999`) since port 8888 is typically already in use during development.

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
TBD, follow what currently exists

---

## Output Format
* Include the **file path** as a comment on the first line of every code block.
* If a change involves multiple files, provide a **bulleted summary** of changes before the code.

---

## Pull Requests
* **PR body:** Keep concise. Include a brief summary only. Do not include test plans, commit lists, or AI attribution.
* **Commit messages:** Keep short and use conventional commit format.
