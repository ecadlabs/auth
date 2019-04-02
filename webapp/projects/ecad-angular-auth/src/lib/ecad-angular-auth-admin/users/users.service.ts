import { Injectable, Inject } from '@angular/core';
import {
  ResourcesService,
  PagedResult,
  FilterCondition
} from '../../resource-util/resources.service';
import { AuthAdminConfig } from '../interfaces/auth-admin-config.i';
import { CreateUser } from '../interfaces/create-user.i';
import { UpdateUser } from '../interfaces/update-user.i';
import { User } from '../interfaces/user.i';
import { AUTH_ADMIN_CONFIG } from '../tokens';
import { Observable } from 'rxjs';
import { IUsersService } from '../interfaces/user-service.i';
import { HttpClient } from '@angular/common/http';
import {
  getPatchOpsFromObj,
  getPatchAddRemoveOpsFromObj,
  SubType
} from '../utils';
import { Membership } from '../interfaces/membership.i';
import { UpdateMembership } from '../interfaces/update-membership.i';

type UpdateAddRemoveUser = SubType<UpdateUser, { [key: string]: boolean }>;
type UpdateAddRemoveMember = SubType<
  UpdateMembership,
  { [key: string]: boolean }
>;
@Injectable({
  providedIn: 'root'
})
export class UsersService implements IUsersService {
  constructor(
    private resourcesService: ResourcesService<User, CreateUser>,
    @Inject(AUTH_ADMIN_CONFIG)
    private authAdminConfigVal: AuthAdminConfig,
    private httpClient: HttpClient
  ) {}

  private get apiEndpoint() {
    return `${this.authAdminConfigVal.apiEndpoint}/users`;
  }

  private get apiEndpointMembers() {
    return `${this.authAdminConfigVal.apiEndpoint}/members`;
  }

  private get apiEndpointTenants() {
    return `${this.authAdminConfigVal.apiEndpoint}/tenants`;
  }

  public getRoles() {
    return this.authAdminConfigVal.roles;
  }

  create(payload: CreateUser): Observable<User> {
    return this.resourcesService.create(this.apiEndpoint + '/', payload);
  }

  updateEmail(id: string, email: string) {
    return this.httpClient.post<void>(this.authAdminConfigVal.emailUpdateUrl, {
      id,
      email
    });
  }

  update(payload: UpdateUser): Observable<User> {
    const allowedKeyForReplace: (keyof UpdateUser)[] = ['name'];
    const allowedKeyForAddRemove: (keyof UpdateAddRemoveUser)[] = [
      'address_whitelist'
    ];
    const operations = getPatchOpsFromObj(allowedKeyForReplace, payload);
    const addRemoveoperations = getPatchAddRemoveOpsFromObj<
      UpdateAddRemoveUser
    >(allowedKeyForAddRemove, payload);
    return this.resourcesService.patch(this.apiEndpoint, payload.id, [
      ...addRemoveoperations,
      ...operations
    ]);
  }

  updateMembership(payload: UpdateMembership): Observable<Membership> {
    const allowedKeyForAddRemove: (keyof UpdateAddRemoveMember)[] = ['roles'];
    const addRemoveoperations = getPatchAddRemoveOpsFromObj<
      UpdateAddRemoveMember
    >(allowedKeyForAddRemove, payload);
    return this.httpClient.patch<Membership>(
      `${this.apiEndpointTenants}/${payload.tenantId}/members/${
        payload.userId
      }`,
      addRemoveoperations
    );
  }

  archiveMembership(userId: string, tenantId: string): Observable<{}> {
    return this.httpClient.delete<{}>(
      `${this.apiEndpointTenants}/${tenantId}/members/${userId}`
    );
  }

  delete(id: string): Observable<boolean> {
    return this.resourcesService.delete(this.apiEndpoint, id);
  }

  fetch(
    filter: FilterCondition<User>[] = [],
    sortBy: keyof User = 'added',
    orderBy: 'asc' | 'desc' = 'desc'
  ): Observable<PagedResult<User>> {
    return this.resourcesService.fetch(
      this.apiEndpoint,
      filter,
      sortBy,
      orderBy
    );
  }

  find(id: string): Observable<User> {
    return this.resourcesService.find(this.apiEndpoint, id);
  }

  findByMembership(id: string): Observable<User> {
    return this.resourcesService.fetchAndCache(
      `${this.apiEndpointMembers}/${id}/user`
    );
  }

  fetchNextPage(result: PagedResult<User>): Observable<PagedResult<User>> {
    return this.resourcesService.fetchNextPage(result);
  }

  fetchPreviousPage(result: PagedResult<User>): Observable<PagedResult<User>> {
    return this.resourcesService.fetchPreviousPage(result);
  }
}
