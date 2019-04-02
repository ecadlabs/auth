export interface UpdateUser {
  id: string;
  email?: string;
  name?: string;
  // Analogous to individual tenant roles
  roles?: { [key: string]: boolean };
  address_whitelist?: { [key: string]: boolean };
}
