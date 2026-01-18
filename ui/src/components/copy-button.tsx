import { ActionIcon, CopyButton as MantineCopyButton, Tooltip } from '@mantine/core';
import { IconCheck, IconCopy } from '@tabler/icons-react';

interface CopyButtonProps {
  value: string;
  size?: number;
}

export function CopyButton({ value, size = 16 }: CopyButtonProps): JSX.Element {
  return (
    <MantineCopyButton value={value} timeout={2000}>
      {({ copied, copy }) => (
        <Tooltip label={copied ? 'Copied!' : 'Copy'} withArrow position="right">
          <ActionIcon
            color={copied ? 'teal' : 'gray'}
            variant="subtle"
            onClick={copy}
            size="sm"
          >
            {copied ? <IconCheck size={size} /> : <IconCopy size={size} />}
          </ActionIcon>
        </Tooltip>
      )}
    </MantineCopyButton>
  );
}
