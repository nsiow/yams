import { useState } from 'react';
import {
  ActionIcon,
  Anchor,
  Badge,
  Box,
  Button,
  Card,
  Checkbox,
  Container,
  Divider,
  Grid,
  Group,
  Image,
  Indicator,
  Menu,
  Paper,
  ScrollArea,
  SegmentedControl,
  Stack,
  Table,
  Tabs,
  Text,
  TextInput,
  Title,
  Tooltip,
  UnstyledButton,
} from '@mantine/core';
import {
  IconArrowRight,
  IconBell,
  IconCheck,
  IconChevronDown,
  IconChevronRight,
  IconCircleCheck,
  IconCircleX,
  IconClock,
  IconDeviceFloppy,
  IconExternalLink,
  IconFlask,
  IconGripVertical,
  IconLayersSubtract,
  IconMoon,
  IconPlayerPlay,
  IconPlus,
  IconRefresh,
  IconSearch,
  IconServer,
  IconShieldCheck,
  IconShieldX,
  IconTestPipe,
  IconTrash,
  IconUser,
  IconDatabase,
  IconShield,
  IconBuilding,
  IconX,
  IconZoomCheck,
  IconEye,
  IconFocus2,
  IconAnalyze,
  IconAlertTriangle,
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

// Sample data for simulation results table previews
const sampleResults = [
  { arn: 'arn:aws:iam::123456789012:role/AdminRole', account: 'Production', accountId: '123456789012' },
  { arn: 'arn:aws:iam::123456789012:role/DevRole', account: 'Production', accountId: '123456789012' },
  { arn: 'arn:aws:iam::987654321098:user/alice', account: 'Development', accountId: '987654321098' },
  { arn: 'arn:aws:iam::555555555555:role/ReadOnlyRole', account: 'Staging', accountId: '555555555555' },
  { arn: 'arn:aws:iam::123456789012:user/bob', account: 'Production', accountId: '123456789012' },
];

// Simulation results table designs - focused on density and formatting
interface TableDesignOption {
  name: string;
  label: string;
  description: string;
  render: () => JSX.Element;
}

function TableDesignExpandable(): JSX.Element {
  const [expanded, setExpanded] = useState<string | null>(null);
  return (
    <Table striped highlightOnHover>
      <Table.Thead>
        <Table.Tr>
          <Table.Th style={{ width: 32 }}></Table.Th>
          <Table.Th>Principal</Table.Th>
          <Table.Th style={{ width: 120 }}>Account</Table.Th>
          <Table.Th style={{ width: 100 }}>Actions</Table.Th>
        </Table.Tr>
      </Table.Thead>
      <Table.Tbody>
        {sampleResults.slice(0, 3).map((r) => (
          <>
            <Table.Tr key={r.arn}>
              <Table.Td>
                <UnstyledButton onClick={() => setExpanded(expanded === r.arn ? null : r.arn)}>
                  {expanded === r.arn ? <IconChevronDown size={14} /> : <IconChevronRight size={14} />}
                </UnstyledButton>
              </Table.Td>
              <Table.Td>
                <Anchor size="sm">{r.arn.split('/').pop()}</Anchor>
                <Text size="xs" c="dimmed" ff="monospace" truncate style={{ maxWidth: 300 }}>{r.arn}</Text>
              </Table.Td>
              <Table.Td>
                <Text size="sm">{r.account}</Text>
              </Table.Td>
              <Table.Td>
                <Anchor size="xs"><Group gap={4}><IconExternalLink size={12} />Check</Group></Anchor>
              </Table.Td>
            </Table.Tr>
            {expanded === r.arn && (
              <Table.Tr key={`${r.arn}-exp`}>
                <Table.Td colSpan={4} p={0}>
                  <Box p="sm" bg="gray.0">
                    <Group gap="xs" mb="xs">
                      <Badge color="green" size="sm" leftSection={<IconCheck size={10} />}>ALLOW</Badge>
                    </Group>
                    <Text size="xs" c="dimmed">Access granted via attached policy AmazonS3FullAccess</Text>
                  </Box>
                </Table.Td>
              </Table.Tr>
            )}
          </>
        ))}
      </Table.Tbody>
    </Table>
  );
}

const tableDesigns: TableDesignOption[] = [
  {
    name: '1',
    label: 'Compact Single-Line',
    description: 'Dense layout showing only essential info. Good for scanning many results quickly.',
    render: () => (
      <Table>
        <Table.Thead>
          <Table.Tr>
            <Table.Th>Principal</Table.Th>
            <Table.Th style={{ width: 100 }}>Account</Table.Th>
            <Table.Th style={{ width: 80 }}></Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {sampleResults.map((r) => (
            <Table.Tr key={r.arn}>
              <Table.Td py={4}>
                <Text size="xs" ff="monospace" truncate style={{ maxWidth: 350 }}>{r.arn}</Text>
              </Table.Td>
              <Table.Td py={4}>
                <Text size="xs" c="dimmed">{r.accountId}</Text>
              </Table.Td>
              <Table.Td py={4}>
                <Anchor size="xs">Check →</Anchor>
              </Table.Td>
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>
    ),
  },
  {
    name: '2',
    label: 'Two-Line with Name',
    description: 'Shows friendly name prominently with full ARN below. Balanced density.',
    render: () => (
      <Table striped>
        <Table.Thead>
          <Table.Tr>
            <Table.Th>Principal</Table.Th>
            <Table.Th style={{ width: 130 }}>Account</Table.Th>
            <Table.Th style={{ width: 100 }}>Actions</Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {sampleResults.map((r) => (
            <Table.Tr key={r.arn}>
              <Table.Td>
                <Text size="sm" fw={500}>{r.arn.split('/').pop()}</Text>
                <Text size="xs" c="dimmed" ff="monospace">{r.arn}</Text>
              </Table.Td>
              <Table.Td>
                <Text size="sm">{r.account}</Text>
                <Text size="xs" c="dimmed">{r.accountId}</Text>
              </Table.Td>
              <Table.Td>
                <Anchor size="xs"><Group gap={4}><IconExternalLink size={12} />Access Check</Group></Anchor>
              </Table.Td>
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>
    ),
  },
  {
    name: '3',
    label: 'Linked Name Only',
    description: 'Minimal view with clickable names. ARN shown on hover.',
    render: () => (
      <Table highlightOnHover>
        <Table.Thead>
          <Table.Tr>
            <Table.Th>Principal</Table.Th>
            <Table.Th style={{ width: 120 }}>Account</Table.Th>
            <Table.Th style={{ width: 80 }}></Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {sampleResults.map((r) => (
            <Table.Tr key={r.arn}>
              <Table.Td>
                <Tooltip label={r.arn} multiline maw={400}>
                  <Anchor size="sm">{r.arn.split('/').pop()}</Anchor>
                </Tooltip>
              </Table.Td>
              <Table.Td>
                <Text size="sm" c="dimmed">{r.account}</Text>
              </Table.Td>
              <Table.Td>
                <Anchor size="xs">Check</Anchor>
              </Table.Td>
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>
    ),
  },
  {
    name: '4',
    label: 'Expandable Rows',
    description: 'Click to expand and see simulation details inline. Interactive exploration.',
    render: () => <TableDesignExpandable />,
  },
  {
    name: '5',
    label: 'Badge Account Tags',
    description: 'Account shown as colored badge. Visual grouping by account.',
    render: () => (
      <Table striped>
        <Table.Thead>
          <Table.Tr>
            <Table.Th>Principal</Table.Th>
            <Table.Th style={{ width: 130 }}>Account</Table.Th>
            <Table.Th style={{ width: 100 }}></Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {sampleResults.map((r, i) => (
            <Table.Tr key={r.arn}>
              <Table.Td>
                <Text size="sm">{r.arn.split('/').pop()}</Text>
                <Text size="xs" c="dimmed" ff="monospace">{r.arn}</Text>
              </Table.Td>
              <Table.Td>
                <Badge size="sm" variant="light" color={['blue', 'green', 'orange', 'violet', 'cyan'][i % 5]}>
                  {r.account}
                </Badge>
              </Table.Td>
              <Table.Td>
                <Anchor size="xs"><IconExternalLink size={12} /></Anchor>
              </Table.Td>
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>
    ),
  },
  {
    name: '6',
    label: 'Very Compact',
    description: 'Maximum density. Monospace font, minimal padding. For power users.',
    render: () => (
      <Table style={{ fontSize: 11 }}>
        <Table.Thead>
          <Table.Tr>
            <Table.Th py={2}>ARN</Table.Th>
            <Table.Th py={2} style={{ width: 80 }}>Acct</Table.Th>
            <Table.Th py={2} style={{ width: 40 }}></Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {sampleResults.map((r) => (
            <Table.Tr key={r.arn}>
              <Table.Td py={2} ff="monospace" style={{ fontSize: 10 }}>
                {r.arn}
              </Table.Td>
              <Table.Td py={2} c="dimmed" style={{ fontSize: 10 }}>
                {r.accountId.slice(0, 6)}...
              </Table.Td>
              <Table.Td py={2}>
                <Anchor size="xs">→</Anchor>
              </Table.Td>
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>
    ),
  },
  {
    name: '7',
    label: 'Spacious with Icons',
    description: 'More breathing room. Icons indicate principal type. Clear visual hierarchy.',
    render: () => (
      <Table>
        <Table.Thead>
          <Table.Tr>
            <Table.Th>Principal</Table.Th>
            <Table.Th style={{ width: 150 }}>Account</Table.Th>
            <Table.Th style={{ width: 120 }}>Actions</Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {sampleResults.map((r) => (
            <Table.Tr key={r.arn}>
              <Table.Td py="md">
                <Group gap="sm">
                  <IconUser size={18} color="var(--mantine-color-blue-6)" />
                  <Box>
                    <Anchor size="sm" fw={500}>{r.arn.split('/').pop()}</Anchor>
                    <Text size="xs" c="dimmed" ff="monospace">{r.arn}</Text>
                  </Box>
                </Group>
              </Table.Td>
              <Table.Td py="md">
                <Text size="sm" fw={500}>{r.account}</Text>
                <Text size="xs" c="dimmed">{r.accountId}</Text>
              </Table.Td>
              <Table.Td py="md">
                <Button size="xs" variant="light">Access Check</Button>
              </Table.Td>
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>
    ),
  },
  {
    name: '8',
    label: 'Card-Style Rows',
    description: 'Each row as a mini-card with border. Clear separation between items.',
    render: () => (
      <Stack gap="xs">
        {sampleResults.slice(0, 4).map((r) => (
          <Paper key={r.arn} withBorder p="sm">
            <Group justify="space-between">
              <Box>
                <Group gap="xs">
                  <Anchor size="sm" fw={500}>{r.arn.split('/').pop()}</Anchor>
                  <Badge size="xs" variant="light">{r.account}</Badge>
                </Group>
                <Text size="xs" c="dimmed" ff="monospace">{r.arn}</Text>
              </Box>
              <Anchor size="xs"><IconExternalLink size={14} /> Check</Anchor>
            </Group>
          </Paper>
        ))}
      </Stack>
    ),
  },
  {
    name: '9',
    label: 'Horizontal Scroll',
    description: 'Fixed columns with scroll for long ARNs. Prevents layout shift.',
    render: () => (
      <ScrollArea>
        <Table style={{ minWidth: 600 }}>
          <Table.Thead>
            <Table.Tr>
              <Table.Th style={{ width: 150 }}>Name</Table.Th>
              <Table.Th style={{ minWidth: 350 }}>Full ARN</Table.Th>
              <Table.Th style={{ width: 100 }}>Account</Table.Th>
              <Table.Th style={{ width: 80 }}></Table.Th>
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {sampleResults.map((r) => (
              <Table.Tr key={r.arn}>
                <Table.Td>
                  <Anchor size="sm">{r.arn.split('/').pop()}</Anchor>
                </Table.Td>
                <Table.Td>
                  <Text size="xs" ff="monospace">{r.arn}</Text>
                </Table.Td>
                <Table.Td>
                  <Text size="sm" c="dimmed">{r.account}</Text>
                </Table.Td>
                <Table.Td>
                  <Anchor size="xs">Check</Anchor>
                </Table.Td>
              </Table.Tr>
            ))}
          </Table.Tbody>
        </Table>
      </ScrollArea>
    ),
  },
  {
    name: '10',
    label: 'Grouped by Account',
    description: 'Results grouped under account headers. Good for cross-account results.',
    render: () => {
      const grouped: Record<string, typeof sampleResults> = {};
      sampleResults.forEach((r) => {
        if (!grouped[r.account]) grouped[r.account] = [];
        grouped[r.account].push(r);
      });
      return (
        <Stack gap="md">
          {Object.entries(grouped).map(([account, items]) => (
            <Box key={account}>
              <Group gap="xs" mb="xs">
                <IconBuilding size={14} color="var(--mantine-color-violet-6)" />
                <Text size="sm" fw={600}>{account}</Text>
                <Text size="xs" c="dimmed">({items.length})</Text>
              </Group>
              <Table>
                <Table.Tbody>
                  {items.map((r) => (
                    <Table.Tr key={r.arn}>
                      <Table.Td>
                        <Text size="sm">{r.arn.split('/').pop()}</Text>
                        <Text size="xs" c="dimmed" ff="monospace">{r.arn}</Text>
                      </Table.Td>
                      <Table.Td style={{ width: 80 }}>
                        <Anchor size="xs">Check</Anchor>
                      </Table.Td>
                    </Table.Tr>
                  ))}
                </Table.Tbody>
              </Table>
            </Box>
          ))}
        </Stack>
      );
    },
  },
];

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

// Icon options for the Simulate column
const simulateIconOptions = [
  { name: 'PlayerPlay', icon: IconPlayerPlay, description: 'Play button - indicates running/executing' },
  { name: 'ArrowRight', icon: IconArrowRight, description: 'Arrow - indicates navigation' },
  { name: 'ExternalLink', icon: IconExternalLink, description: 'External link - indicates opening elsewhere' },
  { name: 'Flask', icon: IconFlask, description: 'Flask - science/experiment theme' },
  { name: 'TestPipe', icon: IconTestPipe, description: 'Test tube - testing theme' },
  { name: 'ZoomCheck', icon: IconZoomCheck, description: 'Magnify with check - inspection/verification' },
  { name: 'Eye', icon: IconEye, description: 'Eye - view/inspect' },
  { name: 'Focus2', icon: IconFocus2, description: 'Focus/target - precision analysis' },
  { name: 'Analyze', icon: IconAnalyze, description: 'Analyze - data analysis theme' },
];

// Topbar preview components
interface TopbarPreview {
  name: string;
  label: string;
  description: string;
  render: () => JSX.Element;
}

function TopbarBase({ children, rightSection }: { children?: React.ReactNode; rightSection?: React.ReactNode }): JSX.Element {
  return (
    <Box bg="purple.6" p="md" style={{ borderRadius: 8 }}>
      <Group h={44} justify="space-between">
        <Group>
          <Group gap="sm">
            <Image src="/apple-touch-icon.png" w={44} h={44} />
            <Title order={3} c="white" ff="'Urbanist', sans-serif" fz="xl" style={{ letterSpacing: '0.15em' }}>
              yams
            </Title>
          </Group>
          {children}
        </Group>
        <Group gap="md">
          {rightSection}
          <Menu shadow="md" width={200}>
            <Menu.Target>
              <Button variant="subtle" c="white" rightSection={<IconChevronDown size={16} />}>Docs</Button>
            </Menu.Target>
          </Menu>
        </Group>
      </Group>
    </Box>
  );
}

const topbarPreviews: TopbarPreview[] = [
  {
    name: '1',
    label: 'Health Indicator',
    description: 'Simple green/red dot showing server connection status. Minimal and unobtrusive.',
    render: () => (
      <TopbarBase rightSection={
        <Tooltip label="Server healthy">
          <Badge color="green" variant="dot" size="lg" c="white">Connected</Badge>
        </Tooltip>
      } />
    ),
  },
  {
    name: '2',
    label: 'Entity Counts',
    description: 'Quick overview of loaded data. Helps users know data is loaded without visiting dashboard.',
    render: () => (
      <TopbarBase rightSection={
        <Group gap="xs">
          <Badge variant="light" color="gray" size="sm">20 principals</Badge>
          <Badge variant="light" color="gray" size="sm">53 resources</Badge>
          <Badge variant="light" color="gray" size="sm">4 accounts</Badge>
        </Group>
      } />
    ),
  },
  {
    name: '3',
    label: 'Global Search',
    description: 'Search across all entity types from anywhere. Power user feature for quick navigation.',
    render: () => (
      <TopbarBase>
        <TextInput
          placeholder="Search principals, resources, policies..."
          leftSection={<IconSearch size={16} />}
          size="sm"
          w={300}
          styles={{ input: { backgroundColor: 'rgba(255,255,255,0.1)', border: 'none', color: 'white', '::placeholder': { color: 'rgba(255,255,255,0.6)' } } }}
        />
      </TopbarBase>
    ),
  },
  {
    name: '4',
    label: 'Active Overlay Badge',
    description: 'Show when an overlay is active for simulations. Important context for "what-if" scenarios.',
    render: () => (
      <TopbarBase>
        <Badge
          color="pink"
          variant="filled"
          leftSection={<IconLayersSubtract size={12} />}
          style={{ cursor: 'pointer' }}
        >
          Overlay: My Test Scenario
        </Badge>
      </TopbarBase>
    ),
  },
  {
    name: '5',
    label: 'Quick Simulation Button',
    description: 'One-click access to start a new simulation. Reduces navigation for common workflow.',
    render: () => (
      <TopbarBase rightSection={
        <Button variant="white" color="purple" size="sm" leftSection={<IconFlask size={16} />}>
          New Simulation
        </Button>
      } />
    ),
  },
  {
    name: '6',
    label: 'Theme Toggle',
    description: 'Dark/light mode switch. Good for users working in different environments.',
    render: () => (
      <TopbarBase rightSection={
        <ActionIcon variant="subtle" c="white" size="lg">
          <IconMoon size={20} />
        </ActionIcon>
      } />
    ),
  },
  {
    name: '7',
    label: 'Notifications Bell',
    description: 'Alert users to stale data sources or other issues. Proactive problem awareness.',
    render: () => (
      <TopbarBase rightSection={
        <Indicator color="red" size={8} offset={4}>
          <ActionIcon variant="subtle" c="white" size="lg">
            <IconBell size={20} />
          </ActionIcon>
        </Indicator>
      } />
    ),
  },
  {
    name: '8',
    label: 'Server Address',
    description: 'Show connected server address. Useful when working with multiple environments.',
    render: () => (
      <TopbarBase rightSection={
        <Group gap="xs">
          <IconServer size={16} color="white" />
          <Text size="sm" c="white" ff="monospace">localhost:8888</Text>
          <Badge color="green" size="xs">●</Badge>
        </Group>
      } />
    ),
  },
  {
    name: '9',
    label: 'Combo: Health + Overlay',
    description: 'Combined view showing both connection status and active overlay context.',
    render: () => (
      <TopbarBase rightSection={
        <Group gap="md">
          <Badge
            color="pink"
            variant="filled"
            leftSection={<IconLayersSubtract size={12} />}
            rightSection={<IconX size={10} style={{ cursor: 'pointer' }} />}
          >
            Test Overlay
          </Badge>
          <Tooltip label="Server healthy • 3 sources loaded">
            <Badge color="green" variant="dot" size="lg" c="white">
              <IconCircleCheck size={14} />
            </Badge>
          </Tooltip>
        </Group>
      } />
    ),
  },
  {
    name: '10',
    label: 'Combo: Search + Status',
    description: 'Global search with compact status indicator. Balanced utility and information.',
    render: () => (
      <TopbarBase>
        <TextInput
          placeholder="Search..."
          leftSection={<IconSearch size={16} />}
          size="sm"
          w={250}
          styles={{ input: { backgroundColor: 'rgba(255,255,255,0.1)', border: 'none', color: 'white' } }}
        />
        <Group gap="xs" ml="md">
          <Badge size="xs" variant="light" color="gray">20P</Badge>
          <Badge size="xs" variant="light" color="gray">53R</Badge>
        </Group>
      </TopbarBase>
    ),
  },
  {
    name: '11',
    label: 'Combo: Full Featured',
    description: 'Everything: search, overlay status, notifications, and settings. Maximum functionality.',
    render: () => (
      <TopbarBase>
        <TextInput
          placeholder="Search..."
          leftSection={<IconSearch size={16} />}
          size="sm"
          w={200}
          styles={{ input: { backgroundColor: 'rgba(255,255,255,0.1)', border: 'none', color: 'white' } }}
        />
        <Badge color="pink" variant="filled" size="sm" ml="sm" leftSection={<IconLayersSubtract size={10} />}>
          Overlay Active
        </Badge>
      </TopbarBase>
      ),
  },
  {
    name: '12',
    label: 'Stale Data Warning',
    description: 'Prominent warning when data sources are stale. Prevents decisions based on outdated info.',
    render: () => (
      <TopbarBase rightSection={
        <Tooltip label="2 data sources are more than 1 hour old">
          <Badge color="yellow" variant="filled" leftSection={<IconAlertTriangle size={12} />}>
            Stale Data
          </Badge>
        </Tooltip>
      } />
    ),
  },
];

export function PreviewPage(): JSX.Element {
  return (
    <Container size="xl" py="xl">
      <Stack gap="xl">
        {/* Topbar Enhancement Options */}
        <div>
          <Title order={1} mb="xs">Topbar Enhancement Options</Title>
          <Text c="dimmed" mb="md">
            Ideas for what could be added to the topbar. These can be combined or used individually.
          </Text>
        </div>

        <Stack gap="xl">
          {topbarPreviews.map((preview) => (
            <Card key={preview.name} padding="lg" withBorder>
              <Stack gap="md">
                <Group justify="space-between">
                  <div>
                    <Group gap="xs">
                      <Badge size="lg" variant="filled" color="grape">{preview.name}</Badge>
                      <Text fw={700} size="lg">{preview.label}</Text>
                    </Group>
                    <Text size="sm" c="dimmed" mt={4}>{preview.description}</Text>
                  </div>
                </Group>
                <Box>
                  {preview.render()}
                </Box>
              </Stack>
            </Card>
          ))}
        </Stack>

        <Divider my="xl" />

        {/* Status Pill Styles */}
        <div>
          <Title order={1} mb="xs">Status Pill Styles</Title>
          <Text c="dimmed" mb="md">
            Different visual styles for status indicators: Healthy/Unhealthy, Fresh/Stale, Allow/Deny.
          </Text>
        </div>

        <Stack gap="xl">
          {/* Style 1: Filled Badges (Current) */}
          <Card padding="lg" withBorder>
            <Stack gap="md">
              <Group gap="xs">
                <Badge size="lg" variant="filled" color="teal">1</Badge>
                <Text fw={700} size="lg">Filled Badges (Current)</Text>
              </Group>
              <Text size="sm" c="dimmed">Solid background with white text. High contrast and visibility.</Text>
              <Group gap="xl">
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Health</Text>
                  <Group gap="sm">
                    <Badge color="green" variant="filled" size="lg">Healthy</Badge>
                    <Badge color="red" variant="filled" size="lg">Unhealthy</Badge>
                  </Group>
                </Stack>
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Data Freshness</Text>
                  <Group gap="sm">
                    <Badge color="green" variant="filled" size="lg">Fresh</Badge>
                    <Badge color="yellow" variant="filled" size="lg">Stale</Badge>
                  </Group>
                </Stack>
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Simulation Result</Text>
                  <Group gap="sm">
                    <Badge color="green" variant="filled" size="lg">ALLOW</Badge>
                    <Badge color="red" variant="filled" size="lg">DENY</Badge>
                  </Group>
                </Stack>
              </Group>
            </Stack>
          </Card>

          {/* Style 2: Light/Subtle Badges */}
          <Card padding="lg" withBorder>
            <Stack gap="md">
              <Group gap="xs">
                <Badge size="lg" variant="filled" color="teal">2</Badge>
                <Text fw={700} size="lg">Light/Subtle Badges</Text>
              </Group>
              <Text size="sm" c="dimmed">Tinted background with colored text. Softer, less aggressive.</Text>
              <Group gap="xl">
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Health</Text>
                  <Group gap="sm">
                    <Badge color="green" variant="light" size="lg">Healthy</Badge>
                    <Badge color="red" variant="light" size="lg">Unhealthy</Badge>
                  </Group>
                </Stack>
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Data Freshness</Text>
                  <Group gap="sm">
                    <Badge color="green" variant="light" size="lg">Fresh</Badge>
                    <Badge color="yellow" variant="light" size="lg">Stale</Badge>
                  </Group>
                </Stack>
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Simulation Result</Text>
                  <Group gap="sm">
                    <Badge color="green" variant="light" size="lg">ALLOW</Badge>
                    <Badge color="red" variant="light" size="lg">DENY</Badge>
                  </Group>
                </Stack>
              </Group>
            </Stack>
          </Card>

          {/* Style 3: Outline Badges */}
          <Card padding="lg" withBorder>
            <Stack gap="md">
              <Group gap="xs">
                <Badge size="lg" variant="filled" color="teal">3</Badge>
                <Text fw={700} size="lg">Outline Badges</Text>
              </Group>
              <Text size="sm" c="dimmed">Colored border with colored text. Minimal, clean look.</Text>
              <Group gap="xl">
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Health</Text>
                  <Group gap="sm">
                    <Badge color="green" variant="outline" size="lg">Healthy</Badge>
                    <Badge color="red" variant="outline" size="lg">Unhealthy</Badge>
                  </Group>
                </Stack>
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Data Freshness</Text>
                  <Group gap="sm">
                    <Badge color="green" variant="outline" size="lg">Fresh</Badge>
                    <Badge color="orange" variant="outline" size="lg">Stale</Badge>
                  </Group>
                </Stack>
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Simulation Result</Text>
                  <Group gap="sm">
                    <Badge color="green" variant="outline" size="lg">ALLOW</Badge>
                    <Badge color="red" variant="outline" size="lg">DENY</Badge>
                  </Group>
                </Stack>
              </Group>
            </Stack>
          </Card>

          {/* Style 4: Dot Indicator */}
          <Card padding="lg" withBorder>
            <Stack gap="md">
              <Group gap="xs">
                <Badge size="lg" variant="filled" color="teal">4</Badge>
                <Text fw={700} size="lg">Dot Indicator</Text>
              </Group>
              <Text size="sm" c="dimmed">Small colored dot next to text. Compact and unobtrusive.</Text>
              <Group gap="xl">
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Health</Text>
                  <Group gap="sm">
                    <Badge color="green" variant="dot" size="lg">Healthy</Badge>
                    <Badge color="red" variant="dot" size="lg">Unhealthy</Badge>
                  </Group>
                </Stack>
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Data Freshness</Text>
                  <Group gap="sm">
                    <Badge color="green" variant="dot" size="lg">Fresh</Badge>
                    <Badge color="yellow" variant="dot" size="lg">Stale</Badge>
                  </Group>
                </Stack>
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Simulation Result</Text>
                  <Group gap="sm">
                    <Badge color="green" variant="dot" size="lg">ALLOW</Badge>
                    <Badge color="red" variant="dot" size="lg">DENY</Badge>
                  </Group>
                </Stack>
              </Group>
            </Stack>
          </Card>

          {/* Style 5: Icon + Text */}
          <Card padding="lg" withBorder>
            <Stack gap="md">
              <Group gap="xs">
                <Badge size="lg" variant="filled" color="teal">5</Badge>
                <Text fw={700} size="lg">Icon + Text</Text>
              </Group>
              <Text size="sm" c="dimmed">Icon reinforces meaning. Good for quick scanning.</Text>
              <Group gap="xl">
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Health</Text>
                  <Group gap="sm">
                    <Badge color="green" variant="filled" size="lg" leftSection={<IconCircleCheck size={14} />}>Healthy</Badge>
                    <Badge color="red" variant="filled" size="lg" leftSection={<IconCircleX size={14} />}>Unhealthy</Badge>
                  </Group>
                </Stack>
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Data Freshness</Text>
                  <Group gap="sm">
                    <Badge color="green" variant="filled" size="lg" leftSection={<IconCheck size={14} />}>Fresh</Badge>
                    <Badge color="yellow" variant="filled" size="lg" leftSection={<IconClock size={14} />}>Stale</Badge>
                  </Group>
                </Stack>
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Simulation Result</Text>
                  <Group gap="sm">
                    <Badge color="green" variant="filled" size="lg" leftSection={<IconShieldCheck size={14} />}>ALLOW</Badge>
                    <Badge color="red" variant="filled" size="lg" leftSection={<IconShieldX size={14} />}>DENY</Badge>
                  </Group>
                </Stack>
              </Group>
            </Stack>
          </Card>

          {/* Style 6: Icon Only */}
          <Card padding="lg" withBorder>
            <Stack gap="md">
              <Group gap="xs">
                <Badge size="lg" variant="filled" color="teal">6</Badge>
                <Text fw={700} size="lg">Icon Only</Text>
              </Group>
              <Text size="sm" c="dimmed">Just icons, maximum density. Tooltip for details.</Text>
              <Group gap="xl">
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Health</Text>
                  <Group gap="sm">
                    <Tooltip label="Healthy"><Box><IconCircleCheck size={24} color="var(--mantine-color-green-6)" /></Box></Tooltip>
                    <Tooltip label="Unhealthy"><Box><IconCircleX size={24} color="var(--mantine-color-red-6)" /></Box></Tooltip>
                  </Group>
                </Stack>
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Data Freshness</Text>
                  <Group gap="sm">
                    <Tooltip label="Fresh"><Box><IconRefresh size={24} color="var(--mantine-color-green-6)" /></Box></Tooltip>
                    <Tooltip label="Stale"><Box><IconClock size={24} color="var(--mantine-color-yellow-6)" /></Box></Tooltip>
                  </Group>
                </Stack>
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Simulation Result</Text>
                  <Group gap="sm">
                    <Tooltip label="ALLOW"><Box><IconShieldCheck size={24} color="var(--mantine-color-green-6)" /></Box></Tooltip>
                    <Tooltip label="DENY"><Box><IconShieldX size={24} color="var(--mantine-color-red-6)" /></Box></Tooltip>
                  </Group>
                </Stack>
              </Group>
            </Stack>
          </Card>

          {/* Style 7: Pill with Glow */}
          <Card padding="lg" withBorder>
            <Stack gap="md">
              <Group gap="xs">
                <Badge size="lg" variant="filled" color="teal">7</Badge>
                <Text fw={700} size="lg">Pill with Glow/Shadow</Text>
              </Group>
              <Text size="sm" c="dimmed">Subtle shadow/glow effect. Adds depth and emphasis.</Text>
              <Group gap="xl">
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Health</Text>
                  <Group gap="sm">
                    <Badge color="green" variant="filled" size="lg" style={{ boxShadow: '0 0 8px var(--mantine-color-green-4)' }}>Healthy</Badge>
                    <Badge color="red" variant="filled" size="lg" style={{ boxShadow: '0 0 8px var(--mantine-color-red-4)' }}>Unhealthy</Badge>
                  </Group>
                </Stack>
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Data Freshness</Text>
                  <Group gap="sm">
                    <Badge color="green" variant="filled" size="lg" style={{ boxShadow: '0 0 8px var(--mantine-color-green-4)' }}>Fresh</Badge>
                    <Badge color="yellow" variant="filled" size="lg" style={{ boxShadow: '0 0 8px var(--mantine-color-yellow-4)' }}>Stale</Badge>
                  </Group>
                </Stack>
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Simulation Result</Text>
                  <Group gap="sm">
                    <Badge color="green" variant="filled" size="lg" style={{ boxShadow: '0 0 8px var(--mantine-color-green-4)' }}>ALLOW</Badge>
                    <Badge color="red" variant="filled" size="lg" style={{ boxShadow: '0 0 8px var(--mantine-color-red-4)' }}>DENY</Badge>
                  </Group>
                </Stack>
              </Group>
            </Stack>
          </Card>

          {/* Style 8: Rounded Square */}
          <Card padding="lg" withBorder>
            <Stack gap="md">
              <Group gap="xs">
                <Badge size="lg" variant="filled" color="teal">8</Badge>
                <Text fw={700} size="lg">Rounded Square</Text>
              </Group>
              <Text size="sm" c="dimmed">Less rounded corners for a more modern/sharp look.</Text>
              <Group gap="xl">
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Health</Text>
                  <Group gap="sm">
                    <Badge color="green" variant="filled" size="lg" radius="sm">Healthy</Badge>
                    <Badge color="red" variant="filled" size="lg" radius="sm">Unhealthy</Badge>
                  </Group>
                </Stack>
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Data Freshness</Text>
                  <Group gap="sm">
                    <Badge color="green" variant="filled" size="lg" radius="sm">Fresh</Badge>
                    <Badge color="yellow" variant="filled" size="lg" radius="sm">Stale</Badge>
                  </Group>
                </Stack>
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Simulation Result</Text>
                  <Group gap="sm">
                    <Badge color="green" variant="filled" size="lg" radius="sm">ALLOW</Badge>
                    <Badge color="red" variant="filled" size="lg" radius="sm">DENY</Badge>
                  </Group>
                </Stack>
              </Group>
            </Stack>
          </Card>

          {/* Style 9: Text with Background Block */}
          <Card padding="lg" withBorder>
            <Stack gap="md">
              <Group gap="xs">
                <Badge size="lg" variant="filled" color="teal">9</Badge>
                <Text fw={700} size="lg">Text Block</Text>
              </Group>
              <Text size="sm" c="dimmed">Plain text with colored background block. Simple and direct.</Text>
              <Group gap="xl">
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Health</Text>
                  <Group gap="sm">
                    <Box px="md" py={4} bg="green.1" style={{ borderRadius: 4 }}><Text size="sm" fw={600} c="green.8">Healthy</Text></Box>
                    <Box px="md" py={4} bg="red.1" style={{ borderRadius: 4 }}><Text size="sm" fw={600} c="red.8">Unhealthy</Text></Box>
                  </Group>
                </Stack>
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Data Freshness</Text>
                  <Group gap="sm">
                    <Box px="md" py={4} bg="green.1" style={{ borderRadius: 4 }}><Text size="sm" fw={600} c="green.8">Fresh</Text></Box>
                    <Box px="md" py={4} bg="yellow.1" style={{ borderRadius: 4 }}><Text size="sm" fw={600} c="yellow.8">Stale</Text></Box>
                  </Group>
                </Stack>
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Simulation Result</Text>
                  <Group gap="sm">
                    <Box px="md" py={4} bg="green.1" style={{ borderRadius: 4 }}><Text size="sm" fw={600} c="green.8">ALLOW</Text></Box>
                    <Box px="md" py={4} bg="red.1" style={{ borderRadius: 4 }}><Text size="sm" fw={600} c="red.8">DENY</Text></Box>
                  </Group>
                </Stack>
              </Group>
            </Stack>
          </Card>

          {/* Style 10: Gradient Badge */}
          <Card padding="lg" withBorder>
            <Stack gap="md">
              <Group gap="xs">
                <Badge size="lg" variant="filled" color="teal">10</Badge>
                <Text fw={700} size="lg">Gradient Badge</Text>
              </Group>
              <Text size="sm" c="dimmed">Gradient background for a more dynamic look. Eye-catching.</Text>
              <Group gap="xl">
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Health</Text>
                  <Group gap="sm">
                    <Badge size="lg" variant="gradient" gradient={{ from: 'teal', to: 'green', deg: 90 }}>Healthy</Badge>
                    <Badge size="lg" variant="gradient" gradient={{ from: 'orange', to: 'red', deg: 90 }}>Unhealthy</Badge>
                  </Group>
                </Stack>
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Data Freshness</Text>
                  <Group gap="sm">
                    <Badge size="lg" variant="gradient" gradient={{ from: 'teal', to: 'green', deg: 90 }}>Fresh</Badge>
                    <Badge size="lg" variant="gradient" gradient={{ from: 'yellow', to: 'orange', deg: 90 }}>Stale</Badge>
                  </Group>
                </Stack>
                <Stack gap="xs">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>Simulation Result</Text>
                  <Group gap="sm">
                    <Badge size="lg" variant="gradient" gradient={{ from: 'teal', to: 'lime', deg: 90 }}>ALLOW</Badge>
                    <Badge size="lg" variant="gradient" gradient={{ from: 'pink', to: 'red', deg: 90 }}>DENY</Badge>
                  </Group>
                </Stack>
              </Group>
            </Stack>
          </Card>
        </Stack>

        <Divider my="xl" />

        {/* Simulate Icon Options */}
        <div>
          <Title order={1} mb="xs">Simulate Column Icon Options</Title>
          <Text c="dimmed" mb="md">
            Choose an icon for the "Simulate" column that links to Access Check.
          </Text>
          <Card withBorder p="lg">
            <Table>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th style={{ width: 60 }}>Icon</Table.Th>
                  <Table.Th style={{ width: 120 }}>Name</Table.Th>
                  <Table.Th>Description</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {simulateIconOptions.map((opt) => (
                  <Table.Tr key={opt.name}>
                    <Table.Td>
                      <Anchor>
                        <opt.icon size={20} />
                      </Anchor>
                    </Table.Td>
                    <Table.Td>
                      <Text size="sm" ff="monospace">{opt.name}</Text>
                    </Table.Td>
                    <Table.Td>
                      <Text size="sm" c="dimmed">{opt.description}</Text>
                    </Table.Td>
                  </Table.Tr>
                ))}
              </Table.Tbody>
            </Table>
          </Card>
        </div>

        <Divider />

        {/* Simulation Results Table Designs */}
        <div>
          <Title order={1} mb="xs">Simulation Results Table Designs</Title>
          <Text c="dimmed">
            Choose a table style for displaying which-principals, which-actions, and which-resources results.
            Focus on density and formatting for scanning large result sets.
          </Text>
        </div>

        <Stack gap="xl">
          {tableDesigns.map((design) => (
            <Card key={design.name} padding="lg" withBorder>
              <Stack gap="md">
                <Group justify="space-between">
                  <div>
                    <Group gap="xs">
                      <Badge size="lg" variant="filled" color="blue">{design.name}</Badge>
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

        <Divider my="xl" />

        {/* Overlay Editor Designs */}
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
