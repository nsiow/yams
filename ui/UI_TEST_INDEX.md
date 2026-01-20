# UI Test Index

This document lists all UI tests organized by category.

## Test Framework

- **Framework**: Vitest
- **Testing Library**: React Testing Library
- **Run tests**: `npm test` (from `ui/` directory) or `make ui-test` (from project root)

## Component Tests

### EmptyState (`src/components/empty-state.test.tsx`)
- Renders no-results variant with default message
- Renders no-selection variant with default message
- Renders no-selection variant with entity name
- Renders error variant with default message
- Renders no-data variant with default message
- Renders custom message when provided

### CollapsibleCard (`src/components/collapsible-card.test.tsx`)
- Renders title
- Renders children when open by default
- Hides children when defaultOpen is false
- Toggles content visibility on click

### FilterBar (`src/components/filter-bar.test.tsx`)
- Renders children
- Shows Clear button when hasActiveFilters is true
- Hides Clear button when hasActiveFilters is false
- Calls onClearAll when Clear button is clicked

### CopyButton (`src/components/copy-button.test.tsx`)
- Renders copy button
- Has correct value accessible for clipboard

### ExportButton (`src/components/export-button.test.tsx`)
- Renders download button
- Creates downloadable blob on click

### CopyEntityButton (`src/components/copy-entity-button.test.tsx`)
- Renders button
- Accepts complex data objects

## API Client Tests

### YamsClient (`src/lib/api/client.test.ts`)

#### Health & Status
- Returns health status
- Returns status response

#### Actions
- Lists all actions
- Gets single action
- Searches actions

#### Principals
- Lists all principals
- Gets single principal

#### Resources
- Lists all resources
- Gets single resource

#### Policies
- Lists all policies
- Gets single policy

#### Accounts
- Lists all accounts
- Gets single account

#### Simulation
- Runs simulation
- Runs whichPrincipals query
- Runs whichResources query
- Runs whichActions query

#### Overlays
- Lists overlays
- Lists overlays with query
- Gets single overlay
- Creates overlay
- Updates overlay
- Deletes overlay

#### Error Handling
- Throws YamsApiError on non-ok response
- Includes status code in error
- Handles non-JSON error responses

#### URL Encoding
- Encodes special characters in action key
- Encodes ARNs properly

## Page Tests

### HomePage (`src/pages/home.test.tsx`)
- Shows loading state initially
- Displays dashboard when data loads successfully
- Shows healthy status when healthcheck succeeds
- Shows unhealthy status when healthcheck fails
- Displays error when API fails
- Displays data sources with freshness indicators
- Refreshes data when refresh button is clicked
- Displays environment variables when present

### OverlayEditorPage (`src/pages/overlays/editor.test.tsx`)

#### New Overlay Creation
- Shows new overlay form when id is "new"
- Enables create button when name is confirmed
- Creates overlay when form is submitted

#### Existing Overlay Editing
- Loads existing overlay data
- Displays entity tabs with counts
- Shows principal in the list
- Shows save button disabled when no changes

#### Error Handling
- Displays error when overlay fails to load

#### Entity Selection
- Shows detail panel when entity is selected

## Test Utilities

### Setup (`src/test/setup.ts`)
- Mocks `window.matchMedia` for Mantine components
- Mocks `ResizeObserver` for Mantine components
- Mocks `IntersectionObserver`
- Mocks `scrollIntoView`
- Mocks clipboard API
- Mocks `URL.createObjectURL` and `URL.revokeObjectURL`

### Render Utils (`src/test/utils.tsx`)
- Custom render function with providers (MantineProvider, MemoryRouter)
- Re-exports all testing-library utilities
