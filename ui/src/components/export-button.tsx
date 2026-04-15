import { ActionIcon, Tooltip } from '@mantine/core';
import { IconDownload } from '@tabler/icons-react';

interface ExportButtonProps {
  data: unknown;
  filename: string;
}

export function ExportButton({ data, filename }: ExportButtonProps): JSX.Element {
  const handleExport = (): void => {
    const json = JSON.stringify(data, null, 2);
    const blob = new Blob([json], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${filename}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  return (
    <Tooltip label="Export as JSON" withArrow position="left">
      <ActionIcon variant="subtle" color="gray" onClick={handleExport} size="lg">
        <IconDownload size={18} />
      </ActionIcon>
    </Tooltip>
  );
}
