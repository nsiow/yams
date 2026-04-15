// ui/src/pages/simulate/shared/resource-creation.ts
// Utilities for handling resource-creating actions (Create*, RunInstances, etc.)

import type { Action, ActionResource } from '../../../lib/api';

// Actions that match these patterns are considered "resource creation" actions
// For these actions, the resource may not exist yet, so we show an ARN editor
// instead of a resource search dropdown
const RESOURCE_CREATION_PATTERNS: RegExp[] = [
  /^ec2:RunInstances$/i,
  /^[a-z0-9-]+:Create[A-Z]/i, // service:Create* (action must start with Create followed by uppercase)
];

// Actions that match the above patterns but should be excluded (false positives)
const RESOURCE_CREATION_EXCLUSIONS: RegExp[] = [
  // Empty for now - add patterns here as needed
];

// Check if an action is a resource creation action
export function isResourceCreationAction(action: string): boolean {
  // Check if action matches any creation pattern
  const matchesCreationPattern = RESOURCE_CREATION_PATTERNS.some((pattern) =>
    pattern.test(action)
  );
  if (!matchesCreationPattern) {
    return false;
  }

  // Check if action is excluded
  const isExcluded = RESOURCE_CREATION_EXCLUSIONS.some((pattern) =>
    pattern.test(action)
  );
  return !isExcluded;
}

// Find the primary resource type for an action based on the action name
// For example, ec2:CreateSecurityGroup -> security-group, ec2:RunInstances -> instance
export function findPrimaryResource(action: Action): ActionResource | null {
  const resources = action.ResolvedResources;
  if (!resources || resources.length === 0) {
    return null;
  }

  // If only one resource, use it
  if (resources.length === 1) {
    return resources[0];
  }

  // Extract the action name part (after the colon)
  const actionName = action.Name.toLowerCase();

  // Special case for RunInstances -> instance
  if (actionName === 'runinstances') {
    const instance = resources.find((r) => r.Name.toLowerCase() === 'instance');
    if (instance) {
      return instance;
    }
  }

  // For Create* actions, try to match the resource name from the action name
  // e.g., CreateSecurityGroup -> security-group, CreateBucket -> bucket
  if (actionName.startsWith('create')) {
    const targetName = actionName.slice(6).toLowerCase(); // Remove 'create'

    // Try exact match first
    const exactMatch = resources.find(
      (r) => r.Name.toLowerCase().replace(/-/g, '') === targetName
    );
    if (exactMatch) {
      return exactMatch;
    }

    // Try partial match
    const partialMatch = resources.find((r) =>
      targetName.includes(r.Name.toLowerCase().replace(/-/g, ''))
    );
    if (partialMatch) {
      return partialMatch;
    }
  }

  // Fall back to first resource
  return resources[0];
}

// Default placeholder account ID for when no principal is selected
export const PLACEHOLDER_ACCOUNT_ID = '000000000000';

// Format an ARN template with sensible defaults
// Replaces wildcards (*) in the ARN format with placeholder values
export function formatArnWithDefaults(
  arnFormat: string,
  accountId?: string
): string {
  // ARN format: arn:partition:service:region:account:resource
  // Examples:
  //   arn:*:ec2:*:*:security-group/*
  //   arn:*:s3:::*
  //   arn:*:ec2:*::image/*

  const parts = arnFormat.split(':');
  if (parts.length < 6) {
    // Invalid ARN format, return as-is with wildcards replaced
    return arnFormat.replace(/\*/g, 'placeholder');
  }

  // Replace partition (parts[1])
  if (parts[1] === '*') {
    parts[1] = 'aws';
  }

  // Replace region (parts[3]) - only if it's a wildcard
  if (parts[3] === '*') {
    parts[3] = 'us-east-1';
  }

  // Replace account (parts[4]) - only if it's a wildcard
  if (parts[4] === '*') {
    parts[4] = accountId || PLACEHOLDER_ACCOUNT_ID;
  }

  // Replace resource wildcards (parts[5] and beyond)
  // The resource segment can contain wildcards like "security-group/*"
  const resourceParts = parts.slice(5).join(':');
  const formattedResource = resourceParts.replace(/\*/g, 'placeholder');
  parts.splice(5, parts.length - 5, formattedResource);

  return parts.join(':');
}

// Get the default ARN for a resource creation action
export function getDefaultArnForAction(
  action: Action,
  accountId?: string
): string | null {
  const primaryResource = findPrimaryResource(action);
  if (!primaryResource) {
    return null;
  }

  const arnFormats = primaryResource.ARNFormats;
  if (!arnFormats || arnFormats.length === 0) {
    return null;
  }

  // Use the first ARN format
  return formatArnWithDefaults(arnFormats[0], accountId);
}
