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
} from '@mantine/core';
import { IconChevronDown } from '@tabler/icons-react';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';

interface NavItem {
  label: string;
  path: string;
}

interface NavSection {
  title: string;
  items: NavItem[];
}

const navSections: NavSection[] = [
  {
    title: 'Search',
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
    items: [
      { label: 'Access Check', path: '/simulate/access' },
    ],
  },
];

export function Layout(): JSX.Element {
  const location = useLocation();
  const navigate = useNavigate();

  return (
    <AppShell
      header={{ height: 60 }}
      navbar={{ width: 200, breakpoint: 'sm' }}
      padding="md"
    >
      <AppShell.Header bg="purple.6">
        <Group h="100%" px="md" justify="space-between">
          <Group>
            <Image src="/apple-touch-icon.png" w={44} h={44} />
            <Title order={3} c="white" ff="'Urbanist', sans-serif" fz="xl" style={{ letterSpacing: '0.15em' }}>
              yams
            </Title>
            <Badge color="pink" variant="filled">
              UI Preview
            </Badge>
          </Group>
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
      </AppShell.Header>

      <AppShell.Navbar p="md">
        <Stack gap="lg">
          <Box
            py={8}
            px="sm"
            style={{ borderRadius: 6, cursor: 'pointer' }}
            bg={location.pathname === '/' ? 'purple.1' : undefined}
            onClick={() => navigate('/')}
          >
            <Text
              size="sm"
              c={location.pathname === '/' ? 'purple.8' : undefined}
              fw={location.pathname === '/' ? 600 : undefined}
            >
              Home
            </Text>
          </Box>
          {navSections.map((section) => (
            <div key={section.title}>
              <Text size="xs" c="dimmed" tt="uppercase" fw={700} mb="xs">
                {section.title}
              </Text>
              {section.items.map((item) => {
                const isActive = location.pathname.startsWith(item.path);
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
