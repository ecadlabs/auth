export interface UserToken {
  exp: number;
  email: string;
  name: string;
  roles: any[];
  permissions: any[];
  iat: number;
  iss: string;
  sub: string;
}
