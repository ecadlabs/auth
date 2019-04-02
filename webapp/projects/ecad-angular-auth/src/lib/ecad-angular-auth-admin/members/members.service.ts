import { Observable } from 'rxjs';
import { FilterableService } from '../../filterable-datasource/filtered-datasource';
import {
  FilterCondition,
  PagedResult,
  ResourcesService
} from '../../resource-util/resources.service';
import { AuthAdminConfig } from '../interfaces/auth-admin-config.i';
import { Membership } from '../interfaces/membership.i';
export class UserMembershipsService implements FilterableService<Membership> {
  constructor(
    private userId: string,
    private resourcesService: ResourcesService<Membership, any>,
    private authAdminConfigVal: AuthAdminConfig
  ) {}

  private get apiEndpoint() {
    return `${this.authAdminConfigVal.apiEndpoint}/users/${
      this.userId
    }/memberships/`;
  }

  fetch(
    filter: FilterCondition<Membership>[] = [],
    sortBy: keyof Membership = 'user_id',
    orderBy: 'asc' | 'desc'
  ): Observable<PagedResult<Membership>> {
    return this.resourcesService.fetch(
      this.apiEndpoint,
      filter || [],
      sortBy,
      orderBy
    );
  }
  fetchNextPage(pagedResult: any): Observable<PagedResult<Membership>> {
    return this.resourcesService.fetchNextPage(pagedResult);
  }
  fetchPreviousPage(pagedResult: any): Observable<PagedResult<Membership>> {
    return this.resourcesService.fetchNextPage(pagedResult);
  }
}
