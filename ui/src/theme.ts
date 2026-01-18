import { createTheme, MantineColorsTuple } from '@mantine/core';

// Custom color palette from https://coolors.co/6e44ff-b892ff-ffc2e2-ff90b3-ef7a85
const purple: MantineColorsTuple = [
  '#f3edff',
  '#e1d7fa',
  '#c0abf1',
  '#9d7ce9',
  '#7f54e2',
  '#6e3bde',
  '#6530dd',
  '#5523c4',
  '#4b1eb0',
  '#3f169b',
];

const pink: MantineColorsTuple = [
  '#ffe9f3',
  '#ffd3e4',
  '#ffa5c7',
  '#ff74a7',
  '#fe4b8d',
  '#fe327c',
  '#ff2374',
  '#e41362',
  '#cc0657',
  '#b3004a',
];

export const theme = createTheme({
  primaryColor: 'purple',
  colors: {
    purple,
    pink,
  },
  fontFamily: 'system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
  headings: {
    fontFamily: 'system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
  },
});
