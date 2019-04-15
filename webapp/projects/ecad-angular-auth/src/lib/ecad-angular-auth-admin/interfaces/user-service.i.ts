import {
  PagedResult,
  FilterCondition
} from '../../resource-util/resources.service';
import { User } from '../interfaces/user.i';
import { CreateUser } from '../interfaces/create-user.i';
import { UpdateUser } from '../interfaces/update-user.i';
import { Observable } from 'rxjs';
import { UpdateMembership } from './update-membership.i';
import { Membership } from './membership.i';

export interface IUsersService {
  getRoles(): {
    value: string;
    displayValue: string;
  }[];
  create(payload: CreateUser): Observable<User>;
  updateEmail(id: string, email: string): Observable<void>;

  update(payload: UpdateUser): Observable<User>;
  updateMembership(payload: UpdateMembership): Observable<Membership>;

  delete(id: string): Observable<boolean>;
  archiveMembership(userId: string, tenantId: string): Observable<{}>;

  fetch(
    filter?: FilterCondition<User>[],
    sortBy?: keyof User,
    orderBy?: 'asc' | 'desc'
  ): Observable<PagedResult<User>>;

  findByMembership(id: string): Observable<User>;
  find(id: string, useCache?: boolean): Observable<User>;
  fetchNextPage(result: PagedResult<User>): Observable<PagedResult<User>>;
  fetchPreviousPage(result: PagedResult<User>): Observable<PagedResult<User>>;
}
