import {
  AppShell,
  Badge,
  Box,
  Button,
  Group,
  Image,
  Menu,
  Stack,
  Text,
  Title,
  useMantineColorScheme,
} from '@mantine/core';
import {
  IconChevronDown,
  IconFlask,
  IconLayersSubtract,
  IconSearch,
  IconServer,
} from '@tabler/icons-react';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';

interface NavItem {
  label: string;
  path: string;
}

interface NavSection {
  title: string;
  icon: React.ComponentType<{ size?: number | string }>;
  items: NavItem[];
}

const navSections: NavSection[] = [
  {
    title: 'Server',
    icon: IconServer,
    items: [
      { label: 'Status', path: '/' },
    ],
  },
  {
    title: 'Explore',
    icon: IconSearch,
    items: [
      { label: 'Accounts', path: '/search/accounts' },
      { label: 'Principals', path: '/search/principals' },
      { label: 'Resources', path: '/search/resources' },
      { label: 'Policies', path: '/search/policies' },
      { label: 'Actions', path: '/search/actions' },
    ],
  },
  {
    title: 'Simulate',
    icon: IconFlask,
    items: [
      { label: 'Access Check', path: '/simulate/access' },
      { label: 'Which Principals', path: '/simulate/which-principals' },
      { label: 'Which Actions', path: '/simulate/which-actions' },
      { label: 'Which Resources', path: '/simulate/which-resources' },
    ],
  },
  {
    title: 'Overlays',
    icon: IconLayersSubtract,
    items: [
      { label: 'Manage', path: '/overlays' },
      { label: 'Editor', path: '/overlays/new/edit' },
    ],
  },
];

export function Layout(): JSX.Element {
  const location = useLocation();
  const navigate = useNavigate();
  const { setColorScheme } = useMantineColorScheme();

  return (
    <AppShell
      header={{ height: 60 }}
      navbar={{ width: 200, breakpoint: 'sm' }}
      padding="md"
    >
      <AppShell.Header bg="purple.6">
        <Group h="100%" px="md" justify="space-between">
          <Group>
            <Group gap="sm" style={{ cursor: 'pointer' }} onClick={() => navigate('/')}>
              <Image src="/apple-touch-icon.png" w={44} h={44} />
              <Title order={3} c="white" ff="'Urbanist', sans-serif" fz="xl" style={{ letterSpacing: '0.15em' }}>
                yams
              </Title>
            </Group>
            <Badge
              color="pink"
              variant="filled"
              style={{ cursor: 'pointer' }}
              onClick={() => navigate('/preview')}
            >
              UI Preview
            </Badge>
          </Group>
          <Group gap="md" align="center">
            <Group gap="xs" align="center">
              <Badge color="green" size="xs" circle />
              <Text size="sm" c="white" ff="monospace" style={{ opacity: 0.9 }}>localhost:8888</Text>
            </Group>
            <Menu shadow="md" width={150}>
              <Menu.Target>
                <Button variant="subtle" c="white" rightSection={<IconChevronDown size={16} />}>Theme</Button>
              </Menu.Target>
              <Menu.Dropdown>
                <Menu.Item onClick={() => setColorScheme('light')}>
                  Light
                </Menu.Item>
                <Menu.Item onClick={() => setColorScheme('dark')}>
                  Dark
                </Menu.Item>
                <Menu.Item onClick={() => setColorScheme('auto')}>
                  System
                </Menu.Item>
              </Menu.Dropdown>
            </Menu>
            <Menu shadow="md" width={200}>
              <Menu.Target>
                <Button variant="subtle" c="white" rightSection={<IconChevronDown size={16} />}>Docs</Button>
              </Menu.Target>
              <Menu.Dropdown>
                <Menu.Item
                  component="a"
                  href="https://nsiow.github.io/yams/"
                  target="_blank"
                >
                  API Documentation
                </Menu.Item>
              </Menu.Dropdown>
            </Menu>
          </Group>
        </Group>
      </AppShell.Header>

      <AppShell.Navbar p="md">
        <Stack gap="lg">
          {navSections.map((section) => (
            <div key={section.title}>
              <Group gap={6} mb="xs">
                <Box style={{ opacity: 0.6 }}>
                  <section.icon size={14} />
                </Box>
                <Text size="xs" c="dimmed" tt="uppercase" fw={700}>
                  {section.title}
                </Text>
              </Group>
              {section.items.map((item) => {
                // Special handling: any /overlays/*/edit path should highlight Editor
                const isOverlayEditor = /^\/overlays\/[^/]+\/edit$/.test(location.pathname);
                const isEditorItem = item.path === '/overlays/new/edit';

                // Find the best matching path in this section (longest match wins)
                const matchingPaths = section.items
                  .filter(i => {
                    // Editor item matches any overlay edit path
                    if (i.path === '/overlays/new/edit' && isOverlayEditor) return true;
                    if (i.path === '/') return location.pathname === '/';
                    return location.pathname === i.path || location.pathname.startsWith(i.path + '/');
                  })
                  .sort((a, b) => b.path.length - a.path.length);
                const bestMatch = matchingPaths[0]?.path;
                let isActive = item.path === bestMatch;

                // Override: Editor always wins when on any overlay edit path
                if (isOverlayEditor && isEditorItem) isActive = true;
                if (isOverlayEditor && item.path === '/overlays' && !isEditorItem) isActive = false;
                return (
                  <Box
                    key={item.path}
                    py={8}
                    px="sm"
                    style={{ borderRadius: 6, cursor: 'pointer' }}
                    bg={isActive ? 'purple.1' : undefined}
                    onClick={() => navigate(item.path)}
                  >
                    <Text
                      size="sm"
                      c={isActive ? 'purple.8' : undefined}
                      fw={isActive ? 600 : undefined}
                    >
                      {item.label}
                    </Text>
                  </Box>
                );
              })}
            </div>
          ))}
        </Stack>
      </AppShell.Navbar>

      <AppShell.Main>
        <Outlet />
      </AppShell.Main>
    </AppShell>
  );
}
