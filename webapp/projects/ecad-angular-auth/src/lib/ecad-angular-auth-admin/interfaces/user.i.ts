export interface UserMemberhsip {
  tenant_id: string;
  tenant_type: 'individual' | 'organization';
  tenant_name: string;
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
  account_type: 'service' | 'regular';
  membership: UserMemberhsip[];
  email_verified: boolean;
  address_whitelist: {
    [key: string]: boolean;
  };
}
