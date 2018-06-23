import { Injectable, Inject } from '@angular/core';
import { ResourcesService, PagedResult, FilterCondition } from './resources.service';
import { authAdminConfig, AuthAdminConfig } from '../ecad-angular-auth-admin.module';

export interface CreateUser {
  email: string;
  password: string;
  name?: string;
  roles?: {};
}

export interface UpdateUser {
  id: string;
  email?: string;
  name?: string;
  roles?: {};
}

export interface User {
  id: string;
  email: string;
  name: string;
  added: string;
  modified: string;
  email_verified: boolean;
  roles: string[];
}

@Injectable({
  providedIn: 'root'
})
export class UsersService {

  constructor(
    private resourcesService: ResourcesService<User, CreateUser>,
    @Inject(authAdminConfig)
    private authAdminConfigVal: AuthAdminConfig
  ) { }

  private get apiEndpoint() {
    return this.authAdminConfigVal.apiEndpoint;
  }

  public getRoles() {
    return this.authAdminConfigVal.roles;
  }

  create(payload: CreateUser) {
    return this.resourcesService.create(this.apiEndpoint + '/', payload);
  }

  update(payload: UpdateUser, addedRoles: string[] = [], deletedRoles: string[] = []) {
    const allowedKeyForReplace: (keyof User)[] = ['email', 'name'];
    let operations = (Object.keys(payload) as (keyof User)[])
    .filter(key => allowedKeyForReplace.includes(key))
    .reduce((prev, key) => {
      return [...prev, {
        op: 'replace',
        path: `/${key}`,
        value: payload[key]
      }];
    }, []);
    operations = addedRoles.reduce((prev, role) => {
      return [...prev, {
        op: 'add',
        path: `/roles/${role}`,
      }];
    }, operations);
    operations = deletedRoles.reduce((prev, role) => {
      return [...prev, {
        op: 'remove',
        path: `/roles/${role}`,
      }];
    }, operations);
    return this.resourcesService.patch(this.apiEndpoint, payload.id, operations);
  }

  delete(id: string) {
    return this.resourcesService.delete(this.apiEndpoint, id);
  }

  fetch(filter: FilterCondition<User>[] = [], sortBy: keyof User = 'added', orderBy: 'asc' | 'desc' = 'desc') {
    return this.resourcesService.fetch(this.apiEndpoint, filter, sortBy, orderBy);
  }

  find(id: string) {
    return this.resourcesService.find(this.apiEndpoint, id);
  }

  fetchNextPage(result: PagedResult<User>)  {
    return this.resourcesService.fetchNextPage(result);
  }

  fetchPreviousPage(result: PagedResult<User>) {
    return this.resourcesService.fetchPreviousPage(result);
  }
}
