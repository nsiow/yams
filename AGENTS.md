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
  * Always confirm PR titles and descriptions with the user before creating. Do not include test plans.
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
* **Yams server:** When starting a one-off yams server for testing, use a non-default port (e.g., `-a :9999`) since port 8888 is typically already in use during development.
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

## Development Server
* A yams server is typically already running on port 8888 during development.
* When you need to start a one-off yams server for testing (e.g., to curl an API endpoint), use a different port:
  ```bash
  ./yams server -a :9999 &
  curl -s http://localhost:9999/api/v1/actions/s3:GetObject
  pkill -f "yams server.*9999" || true
  ```

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

---

## Pull Requests
* **PR body:** Keep concise. Include a brief summary only. Do not include test plans, commit lists, or AI attribution.
* **Commit messages:** Keep short and use conventional commit format.

---

## UI Testing Guidelines

When writing or modifying UI tests:

### Avoid Circular Tests
Do not write tests that simply verify the test setup rather than actual behavior:
```tsx
// BAD: Circular test - just verifies the mock
it('returns mocked data', () => {
  vi.mocked(api.getData).mockReturnValue('test');
  expect(api.getData()).toBe('test'); // This tests nothing useful
});
```

### Avoid Useless Tests
Do not write tests that provide no meaningful coverage:
```tsx
// BAD: Tests implementation detail, not behavior
it('calls useState', () => {
  render(<Component />);
  expect(React.useState).toHaveBeenCalled();
});

// BAD: Tests that a component renders without testing anything specific
it('renders', () => {
  render(<Component />);
  // No assertions about what was rendered
});
```

### Good Test Patterns
- Test user-visible behavior and interactions
- Test error states and edge cases
- Test that correct data is displayed
- Test that user actions trigger expected effects

```tsx
// GOOD: Tests actual user behavior
it('displays error when API fails', async () => {
  vi.mocked(api.getData).mockRejectedValue(new Error('Network error'));
  render(<Component />);
  await waitFor(() => {
    expect(screen.getByText('Network error')).toBeInTheDocument();
  });
});

// GOOD: Tests user interaction
it('submits form when button clicked', async () => {
  const user = userEvent.setup();
  render(<Form />);
  await user.type(screen.getByLabelText('Name'), 'Test');
  await user.click(screen.getByRole('button', { name: /submit/i }));
  expect(api.submitForm).toHaveBeenCalledWith({ name: 'Test' });
});
```

### Test Organization
- Group related tests in `describe` blocks
- Use clear, descriptive test names
- Keep tests focused on single behaviors
- Prefer testing public API over implementation details
