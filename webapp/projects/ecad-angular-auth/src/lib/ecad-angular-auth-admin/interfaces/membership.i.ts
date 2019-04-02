export interface Membership {
  id: string;
  roles: {};
  tenant_id: string;
  user_id: string;
  type: 'owner' | 'member';
}
