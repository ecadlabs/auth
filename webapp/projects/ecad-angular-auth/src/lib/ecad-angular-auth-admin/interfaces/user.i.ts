export interface UserMemberhsip {
  tenant_id: string;
  tenant_type: 'individual' | 'organization';
  type: 'owner' | 'member';
  roles: {
    [key: string]: boolean;
  };
}

export interface User {
  id: string;
  email: string;
  name: string;
  added: string;
  modified: string;
  membership: UserMemberhsip[];
  email_verified: boolean;
  address_whitelist: {
    [key: string]: boolean;
  };
}
