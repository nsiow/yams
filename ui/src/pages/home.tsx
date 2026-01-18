import { Container, Stack, Text, Title } from '@mantine/core';

export function HomePage(): JSX.Element {
  return (
    <Container size="md" py="xl">
      <Stack gap="lg">
        <Title order={1}>Welcome to yams</Title>
        <Text c="dimmed">
          Yet Another Management System for AWS IAM
        </Text>
      </Stack>
    </Container>
  );
}
