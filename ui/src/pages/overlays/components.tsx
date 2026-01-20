import { useCallback, useEffect, useState } from 'react';
import {
  ActionIcon,
  Badge,
  Box,
  Button,
  Combobox,
  Divider,
  Group,
  InputBase,
  ScrollArea,
  Stack,
  Text,
  TextInput,
  Title,
  Tooltip,
  useCombobox,
} from '@mantine/core';
import { useDebouncedValue } from '@mantine/hooks';
import {
  IconPlus,
  IconTrash,
  IconDeviceFloppy,
  IconUser,
  IconDatabase,
  IconShield,
  IconBuilding,
  IconSearch,
  IconX,
} from '@tabler/icons-react';
import CodeMirror from '@uiw/react-codemirror';
import { json } from '@codemirror/lang-json';
import type { Principal, Resource, Policy, Account, PolicyDocument, PrincipalTag, ResourceTag } from '../../lib/api';
import { yamsApi } from '../../lib/api';
import { CopyButton } from '../../components';

// Section heading component for consistent styling
interface SectionHeadingProps {
  title: string;
  children?: React.ReactNode;
}

function SectionHeading({ title, children }: SectionHeadingProps): JSX.Element {
  return (
    <Box
      p="sm"
      style={{
        backgroundColor: 'var(--mantine-color-gray-1)',
        borderRadius: 'var(--mantine-radius-sm)',
        borderLeft: '3px solid var(--mantine-color-violet-5)',
      }}
    >
      <Group justify="space-between" wrap="nowrap">
        <Title order={6} fw={600} c="dark">
          {title}
        </Title>
        {children}
      </Group>
    </Box>
  );
}

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

// String list editor for Groups, etc.
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

// Helper to extract account ID from policy ARN
function extractPolicyAccountId(arn: string): string | null {
  // AWS managed policies: arn:aws:iam::aws:policy/...
  if (arn.includes(':aws:policy/')) {
    return 'aws';
  }
  // Customer managed policies: arn:aws:iam::ACCOUNT_ID:policy/...
  const parts = arn.split(':');
  if (parts.length >= 5 && parts[4]) {
    return parts[4];
  }
  return null;
}

// Policy selector for AttachedPolicies and PermissionsBoundary
interface PolicySelectorProps {
  value: string | null;
  onChange: (value: string | null) => void;
  overlayPolicies: string[];
  accountId: string;
  placeholder?: string;
  disabled?: boolean;
}

function PolicySelector({
  value,
  onChange,
  overlayPolicies,
  accountId,
  placeholder = 'Search policies...',
  disabled,
}: PolicySelectorProps): JSX.Element {
  const combobox = useCombobox({
    onDropdownClose: () => combobox.resetSelectedOption(),
  });
  const [search, setSearch] = useState('');
  const [debouncedSearch] = useDebouncedValue(search, 200);
  const [universePolicies, setUniversePolicies] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);

  // Filter to only include policies from the same account or AWS managed policies
  const filterByAccount = useCallback((policies: string[]): string[] => {
    return policies.filter((arn) => {
      const policyAccountId = extractPolicyAccountId(arn);
      return policyAccountId === accountId || policyAccountId === 'aws';
    });
  }, [accountId]);

  // Fetch policies from universe when searching
  useEffect(() => {
    if (!debouncedSearch) {
      setUniversePolicies([]);
      return;
    }
    setLoading(true);
    yamsApi.searchPolicies(debouncedSearch)
      .then((results) => setUniversePolicies(filterByAccount(results).slice(0, 20)))
      .catch(() => setUniversePolicies([]))
      .finally(() => setLoading(false));
  }, [debouncedSearch, filterByAccount]);

  // Combine overlay policies and universe policies, deduplicated, filtered by account
  const sameAccountOverlayPolicies = filterByAccount(overlayPolicies);
  const allPolicies = Array.from(new Set([...sameAccountOverlayPolicies, ...universePolicies]));
  const filteredPolicies = debouncedSearch
    ? allPolicies.filter((p) => p.toLowerCase().includes(debouncedSearch.toLowerCase()))
    : sameAccountOverlayPolicies;

  const handleSelect = (arn: string): void => {
    onChange(arn);
    setSearch('');
    combobox.closeDropdown();
  };

  const handleClear = (): void => {
    onChange(null);
    setSearch('');
  };

  return (
    <Combobox
      store={combobox}
      onOptionSubmit={handleSelect}
      styles={{
        dropdown: {
          border: '2px solid var(--mantine-color-gray-4)',
          boxShadow: 'var(--mantine-shadow-md)',
        },
      }}
    >
      <Combobox.Target>
        {value ? (
          <Group
            gap="xs"
            justify="space-between"
            wrap="nowrap"
            style={{
              border: '1px solid var(--mantine-color-gray-4)',
              borderRadius: 'var(--mantine-radius-sm)',
              padding: '6px 10px',
              minHeight: '34px',
              cursor: disabled ? 'not-allowed' : 'pointer',
              backgroundColor: disabled ? 'var(--mantine-color-gray-1)' : undefined,
            }}
            onClick={() => !disabled && combobox.openDropdown()}
          >
            <Text size="sm" ff="monospace" truncate style={{ flex: 1 }}>
              {extractName(value)}
            </Text>
            {!disabled && (
              <ActionIcon
                size="xs"
                variant="subtle"
                color="gray"
                onClick={(e) => {
                  e.stopPropagation();
                  handleClear();
                }}
              >
                <IconX size={12} />
              </ActionIcon>
            )}
          </Group>
        ) : (
          <InputBase
            value={search}
            onChange={(e) => {
              setSearch(e.currentTarget.value);
              combobox.openDropdown();
            }}
            onClick={() => combobox.openDropdown()}
            onFocus={() => combobox.openDropdown()}
            onBlur={() => setTimeout(() => combobox.closeDropdown(), 150)}
            placeholder={placeholder}
            rightSection={loading ? null : <IconSearch size={14} color="var(--mantine-color-dimmed)" />}
            size="sm"
            disabled={disabled}
          />
        )}
      </Combobox.Target>

      <Combobox.Dropdown>
        <Combobox.Options>
          <ScrollArea.Autosize mah={200} type="scroll">
            {filteredPolicies.length === 0 ? (
              <Combobox.Empty>
                {loading ? 'Searching...' : debouncedSearch ? 'No policies found' : 'Type to search'}
              </Combobox.Empty>
            ) : (
              filteredPolicies.map((arn) => (
                <Combobox.Option value={arn} key={arn}>
                  <Box>
                    <Text size="sm" fw={500}>{extractName(arn)}</Text>
                    <Text size="xs" c="dimmed" ff="monospace" truncate>
                      {arn}
                    </Text>
                    {overlayPolicies.includes(arn) && (
                      <Badge size="xs" variant="light" color="violet" mt={2}>
                        in overlay
                      </Badge>
                    )}
                  </Box>
                </Combobox.Option>
              ))
            )}
          </ScrollArea.Autosize>
        </Combobox.Options>
      </Combobox.Dropdown>
    </Combobox>
  );
}

// Multi-policy selector for AttachedPolicies
interface MultiPolicySelectorProps {
  values: string[] | undefined | null;
  onChange: (values: string[]) => void;
  overlayPolicies: string[];
  accountId: string;
  disabled?: boolean;
}

function MultiPolicySelector({
  values,
  onChange,
  overlayPolicies,
  accountId,
  disabled,
}: MultiPolicySelectorProps): JSX.Element {
  const list = values || [];

  const addPolicy = (arn: string): void => {
    if (!list.includes(arn)) {
      onChange([...list, arn]);
    }
  };

  const removePolicy = (index: number): void => {
    onChange(list.filter((_, i) => i !== index));
  };

  return (
    <Stack gap="xs">
      {list.map((arn, idx) => (
        <Group key={arn} gap="xs" wrap="nowrap" p="xs" style={{
          border: '1px solid var(--mantine-color-gray-3)',
          borderRadius: 'var(--mantine-radius-sm)',
          backgroundColor: 'var(--mantine-color-gray-0)',
        }}>
          <Box style={{ flex: 1, minWidth: 0 }}>
            <Text size="sm" fw={500} truncate>{extractName(arn)}</Text>
            <Text size="xs" c="dimmed" ff="monospace" truncate>{arn}</Text>
          </Box>
          <ActionIcon
            variant="subtle"
            color="red"
            size="sm"
            onClick={() => removePolicy(idx)}
            disabled={disabled}
          >
            <IconTrash size={14} />
          </ActionIcon>
        </Group>
      ))}
      <PolicySelector
        value={null}
        onChange={(arn) => arn && addPolicy(arn)}
        overlayPolicies={overlayPolicies.filter((p) => !list.includes(p))}
        accountId={accountId}
        placeholder="Add policy..."
        disabled={disabled}
      />
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
  overlayPolicies?: string[];
}

export function EntityDetailPanel({
  entity,
  type,
  onChange,
  onSave,
  isDirty,
  overlayPolicies = [],
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
        return <PrincipalFields entity={entity as Principal} onChange={onChange} overlayPolicies={overlayPolicies} />;
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

      <Divider />

      {/* Fields */}
      <ScrollArea style={{ flex: 1 }}>
        <Stack gap="lg" pr="sm">
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
  overlayPolicies: string[];
}

function PrincipalFields({ entity, onChange, overlayPolicies }: PrincipalFieldsProps): JSX.Element {
  const update = <K extends keyof Principal>(field: K, value: Principal[K]): void => {
    onChange({ ...entity, [field]: value });
  };

  return (
    <>
      {/* Identity Section */}
      <Box>
        <SectionHeading title="Identity" />
        <Stack gap="sm" mt="sm">
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
        </Stack>
      </Box>

      {/* Permissions Boundary Section */}
      <Box>
        <SectionHeading title="Permissions Boundary" />
        <Box mt="sm">
          <PolicySelector
            value={entity.PermissionsBoundary || null}
            onChange={(val) => update('PermissionsBoundary', val || undefined)}
            overlayPolicies={overlayPolicies}
            accountId={entity.AccountId}
            placeholder="Select permissions boundary..."
          />
        </Box>
      </Box>

      {/* Attached Policies Section */}
      <Box>
        <SectionHeading title="Attached Policies" />
        <Box mt="sm">
          <MultiPolicySelector
            values={entity.AttachedPolicies}
            onChange={(values) => update('AttachedPolicies', values)}
            overlayPolicies={overlayPolicies}
            accountId={entity.AccountId}
          />
        </Box>
      </Box>

      {/* Groups Section */}
      <Box>
        <SectionHeading title="Groups" />
        <Box mt="sm">
          <StringListEditor
            label="Group"
            values={entity.Groups}
            onChange={(values) => update('Groups', values)}
            placeholder="group-name"
          />
        </Box>
      </Box>

      {/* Tags Section */}
      <Box>
        <SectionHeading title="Tags" />
        <Box mt="sm">
          <TagEditor
            tags={entity.Tags}
            onChange={(tags) => update('Tags', tags as PrincipalTag[])}
          />
        </Box>
      </Box>

      {/* Inline Policies Section */}
      <Box>
        <SectionHeading title="Inline Policies" />
        <Box mt="sm">
          <InlinePoliciesEditor
            policies={entity.InlinePolicies}
            onChange={(policies) => update('InlinePolicies', policies)}
          />
        </Box>
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
      {/* Identity Section */}
      <Box>
        <SectionHeading title="Identity" />
        <Stack gap="sm" mt="sm">
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
        </Stack>
      </Box>

      {/* Tags Section */}
      <Box>
        <SectionHeading title="Tags" />
        <Box mt="sm">
          <TagEditor
            tags={entity.Tags}
            onChange={(tags) => update('Tags', tags as ResourceTag[])}
          />
        </Box>
      </Box>

      {/* Resource Policy Section */}
      <Box>
        <SectionHeading title="Resource Policy" />
        <Box mt="sm">
          <PolicyEditor
            value={entity.Policy}
            onChange={(policy) => update('Policy', policy)}
            height="300px"
          />
        </Box>
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
      {/* Identity Section */}
      <Box>
        <SectionHeading title="Identity" />
        <Stack gap="sm" mt="sm">
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
        </Stack>
      </Box>

      {/* Policy Document Section */}
      <Box>
        <SectionHeading title="Policy Document" />
        <Box mt="sm">
          <PolicyEditor
            value={entity.Policy}
            onChange={(policy) => update('Policy', policy ?? { Version: '2012-10-17', Statement: [] })}
            height="400px"
          />
        </Box>
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
      {/* Identity Section */}
      <Box>
        <SectionHeading title="Identity" />
        <Stack gap="sm" mt="sm">
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
        </Stack>
      </Box>

      {/* Organization Paths Section */}
      {entity.OrgPaths && entity.OrgPaths.length > 0 && (
        <Box>
          <SectionHeading title="Organization Paths" />
          <Stack gap="xs" mt="sm">
            {entity.OrgPaths.map((path, idx) => (
              <Text key={idx} size="xs" ff="monospace" c="dimmed">
                {path}
              </Text>
            ))}
          </Stack>
        </Box>
      )}

      {/* Organization Nodes Section */}
      {entity.OrgNodes && entity.OrgNodes.length > 0 && (
        <Box>
          <SectionHeading title="Organization Nodes" />
          <Stack gap="xs" mt="sm">
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
