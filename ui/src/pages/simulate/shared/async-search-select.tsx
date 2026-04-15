// ui/src/pages/simulate/shared/async-search-select.tsx
import { useEffect, useState } from 'react';
import {
  ActionIcon,
  Box,
  Combobox,
  Group,
  Input,
  InputBase,
  Loader,
  ScrollArea,
  Text,
  Tooltip,
  useCombobox,
} from '@mantine/core';
import { useDebouncedValue } from '@mantine/hooks';
import { IconSearch, IconX } from '@tabler/icons-react';
import { extractAccountId, extractService, highlightMatch } from './utils';

export interface AsyncSearchSelectProps {
  label: string;
  placeholder: string;
  value: string | null;
  onChange: (value: string | null) => void;
  onSearch: (query: string) => Promise<string[]>;
  formatLabel?: (value: string) => string;
  accountNames?: Record<string, string>;
  resourceAccounts?: Record<string, string>;
  accessLevels?: Record<string, string>;
  showAccountName?: boolean;
  showResourceType?: boolean;
  showAccessLevel?: boolean;
  disabled?: boolean;
  disabledMessage?: string;
}

export function AsyncSearchSelect({
  label,
  placeholder,
  value,
  onChange,
  onSearch,
  formatLabel,
  accountNames,
  resourceAccounts,
  accessLevels,
  showAccountName,
  showResourceType,
  showAccessLevel,
  disabled,
  disabledMessage,
}: AsyncSearchSelectProps): JSX.Element {
  const combobox = useCombobox({
    onDropdownClose: () => combobox.resetSelectedOption(),
  });

  const [search, setSearch] = useState('');
  const [debouncedSearch] = useDebouncedValue(search, 300);
  const [loading, setLoading] = useState(false);
  const [options, setOptions] = useState<string[]>([]);

  // Search when debounced value changes
  useEffect(() => {
    if (debouncedSearch.length < 2) {
      setOptions([]);
      return;
    }

    setLoading(true);
    onSearch(debouncedSearch)
      .then((results) => {
        setOptions(results.slice(0, 100));
      })
      .catch((err) => {
        console.error('Search failed:', err);
        setOptions([]);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [debouncedSearch, onSearch]);

  const displayValue = value ? (formatLabel ? formatLabel(value) : value) : '';

  const handleClear = (e: React.MouseEvent): void => {
    e.stopPropagation();
    onChange(null);
    setSearch('');
  };

  // Render disabled state
  if (disabled) {
    return (
      <Input.Wrapper label={label}>
        <Box
          style={{
            border: '1px solid var(--mantine-color-gray-3)',
            borderRadius: 'var(--mantine-radius-sm)',
            padding: '12px',
            height: '58px',
            display: 'flex',
            alignItems: 'center',
            backgroundColor: 'var(--mantine-color-gray-1)',
            cursor: 'not-allowed',
          }}
        >
          <Text size="sm" c="dimmed">
            {disabledMessage || 'Not available'}
          </Text>
        </Box>
      </Input.Wrapper>
    );
  }

  return (
    <Combobox
      store={combobox}
      onOptionSubmit={(val) => {
        onChange(val);
        setSearch('');
        combobox.closeDropdown();
      }}
    >
      <Combobox.Target>
        <Input.Wrapper label={label}>
          {value && !combobox.dropdownOpened ? (
            (() => {
              const accountId = showAccountName
                ? (extractAccountId(value) || (resourceAccounts ? resourceAccounts[value] : null))
                : null;
              const accountName = accountId && accountNames ? accountNames[accountId] : null;
              const service = showResourceType ? extractService(value) : null;
              const accessLevel = showAccessLevel && accessLevels ? accessLevels[value] : null;
              return (
                <Tooltip label={value} multiline maw={400} openDelay={500}>
                  <Group
                    gap="xs"
                    justify="space-between"
                    wrap="nowrap"
                    onClick={() => combobox.openDropdown()}
                    style={{
                      border: '1px solid var(--mantine-color-gray-4)',
                      borderRadius: 'var(--mantine-radius-sm)',
                      padding: '8px 12px',
                      height: '58px',
                      cursor: 'pointer',
                    }}
                  >
                    <Box style={{ overflow: 'hidden', flex: 1 }}>
                      <Text size="sm" truncate>
                        {displayValue}
                      </Text>
                      {showResourceType && (accountName || service) && (
                        <Text size="xs" c="dimmed" truncate>
                          {[accountName ? `${accountName} (${accountId})` : accountId, service].filter(Boolean).join(' · ')}
                        </Text>
                      )}
                      {showAccountName && !showResourceType && accountName && (
                        <Text size="xs" c="dimmed" truncate>
                          {accountName} ({accountId})
                        </Text>
                      )}
                      {showAccessLevel && accessLevel && (
                        <Text size="xs" c="dimmed" truncate>
                          {accessLevel}
                        </Text>
                      )}
                      {!showAccountName && !showResourceType && !showAccessLevel && formatLabel && (
                        <Text size="xs" c="dimmed" truncate>
                          {value}
                        </Text>
                      )}
                    </Box>
                    <ActionIcon
                      size="sm"
                      variant="subtle"
                      color="gray"
                      onClick={handleClear}
                      aria-label="Clear selection"
                    >
                      <IconX size={14} />
                    </ActionIcon>
                  </Group>
                </Tooltip>
              );
            })()
          ) : (
            <InputBase
              size="lg"
              rightSection={loading ? <Loader size={18} /> : <IconSearch size={18} color="var(--mantine-color-dimmed)" />}
              value={search}
              onChange={(event) => {
                setSearch(event.currentTarget.value);
                combobox.openDropdown();
                combobox.updateSelectedOptionIndex();
              }}
              onClick={() => combobox.openDropdown()}
              onFocus={() => combobox.openDropdown()}
              onBlur={() => {
                combobox.closeDropdown();
                setSearch('');
              }}
              placeholder={placeholder}
              styles={{ input: { height: '58px' } }}
              rightSectionPointerEvents="none"
            />
          )}
        </Input.Wrapper>
      </Combobox.Target>

      <Combobox.Dropdown>
        <Combobox.Options>
          <ScrollArea.Autosize mah={300} type="scroll">
            {loading ? (
              <Combobox.Empty>Searching...</Combobox.Empty>
            ) : options.length === 0 ? (
              <Combobox.Empty>
                {debouncedSearch.length < 2 ? 'Type at least 2 characters to search' : 'No results found'}
              </Combobox.Empty>
            ) : (
              options.map((option) => {
                const accountId = showAccountName
                  ? (extractAccountId(option) || (resourceAccounts ? resourceAccounts[option] : null))
                  : null;
                const accountName = accountId && accountNames ? accountNames[accountId] : null;
                const service = showResourceType ? extractService(option) : null;
                const accessLevel = showAccessLevel && accessLevels ? accessLevels[option] : null;
                return (
                  <Combobox.Option value={option} key={option}>
                    <Text size="sm" truncate>
                      {highlightMatch(formatLabel ? formatLabel(option) : option, debouncedSearch)}
                    </Text>
                    {showResourceType && (accountName || service) && (
                      <Text size="xs" c="dimmed" truncate>
                        {[accountName ? `${accountName} (${accountId})` : accountId, service].filter(Boolean).join(' · ')}
                      </Text>
                    )}
                    {showAccountName && !showResourceType && accountName && (
                      <Text size="xs" c="dimmed" truncate>
                        {accountName} ({accountId})
                      </Text>
                    )}
                    {showAccessLevel && accessLevel && (
                      <Text size="xs" c="dimmed" truncate>
                        {accessLevel}
                      </Text>
                    )}
                    {!showAccountName && !showResourceType && !showAccessLevel && formatLabel && (
                      <Text size="xs" c="dimmed" truncate>
                        {option}
                      </Text>
                    )}
                  </Combobox.Option>
                );
              })
            )}
          </ScrollArea.Autosize>
        </Combobox.Options>
      </Combobox.Dropdown>
    </Combobox>
  );
}
