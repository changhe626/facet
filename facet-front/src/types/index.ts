export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export interface PaginatedResponse<T> {
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
  items: T[];
}

export interface User {
  id: string;
  email: string;
  displayName: string;
  role: string;
  lastLoginAt?: string | null;
  createdAt: string;
  updatedAt?: string;
  createdBy?: string;
  updatedBy?: string;
}

export interface Environment {
  id: string;
  tenantId?: string;
  name: string;
  description?: string;
  sdkKey?: string;
  isActive: boolean;
  createdAt: string;
  updatedAt?: string;
}

export interface WhitelistRule {
  id?: string;
  accountId?: string;
  userIDs?: string[];
}

export interface FeatureFlag {
  id: string;
  name: string;
  description?: string;
  enabled: boolean;
  flagType: string;
  expiresAt?: string | null;
  rolloutPercentage?: number;
  archivedAt?: string | null;
  createdAt: string;
  updatedAt: string;
  createdBy?: string;
  updatedBy?: string;
  whitelistRules?: WhitelistRule[];
  segmentIds?: string[];
}

export interface FeatureFlagDetail extends FeatureFlag {
  rules?: WhitelistRule[];
  history?: FeatureFlagHistory[];
  rolloutPercentage?: number;
  segmentIds?: string[];
}

export interface FeatureFlagHistory {
  historyId: string;
  changeType: string;
  changedBy: string;
  changeTime: string;
  name: string;
  description: string;
  enabled: boolean;
  flagType: string;
  expiresAt?: string | null;
  rulesSnapshot?: WhitelistRule[];
}

export interface Tenant {
  id: string;
  name: string;
  description?: string;
  website?: string;
  status: string;
  levelId: string;
  levelName?: string;
  createdAt: string;
  createdBy?: string;
  updatedBy?: string;
}

export interface Notification {
  id: string;
  level: string;
  type: string;
  title: string;
  content?: string;
  createdAt: string;
  isRead: boolean;
  readAt?: string | null;
}

export interface SystemNotification {
  id: string;
  level: string;
  type: string;
  title: string;
  content: string;
  createdAt: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface NotificationStatus {
  totalCount: number;
  unreadCount: number;
}

export interface TenantLevel {
  id: string;
  name: string;
  maxEnvironments: number;
  maxFeatureFlags: number;
  maxTenantAdmins: number;
  maxTenantViewers: number;
}



export interface ListParams {
  page?: number;
  pageSize?: number;
  keyword?: string;
  is_active?: boolean;
}

export interface CreateEnvironmentRequest {
  name: string;
  description?: string;
}

export interface UpdateEnvironmentRequest {
  name: string;
  description?: string;
  isActive?: boolean;
}

export interface CreateFeatureFlagRequest {
  name: string;
  enabled?: boolean;
  description?: string;
  flagType?: string;
  whitelistRules?: WhitelistRule[];
}

export interface UpdateFeatureFlagRequest {
  enabled?: boolean;
  description?: string;
  whitelistRules?: WhitelistRule[];
}

export interface CreateUserRequest {
  email: string;
  displayName?: string;
  role: string;
  initialPassword: string;
}

export interface UpdateUserRequest {
  email: string;
  displayName?: string;
  role: string;
}

export interface CreateTenantRequest {
  name: string;
  levelId?: string;
  description?: string;
  website?: string;
  status?: string;
  adminEmail: string;
  adminPassword: string;
  adminRole: string;
}

export interface CreateSystemNotificationRequest {
  level: string;
  type: string;
  title: string;
  content: string;
}

export interface Segment {
  id: string;
  name: string;
  description?: string;
  matchType?: string;
  conditions?: SegmentCondition[];
  createdAt: string;
  updatedAt?: string;
}

export interface SegmentCondition {
  id?: string;
  attributeKey: string;
  operator: string;
  value: string;
}