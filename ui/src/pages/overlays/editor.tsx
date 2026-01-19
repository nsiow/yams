import { useCallback, useEffect, useState } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import {
  ActionIcon,
  Anchor,
  Box,
  Breadcrumbs,
  Button,
  Card,
  Checkbox,
  Combobox,
  Grid,
  Group,
  Input,
  InputBase,
  Modal,
  ScrollArea,
  Stack,
  Tabs,
  Text,
  Textarea,
  TextInput,
  Title,
  useCombobox,
} from '@mantine/core';
import { useDebouncedValue } from '@mantine/hooks';
import { useDisclosure } from '@mantine/hooks';
import { CodeHighlight } from '@mantine/code-highlight';
import {
  IconChevronRight,
  IconEdit,
  IconPlus,
  IconTrash,
  IconDeviceFloppy,
  IconUser,
  IconDatabase,
  IconShield,
  IconBuilding,
} from '@tabler/icons-react';
import { yamsApi } from '../../lib/api';
import type { OverlayData, OverlaySummary } from '../../lib/api';
import { IconSearch } from '@tabler/icons-react';
import {
  CopyButton,
  EmptyState,
  DetailSkeleton,
} from '../../components';

import '@mantine/code-highlight/styles.css';

// Extract account ID from ARN (5th segment)
function extractAccountId(arn: string): string | null {
  const parts = arn.split(':');
  if (parts.length >= 5 && parts[4]) {
    return parts[4];
  }
  return null;
}

// Extract service type from ARN (3rd segment)
function extractService(arn: string): string | null {
  const parts = arn.split(':');
  if (parts.length >= 3 && parts[2]) {
    return parts[2];
  }
  return null;
}

// Format display name from ARN
function formatArnLabel(arn: string): string {
  const parts = arn.split('/');
  return parts[parts.length - 1] || arn;
}

interface AddEntityModalProps {
  opened: boolean;
  onClose: () => void;
  entityType: 'principal' | 'resource' | 'policy' | 'account' | 'group';
  onAdd: (arns: string[]) => void;
  existingArns: string[];
}

function AddEntityModal({ opened, onClose, entityType, onAdd, existingArns }: AddEntityModalProps): JSX.Element {
  const [searchQuery, setSearchQuery] = useState('');
  const [results, setResults] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedArns, setSelectedArns] = useState<Set<string>>(new Set());
  const [accountNames, setAccountNames] = useState<Record<string, string>>({});
  const [resourceAccounts, setResourceAccounts] = useState<Record<string, string>>({});

  // Fetch account info on mount
  useEffect(() => {
    yamsApi.accountNames()
      .then(setAccountNames)
      .catch((err) => console.error('Failed to fetch account names:', err));
    yamsApi.resourceAccounts()
      .then(setResourceAccounts)
      .catch((err) => console.error('Failed to fetch resource accounts:', err));
  }, []);

  useEffect(() => {
    if (!opened) {
      setSearchQuery('');
      setResults([]);
      setSelectedArns(new Set());
      return;
    }

    async function fetchResults(): Promise<void> {
      setLoading(true);
      try {
        let data: string[] = [];
        switch (entityType) {
          case 'principal':
            data = searchQuery
              ? await yamsApi.searchPrincipals(searchQuery)
              : await yamsApi.listPrincipals();
            break;
          case 'resource':
            data = searchQuery
              ? await yamsApi.searchResources(searchQuery)
              : await yamsApi.listResources();
            break;
          case 'policy':
            data = searchQuery
              ? await yamsApi.searchPolicies(searchQuery)
              : await yamsApi.listPolicies();
            break;
          case 'account':
            data = searchQuery
              ? await yamsApi.searchAccounts(searchQuery)
              : await yamsApi.listAccounts();
            break;
        }
        // Filter out already added entities
        setResults(data.filter(arn => !existingArns.includes(arn)).slice(0, 50));
      } catch (err) {
        console.error('Failed to fetch results:', err);
      } finally {
        setLoading(false);
      }
    }

    const timer = setTimeout(fetchResults, 200);
    return () => clearTimeout(timer);
  }, [opened, searchQuery, entityType, existingArns]);

  const toggleArn = (arn: string): void => {
    setSelectedArns((prev) => {
      const next = new Set(prev);
      if (next.has(arn)) {
        next.delete(arn);
      } else {
        next.add(arn);
      }
      return next;
    });
  };

  const handleAdd = (): void => {
    if (selectedArns.size > 0) {
      onAdd(Array.from(selectedArns));
      onClose();
    }
  };

  // Get account info for an ARN
  const getAccountInfo = (arn: string): { accountId: string | null; accountName: string | null; service: string | null } => {
    const accountId = extractAccountId(arn) || resourceAccounts[arn] || null;
    const accountName = accountId ? accountNames[accountId] || null : null;
    const service = extractService(arn);
    return { accountId, accountName, service };
  };

  return (
    <Modal opened={opened} onClose={onClose} title={`Add ${entityType}s`} size="lg">
      <Stack gap="md">
        <TextInput
          placeholder={`Search ${entityType}s...`}
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.currentTarget.value)}
          leftSection={<IconSearch size={16} />}
          data-autofocus
        />
        <ScrollArea h={350}>
          {loading ? (
            <Text c="dimmed" ta="center" py="xl">Loading...</Text>
          ) : results.length === 0 ? (
            <Text c="dimmed" ta="center" py="xl">
              {searchQuery ? 'No results found' : `Type to search ${entityType}s`}
            </Text>
          ) : (
            <Stack gap={4}>
              {results.map((arn) => {
                const isSelected = selectedArns.has(arn);
                const { accountId, accountName, service } = getAccountInfo(arn);
                return (
                  <Group
                    key={arn}
                    gap="sm"
                    p="sm"
                    wrap="nowrap"
                    onClick={() => toggleArn(arn)}
                    style={{
                      cursor: 'pointer',
                      borderRadius: 'var(--mantine-radius-sm)',
                      backgroundColor: isSelected ? 'var(--mantine-color-violet-0)' : undefined,
                      border: isSelected ? '1px solid var(--mantine-color-violet-4)' : '1px solid transparent',
                    }}
                  >
                    <Checkbox
                      checked={isSelected}
                      onChange={() => toggleArn(arn)}
                      onClick={(e) => e.stopPropagation()}
                    />
                    <Box style={{ flex: 1, minWidth: 0 }}>
                      <Text size="sm" fw={500} truncate>
                        {formatArnLabel(arn)}
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
                            {(accountName || accountId) && <Text size="xs" c="dimmed">·</Text>}
                            <Text size="xs" c="dimmed">{service}</Text>
                          </>
                        )}
                      </Group>
                      <Text size="xs" c="dimmed" ff="monospace" truncate>
                        {arn}
                      </Text>
                    </Box>
                  </Group>
                );
              })}
            </Stack>
          )}
        </ScrollArea>
        <Group justify="space-between">
          <Text size="sm" c="dimmed">
            {selectedArns.size} selected
          </Text>
          <Group gap="sm">
            <Button variant="default" onClick={onClose}>Cancel</Button>
            <Button onClick={handleAdd} disabled={selectedArns.size === 0}>
              Add {selectedArns.size > 0 ? `(${selectedArns.size})` : ''}
            </Button>
          </Group>
        </Group>
      </Stack>
    </Modal>
  );
}

// Edit entity modal for modifying entity properties
interface EditEntityModalProps {
  opened: boolean;
  onClose: () => void;
  entityType: 'principal' | 'resource' | 'policy' | 'account';
  entity: unknown | null;
  onSave: (updated: unknown) => void;
}

function EditEntityModal({ opened, onClose, entityType, entity, onSave }: EditEntityModalProps): JSX.Element {
  const [jsonValue, setJsonValue] = useState('');
  const [parseError, setParseError] = useState<string | null>(null);

  // Reset when modal opens with new entity
  useEffect(() => {
    if (opened && entity) {
      setJsonValue(JSON.stringify(entity, null, 2));
      setParseError(null);
    }
  }, [opened, entity]);

  const handleSave = (): void => {
    try {
      const parsed = JSON.parse(jsonValue);
      setParseError(null);
      onSave(parsed);
      onClose();
    } catch {
      setParseError('Invalid JSON format');
    }
  };

  return (
    <Modal opened={opened} onClose={onClose} title={`Edit ${entityType}`} size="lg">
      <Stack gap="md">
        <Text size="sm" c="dimmed">
          Edit the JSON representation of this {entityType}. Changes will be applied to the overlay.
        </Text>
        <Textarea
          value={jsonValue}
          onChange={(e) => {
            setJsonValue(e.currentTarget.value);
            setParseError(null);
          }}
          minRows={15}
          maxRows={20}
          autosize
          ff="monospace"
          styles={{ input: { fontSize: '12px' } }}
          error={parseError}
        />
        <Group justify="flex-end" gap="sm">
          <Button variant="default" onClick={onClose}>Cancel</Button>
          <Button onClick={handleSave}>Save Changes</Button>
        </Group>
      </Stack>
    </Modal>
  );
}

// Overlay name input with autocomplete for existing overlays
interface OverlayNameInputProps {
  value: string;
  onChange: (value: string) => void;
  overlays: OverlaySummary[];
  error?: string;
  currentOverlayId?: string;
}

function OverlayNameInput({ value, onChange, overlays, error, currentOverlayId }: OverlayNameInputProps): JSX.Element {
  const combobox = useCombobox({
    onDropdownClose: () => combobox.resetSelectedOption(),
  });
  const [search, setSearch] = useState(value);
  const [debouncedSearch] = useDebouncedValue(search, 150);

  // Filter overlays by search, excluding current overlay
  const filteredOverlays = overlays.filter(
    (o) => o.id !== currentOverlayId && o.name.toLowerCase().includes(debouncedSearch.toLowerCase())
  );

  // Sync external value changes
  useEffect(() => {
    setSearch(value);
  }, [value]);

  const handleSelect = (overlayName: string): void => {
    onChange(overlayName);
    setSearch(overlayName);
    combobox.closeDropdown();
  };

  return (
    <Combobox
      store={combobox}
      onOptionSubmit={(val) => handleSelect(val)}
    >
      <Combobox.Target>
        <Input.Wrapper label="Overlay Name" error={error}>
          <InputBase
            value={search}
            onChange={(e) => {
              const newValue = e.currentTarget.value;
              setSearch(newValue);
              onChange(newValue);
              combobox.openDropdown();
              combobox.updateSelectedOptionIndex();
            }}
            onClick={() => combobox.openDropdown()}
            onFocus={() => combobox.openDropdown()}
            onBlur={() => combobox.closeDropdown()}
            placeholder="Enter overlay name or search existing..."
            rightSection={<IconSearch size={16} color="var(--mantine-color-dimmed)" />}
            rightSectionPointerEvents="none"
          />
        </Input.Wrapper>
      </Combobox.Target>

      <Combobox.Dropdown>
        <Combobox.Options>
          <ScrollArea.Autosize mah={200} type="scroll">
            {filteredOverlays.length === 0 ? (
              <Combobox.Empty>
                {debouncedSearch ? 'No matching overlays - this will create a new one' : 'Type to search or enter a new name'}
              </Combobox.Empty>
            ) : (
              filteredOverlays.slice(0, 10).map((overlay) => (
                <Combobox.Option value={overlay.name} key={overlay.id}>
                  <Group gap="xs">
                    <Text size="sm" fw={500}>{overlay.name}</Text>
                    <Text size="xs" c="dimmed">
                      {overlay.numPrincipals}P · {overlay.numResources}R · {overlay.numPolicies}Po
                    </Text>
                  </Group>
                </Combobox.Option>
              ))
            )}
          </ScrollArea.Autosize>
        </Combobox.Options>
      </Combobox.Dropdown>
    </Combobox>
  );
}

export function OverlayEditorPage(): JSX.Element {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const isNew = id === 'new';

  const [loading, setLoading] = useState(!isNew);
  const [saving, setSaving] = useState(false);
  const [overlay, setOverlay] = useState<OverlayData | null>(
    isNew
      ? { name: '', id: '', createdAt: '', principals: [], resources: [], policies: [], accounts: [], groups: [] }
      : null
  );
  const [error, setError] = useState<string | null>(null);
  const [hasChanges, setHasChanges] = useState(false);
  const [existingOverlays, setExistingOverlays] = useState<OverlaySummary[]>([]);

  // Fetch existing overlays for name autocomplete
  useEffect(() => {
    yamsApi.listOverlays()
      .then(setExistingOverlays)
      .catch((err) => console.error('Failed to fetch overlays:', err));
  }, []);

  // Add modal state
  const [addModalOpened, { open: openAddModal, close: closeAddModal }] = useDisclosure(false);
  const [addEntityType, setAddEntityType] = useState<'principal' | 'resource' | 'policy' | 'account' | 'group'>('principal');

  // Edit modal state
  const [editModalOpened, { open: openEditModal, close: closeEditModal }] = useDisclosure(false);
  const [editEntityType, setEditEntityType] = useState<'principal' | 'resource' | 'policy' | 'account'>('principal');
  const [editingEntity, setEditingEntity] = useState<unknown | null>(null);
  const [editingEntityIndex, setEditingEntityIndex] = useState<number>(-1);

  const fetchOverlay = useCallback(async (): Promise<void> => {
    if (!id || isNew) return;
    setLoading(true);
    try {
      const data = await yamsApi.getOverlay(id);
      setOverlay(data);
      setError(null);
    } catch (err) {
      console.error('Failed to fetch overlay:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch overlay');
    } finally {
      setLoading(false);
    }
  }, [id, isNew]);

  useEffect(() => {
    fetchOverlay();
  }, [fetchOverlay]);

  const handleOpenAddModal = (type: 'principal' | 'resource' | 'policy' | 'account' | 'group'): void => {
    setAddEntityType(type);
    openAddModal();
  };

  const handleAddEntities = async (arns: string[]): Promise<void> => {
    if (!overlay || arns.length === 0) return;

    try {
      switch (addEntityType) {
        case 'principal': {
          const principals = await Promise.all(arns.map((arn) => yamsApi.getPrincipal(arn)));
          setOverlay({
            ...overlay,
            principals: [...(overlay.principals || []), ...principals],
          });
          break;
        }
        case 'resource': {
          const resources = await Promise.all(arns.map((arn) => yamsApi.getResource(arn)));
          setOverlay({
            ...overlay,
            resources: [...(overlay.resources || []), ...resources],
          });
          break;
        }
        case 'policy': {
          const policies = await Promise.all(arns.map((arn) => yamsApi.getPolicy(arn)));
          setOverlay({
            ...overlay,
            policies: [...(overlay.policies || []), ...policies],
          });
          break;
        }
        case 'account': {
          const accounts = await Promise.all(arns.map((arn) => yamsApi.getAccount(arn)));
          setOverlay({
            ...overlay,
            accounts: [...(overlay.accounts || []), ...accounts],
          });
          break;
        }
      }
      setHasChanges(true);
    } catch (err) {
      console.error('Failed to add entities:', err);
    }
  };

  const handleRemovePrincipal = (arn: string): void => {
    if (!overlay) return;
    setOverlay({
      ...overlay,
      principals: overlay.principals?.filter(p => p.Arn !== arn),
    });
    setHasChanges(true);
  };

  const handleRemoveResource = (arn: string): void => {
    if (!overlay) return;
    setOverlay({
      ...overlay,
      resources: overlay.resources?.filter(r => r.Arn !== arn),
    });
    setHasChanges(true);
  };

  const handleRemovePolicy = (arn: string): void => {
    if (!overlay) return;
    setOverlay({
      ...overlay,
      policies: overlay.policies?.filter(p => p.Arn !== arn),
    });
    setHasChanges(true);
  };

  const handleRemoveAccount = (accountId: string): void => {
    if (!overlay) return;
    setOverlay({
      ...overlay,
      accounts: overlay.accounts?.filter(a => a.Id !== accountId),
    });
    setHasChanges(true);
  };

  // Edit handlers
  const handleOpenEditModal = (type: 'principal' | 'resource' | 'policy' | 'account', entity: unknown, index: number): void => {
    setEditEntityType(type);
    setEditingEntity(entity);
    setEditingEntityIndex(index);
    openEditModal();
  };

  const handleSaveEditedEntity = (updated: unknown): void => {
    if (!overlay || editingEntityIndex < 0) return;

    switch (editEntityType) {
      case 'principal':
        setOverlay({
          ...overlay,
          principals: overlay.principals?.map((p, i) => (i === editingEntityIndex ? updated : p)) as typeof overlay.principals,
        });
        break;
      case 'resource':
        setOverlay({
          ...overlay,
          resources: overlay.resources?.map((r, i) => (i === editingEntityIndex ? updated : r)) as typeof overlay.resources,
        });
        break;
      case 'policy':
        setOverlay({
          ...overlay,
          policies: overlay.policies?.map((p, i) => (i === editingEntityIndex ? updated : p)) as typeof overlay.policies,
        });
        break;
      case 'account':
        setOverlay({
          ...overlay,
          accounts: overlay.accounts?.map((a, i) => (i === editingEntityIndex ? updated : a)) as typeof overlay.accounts,
        });
        break;
    }
    setHasChanges(true);
    setEditingEntity(null);
    setEditingEntityIndex(-1);
  };

  const handleSave = async (): Promise<void> => {
    if (!overlay) return;
    if (!overlay.name.trim()) {
      setError('Overlay name is required');
      return;
    }
    setSaving(true);
    setError(null);
    try {
      if (isNew) {
        const created = await yamsApi.createOverlay({
          name: overlay.name,
          principals: overlay.principals,
          resources: overlay.resources,
          policies: overlay.policies,
          accounts: overlay.accounts,
          groups: overlay.groups,
        });
        navigate(`/overlays/${created.id}/edit`, { replace: true });
      } else {
        await yamsApi.updateOverlay(id!, {
          name: overlay.name,
          principals: overlay.principals,
          resources: overlay.resources,
          policies: overlay.policies,
          accounts: overlay.accounts,
          groups: overlay.groups,
        });
      }
      setHasChanges(false);
    } catch (err) {
      console.error('Failed to save overlay:', err);
      setError(err instanceof Error ? err.message : 'Failed to save overlay');
    } finally {
      setSaving(false);
    }
  };

  const breadcrumbItems = (isNew
    ? [
        { title: 'Overlays', href: '/overlays' },
        { title: 'New Overlay', href: '' },
      ]
    : [
        { title: 'Overlays', href: '/overlays' },
        { title: overlay?.name || 'Loading...', href: `/overlays/${id}` },
        { title: 'Edit', href: '' },
      ]
  ).map((item, index, arr) => {
    const isLast = index === arr.length - 1;
    if (isLast) {
      return <Text key={item.title} size="sm" c="dimmed">{item.title}</Text>;
    }
    return (
      <Anchor key={item.title} component={Link} to={item.href} size="sm">
        {item.title}
      </Anchor>
    );
  });

  if (error) {
    return (
      <Box p="md">
        <EmptyState variant="error" message={error} />
      </Box>
    );
  }

  if (loading) {
    return (
      <Box p="md">
        <DetailSkeleton />
      </Box>
    );
  }

  if (!overlay) {
    return (
      <Box p="md">
        <EmptyState variant="error" message="Overlay not found" />
      </Box>
    );
  }

  const existingArns = [
    ...(overlay.principals?.map(p => p.Arn) || []),
    ...(overlay.resources?.map(r => r.Arn) || []),
    ...(overlay.policies?.map(p => p.Arn) || []),
    ...(overlay.accounts?.map(a => a.Id) || []),
    ...(overlay.groups?.map(g => g.Arn) || []),
  ];

  return (
    <Box p="md" h="100%">
      <Grid gutter="md" h="100%">
        {/* Left column - Entity list */}
        <Grid.Col span={6}>
          <Stack gap="md" h="100%">
            <Breadcrumbs separator={<IconChevronRight size={14} />}>{breadcrumbItems}</Breadcrumbs>
            <Group justify="space-between" align="flex-end">
              <Box style={{ flex: 1, maxWidth: 400 }}>
                <OverlayNameInput
                  value={overlay.name}
                  onChange={(name) => {
                    setOverlay({ ...overlay, name });
                    setHasChanges(true);
                  }}
                  overlays={existingOverlays}
                  currentOverlayId={isNew ? undefined : id}
                  error={error && !overlay.name.trim() ? 'Name is required' : undefined}
                />
              </Box>
              <Button
                leftSection={<IconDeviceFloppy size={16} />}
                onClick={handleSave}
                loading={saving}
                disabled={!overlay.name.trim() || (!hasChanges && !isNew)}
              >
                {isNew ? 'Create Overlay' : 'Save Changes'}
              </Button>
            </Group>

            <Card withBorder style={{ flex: 1, overflow: 'hidden' }}>
              <Tabs defaultValue="principals" style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
                <Tabs.List>
                  <Tabs.Tab value="principals" leftSection={<IconUser size={14} />}>
                    Principals ({overlay.principals?.length || 0})
                  </Tabs.Tab>
                  <Tabs.Tab value="resources" leftSection={<IconDatabase size={14} />}>
                    Resources ({overlay.resources?.length || 0})
                  </Tabs.Tab>
                  <Tabs.Tab value="policies" leftSection={<IconShield size={14} />}>
                    Policies ({overlay.policies?.length || 0})
                  </Tabs.Tab>
                  <Tabs.Tab value="accounts" leftSection={<IconBuilding size={14} />}>
                    Accounts ({overlay.accounts?.length || 0})
                  </Tabs.Tab>
                </Tabs.List>

                <Tabs.Panel value="principals" style={{ flex: 1, overflow: 'hidden' }}>
                  <Stack gap="md" p="md" h="100%">
                    <Button
                      variant="light"
                      leftSection={<IconPlus size={16} />}
                      onClick={() => handleOpenAddModal('principal')}
                    >
                      Add Principal
                    </Button>
                    <ScrollArea style={{ flex: 1 }}>
                      {overlay.principals?.length === 0 ? (
                        <Text c="dimmed" ta="center" py="xl">No principals in overlay</Text>
                      ) : (
                        <Stack gap="xs">
                          {overlay.principals?.map((p, idx) => (
                            <Group key={p.Arn} gap="xs" p="xs" style={{ border: '1px solid var(--mantine-color-default-border)', borderRadius: 4 }}>
                              <Text size="sm" ff="monospace" style={{ flex: 1, wordBreak: 'break-all' }}>
                                {p.Arn}
                              </Text>
                              <CopyButton value={p.Arn} />
                              <ActionIcon variant="light" onClick={() => handleOpenEditModal('principal', p, idx)}>
                                <IconEdit size={14} />
                              </ActionIcon>
                              <ActionIcon variant="light" color="red" onClick={() => handleRemovePrincipal(p.Arn)}>
                                <IconTrash size={14} />
                              </ActionIcon>
                            </Group>
                          ))}
                        </Stack>
                      )}
                    </ScrollArea>
                  </Stack>
                </Tabs.Panel>

                <Tabs.Panel value="resources" style={{ flex: 1, overflow: 'hidden' }}>
                  <Stack gap="md" p="md" h="100%">
                    <Button
                      variant="light"
                      leftSection={<IconPlus size={16} />}
                      onClick={() => handleOpenAddModal('resource')}
                    >
                      Add Resource
                    </Button>
                    <ScrollArea style={{ flex: 1 }}>
                      {overlay.resources?.length === 0 ? (
                        <Text c="dimmed" ta="center" py="xl">No resources in overlay</Text>
                      ) : (
                        <Stack gap="xs">
                          {overlay.resources?.map((r, idx) => (
                            <Group key={r.Arn} gap="xs" p="xs" style={{ border: '1px solid var(--mantine-color-default-border)', borderRadius: 4 }}>
                              <Text size="sm" ff="monospace" style={{ flex: 1, wordBreak: 'break-all' }}>
                                {r.Arn}
                              </Text>
                              <CopyButton value={r.Arn} />
                              <ActionIcon variant="light" onClick={() => handleOpenEditModal('resource', r, idx)}>
                                <IconEdit size={14} />
                              </ActionIcon>
                              <ActionIcon variant="light" color="red" onClick={() => handleRemoveResource(r.Arn)}>
                                <IconTrash size={14} />
                              </ActionIcon>
                            </Group>
                          ))}
                        </Stack>
                      )}
                    </ScrollArea>
                  </Stack>
                </Tabs.Panel>

                <Tabs.Panel value="policies" style={{ flex: 1, overflow: 'hidden' }}>
                  <Stack gap="md" p="md" h="100%">
                    <Button
                      variant="light"
                      leftSection={<IconPlus size={16} />}
                      onClick={() => handleOpenAddModal('policy')}
                    >
                      Add Policy
                    </Button>
                    <ScrollArea style={{ flex: 1 }}>
                      {overlay.policies?.length === 0 ? (
                        <Text c="dimmed" ta="center" py="xl">No policies in overlay</Text>
                      ) : (
                        <Stack gap="xs">
                          {overlay.policies?.map((p, idx) => (
                            <Group key={p.Arn} gap="xs" p="xs" style={{ border: '1px solid var(--mantine-color-default-border)', borderRadius: 4 }}>
                              <Text size="sm" ff="monospace" style={{ flex: 1, wordBreak: 'break-all' }}>
                                {p.Arn}
                              </Text>
                              <CopyButton value={p.Arn} />
                              <ActionIcon variant="light" onClick={() => handleOpenEditModal('policy', p, idx)}>
                                <IconEdit size={14} />
                              </ActionIcon>
                              <ActionIcon variant="light" color="red" onClick={() => handleRemovePolicy(p.Arn)}>
                                <IconTrash size={14} />
                              </ActionIcon>
                            </Group>
                          ))}
                        </Stack>
                      )}
                    </ScrollArea>
                  </Stack>
                </Tabs.Panel>

                <Tabs.Panel value="accounts" style={{ flex: 1, overflow: 'hidden' }}>
                  <Stack gap="md" p="md" h="100%">
                    <Button
                      variant="light"
                      leftSection={<IconPlus size={16} />}
                      onClick={() => handleOpenAddModal('account')}
                    >
                      Add Account
                    </Button>
                    <ScrollArea style={{ flex: 1 }}>
                      {overlay.accounts?.length === 0 ? (
                        <Text c="dimmed" ta="center" py="xl">No accounts in overlay</Text>
                      ) : (
                        <Stack gap="xs">
                          {overlay.accounts?.map((a, idx) => (
                            <Group key={a.Id} gap="xs" p="xs" style={{ border: '1px solid var(--mantine-color-default-border)', borderRadius: 4 }}>
                              <Text size="sm" ff="monospace" style={{ flex: 1 }}>
                                {a.Name} ({a.Id})
                              </Text>
                              <CopyButton value={a.Id} />
                              <ActionIcon variant="light" onClick={() => handleOpenEditModal('account', a, idx)}>
                                <IconEdit size={14} />
                              </ActionIcon>
                              <ActionIcon variant="light" color="red" onClick={() => handleRemoveAccount(a.Id)}>
                                <IconTrash size={14} />
                              </ActionIcon>
                            </Group>
                          ))}
                        </Stack>
                      )}
                    </ScrollArea>
                  </Stack>
                </Tabs.Panel>
              </Tabs>
            </Card>
          </Stack>
        </Grid.Col>

        {/* Right column - JSON preview */}
        <Grid.Col span={6}>
          <Card withBorder h="100%" p="md">
            <Stack gap="md" h="100%">
              <Title order={4}>JSON Preview</Title>
              <ScrollArea style={{ flex: 1 }}>
                <CodeHighlight
                  code={JSON.stringify(overlay, null, 2)}
                  language="json"
                  withCopyButton
                />
              </ScrollArea>
            </Stack>
          </Card>
        </Grid.Col>
      </Grid>

      <AddEntityModal
        opened={addModalOpened}
        onClose={closeAddModal}
        entityType={addEntityType}
        onAdd={handleAddEntities}
        existingArns={existingArns}
      />

      <EditEntityModal
        opened={editModalOpened}
        onClose={closeEditModal}
        entityType={editEntityType}
        entity={editingEntity}
        onSave={handleSaveEditedEntity}
      />
    </Box>
  );
}
