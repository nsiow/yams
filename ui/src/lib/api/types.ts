// API Response Types

export interface SourceStatus {
  source: string;
  updated: string;
}

export interface StatusResponse {
  accounts: number;
  entities: number;
  groups: number;
  policies: number;
  principals: number;
  resources: number;
  sources: SourceStatus[];
  env?: Record<string, string>;
}

export interface ActionResource {
  Name: string;
  ARNFormats?: string[];
  ConditionKeys?: string[];
}

export interface Action {
  Name: string;
  Service: string;
  AccessLevel?: string;
  ActionConditionKeys?: string[];
  ResolvedResources?: ActionResource[];
}

export interface PrincipalTag {
  Key: string;
  Value: string;
}

export interface Principal {
  Type: string;
  AccountId: string;
  Name: string;
  Arn: string;
  Tags?: PrincipalTag[];
  InlinePolicies?: PolicyDocument[];
  AttachedPolicies?: string[];
  Groups?: string[] | null;
  PermissionsBoundary?: string;
}

export interface ResourceTag {
  Key: string;
  Value: string;
}

export interface Resource {
  Type: string;
  AccountId: string;
  Region: string;
  Name: string;
  Arn: string;
  Tags?: ResourceTag[];
  Policy?: PolicyDocument;
}

export interface Policy {
  Type: string;
  AccountId: string;
  Arn: string;
  Name: string;
  Policy: PolicyDocument;
}

export interface PolicyDocument {
  _Name?: string;
  Version: string;
  Id?: string;
  Statement: PolicyStatement[];
}

export interface PolicyStatement {
  Sid?: string;
  Effect: 'Allow' | 'Deny';
  Action: string | string[];
  Resource: string | string[];
  Condition?: Record<string, Record<string, string | string[]>>;
  Principal?: string | Record<string, string | string[]>;
}

export interface OrgNode {
  Id: string;
  Type: 'ROOT' | 'ORGANIZATIONAL_UNIT' | 'ACCOUNT';
  Arn: string;
  Name: string;
  SCPs?: string[];
  RCPs?: string[];
}

export interface Account {
  Id: string;
  Name: string;
  OrgId: string;
  OrgPaths?: string[];
  OrgNodes?: OrgNode[];
  // TODO(nsiow): Add Tags field when available in API
}

// Simulation Types

export interface SimulationRequest {
  principal: string;
  action: string;
  resource: string;
  context?: Record<string, string>;
  fuzzy?: boolean;
  explain?: boolean;
  trace?: boolean;
  overlay?: SimulationOverlay;
}

export interface SimulationOverlay {
  accounts?: Account[];
  groups?: Group[];
  policies?: Policy[];
  principals?: Principal[];
  resources?: Resource[];
}

export interface SimulationResponse {
  result: 'ALLOW' | 'DENY';
  principal: string;
  action: string;
  resource?: string;
  explain?: string[];
  trace?: string[];
}

export interface WhichPrincipalsRequest {
  action: string;
  resource?: string;
  context?: Record<string, string>;
  overlay?: SimulationOverlay;
  fuzzy?: boolean;
}

export interface WhichResourcesRequest {
  principal: string;
  action?: string;
  context?: Record<string, string>;
  overlay?: SimulationOverlay;
  fuzzy?: boolean;
}

export interface WhichActionsRequest {
  principal: string;
  resource: string;
  context?: Record<string, string>;
  overlay?: SimulationOverlay;
  fuzzy?: boolean;
}

// Server returns bare arrays for which-* endpoints
export type WhichPrincipalsResponse = string[];
export type WhichResourcesResponse = string[];
export type WhichActionsResponse = string[];

// Overlay Types

export interface OverlaySummary {
  name: string;
  id: string;
  createdAt: string;
  numPrincipals: number;
  numResources: number;
  numPolicies: number;
  numAccounts: number;
  numGroups: number;
}

export interface OverlayData {
  name: string;
  id: string;
  createdAt: string;
  accounts?: Account[];
  groups?: Group[];
  policies?: Policy[];
  principals?: Principal[];
  resources?: Resource[];
}

export interface CreateOverlayRequest {
  name: string;
  accounts?: Account[];
  groups?: Group[];
  policies?: Policy[];
  principals?: Principal[];
  resources?: Resource[];
}

export interface UpdateOverlayRequest {
  name?: string;
  accounts?: Account[];
  groups?: Group[];
  policies?: Policy[];
  principals?: Principal[];
  resources?: Resource[];
}

export interface Group {
  Type?: string;
  AccountId?: string;
  Arn: string;
  InlinePolicies?: PolicyDocument[];
  AttachedPolicies?: string[];
}

// Action Targeting

export interface ActionTargeting {
  action: string;
  arnFormats: string[];
  customHandling?: string[];
  hasTargets: boolean;
}

// API Error

export interface ApiError {
  message: string;
  status: number;
}
