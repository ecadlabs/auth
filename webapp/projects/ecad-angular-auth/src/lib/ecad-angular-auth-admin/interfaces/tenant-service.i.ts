import { Observable } from 'rxjs';
import {
  FilterCondition,
  PagedResult
} from '../../resource-util/resources.service';
import { Tenant } from './tenant.i';
import { Membership } from './membership.i';

export interface ITenantService {
  fetch(
    filter?: FilterCondition<Tenant>[],
    sortBy?: keyof Tenant,
    orderBy?: 'asc' | 'desc'
  ): Observable<PagedResult<Tenant>>;
  fetchNextPage(result: PagedResult<Tenant>): Observable<PagedResult<Tenant>>;
  fetchPreviousPage(
    result: PagedResult<Tenant>
  ): Observable<PagedResult<Tenant>>;

  findMembers(id: string): Observable<PagedResult<Membership[]>>;

  find(id: string): Observable<Tenant>;
}
