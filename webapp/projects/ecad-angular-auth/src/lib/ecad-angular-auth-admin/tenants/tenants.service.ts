import { Injectable, Inject } from '@angular/core';
import { Observable } from 'rxjs';
import { Tenant } from '../interfaces/tenant.i';
import { AUTH_ADMIN_CONFIG } from '../tokens';
import { FilterableService } from '../../filterable-datasource/filtered-datasource';
import {
  FilterCondition,
  PagedResult,
  ResourcesService
} from '../../resource-util/resources.service';
import { AuthAdminConfig } from '../interfaces/auth-admin-config.i';
import { HttpClient } from '@angular/common/http';

@Injectable({
  providedIn: 'root'
})
export class TenantsService implements FilterableService<Tenant> {
  constructor(
    @Inject(AUTH_ADMIN_CONFIG)
    private authAdminConfigVal: AuthAdminConfig,
    private resourcesService: ResourcesService<Tenant, any>,
    private http: HttpClient
  ) {}

  private get apiEndpoint() {
    return `${this.authAdminConfigVal.apiEndpoint}/tenants`;
  }

  fetch(
    filter: FilterCondition<Tenant>[],
    sortBy: keyof Tenant = 'name',
    orderBy: 'asc' | 'desc'
  ): Observable<any> {
    return this.resourcesService.fetch(
      this.apiEndpoint,
      filter,
      sortBy,
      orderBy
    );
  }
  fetchNextPage(pagedResult: PagedResult<Tenant>): Observable<any> {
    return this.resourcesService.fetchNextPage(pagedResult);
  }
  fetchPreviousPage(pagedResult: PagedResult<Tenant>): Observable<any> {
    return this.resourcesService.fetchPreviousPage(pagedResult);
  }

  findMembers(id: string) {
    return this.http.get(`${this.apiEndpoint}/${id}/members/`);
  }

  find(id: string) {
    return this.http.get(`${this.apiEndpoint}/${id}`);
  }
}
