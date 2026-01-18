import { Box, Group, Skeleton, Stack } from '@mantine/core';

interface ListSkeletonProps {
  count?: number;
}

export function ListSkeleton({ count = 10 }: ListSkeletonProps): JSX.Element {
  return (
    <Stack gap={0}>
      {Array.from({ length: count }).map((_, i) => (
        <Box
          key={i}
          p="sm"
          style={{ borderBottom: '1px solid var(--mantine-color-default-border)' }}
        >
          <Group gap="md">
            <Skeleton height={16} width={16} radius="sm" />
            <Stack gap={6} style={{ flex: 1 }}>
              <Skeleton height={14} width="60%" radius="sm" />
              <Skeleton height={10} width="40%" radius="sm" />
            </Stack>
          </Group>
        </Box>
      ))}
    </Stack>
  );
}

export function DetailSkeleton(): JSX.Element {
  return (
    <Stack gap="md" p="md">
      <Group justify="space-between">
        <Skeleton height={28} width="50%" radius="sm" />
        <Skeleton height={24} width={80} radius="xl" />
      </Group>
      <Skeleton height={120} radius="sm" />
      <Skeleton height={80} radius="sm" />
      <Skeleton height={100} radius="sm" />
    </Stack>
  );
}
