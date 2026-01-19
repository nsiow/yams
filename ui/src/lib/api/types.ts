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
  principal?: {
    tags?: Record<string, string>;
    inlinePolicies?: Record<string, PolicyDocument>;
    attachedPolicies?: string[];
  };
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
  resource: string;
  explain?: boolean;
}

export interface WhichResourcesRequest {
  principal: string;
  action: string;
  explain?: boolean;
}

export interface WhichActionsRequest {
  principal: string;
  resource: string;
  explain?: boolean;
}

export interface WhichPrincipalsResponse {
  principals: string[];
}

export interface WhichResourcesResponse {
  resources: string[];
}

export interface WhichActionsResponse {
  actions: string[];
}

// API Error

export interface ApiError {
  message: string;
  status: number;
}
