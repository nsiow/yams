import { ActionIcon, CopyButton, Tooltip } from '@mantine/core';
import { IconCheck, IconCopy } from '@tabler/icons-react';

interface CopyEntityButtonProps {
  data: unknown;
}

export function CopyEntityButton({ data }: CopyEntityButtonProps): JSX.Element {
  const json = JSON.stringify(data, null, 2);

  return (
    <CopyButton value={json} timeout={2000}>
      {({ copied, copy }) => (
        <Tooltip label={copied ? 'Copied!' : 'Copy to clipboard'} withArrow position="left">
          <ActionIcon
            color={copied ? 'teal' : 'gray'}
            variant="subtle"
            onClick={copy}
            size="lg"
          >
            {copied ? <IconCheck size={18} /> : <IconCopy size={18} />}
          </ActionIcon>
        </Tooltip>
      )}
    </CopyButton>
  );
}
