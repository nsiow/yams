// ui/src/pages/simulate/shared/subresource-config.ts
// Configuration for subresource path customization

export interface SubresourceConfig {
  type: string; // e.g., "s3-object"
  arnPattern: RegExp; // Pattern to detect this subresource type
  defaultPath: string; // Default path value
  label: string; // Display label
}

export const SUBRESOURCE_TYPES: SubresourceConfig[] = [
  {
    type: 's3-object',
    arnPattern: /^arn:aws:s3:::[^/]+\/.+$/,
    defaultPath: 'object.txt',
    label: 'S3 Object Path',
  },
];

export function getSubresourceConfig(arn: string): SubresourceConfig | null {
  return SUBRESOURCE_TYPES.find((c) => c.arnPattern.test(arn)) || null;
}

export function getBaseArn(arn: string, config: SubresourceConfig): string {
  // For S3 objects, return the bucket ARN
  if (config.type === 's3-object') {
    const idx = arn.indexOf('/');
    return idx > -1 ? arn.substring(0, idx) : arn;
  }
  return arn;
}

export function getSubresourcePath(arn: string, config: SubresourceConfig): string {
  if (config.type === 's3-object') {
    const idx = arn.indexOf('/');
    return idx > -1 ? arn.substring(idx + 1) : config.defaultPath;
  }
  return config.defaultPath;
}

export function buildArnWithPath(baseArn: string, path: string): string {
  return `${baseArn}/${path}`;
}
