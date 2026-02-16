// ui/src/pages/simulate/shared/arn-editor.tsx
// A text input for editing ARNs for resource creation actions

import { useEffect, useState } from 'react';
import { Box, Input, Text, TextInput, Tooltip } from '@mantine/core';
import { IconPencil } from '@tabler/icons-react';
import { yamsApi } from '../../../lib/api';
import type { Action } from '../../../lib/api';
import {
  isResourceCreationAction,
  getDefaultArnForAction,
  PLACEHOLDER_ACCOUNT_ID,
} from './resource-creation';

export interface ArnEditorProps {
  action: string;
  value: string | null;
  onChange: (value: string | null) => void;
  accountId?: string; // Account ID from the selected principal (if available)
  label?: string;
}

export function ArnEditor({
  action,
  value,
  onChange,
  accountId,
  label = 'Resource ARN',
}: ArnEditorProps): JSX.Element {
  const [actionDetails, setActionDetails] = useState<Action | null>(null);
  const [loading, setLoading] = useState(false);
  const [initialized, setInitialized] = useState(false);

  // Fetch action details when action changes
  useEffect(() => {
    if (!action || !isResourceCreationAction(action)) {
      setActionDetails(null);
      setInitialized(false);
      return;
    }

    setLoading(true);
    yamsApi
      .getAction(action)
      .then((details) => {
        setActionDetails(details);
        setInitialized(false); // Reset so we can compute new default
      })
      .catch((err) => {
        console.error('Failed to fetch action details:', err);
        setActionDetails(null);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [action]);

  // Set default ARN when action details are loaded (only once per action)
  useEffect(() => {
    if (!actionDetails || initialized) {
      return;
    }

    const effectiveAccountId = accountId || PLACEHOLDER_ACCOUNT_ID;
    const defaultArn = getDefaultArnForAction(actionDetails, effectiveAccountId);

    if (defaultArn) {
      onChange(defaultArn);
      setInitialized(true);
    }
  }, [actionDetails, accountId, initialized, onChange]);

  // Update ARN with new account ID when it changes (but keep user edits)
  useEffect(() => {
    if (!actionDetails || !value || !accountId) {
      return;
    }

    // Only update if the current value still has the placeholder account
    if (value.includes(PLACEHOLDER_ACCOUNT_ID)) {
      const updatedArn = value.replace(PLACEHOLDER_ACCOUNT_ID, accountId);
      onChange(updatedArn);
    }
  }, [accountId, actionDetails, value, onChange]);

  if (loading) {
    return (
      <Input.Wrapper label={label}>
        <Box
          style={{
            border: '1px solid var(--mantine-color-gray-3)',
            borderRadius: 'var(--mantine-radius-sm)',
            padding: '12px',
            height: '58px',
            display: 'flex',
            alignItems: 'center',
            backgroundColor: 'var(--mantine-color-gray-0)',
          }}
        >
          <Text size="sm" c="dimmed">
            Loading ARN format...
          </Text>
        </Box>
      </Input.Wrapper>
    );
  }

  return (
    <Tooltip
      label="This action creates new resources. Enter the ARN of the resource to be created."
      multiline
      maw={300}
      openDelay={500}
    >
      <TextInput
        label={label}
        placeholder="arn:aws:service:region:account:resource"
        value={value || ''}
        onChange={(e) => onChange(e.currentTarget.value || null)}
        leftSection={<IconPencil size={16} color="var(--mantine-color-violet-6)" />}
        inputWrapperOrder={['label', 'input', 'description', 'error']}
        styles={{
          input: {
            height: '58px',
            fontFamily: 'var(--mantine-font-family-monospace)',
            fontSize: 'var(--mantine-font-size-sm)',
          },
        }}
        description={
          <Text size="xs" c="violet.6">
            Resource creation action — enter the ARN for the new resource
          </Text>
        }
      />
    </Tooltip>
  );
}
