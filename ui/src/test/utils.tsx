// ui/src/test/utils.tsx
/* eslint-disable react-refresh/only-export-components */
import { MantineProvider } from '@mantine/core';
import { render, RenderOptions } from '@testing-library/react';
import { MemoryRouter, MemoryRouterProps } from 'react-router-dom';
import { ReactElement, ReactNode } from 'react';

interface WrapperProps {
  children: ReactNode;
}

interface CustomRenderOptions extends Omit<RenderOptions, 'wrapper'> {
  routerProps?: MemoryRouterProps;
  withRouter?: boolean;
}

// Wrapper component that provides Mantine only
function MantineOnlyWrapper({ children }: WrapperProps): JSX.Element {
  return (
    <MantineProvider>
      {children}
    </MantineProvider>
  );
}

// Wrapper component that provides all necessary providers
function createWrapper(routerProps?: MemoryRouterProps) {
  return function AllProviders({ children }: WrapperProps): JSX.Element {
    return (
      <MemoryRouter {...routerProps}>
        <MantineProvider>
          {children}
        </MantineProvider>
      </MemoryRouter>
    );
  };
}

// Custom render function that wraps components with providers
function customRender(
  ui: ReactElement,
  options?: CustomRenderOptions
): ReturnType<typeof render> {
  const { routerProps, withRouter = true, ...renderOptions } = options ?? {};

  if (withRouter) {
    return render(ui, { wrapper: createWrapper(routerProps), ...renderOptions });
  }
  return render(ui, { wrapper: MantineOnlyWrapper, ...renderOptions });
}

// Re-export everything from testing-library
export * from '@testing-library/react';
export { customRender as render };
