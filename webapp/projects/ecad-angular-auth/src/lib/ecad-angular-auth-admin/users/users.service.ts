import { Injectable, Inject } from '@angular/core';
import { ResourcesService, PagedResult, FilterCondition } from '../../resource-util/resources.service';
import { AuthAdminConfig, CreateUser, User, UpdateUser } from '../interfaces';
import { authAdminConfig } from '../tokens';
import { Observable } from 'rxjs';

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

  create(payload: CreateUser): Observable<User> {
    return this.resourcesService.create(this.apiEndpoint + '/', payload);
  }

  update(payload: UpdateUser, addedRoles: string[] = [], deletedRoles: string[] = []): Observable<User> {
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

  delete(id: string): Observable<boolean> {
    return this.resourcesService.delete(this.apiEndpoint, id);
  }

  fetch(
    filter: FilterCondition<User>[] = [],
    sortBy: keyof User = 'added',
    orderBy: 'asc' | 'desc' = 'desc'
  ): Observable<PagedResult<User>> {
    return this.resourcesService.fetch(this.apiEndpoint, filter, sortBy, orderBy);
  }

  find(id: string): Observable<User> {
    return this.resourcesService.find(this.apiEndpoint, id);
  }

  fetchNextPage(result: PagedResult<User>): Observable<PagedResult<User>> {
    return this.resourcesService.fetchNextPage(result);
  }

  fetchPreviousPage(result: PagedResult<User>): Observable<PagedResult<User>> {
    return this.resourcesService.fetchPreviousPage(result);
  }
}
