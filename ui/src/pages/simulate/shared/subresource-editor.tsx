// ui/src/pages/simulate/shared/subresource-editor.tsx
import { useState, useEffect } from 'react';
import { Box, Group, Text, TextInput, ActionIcon } from '@mantine/core';
import { IconFolder, IconPencil, IconCheck, IconX } from '@tabler/icons-react';
import {
  getSubresourceConfig,
  getBaseArn,
  getSubresourcePath,
  buildArnWithPath,
} from './subresource-config';

interface SubresourceEditorProps {
  arn: string;
  onArnChange: (newArn: string) => void;
}

export function SubresourceEditor({ arn, onArnChange }: SubresourceEditorProps): JSX.Element | null {
  const config = getSubresourceConfig(arn);
  const [isEditing, setIsEditing] = useState(false);
  const [editPath, setEditPath] = useState('');

  // Update editPath when arn changes externally
  useEffect(() => {
    if (config) {
      setEditPath(getSubresourcePath(arn, config));
    }
  }, [arn, config]);

  if (!config) return null;

  const baseArn = getBaseArn(arn, config);
  const currentPath = getSubresourcePath(arn, config);

  const handleStartEdit = (): void => {
    setEditPath(currentPath);
    setIsEditing(true);
  };

  const handleSave = (): void => {
    const trimmedPath = editPath.trim();
    if (trimmedPath && trimmedPath !== currentPath) {
      onArnChange(buildArnWithPath(baseArn, trimmedPath));
    }
    setIsEditing(false);
  };

  const handleCancel = (): void => {
    setEditPath(currentPath);
    setIsEditing(false);
  };

  const handleKeyDown = (e: React.KeyboardEvent): void => {
    if (e.key === 'Enter') {
      handleSave();
    } else if (e.key === 'Escape') {
      handleCancel();
    }
  };

  return (
    <Box
      p="xs"
      mt="xs"
      style={{
        backgroundColor: 'var(--mantine-color-gray-0)',
        borderRadius: 'var(--mantine-radius-sm)',
        border: '1px solid var(--mantine-color-gray-3)',
      }}
    >
      <Group gap="xs" wrap="nowrap">
        <IconFolder size={14} color="var(--mantine-color-dimmed)" />
        <Text size="xs" c="dimmed" style={{ flexShrink: 0 }}>
          {config.label}:
        </Text>
        {isEditing ? (
          <>
            <TextInput
              size="xs"
              value={editPath}
              onChange={(e) => setEditPath(e.currentTarget.value)}
              onKeyDown={handleKeyDown}
              placeholder={config.defaultPath}
              style={{ flex: 1 }}
              autoFocus
            />
            <ActionIcon size="xs" variant="subtle" color="green" onClick={handleSave}>
              <IconCheck size={14} />
            </ActionIcon>
            <ActionIcon size="xs" variant="subtle" color="gray" onClick={handleCancel}>
              <IconX size={14} />
            </ActionIcon>
          </>
        ) : (
          <>
            <Text size="xs" ff="monospace" style={{ flex: 1 }} truncate>
              {currentPath}
            </Text>
            <ActionIcon
              size="xs"
              variant="subtle"
              color="gray"
              onClick={handleStartEdit}
              title="Edit object path"
            >
              <IconPencil size={14} />
            </ActionIcon>
          </>
        )}
      </Group>
    </Box>
  );
}
