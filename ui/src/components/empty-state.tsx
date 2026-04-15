import { Stack, Text, ThemeIcon } from '@mantine/core';
import { IconSearch, IconFilter, IconDatabaseOff } from '@tabler/icons-react';

type EmptyStateVariant = 'no-results' | 'no-selection' | 'error' | 'no-data';

interface EmptyStateProps {
  variant: EmptyStateVariant;
  entityName?: string;
  message?: string;
}

const variants: Record<EmptyStateVariant, { icon: typeof IconSearch; color: string; defaultMessage: string }> = {
  'no-results': {
    icon: IconFilter,
    color: 'gray',
    defaultMessage: 'No results match your filters. Try adjusting your search or clearing filters.',
  },
  'no-selection': {
    icon: IconSearch,
    color: 'gray',
    defaultMessage: 'Select an item to view details',
  },
  'error': {
    icon: IconDatabaseOff,
    color: 'red',
    defaultMessage: 'Failed to load data',
  },
  'no-data': {
    icon: IconDatabaseOff,
    color: 'gray',
    defaultMessage: 'No data available',
  },
};

export function EmptyState({ variant, entityName, message }: EmptyStateProps): JSX.Element {
  const config = variants[variant];
  const Icon = config.icon;

  let displayMessage = message || config.defaultMessage;
  if (entityName && variant === 'no-selection') {
    displayMessage = `Select a ${entityName} to view details`;
  }

  return (
    <Stack align="center" justify="center" h="100%" gap="md">
      <ThemeIcon size={48} radius="xl" color={config.color} variant="light">
        <Icon size={24} />
      </ThemeIcon>
      <Text c="dimmed" ta="center" maw={300}>
        {displayMessage}
      </Text>
    </Stack>
  );
}
