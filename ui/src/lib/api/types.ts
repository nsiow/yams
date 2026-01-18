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

export interface Action {
  key: string;
  service: string;
  action: string;
  description?: string;
  resourceTypes?: string[];
  conditionKeys?: string[];
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

export interface Resource {
  arn: string;
  type: string;
  accountId: string;
  region?: string;
  name?: string;
  policy?: unknown;
}

export interface Policy {
  arn: string;
  name: string;
  type: string;
  document: PolicyDocument;
}

export interface PolicyDocument {
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

export interface Account {
  id: string;
  name?: string;
  email?: string;
}

// Simulation Types

export interface SimulationRequest {
  principal: string;
  action: string;
  resource: string;
  explain?: boolean;
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
  decision: 'Allow' | 'Deny';
  explanation?: SimulationExplanation;
}

export interface SimulationExplanation {
  matchedStatements: MatchedStatement[];
  reason: string;
}

export interface MatchedStatement {
  policyArn: string;
  statementId?: string;
  effect: 'Allow' | 'Deny';
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
