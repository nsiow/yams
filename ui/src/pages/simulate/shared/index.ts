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
export {
  isResourceCreationAction,
  findPrimaryResource,
  formatArnWithDefaults,
  getDefaultArnForAction,
  PLACEHOLDER_ACCOUNT_ID,
} from './resource-creation';
export { ArnEditor } from './arn-editor';
export type { ArnEditorProps } from './arn-editor';
export { useSharedContext } from './use-shared-context';
