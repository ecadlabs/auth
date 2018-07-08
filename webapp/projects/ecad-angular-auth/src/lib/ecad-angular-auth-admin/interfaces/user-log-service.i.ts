import { FilterCondition, PagedResult } from '../../resource-util/resources.service';
import { UserLogEntry } from './user-log-entry.i';
import { Observable } from 'rxjs';

export interface IUserLogService {
    fetch(
        filter?: FilterCondition<UserLogEntry>[],
        sortBy?: keyof UserLogEntry,
        orderBy?: 'asc' | 'desc'
    ): Observable<PagedResult<UserLogEntry>>;

    find(id: string): Observable<UserLogEntry>;
    fetchNextPage(result: PagedResult<UserLogEntry>): Observable<PagedResult<UserLogEntry>>;
    fetchPreviousPage(result: PagedResult<UserLogEntry>): Observable<PagedResult<UserLogEntry>>;
}
