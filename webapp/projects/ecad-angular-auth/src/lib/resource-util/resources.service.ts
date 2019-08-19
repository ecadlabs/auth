import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable, of as observableOf } from 'rxjs';
import { catchError, map, shareReplay } from 'rxjs/operators';

export interface FilterCondition<T> {
  operation: FilterOperation;
  field: keyof T;
  value: string;
}

export interface FilterExpression<T> {
  condition: 'and' | 'or';
  filterConditions: FilterCondition<T>[];
}

export type FilterOperation =
  | 'eq'
  | 'ne'
  | 'lt'
  | 'gt'
  | 'le'
  | 'ge'
  | 're'
  | 'l'
  | 'p'
  | 's'
  | 'sub'
  | 'has'
  | '!eq'
  | '!ne'
  | '!lt'
  | '!gt'
  | '!le'
  | '!ge'
  | '!re'
  | '!l'
  | '!p'
  | '!s'
  | '!sub'
  | '!has';

export type PagedResult<T> = any;
export interface PatchPayload<T> {
  op: 'replace' | 'add' | 'remove';
  path: string;
  value: {};
}

@Injectable({
  providedIn: 'root'
})
export class ResourcesService<T, U> {
  private cache = new Map<string, Observable<any>>();

  constructor(private httpClient: HttpClient) {}

  private previousMap = new Map();

  create(resourceUrl: string, payload: U): Observable<T> {
    return this.httpClient.post<T>(`${resourceUrl}`, payload);
  }

  patch(
    resourceUrl: string,
    id: string,
    payload: PatchPayload<T>[]
  ): Observable<T> {
    return this.httpClient.patch<T>(`${resourceUrl}/${id}`, payload);
  }

  delete(resourceUrl: string, id: string): Observable<boolean> {
    return this.httpClient.delete(`${resourceUrl}/${id}`).pipe(map(() => true));
  }

  fetch(
    resourceUrl: string,
    filters: FilterCondition<T>[] = [],
    sortBy: keyof T,
    order: 'asc' | 'desc'
  ): Observable<PagedResult<T>> {
    const query = { and: this.convertExpression(filters) };
    const queryStr = filters.length > 0 ? `&q=${JSON.stringify(query)}` : '';
    return this.httpClient
      .get<PagedResult<T>>(
        `${resourceUrl}/?count=true&sortBy=${sortBy}&order=${order}${queryStr}`
      )
      .pipe(map((page: PagedResult<T>) => ({ ...page, currentPage: 1 })));
  }

  private createEmptyPagedResult(
    current: string,
    result?: PagedResult<T>,
    currentPage = 1
  ) {
    return {
      current,
      currentPage,
      value: [],
      total_count: result ? result.total_count : 0,
      next: ''
    };
  }

  private convertExpression(
    expressions: (FilterExpression<T> | FilterCondition<T>)[]
  ) {
    const query = [] as any[];
    for (const exp of expressions) {
      if ('filterConditions' in exp) {
        query.push({
          [exp.condition]: this.convertExpression(exp.filterConditions)
        });
      } else {
        query.push({
          [exp.operation]: {
            [exp.field]: exp.value
          }
        });
      }
    }
    return query;
  }

  fetchPreviousPage(result: PagedResult<T>): Observable<PagedResult<T>> {
    return observableOf(this.previousMap.get(result.current) || result);
  }

  fetchNextPage(result: PagedResult<T>): Observable<PagedResult<T>> {
    return this.httpClient.get(result.next, { observe: 'response' }).pipe(
      map(res => {
        if (res.status === 204) {
          return this.createEmptyPagedResult(result.next, result);
        } else {
          return res.body as PagedResult<T>;
        }
      }),
      map(newResult => {
        this.previousMap.set(result.next, result);
        return Object.assign(newResult, {
          currentPage: result.currentPage + 1,
          current: result.next,
          previous: result.current,
          total_count: result.total_count
        }) as PagedResult<T>;
      }),
      catchError(() => {
        return observableOf(result);
      })
    );
  }

  private execCache<C>(
    url: string,
    obsFactory: () => Observable<C>,
    bypass: boolean = false,
    cacheDuration = 20000
  ) {
    if (bypass || !this.cache.has(url)) {
      this.cache.set(url, obsFactory().pipe(shareReplay(1, cacheDuration)));
    }
    return this.cache.get(url) as Observable<C>;
  }

  find(resourceUrl: string, id: string, useCache = true): Observable<T> {
    const url = `${resourceUrl}/${id}`;
    return this.fetchAndCache(url, useCache);
  }

  fetchAndCache(url: string, useCache: boolean): Observable<T> {
    return this.execCache(url, () => this.httpClient.get<T>(url), !useCache);
  }
}
