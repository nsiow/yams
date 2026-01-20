// ui/src/pages/simulate/shared/context-editor.tsx
/* eslint-disable react-refresh/only-export-components */
import {
  ActionIcon,
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
}

export function ContextEditor({
  contextVars,
  onChange,
  onRerun,
  showRerunButton,
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

  return (
    <Card withBorder p="lg">
      <Group justify="space-between" mb={contextVars.length > 0 ? 'md' : undefined}>
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
      {contextVars.length === 0 && (
        <Text size="sm" c="dimmed">
          Add context variables to test conditions like aws:SourceIp, aws:RequestTag/*, etc.
        </Text>
      )}
    </Card>
  );
}

// Helper to build context object from key-value pairs
export function buildContext(contextVars: ContextVariable[]): Record<string, string> | undefined {
  const validPairs = contextVars.filter((cv) => cv.key.trim() && cv.value.trim());
  if (validPairs.length === 0) return undefined;
  return Object.fromEntries(validPairs.map((cv) => [cv.key.trim(), cv.value.trim()]));
}
