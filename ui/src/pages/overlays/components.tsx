import { useCallback, useEffect, useState } from 'react';
import {
  ActionIcon,
  Badge,
  Box,
  Button,
  Group,
  ScrollArea,
  Stack,
  Text,
  TextInput,
  Tooltip,
} from '@mantine/core';
import {
  IconPlus,
  IconTrash,
  IconDeviceFloppy,
  IconUser,
  IconDatabase,
  IconShield,
  IconBuilding,
} from '@tabler/icons-react';
import CodeMirror from '@uiw/react-codemirror';
import { json } from '@codemirror/lang-json';
import type { Principal, Resource, Policy, Account, PolicyDocument, PrincipalTag, ResourceTag } from '../../lib/api';
import { CopyButton } from '../../components';

// Helper functions for ARN parsing
function extractAccountId(arn: string): string | null {
  const parts = arn.split(':');
  if (parts.length >= 5 && parts[4]) {
    return parts[4];
  }
  return null;
}

function extractService(arn: string): string | null {
  const parts = arn.split(':');
  if (parts.length >= 3 && parts[2]) {
    return parts[2];
  }
  return null;
}

function extractName(arn: string): string {
  const parts = arn.split('/');
  return parts[parts.length - 1] || arn;
}

// Entity type icons
const entityIcons = {
  principal: IconUser,
  resource: IconDatabase,
  policy: IconShield,
  account: IconBuilding,
};

// Entity type colors
const entityColors = {
  principal: 'violet',
  resource: 'blue',
  policy: 'green',
  account: 'orange',
};

// Get entity identifier (ARN or ID)
type EntityType = 'principal' | 'resource' | 'policy' | 'account';

function getEntityId(entity: Principal | Resource | Policy | Account, type: EntityType): string {
  if (type === 'account') {
    return (entity as Account).Id;
  }
  return (entity as Principal | Resource | Policy).Arn;
}

function getEntityName(entity: Principal | Resource | Policy | Account, type: EntityType): string {
  if (type === 'account') {
    return (entity as Account).Name;
  }
  return extractName((entity as Principal | Resource | Policy).Arn);
}

// Props for EntityListItem
interface EntityListItemProps {
  entity: Principal | Resource | Policy | Account;
  type: EntityType;
  isSelected: boolean;
  onClick: () => void;
  onDelete: () => void;
  accountNames?: Record<string, string>;
}

export function EntityListItem({
  entity,
  type,
  isSelected,
  onClick,
  onDelete,
  accountNames,
}: EntityListItemProps): JSX.Element {
  const Icon = entityIcons[type];
  const color = entityColors[type];
  const id = getEntityId(entity, type);
  const name = getEntityName(entity, type);

  // Get account info for display
  let accountId: string | null = null;
  let accountName: string | null = null;
  let service: string | null = null;

  if (type === 'account') {
    accountId = (entity as Account).Id;
    accountName = (entity as Account).Name;
  } else {
    const arn = (entity as Principal | Resource | Policy).Arn;
    accountId = extractAccountId(arn);
    accountName = accountId && accountNames ? accountNames[accountId] : null;
    service = extractService(arn);
  }

  return (
    <Group
      gap="sm"
      p="sm"
      wrap="nowrap"
      onClick={onClick}
      style={{
        cursor: 'pointer',
        borderRadius: 'var(--mantine-radius-sm)',
        backgroundColor: isSelected ? `var(--mantine-color-${color}-0)` : undefined,
        border: isSelected
          ? `1px solid var(--mantine-color-${color}-4)`
          : '1px solid var(--mantine-color-gray-2)',
      }}
    >
      <Icon size={18} color={`var(--mantine-color-${color}-6)`} style={{ flexShrink: 0 }} />
      <Box style={{ flex: 1, minWidth: 0 }}>
        <Text size="sm" fw={500} truncate>
          {name}
        </Text>
        <Group gap="xs">
          {accountName && (
            <Text size="xs" c="dimmed">
              {accountName} ({accountId})
            </Text>
          )}
          {!accountName && accountId && (
            <Text size="xs" c="dimmed">{accountId}</Text>
          )}
          {service && (
            <>
              {accountId && <Text size="xs" c="dimmed">·</Text>}
              <Text size="xs" c="dimmed">{service}</Text>
            </>
          )}
        </Group>
        {type !== 'account' && (
          <Text size="xs" c="dimmed" ff="monospace" truncate>
            {id}
          </Text>
        )}
      </Box>
      <ActionIcon
        variant="subtle"
        color="red"
        size="sm"
        onClick={(e) => {
          e.stopPropagation();
          onDelete();
        }}
      >
        <IconTrash size={14} />
      </ActionIcon>
    </Group>
  );
}

// Tag editor for Principal and Resource tags
interface TagEditorProps {
  tags: PrincipalTag[] | ResourceTag[] | undefined;
  onChange: (tags: PrincipalTag[] | ResourceTag[]) => void;
  disabled?: boolean;
}

export function TagEditor({ tags, onChange, disabled }: TagEditorProps): JSX.Element {
  const tagList = tags || [];

  const addTag = (): void => {
    onChange([...tagList, { Key: '', Value: '' }]);
  };

  const removeTag = (index: number): void => {
    onChange(tagList.filter((_, i) => i !== index));
  };

  const updateTag = (index: number, field: 'Key' | 'Value', value: string): void => {
    onChange(tagList.map((tag, i) => (i === index ? { ...tag, [field]: value } : tag)));
  };

  return (
    <Stack gap="xs">
      {tagList.map((tag, idx) => (
        <Group key={idx} gap="xs" wrap="nowrap">
          <TextInput
            placeholder="Key"
            value={tag.Key}
            onChange={(e) => updateTag(idx, 'Key', e.currentTarget.value)}
            style={{ flex: 1 }}
            size="sm"
            disabled={disabled}
          />
          <TextInput
            placeholder="Value"
            value={tag.Value}
            onChange={(e) => updateTag(idx, 'Value', e.currentTarget.value)}
            style={{ flex: 1 }}
            size="sm"
            disabled={disabled}
          />
          <ActionIcon
            variant="subtle"
            color="red"
            onClick={() => removeTag(idx)}
            disabled={disabled}
          >
            <IconTrash size={14} />
          </ActionIcon>
        </Group>
      ))}
      <Button
        variant="subtle"
        size="xs"
        leftSection={<IconPlus size={14} />}
        onClick={addTag}
        disabled={disabled}
        style={{ alignSelf: 'flex-start' }}
      >
        Add Tag
      </Button>
    </Stack>
  );
}

// String list editor for AttachedPolicies, Groups, etc.
interface StringListEditorProps {
  label: string;
  values: string[] | undefined | null;
  onChange: (values: string[]) => void;
  placeholder?: string;
  disabled?: boolean;
}

export function StringListEditor({
  label,
  values,
  onChange,
  placeholder,
  disabled,
}: StringListEditorProps): JSX.Element {
  const list = values || [];

  const addItem = (): void => {
    onChange([...list, '']);
  };

  const removeItem = (index: number): void => {
    onChange(list.filter((_, i) => i !== index));
  };

  const updateItem = (index: number, value: string): void => {
    onChange(list.map((v, i) => (i === index ? value : v)));
  };

  return (
    <Stack gap="xs">
      {list.map((value, idx) => (
        <Group key={idx} gap="xs" wrap="nowrap">
          <TextInput
            placeholder={placeholder}
            value={value}
            onChange={(e) => updateItem(idx, e.currentTarget.value)}
            style={{ flex: 1 }}
            size="sm"
            disabled={disabled}
          />
          <ActionIcon
            variant="subtle"
            color="red"
            onClick={() => removeItem(idx)}
            disabled={disabled}
          >
            <IconTrash size={14} />
          </ActionIcon>
        </Group>
      ))}
      <Button
        variant="subtle"
        size="xs"
        leftSection={<IconPlus size={14} />}
        onClick={addItem}
        disabled={disabled}
        style={{ alignSelf: 'flex-start' }}
      >
        Add {label}
      </Button>
    </Stack>
  );
}

// JSON/Policy editor using CodeMirror
interface PolicyEditorProps {
  value: PolicyDocument | undefined;
  onChange: (value: PolicyDocument | undefined) => void;
  error?: string;
  disabled?: boolean;
  height?: string;
}

export function PolicyEditor({
  value,
  onChange,
  error,
  disabled,
  height = '300px',
}: PolicyEditorProps): JSX.Element {
  const [jsonText, setJsonText] = useState(value ? JSON.stringify(value, null, 2) : '');
  const [parseError, setParseError] = useState<string | null>(null);

  // Sync external value changes
  useEffect(() => {
    setJsonText(value ? JSON.stringify(value, null, 2) : '');
    setParseError(null);
  }, [value]);

  const handleChange = useCallback((val: string) => {
    setJsonText(val);
    if (!val.trim()) {
      setParseError(null);
      onChange(undefined);
      return;
    }
    try {
      const parsed = JSON.parse(val);
      setParseError(null);
      onChange(parsed);
    } catch {
      setParseError('Invalid JSON');
    }
  }, [onChange]);

  return (
    <Stack gap="xs">
      <Box
        style={{
          border: `1px solid var(--mantine-color-${parseError || error ? 'red' : 'gray'}-4)`,
          borderRadius: 'var(--mantine-radius-sm)',
          overflow: 'hidden',
        }}
      >
        <CodeMirror
          value={jsonText}
          height={height}
          extensions={[json()]}
          onChange={handleChange}
          editable={!disabled}
          basicSetup={{
            lineNumbers: true,
            foldGutter: true,
            bracketMatching: true,
            autocompletion: false,
          }}
        />
      </Box>
      {(parseError || error) && (
        <Text size="xs" c="red">
          {parseError || error}
        </Text>
      )}
    </Stack>
  );
}

// Inline policy array editor
interface InlinePoliciesEditorProps {
  policies: PolicyDocument[] | undefined;
  onChange: (policies: PolicyDocument[]) => void;
  disabled?: boolean;
}

export function InlinePoliciesEditor({
  policies,
  onChange,
  disabled,
}: InlinePoliciesEditorProps): JSX.Element {
  const list = policies || [];

  const addPolicy = (): void => {
    const newPolicy: PolicyDocument = {
      Version: '2012-10-17',
      Statement: [],
    };
    onChange([...list, newPolicy]);
  };

  const removePolicy = (index: number): void => {
    onChange(list.filter((_, i) => i !== index));
  };

  const updatePolicy = (index: number, policy: PolicyDocument | undefined): void => {
    if (policy) {
      onChange(list.map((p, i) => (i === index ? policy : p)));
    }
  };

  return (
    <Stack gap="md">
      {list.map((policy, idx) => (
        <Box key={idx}>
          <Group justify="space-between" mb="xs">
            <Text size="sm" fw={500}>
              {policy._Name || `Policy ${idx + 1}`}
            </Text>
            <ActionIcon
              variant="subtle"
              color="red"
              onClick={() => removePolicy(idx)}
              disabled={disabled}
            >
              <IconTrash size={14} />
            </ActionIcon>
          </Group>
          <PolicyEditor
            value={policy}
            onChange={(val) => updatePolicy(idx, val)}
            disabled={disabled}
            height="200px"
          />
        </Box>
      ))}
      <Button
        variant="subtle"
        size="xs"
        leftSection={<IconPlus size={14} />}
        onClick={addPolicy}
        disabled={disabled}
        style={{ alignSelf: 'flex-start' }}
      >
        Add Inline Policy
      </Button>
    </Stack>
  );
}

// Detail panel for editing entities
interface EntityDetailPanelProps {
  entity: Principal | Resource | Policy | Account | null;
  type: EntityType | null;
  onChange: (entity: Principal | Resource | Policy | Account) => void;
  onSave: () => void;
  isDirty: boolean;
}

export function EntityDetailPanel({
  entity,
  type,
  onChange,
  onSave,
  isDirty,
}: EntityDetailPanelProps): JSX.Element {
  if (!entity || !type) {
    return (
      <Box
        h="100%"
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <Text c="dimmed" ta="center">
          Select an entity from the list to view and edit its details
        </Text>
      </Box>
    );
  }

  const Icon = entityIcons[type];
  const color = entityColors[type];
  const id = getEntityId(entity, type);
  const name = getEntityName(entity, type);

  // Render fields based on entity type
  const renderFields = (): JSX.Element => {
    switch (type) {
      case 'principal':
        return <PrincipalFields entity={entity as Principal} onChange={onChange} />;
      case 'resource':
        return <ResourceFields entity={entity as Resource} onChange={onChange} />;
      case 'policy':
        return <PolicyFields entity={entity as Policy} onChange={onChange} />;
      case 'account':
        return <AccountFields entity={entity as Account} onChange={onChange} />;
      default:
        return <Text c="dimmed">Unknown entity type</Text>;
    }
  };

  return (
    <Stack gap="md" h="100%">
      {/* Header */}
      <Group justify="space-between" wrap="nowrap">
        <Group gap="sm" wrap="nowrap" style={{ minWidth: 0 }}>
          <Icon size={20} color={`var(--mantine-color-${color}-6)`} />
          <Box style={{ minWidth: 0 }}>
            <Text size="lg" fw={600} truncate>
              {name}
            </Text>
            <Badge size="xs" color={color} variant="light">
              {type}
            </Badge>
          </Box>
        </Group>
        <Group gap="xs" wrap="nowrap">
          <CopyButton value={id} />
          <Tooltip label={isDirty ? 'Save changes' : 'No changes to save'}>
            <Button
              size="sm"
              leftSection={<IconDeviceFloppy size={16} />}
              disabled={!isDirty}
              onClick={onSave}
            >
              Save
            </Button>
          </Tooltip>
        </Group>
      </Group>

      {/* Fields */}
      <ScrollArea style={{ flex: 1 }}>
        <Stack gap="md" pr="sm">
          {renderFields()}
        </Stack>
      </ScrollArea>
    </Stack>
  );
}

// Principal-specific fields
interface PrincipalFieldsProps {
  entity: Principal;
  onChange: (entity: Principal) => void;
}

function PrincipalFields({ entity, onChange }: PrincipalFieldsProps): JSX.Element {
  const update = <K extends keyof Principal>(field: K, value: Principal[K]): void => {
    onChange({ ...entity, [field]: value });
  };

  return (
    <>
      <TextInput
        label="ARN"
        value={entity.Arn}
        onChange={(e) => update('Arn', e.currentTarget.value)}
        ff="monospace"
        size="sm"
      />
      <TextInput
        label="Name"
        value={entity.Name}
        onChange={(e) => update('Name', e.currentTarget.value)}
        size="sm"
      />
      <Group grow>
        <TextInput
          label="Type"
          value={entity.Type}
          readOnly
          size="sm"
          styles={{ input: { backgroundColor: 'var(--mantine-color-gray-1)' } }}
        />
        <TextInput
          label="Account ID"
          value={entity.AccountId}
          readOnly
          size="sm"
          styles={{ input: { backgroundColor: 'var(--mantine-color-gray-1)' } }}
        />
      </Group>
      <TextInput
        label="Permissions Boundary"
        value={entity.PermissionsBoundary || ''}
        onChange={(e) => update('PermissionsBoundary', e.currentTarget.value || undefined)}
        placeholder="arn:aws:iam::..."
        ff="monospace"
        size="sm"
      />
      <Box>
        <Text size="sm" fw={500} mb="xs">Tags</Text>
        <TagEditor
          tags={entity.Tags}
          onChange={(tags) => update('Tags', tags as PrincipalTag[])}
        />
      </Box>
      <Box>
        <Text size="sm" fw={500} mb="xs">Attached Policies</Text>
        <StringListEditor
          label="Policy"
          values={entity.AttachedPolicies}
          onChange={(values) => update('AttachedPolicies', values)}
          placeholder="arn:aws:iam::..."
        />
      </Box>
      <Box>
        <Text size="sm" fw={500} mb="xs">Groups</Text>
        <StringListEditor
          label="Group"
          values={entity.Groups}
          onChange={(values) => update('Groups', values)}
          placeholder="group-name"
        />
      </Box>
      <Box>
        <Text size="sm" fw={500} mb="xs">Inline Policies</Text>
        <InlinePoliciesEditor
          policies={entity.InlinePolicies}
          onChange={(policies) => update('InlinePolicies', policies)}
        />
      </Box>
    </>
  );
}

// Resource-specific fields
interface ResourceFieldsProps {
  entity: Resource;
  onChange: (entity: Resource) => void;
}

function ResourceFields({ entity, onChange }: ResourceFieldsProps): JSX.Element {
  const update = <K extends keyof Resource>(field: K, value: Resource[K]): void => {
    onChange({ ...entity, [field]: value });
  };

  return (
    <>
      <TextInput
        label="ARN"
        value={entity.Arn}
        onChange={(e) => update('Arn', e.currentTarget.value)}
        ff="monospace"
        size="sm"
      />
      <TextInput
        label="Name"
        value={entity.Name}
        onChange={(e) => update('Name', e.currentTarget.value)}
        size="sm"
      />
      <Group grow>
        <TextInput
          label="Type"
          value={entity.Type}
          onChange={(e) => update('Type', e.currentTarget.value)}
          size="sm"
        />
        <TextInput
          label="Region"
          value={entity.Region}
          onChange={(e) => update('Region', e.currentTarget.value)}
          size="sm"
        />
      </Group>
      <TextInput
        label="Account ID"
        value={entity.AccountId}
        readOnly
        size="sm"
        styles={{ input: { backgroundColor: 'var(--mantine-color-gray-1)' } }}
      />
      <Box>
        <Text size="sm" fw={500} mb="xs">Tags</Text>
        <TagEditor
          tags={entity.Tags}
          onChange={(tags) => update('Tags', tags as ResourceTag[])}
        />
      </Box>
      <Box>
        <Text size="sm" fw={500} mb="xs">Resource Policy</Text>
        <PolicyEditor
          value={entity.Policy}
          onChange={(policy) => update('Policy', policy)}
          height="300px"
        />
      </Box>
    </>
  );
}

// Policy-specific fields
interface PolicyFieldsProps {
  entity: Policy;
  onChange: (entity: Policy) => void;
}

function PolicyFields({ entity, onChange }: PolicyFieldsProps): JSX.Element {
  const update = <K extends keyof Policy>(field: K, value: Policy[K]): void => {
    onChange({ ...entity, [field]: value });
  };

  return (
    <>
      <TextInput
        label="ARN"
        value={entity.Arn}
        onChange={(e) => update('Arn', e.currentTarget.value)}
        ff="monospace"
        size="sm"
      />
      <TextInput
        label="Name"
        value={entity.Name}
        onChange={(e) => update('Name', e.currentTarget.value)}
        size="sm"
      />
      <Group grow>
        <TextInput
          label="Type"
          value={entity.Type}
          readOnly
          size="sm"
          styles={{ input: { backgroundColor: 'var(--mantine-color-gray-1)' } }}
        />
        <TextInput
          label="Account ID"
          value={entity.AccountId}
          readOnly
          size="sm"
          styles={{ input: { backgroundColor: 'var(--mantine-color-gray-1)' } }}
        />
      </Group>
      <Box>
        <Text size="sm" fw={500} mb="xs">Policy Document</Text>
        <PolicyEditor
          value={entity.Policy}
          onChange={(policy) => update('Policy', policy!)}
          height="400px"
        />
      </Box>
    </>
  );
}

// Account-specific fields
interface AccountFieldsProps {
  entity: Account;
  onChange: (entity: Account) => void;
}

function AccountFields({ entity, onChange }: AccountFieldsProps): JSX.Element {
  const update = <K extends keyof Account>(field: K, value: Account[K]): void => {
    onChange({ ...entity, [field]: value });
  };

  return (
    <>
      <TextInput
        label="Account ID"
        value={entity.Id}
        readOnly
        ff="monospace"
        size="sm"
        styles={{ input: { backgroundColor: 'var(--mantine-color-gray-1)' } }}
      />
      <TextInput
        label="Name"
        value={entity.Name}
        onChange={(e) => update('Name', e.currentTarget.value)}
        size="sm"
      />
      <TextInput
        label="Organization ID"
        value={entity.OrgId || ''}
        readOnly
        size="sm"
        styles={{ input: { backgroundColor: 'var(--mantine-color-gray-1)' } }}
      />
      {entity.OrgPaths && entity.OrgPaths.length > 0 && (
        <Box>
          <Text size="sm" fw={500} mb="xs">Organization Paths</Text>
          <Stack gap="xs">
            {entity.OrgPaths.map((path, idx) => (
              <Text key={idx} size="xs" ff="monospace" c="dimmed">
                {path}
              </Text>
            ))}
          </Stack>
        </Box>
      )}
      {entity.OrgNodes && entity.OrgNodes.length > 0 && (
        <Box>
          <Text size="sm" fw={500} mb="xs">Organization Nodes</Text>
          <Stack gap="xs">
            {entity.OrgNodes.map((node, idx) => (
              <Box key={idx} p="xs" style={{ backgroundColor: 'var(--mantine-color-gray-0)', borderRadius: 'var(--mantine-radius-sm)' }}>
                <Text size="sm" fw={500}>{node.Name}</Text>
                <Text size="xs" c="dimmed">{node.Type} · {node.Id}</Text>
              </Box>
            ))}
          </Stack>
        </Box>
      )}
    </>
  );
}
