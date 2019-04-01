interface CreateRegularUser {
  email: string;
  password: string;
  name?: string;
  roles?: {};
  type?: 'regular';
}

interface CreateServiceUser {
  name?: string;
  roles?: {};
  type: 'service';
  address_whitelist?: string[];
}

export type CreateUser = CreateRegularUser | CreateServiceUser;
