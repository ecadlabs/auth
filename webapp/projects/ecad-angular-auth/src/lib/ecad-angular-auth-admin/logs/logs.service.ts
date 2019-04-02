import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { FilterableService } from '../../filterable-datasource/filtered-datasource';
import {
  FilterCondition,
  PagedResult,
  ResourcesService
} from '../../resource-util/resources.service';
import { UserLogEntry } from '../interfaces/user-log-entry.i';
import { IUserLogService } from '../interfaces/user-log-service.i';

@Injectable({
  providedIn: 'root'
})
export class LogsService
  implements IUserLogService, FilterableService<UserLogEntry> {
  constructor(private resourcesService: ResourcesService<UserLogEntry, void>) {}

  private readonly userLogsUrl = '/api/v1/logs';

  fetch(
    filter: FilterCondition<UserLogEntry>[] = [],
    sortBy: keyof UserLogEntry = 'ts',
    orderBy: 'asc' | 'desc' = 'desc'
  ): Observable<PagedResult<UserLogEntry>> {
    return this.resourcesService.fetch(
      this.userLogsUrl,
      filter,
      sortBy,
      orderBy
    );
  }

  find(id: string): Observable<UserLogEntry> {
    return this.resourcesService.find(`${this.userLogsUrl}`, id);
  }
  fetchNextPage(
    result: PagedResult<UserLogEntry>
  ): Observable<PagedResult<UserLogEntry>> {
    return this.fetchNextPage(result);
  }
  fetchPreviousPage(
    result: PagedResult<UserLogEntry>
  ): Observable<PagedResult<UserLogEntry>> {
    return this.fetchPreviousPage(result);
  }
}
