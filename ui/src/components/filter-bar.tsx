import { Button, Group } from '@mantine/core';
import { IconX } from '@tabler/icons-react';

interface FilterBarProps {
  children: React.ReactNode;
  hasActiveFilters: boolean;
  onClearAll: () => void;
}

export function FilterBar({ children, hasActiveFilters, onClearAll }: FilterBarProps): JSX.Element {
  return (
    <Group gap="sm" align="flex-end">
      {children}
      {hasActiveFilters && (
        <Button
          variant="subtle"
          color="gray"
          size="sm"
          leftSection={<IconX size={14} />}
          onClick={onClearAll}
        >
          Clear
        </Button>
      )}
    </Group>
  );
}
