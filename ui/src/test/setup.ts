// ui/src/test/setup.ts
import '@testing-library/jest-dom';
import { vi } from 'vitest';

// Mock matchMedia for Mantine components
const createMatchMediaMock = () => {
  return (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  });
};

window.matchMedia = createMatchMediaMock();

// Mock ResizeObserver for Mantine components (must be a class)
class ResizeObserverMock {
  observe = vi.fn();
  unobserve = vi.fn();
  disconnect = vi.fn();
}

global.ResizeObserver = ResizeObserverMock;

// Mock IntersectionObserver
class IntersectionObserverMock {
  observe = vi.fn();
  unobserve = vi.fn();
  disconnect = vi.fn();
  root = null;
  rootMargin = '';
  thresholds = [];
}

global.IntersectionObserver = IntersectionObserverMock as unknown as typeof IntersectionObserver;

// Mock scrollIntoView
Element.prototype.scrollIntoView = vi.fn();

// Mock getComputedStyle for Mantine
const originalGetComputedStyle = window.getComputedStyle;
window.getComputedStyle = (element: Element) => {
  return originalGetComputedStyle(element);
};

// Mock clipboard API
Object.assign(navigator, {
  clipboard: {
    writeText: vi.fn().mockResolvedValue(undefined),
    readText: vi.fn().mockResolvedValue(''),
  },
});

// Mock URL.createObjectURL and revokeObjectURL for export tests
global.URL.createObjectURL = vi.fn(() => 'blob:test');
global.URL.revokeObjectURL = vi.fn();
