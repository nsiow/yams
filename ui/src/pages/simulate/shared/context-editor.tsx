// ui/src/pages/simulate/shared/context-editor.tsx
/* eslint-disable react-refresh/only-export-components */
import {
  ActionIcon,
  Box,
  Button,
  Card,
  Group,
  Stack,
  Text,
  TextInput,
  Title,
} from '@mantine/core';
import { IconPlus, IconPlayerPlay, IconX } from '@tabler/icons-react';

export interface ContextVariable {
  key: string;
  value: string;
}

interface ContextEditorProps {
  contextVars: ContextVariable[];
  onChange: (vars: ContextVariable[]) => void;
  onRerun?: () => void;
  showRerunButton?: boolean;
  sharedContextVars?: ContextVariable[];
}

export function ContextEditor({
  contextVars,
  onChange,
  onRerun,
  showRerunButton,
  sharedContextVars,
}: ContextEditorProps): JSX.Element {
  const addContextVar = (): void => {
    onChange([...contextVars, { key: '', value: '' }]);
  };

  const removeContextVar = (index: number): void => {
    onChange(contextVars.filter((_, i) => i !== index));
  };

  const updateContextVar = (index: number, field: 'key' | 'value', val: string): void => {
    onChange(contextVars.map((cv, i) => (i === index ? { ...cv, [field]: val } : cv)));
  };

  const hasShared = sharedContextVars && sharedContextVars.length > 0;
  const hasContent = contextVars.length > 0 || hasShared;

  return (
    <Card withBorder p="lg">
      <Group justify="space-between" mb={hasContent ? 'md' : undefined}>
        <Title order={5}>Request Context</Title>
        <Button
          variant="subtle"
          size="xs"
          leftSection={<IconPlus size={14} />}
          onClick={addContextVar}
        >
          Add Variable
        </Button>
      </Group>

      {/* Shared context variables (read-only) */}
      {hasShared && (
        <Stack gap="xs" mb={contextVars.length > 0 ? 'sm' : undefined}>
          {sharedContextVars.map((cv, idx) => (
            <Group key={`shared-${idx}`} gap="xs" wrap="nowrap">
              <TextInput
                value={cv.key}
                readOnly
                style={{ flex: 1 }}
                size="sm"
                styles={{ input: { backgroundColor: 'var(--mantine-color-violet-0)', color: 'var(--mantine-color-violet-7)' } }}
              />
              <TextInput
                value={cv.value}
                readOnly
                style={{ flex: 1 }}
                size="sm"
                styles={{ input: { backgroundColor: 'var(--mantine-color-violet-0)', color: 'var(--mantine-color-violet-7)' } }}
              />
              <Box w={28} style={{ flexShrink: 0 }}>
                <Text size="xs" c="dimmed" ta="center">&nbsp;</Text>
              </Box>
            </Group>
          ))}
          <Text size="xs" c="dimmed" fs="italic">Shared request context</Text>
        </Stack>
      )}

      {/* User-defined context variables */}
      {contextVars.length > 0 && (
        <Stack gap="xs">
          {contextVars.map((cv, idx) => (
            <Group key={idx} gap="xs" wrap="nowrap">
              <TextInput
                placeholder="Key (e.g., aws:SourceIp)"
                value={cv.key}
                onChange={(e) => updateContextVar(idx, 'key', e.currentTarget.value)}
                style={{ flex: 1 }}
                size="sm"
              />
              <TextInput
                placeholder="Value"
                value={cv.value}
                onChange={(e) => updateContextVar(idx, 'value', e.currentTarget.value)}
                style={{ flex: 1 }}
                size="sm"
              />
              <ActionIcon
                variant="subtle"
                color="gray"
                onClick={() => removeContextVar(idx)}
                aria-label="Remove variable"
              >
                <IconX size={16} />
              </ActionIcon>
            </Group>
          ))}
          {showRerunButton && onRerun && (
            <Button size="xs" variant="light" onClick={onRerun} mt="xs" leftSection={<IconPlayerPlay size={14} />}>
              Re-run Simulation
            </Button>
          )}
        </Stack>
      )}
      {!hasContent && (
        <Text size="sm" c="dimmed">
          Add context variables to test conditions like aws:SourceIp, aws:RequestTag/*, etc.
        </Text>
      )}
    </Card>
  );
}

// Helper to build context object from key-value pairs, with optional shared context merged in.
// Shared vars are included first, then user vars override on key conflict.
export function buildContext(
  contextVars: ContextVariable[],
  sharedVars?: ContextVariable[],
): Record<string, string> | undefined {
  const result: Record<string, string> = {};

  // Shared vars first
  if (sharedVars) {
    for (const cv of sharedVars) {
      const k = cv.key.trim();
      const v = cv.value.trim();
      if (k && v) result[k] = v;
    }
  }

  // User vars override
  for (const cv of contextVars) {
    const k = cv.key.trim();
    const v = cv.value.trim();
    if (k && v) result[k] = v;
  }

  if (Object.keys(result).length === 0) return undefined;
  return result;
}
