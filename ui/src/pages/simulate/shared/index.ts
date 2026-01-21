// ui/src/pages/simulate/shared/index.ts
export { AsyncSearchSelect } from './async-search-select';
export type { AsyncSearchSelectProps } from './async-search-select';
export { OverlaySelector, buildCombinedOverlay } from './overlay-selector';
export { ContextEditor, buildContext } from './context-editor';
export type { ContextVariable } from './context-editor';
export {
  extractService,
  extractAccountId,
  formatPrincipalLabel,
  formatResourceLabel,
  formatRelativeTime,
  highlightMatch,
  buildAccessCheckUrl,
  isS3Object,
  isS3Bucket,
  getS3BucketFromObject,
  getS3ObjectPath,
} from './utils.tsx';
export {
  getSubresourceConfig,
  getBaseArn,
  getSubresourcePath,
  buildArnWithPath,
  SUBRESOURCE_TYPES,
} from './subresource-config';
export type { SubresourceConfig } from './subresource-config';
export { SubresourceEditor } from './subresource-editor';
