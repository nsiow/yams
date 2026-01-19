import { useState } from 'react';
import {
  ActionIcon,
  Badge,
  Box,
  Button,
  Card,
  Checkbox,
  Container,
  Divider,
  Grid,
  Group,
  Paper,
  ScrollArea,
  SegmentedControl,
  Stack,
  Tabs,
  Text,
  TextInput,
  Title,
  UnstyledButton,
} from '@mantine/core';
import {
  IconChevronDown,
  IconChevronRight,
  IconDeviceFloppy,
  IconGripVertical,
  IconPlus,
  IconSearch,
  IconTrash,
  IconUser,
  IconDatabase,
  IconShield,
  IconBuilding,
  IconX,
} from '@tabler/icons-react';

// Sample data for editor previews
const samplePrincipals = [
  'arn:aws:iam::123456789012:role/AdminRole',
  'arn:aws:iam::123456789012:role/DevRole',
  'arn:aws:iam::123456789012:user/alice',
];
const sampleResources = [
  'arn:aws:s3:::my-bucket',
  'arn:aws:s3:::my-bucket/*',
  'arn:aws:dynamodb:us-east-1:123456789012:table/Users',
];
const samplePolicies = [
  'arn:aws:iam::123456789012:policy/ReadOnly',
  'arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess',
];

const entityTypes = [
  { key: 'principals', label: 'Principals', icon: IconUser, color: 'blue', count: 3 },
  { key: 'resources', label: 'Resources', icon: IconDatabase, color: 'green', count: 3 },
  { key: 'policies', label: 'Policies', icon: IconShield, color: 'orange', count: 2 },
  { key: 'accounts', label: 'Accounts', icon: IconBuilding, color: 'violet', count: 0 },
];

interface DesignOption {
  name: string;
  label: string;
  description: string;
  render: () => JSX.Element;
}

// Extract stateful components to avoid hooks-in-render lint errors
function DesignAccordion(): JSX.Element {
  const [expanded, setExpanded] = useState<string | null>('principals');
  return (
    <Card withBorder p="md">
      <Group justify="space-between" mb="md">
        <Text fw={600}>My Overlay</Text>
        <Button size="sm" leftSection={<IconDeviceFloppy size={14} />}>Save</Button>
      </Group>
      <Stack gap={0}>
        {entityTypes.map((t) => (
          <Box key={t.key} style={{ borderBottom: '1px solid var(--mantine-color-gray-3)' }}>
            <UnstyledButton w="100%" p="sm" onClick={() => setExpanded(expanded === t.key ? null : t.key)}>
              <Group justify="space-between">
                <Group gap="xs">
                  {expanded === t.key ? <IconChevronDown size={14} /> : <IconChevronRight size={14} />}
                  <t.icon size={16} color={`var(--mantine-color-${t.color}-6)`} />
                  <Text size="sm" fw={500}>{t.label}</Text>
                </Group>
                <Badge size="sm" variant="light">{t.count}</Badge>
              </Group>
            </UnstyledButton>
            {expanded === t.key && (
              <Box p="sm" pt={0} bg="gray.0">
                <Button size="xs" variant="light" mb="xs" leftSection={<IconPlus size={12} />}>Add</Button>
                {t.key === 'principals' && samplePrincipals.slice(0, 2).map((p) => (
                  <Text key={p} size="xs" ff="monospace" mb={4}>{p}</Text>
                ))}
                {t.count === 0 && <Text size="xs" c="dimmed">No {t.label.toLowerCase()} added</Text>}
              </Box>
            )}
          </Box>
        ))}
      </Stack>
    </Card>
  );
}

function DesignUnifiedList(): JSX.Element {
  const [filter, setFilter] = useState('all');
  const allItems = [
    ...samplePrincipals.map(p => ({ arn: p, type: 'principal' })),
    ...sampleResources.map(r => ({ arn: r, type: 'resource' })),
    ...samplePolicies.map(p => ({ arn: p, type: 'policy' })),
  ];
  return (
    <Card withBorder p="md">
      <Group justify="space-between" mb="md">
        <Text fw={600}>My Overlay</Text>
        <Button size="sm" leftSection={<IconDeviceFloppy size={14} />}>Save</Button>
      </Group>
      <Group mb="md" gap="xs">
        <SegmentedControl
          size="xs"
          value={filter}
          onChange={setFilter}
          data={[
            { label: `All (${allItems.length})`, value: 'all' },
            { label: 'Principals', value: 'principal' },
            { label: 'Resources', value: 'resource' },
            { label: 'Policies', value: 'policy' },
          ]}
        />
        <Button size="xs" variant="light" leftSection={<IconPlus size={12} />}>Add</Button>
      </Group>
      <ScrollArea h={150}>
        <Stack gap="xs">
          {allItems.filter(i => filter === 'all' || i.type === filter).map((item) => (
            <Group key={item.arn} justify="space-between" p="xs" style={{ border: '1px solid var(--mantine-color-gray-3)', borderRadius: 4 }}>
              <Group gap="xs">
                <Badge size="xs" variant="dot" color={item.type === 'principal' ? 'blue' : item.type === 'resource' ? 'green' : 'orange'}>
                  {item.type}
                </Badge>
                <Text size="xs" ff="monospace" truncate style={{ maxWidth: 200 }}>{item.arn.split(':').pop()}</Text>
              </Group>
              <ActionIcon size="sm" variant="subtle" color="red"><IconTrash size={14} /></ActionIcon>
            </Group>
          ))}
        </Stack>
      </ScrollArea>
    </Card>
  );
}

function DesignSidebar(): JSX.Element {
  const [active, setActive] = useState('principals');
  return (
    <Card withBorder p={0} style={{ overflow: 'hidden' }}>
      <Grid gutter={0}>
        <Grid.Col span={3} bg="gray.0" p="sm" style={{ borderRight: '1px solid var(--mantine-color-gray-3)' }}>
          <Stack gap={4}>
            {entityTypes.map((t) => (
              <UnstyledButton
                key={t.key}
                p="xs"
                onClick={() => setActive(t.key)}
                style={{
                  borderRadius: 4,
                  backgroundColor: active === t.key ? 'white' : undefined,
                  border: active === t.key ? '1px solid var(--mantine-color-gray-3)' : '1px solid transparent',
                }}
              >
                <Group gap="xs">
                  <t.icon size={14} color={`var(--mantine-color-${t.color}-6)`} />
                  <Text size="xs">{t.label}</Text>
                </Group>
              </UnstyledButton>
            ))}
          </Stack>
        </Grid.Col>
        <Grid.Col span={9} p="sm">
          <Group justify="space-between" mb="sm">
            <Text size="sm" fw={600}>Principals (3)</Text>
            <Group gap="xs">
              <Button size="xs" variant="light" leftSection={<IconPlus size={12} />}>Add</Button>
              <Button size="xs" leftSection={<IconDeviceFloppy size={12} />}>Save</Button>
            </Group>
          </Group>
          <Stack gap="xs">
            {samplePrincipals.map((p) => (
              <Group key={p} justify="space-between" p="xs" style={{ border: '1px solid var(--mantine-color-gray-3)', borderRadius: 4 }}>
                <Text size="xs" ff="monospace">{p}</Text>
                <ActionIcon size="sm" variant="subtle" color="red"><IconTrash size={14} /></ActionIcon>
              </Group>
            ))}
          </Stack>
        </Grid.Col>
      </Grid>
    </Card>
  );
}

const designs: DesignOption[] = [
  {
    name: 'A',
    label: 'Tabbed Panels',
    description: 'Separate tabs for each entity type. Clean organization, easy to focus on one type at a time.',
    render: () => (
      <Card withBorder p="md">
        <Group justify="space-between" mb="md">
          <TextInput placeholder="Overlay name..." defaultValue="My Overlay" style={{ width: 200 }} />
          <Button size="sm" leftSection={<IconDeviceFloppy size={14} />}>Save</Button>
        </Group>
        <Tabs defaultValue="principals">
          <Tabs.List>
            {entityTypes.map((t) => (
              <Tabs.Tab key={t.key} value={t.key} leftSection={<t.icon size={14} />}>
                {t.label} ({t.count})
              </Tabs.Tab>
            ))}
          </Tabs.List>
          <Tabs.Panel value="principals" pt="md">
            <Stack gap="xs">
              <Button size="xs" variant="light" leftSection={<IconPlus size={14} />}>Add Principal</Button>
              {samplePrincipals.map((p) => (
                <Group key={p} justify="space-between" p="xs" style={{ border: '1px solid var(--mantine-color-gray-3)', borderRadius: 4 }}>
                  <Text size="xs" ff="monospace" truncate style={{ maxWidth: 280 }}>{p}</Text>
                  <ActionIcon size="sm" variant="subtle" color="red"><IconTrash size={14} /></ActionIcon>
                </Group>
              ))}
            </Stack>
          </Tabs.Panel>
        </Tabs>
      </Card>
    ),
  },
  {
    name: 'B',
    label: 'Split Panel',
    description: 'Available entities on left, selected on right. Good for browsing and selecting.',
    render: () => (
      <Card withBorder p="md">
        <Group justify="space-between" mb="md">
          <Text fw={600}>My Overlay</Text>
          <Button size="sm" leftSection={<IconDeviceFloppy size={14} />}>Save</Button>
        </Group>
        <Grid gutter="md">
          <Grid.Col span={6}>
            <Card withBorder p="sm" bg="gray.0">
              <Text size="sm" fw={600} mb="xs">Available Principals</Text>
              <TextInput size="xs" placeholder="Search..." leftSection={<IconSearch size={12} />} mb="xs" />
              <ScrollArea h={120}>
                <Stack gap={4}>
                  {['role/NewRole', 'user/bob', 'user/charlie'].map((p) => (
                    <UnstyledButton key={p} p={4} style={{ borderRadius: 4, border: '1px solid var(--mantine-color-gray-3)', background: 'white' }}>
                      <Group gap="xs">
                        <IconPlus size={12} color="var(--mantine-color-blue-6)" />
                        <Text size="xs" ff="monospace">{p}</Text>
                      </Group>
                    </UnstyledButton>
                  ))}
                </Stack>
              </ScrollArea>
            </Card>
          </Grid.Col>
          <Grid.Col span={6}>
            <Card withBorder p="sm">
              <Text size="sm" fw={600} mb="xs">Selected ({samplePrincipals.length})</Text>
              <ScrollArea h={150}>
                <Stack gap={4}>
                  {samplePrincipals.map((p) => (
                    <Group key={p} justify="space-between" p={4} style={{ borderRadius: 4, background: 'var(--mantine-color-blue-0)' }}>
                      <Text size="xs" ff="monospace" truncate style={{ maxWidth: 140 }}>{p.split('/').pop()}</Text>
                      <ActionIcon size="xs" variant="subtle" color="red"><IconX size={12} /></ActionIcon>
                    </Group>
                  ))}
                </Stack>
              </ScrollArea>
            </Card>
          </Grid.Col>
        </Grid>
      </Card>
    ),
  },
  {
    name: 'C',
    label: 'Accordion Sections',
    description: 'Expandable sections for each entity type. Compact when collapsed, detailed when expanded.',
    render: () => <DesignAccordion />,
  },
  {
    name: 'D',
    label: 'Card Grid',
    description: 'Entity types as cards in a grid. Visual overview with quick actions.',
    render: () => (
      <Card withBorder p="md">
        <Group justify="space-between" mb="md">
          <Text fw={600}>My Overlay</Text>
          <Button size="sm" leftSection={<IconDeviceFloppy size={14} />}>Save</Button>
        </Group>
        <Grid gutter="sm">
          {entityTypes.map((t) => (
            <Grid.Col span={6} key={t.key}>
              <Card withBorder p="sm" style={{ height: '100%' }}>
                <Group justify="space-between" mb="xs">
                  <Group gap="xs">
                    <t.icon size={16} color={`var(--mantine-color-${t.color}-6)`} />
                    <Text size="sm" fw={600}>{t.label}</Text>
                  </Group>
                  <Badge size="sm">{t.count}</Badge>
                </Group>
                {t.count > 0 ? (
                  <Text size="xs" c="dimmed" mb="xs">{t.count} items configured</Text>
                ) : (
                  <Text size="xs" c="dimmed" mb="xs">None added yet</Text>
                )}
                <Button size="xs" variant="light" fullWidth leftSection={<IconPlus size={12} />}>
                  Add {t.label}
                </Button>
              </Card>
            </Grid.Col>
          ))}
        </Grid>
      </Card>
    ),
  },
  {
    name: 'E',
    label: 'Unified List',
    description: 'Single list with type filter. Simple and streamlined for smaller overlays.',
    render: () => <DesignUnifiedList />,
  },
  {
    name: 'F',
    label: 'Drag & Drop Builder',
    description: 'Visual builder with draggable items. Interactive and intuitive.',
    render: () => (
      <Card withBorder p="md">
        <Group justify="space-between" mb="md">
          <Text fw={600}>My Overlay</Text>
          <Button size="sm" leftSection={<IconDeviceFloppy size={14} />}>Save</Button>
        </Group>
        <Grid gutter="md">
          <Grid.Col span={4}>
            <Text size="xs" fw={600} c="dimmed" mb="xs">DRAG TO ADD</Text>
            <Stack gap={4}>
              {['Principals', 'Resources', 'Policies'].map((type) => (
                <Paper key={type} withBorder p="xs" style={{ cursor: 'grab' }}>
                  <Group gap="xs">
                    <IconGripVertical size={14} color="var(--mantine-color-gray-5)" />
                    <Text size="xs">{type}</Text>
                  </Group>
                </Paper>
              ))}
            </Stack>
          </Grid.Col>
          <Grid.Col span={8}>
            <Box p="md" style={{ border: '2px dashed var(--mantine-color-gray-4)', borderRadius: 8, minHeight: 150 }}>
              <Text size="xs" c="dimmed" ta="center" mb="md">Drop items here or click to add</Text>
              <Stack gap="xs">
                {samplePrincipals.slice(0, 2).map((p) => (
                  <Paper key={p} withBorder p="xs" bg="blue.0">
                    <Group justify="space-between">
                      <Group gap="xs">
                        <IconGripVertical size={14} color="var(--mantine-color-gray-5)" style={{ cursor: 'grab' }} />
                        <IconUser size={14} color="var(--mantine-color-blue-6)" />
                        <Text size="xs" ff="monospace">{p.split('/').pop()}</Text>
                      </Group>
                      <ActionIcon size="xs" variant="subtle" color="red"><IconX size={12} /></ActionIcon>
                    </Group>
                  </Paper>
                ))}
              </Stack>
            </Box>
          </Grid.Col>
        </Grid>
      </Card>
    ),
  },
  {
    name: 'G',
    label: 'Checklist Style',
    description: 'Checkbox-based selection from available entities. Quick multi-select.',
    render: () => (
      <Card withBorder p="md">
        <Group justify="space-between" mb="md">
          <Text fw={600}>My Overlay</Text>
          <Button size="sm" leftSection={<IconDeviceFloppy size={14} />}>Save</Button>
        </Group>
        <Tabs defaultValue="principals">
          <Tabs.List>
            <Tabs.Tab value="principals">Principals</Tabs.Tab>
            <Tabs.Tab value="resources">Resources</Tabs.Tab>
          </Tabs.List>
          <Tabs.Panel value="principals" pt="md">
            <TextInput size="xs" placeholder="Search principals..." leftSection={<IconSearch size={12} />} mb="sm" />
            <ScrollArea h={130}>
              <Stack gap={4}>
                {[...samplePrincipals, 'arn:aws:iam::123456789012:role/NewRole', 'arn:aws:iam::123456789012:user/bob'].map((p, i) => (
                  <Checkbox
                    key={p}
                    label={<Text size="xs" ff="monospace">{p.split('/').pop()}</Text>}
                    defaultChecked={i < 3}
                    size="xs"
                  />
                ))}
              </Stack>
            </ScrollArea>
            <Divider my="sm" />
            <Text size="xs" c="dimmed">3 of 5 selected</Text>
          </Tabs.Panel>
        </Tabs>
      </Card>
    ),
  },
  {
    name: 'H',
    label: 'Inline Editor',
    description: 'Edit entities inline with quick add. Minimal, fast editing experience.',
    render: () => (
      <Card withBorder p="md">
        <Group justify="space-between" mb="md">
          <TextInput placeholder="Overlay name..." defaultValue="My Overlay" variant="unstyled" styles={{ input: { fontWeight: 600, fontSize: 16 } }} />
          <Button size="sm" leftSection={<IconDeviceFloppy size={14} />}>Save</Button>
        </Group>
        <Stack gap="md">
          {entityTypes.slice(0, 3).map((t) => (
            <Box key={t.key}>
              <Group gap="xs" mb="xs">
                <t.icon size={14} color={`var(--mantine-color-${t.color}-6)`} />
                <Text size="sm" fw={600}>{t.label}</Text>
              </Group>
              <Group gap="xs">
                {(t.key === 'principals' ? samplePrincipals : t.key === 'resources' ? sampleResources : samplePolicies).slice(0, 2).map((item) => (
                  <Badge key={item} variant="light" rightSection={<IconX size={10} style={{ cursor: 'pointer' }} />}>
                    {item.split('/').pop() || item.split(':').pop()}
                  </Badge>
                ))}
                <Badge variant="outline" style={{ cursor: 'pointer' }} leftSection={<IconPlus size={10} />}>
                  Add
                </Badge>
              </Group>
            </Box>
          ))}
        </Stack>
      </Card>
    ),
  },
  {
    name: 'I',
    label: 'Sidebar Navigation',
    description: 'Vertical sidebar for entity types, content on right. Good for many entities.',
    render: () => <DesignSidebar />,
  },
  {
    name: 'J',
    label: 'Minimal Form',
    description: 'Simple form-based approach with text inputs. Direct ARN entry for power users.',
    render: () => (
      <Card withBorder p="md">
        <Stack gap="md">
          <TextInput label="Overlay Name" defaultValue="My Overlay" />
          <div>
            <Text size="sm" fw={500} mb="xs">Principals</Text>
            <Stack gap="xs">
              {samplePrincipals.map((p) => (
                <Group key={p} gap="xs">
                  <TextInput size="xs" defaultValue={p} style={{ flex: 1 }} ff="monospace" />
                  <ActionIcon variant="subtle" color="red"><IconTrash size={14} /></ActionIcon>
                </Group>
              ))}
              <Button size="xs" variant="subtle" leftSection={<IconPlus size={12} />}>Add Principal ARN</Button>
            </Stack>
          </div>
          <div>
            <Text size="sm" fw={500} mb="xs">Resources</Text>
            <Stack gap="xs">
              {sampleResources.slice(0, 2).map((r) => (
                <Group key={r} gap="xs">
                  <TextInput size="xs" defaultValue={r} style={{ flex: 1 }} ff="monospace" />
                  <ActionIcon variant="subtle" color="red"><IconTrash size={14} /></ActionIcon>
                </Group>
              ))}
              <Button size="xs" variant="subtle" leftSection={<IconPlus size={12} />}>Add Resource ARN</Button>
            </Stack>
          </div>
          <Group justify="flex-end">
            <Button leftSection={<IconDeviceFloppy size={14} />}>Save Overlay</Button>
          </Group>
        </Stack>
      </Card>
    ),
  },
];

export function PreviewPage(): JSX.Element {
  return (
    <Container size="xl" py="xl">
      <Stack gap="xl">
        <div>
          <Title order={1} mb="xs">Overlay Editor Designs</Title>
          <Text c="dimmed">
            Choose a design style for the Overlay Editor. Each option optimizes for different workflows.
          </Text>
        </div>

        <Stack gap="xl">
          {designs.map((design) => (
            <Card key={design.name} padding="lg" withBorder>
              <Stack gap="md">
                <Group justify="space-between">
                  <div>
                    <Group gap="xs">
                      <Badge size="lg" variant="filled" color="violet">{design.name}</Badge>
                      <Text fw={700} size="lg">{design.label}</Text>
                    </Group>
                    <Text size="sm" c="dimmed" mt={4}>{design.description}</Text>
                  </div>
                </Group>
                <Box
                  p="md"
                  style={{
                    border: '1px solid var(--mantine-color-gray-3)',
                    borderRadius: 8,
                    backgroundColor: 'var(--mantine-color-gray-0)',
                  }}
                >
                  {design.render()}
                </Box>
              </Stack>
            </Card>
          ))}
        </Stack>
      </Stack>
    </Container>
  );
}
