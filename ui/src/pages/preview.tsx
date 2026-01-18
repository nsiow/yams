import {
  Box,
  Card,
  Container,
  NavLink,
  Stack,
  Text,
  Title,
} from '@mantine/core';

interface SidebarStyle {
  name: string;
  label: string;
  render: () => JSX.Element;
}

const sampleNav = [
  { title: 'Search', items: ['Principals', 'Policies', 'Resources'] },
  { title: 'Simulate', items: ['Access Check', 'Policy Evaluation'] },
];

const sidebarStyles: SidebarStyle[] = [
  {
    name: 'A',
    label: 'Minimal - Simple text links',
    render: () => (
      <Box p="md">
        {sampleNav.map((section) => (
          <div key={section.title}>
            <Text size="xs" c="dimmed" tt="uppercase" fw={700} mb="xs" mt="md">
              {section.title}
            </Text>
            {section.items.map((item, i) => (
              <Text
                key={item}
                size="sm"
                py={6}
                px="sm"
                style={{ cursor: 'pointer' }}
                c={i === 0 ? 'purple.6' : undefined}
              >
                {item}
              </Text>
            ))}
          </div>
        ))}
      </Box>
    ),
  },
  {
    name: 'B',
    label: 'Highlighted - Background on active',
    render: () => (
      <Box p="md">
        {sampleNav.map((section) => (
          <div key={section.title}>
            <Text size="xs" c="dimmed" tt="uppercase" fw={700} mb="xs" mt="md">
              {section.title}
            </Text>
            {section.items.map((item, i) => (
              <Box
                key={item}
                py={8}
                px="sm"
                style={{ borderRadius: 6, cursor: 'pointer' }}
                bg={i === 0 ? 'purple.1' : undefined}
              >
                <Text size="sm" c={i === 0 ? 'purple.8' : undefined} fw={i === 0 ? 600 : undefined}>
                  {item}
                </Text>
              </Box>
            ))}
          </div>
        ))}
      </Box>
    ),
  },
  {
    name: 'C',
    label: 'Left border - Active indicator line',
    render: () => (
      <Box p="md">
        {sampleNav.map((section) => (
          <div key={section.title}>
            <Text size="xs" c="dimmed" tt="uppercase" fw={700} mb="xs" mt="md">
              {section.title}
            </Text>
            {section.items.map((item, i) => (
              <Box
                key={item}
                py={8}
                px="sm"
                style={{
                  borderLeft: i === 0 ? '3px solid var(--mantine-color-purple-6)' : '3px solid transparent',
                  cursor: 'pointer',
                }}
              >
                <Text size="sm" c={i === 0 ? 'purple.6' : undefined} fw={i === 0 ? 600 : undefined}>
                  {item}
                </Text>
              </Box>
            ))}
          </div>
        ))}
      </Box>
    ),
  },
  {
    name: 'D',
    label: 'Mantine NavLink - Default component',
    render: () => (
      <Box p="md">
        {sampleNav.map((section) => (
          <div key={section.title}>
            <Text size="xs" c="dimmed" tt="uppercase" fw={700} mb="xs" mt="md">
              {section.title}
            </Text>
            {section.items.map((item, i) => (
              <NavLink key={item} label={item} active={i === 0} />
            ))}
          </div>
        ))}
      </Box>
    ),
  },
  {
    name: 'E',
    label: 'Pill style - Rounded active background',
    render: () => (
      <Box p="md">
        {sampleNav.map((section) => (
          <div key={section.title}>
            <Text size="xs" c="dimmed" tt="uppercase" fw={700} mb="xs" mt="md">
              {section.title}
            </Text>
            {section.items.map((item, i) => (
              <Box
                key={item}
                py={8}
                px="md"
                my={2}
                style={{ borderRadius: 20, cursor: 'pointer' }}
                bg={i === 0 ? 'purple.6' : undefined}
              >
                <Text size="sm" c={i === 0 ? 'white' : undefined} fw={i === 0 ? 500 : undefined}>
                  {item}
                </Text>
              </Box>
            ))}
          </div>
        ))}
      </Box>
    ),
  },
];

export function PreviewPage(): JSX.Element {
  return (
    <Container size="lg" py="xl">
      <Stack gap="lg">
        <Title order={1}>UI Preview</Title>
        <Text c="dimmed">
          Choose a sidebar style (A-E):
        </Text>

        <Stack gap="md">
          {sidebarStyles.map((style) => (
            <Card key={style.name} padding="md" withBorder>
              <Stack gap="xs">
                <Text fw={700}>Option {style.name}: {style.label}</Text>
                <Box
                  style={{ border: '1px solid var(--mantine-color-gray-3)', borderRadius: 8 }}
                  w={250}
                  bg="white"
                >
                  {style.render()}
                </Box>
              </Stack>
            </Card>
          ))}
        </Stack>
      </Stack>
    </Container>
  );
}
