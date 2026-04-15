import { useState } from 'react';
import { Card, Collapse, Group, Title, UnstyledButton } from '@mantine/core';
import { IconChevronDown, IconChevronRight } from '@tabler/icons-react';

interface CollapsibleCardProps {
  title: string;
  children: React.ReactNode;
  defaultOpen?: boolean;
}

export function CollapsibleCard({
  title,
  children,
  defaultOpen = true,
}: CollapsibleCardProps): JSX.Element {
  const [opened, setOpened] = useState(defaultOpen);

  return (
    <Card withBorder p="sm">
      <UnstyledButton onClick={() => setOpened((o) => !o)} w="100%">
        <Group gap="xs">
          {opened ? <IconChevronDown size={16} /> : <IconChevronRight size={16} />}
          <Title order={5}>{title}</Title>
        </Group>
      </UnstyledButton>
      <Collapse in={opened}>
        <div style={{ paddingTop: '12px' }}>{children}</div>
      </Collapse>
    </Card>
  );
}
