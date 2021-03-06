import { DataSource } from '@angular/cdk/table';
import {
  of,
  Observable,
  Subject,
  BehaviorSubject,
  combineLatest,
  merge,
  ReplaySubject
} from 'rxjs';
import { CollectionViewer } from '@angular/cdk/collections';
import { Sort } from '@angular/material';
import {
  FilterCondition,
  PagedResult
} from '../resource-util/resources.service';
import {
  startWith,
  switchMap,
  map,
  debounceTime,
  tap,
  catchError
} from 'rxjs/operators';

export interface FilterableService<T> {
  fetch(
    filter: FilterCondition<T>[],
    sortBy: keyof T,
    orderBy: 'asc' | 'desc'
  ): Observable<PagedResult<T>>;
  fetchNextPage(pagedResult: PagedResult<T>): Observable<PagedResult<T>>;
  fetchPreviousPage(pagedResult: PagedResult<T>): Observable<PagedResult<T>>;
}

export class FilteredDatasource<T> extends DataSource<T> {
  private refreshSubject = new BehaviorSubject({});
  private newFilters = new BehaviorSubject(of([]));

  private filteredConditions$ = [];
  public isLoading$: Subject<boolean> = new Subject<boolean>();
  public pageInfo$: ReplaySubject<{
    currentPage: number;
    totalCount: number;
    totalPage: number;
  }> = new ReplaySubject(1);
  constructor(
    private service: FilterableService<T>,
    private sort$: Observable<Sort>,
    private nextPage: Observable<void>,
    private previousPage: Observable<void>,
    private debounceDuration: number = 250
  ) {
    super();
  }

  addFilterConditionObservable(observable: Observable<FilterCondition<T>[]>) {
    this.filteredConditions$.push(observable);
    this.newFilters.next(this.createFilterObservable());
  }

  private createFilterObservable() {
    const observables: Observable<
      FilterCondition<T>[]
    >[] = this.filteredConditions$.map(
      (obs: Observable<FilterCondition<T>[]>) => {
        return obs.pipe(startWith([] as FilterCondition<T>[]));
      }
    );
    return combineLatest(...observables).pipe(
      map(filterConditions => [].concat(...filterConditions))
    ) as Observable<FilterCondition<T>[]>;
  }

  refresh() {
    this.refreshSubject.next({});
  }

  connect(collectionViewer: CollectionViewer) {
    let cachedPagedResult;
    return this.newFilters.pipe(
      switchMap(filterConditions$ => {
        return this.refreshSubject
          .pipe(
            startWith({}),
            switchMap(() =>
              combineLatest(filterConditions$, this.sort$.pipe(startWith({})))
            )
          )
          .pipe(
            tap(() => this.isLoading$.next(true)),
            debounceTime(this.debounceDuration),
            switchMap((filterConditionsSort: [FilterCondition<T>[], Sort]) => {
              let currentPagedResult;
              return this.service
                .fetch(
                  filterConditionsSort[0],
                  filterConditionsSort[1].active as keyof T,
                  filterConditionsSort[1].direction || 'desc'
                )
                .pipe(
                  switchMap(pagedResult => {
                    return merge(
                      of(pagedResult) as Observable<PagedResult<T>>,
                      this.nextPage.pipe(
                        switchMap(() =>
                          this.service.fetchNextPage(currentPagedResult)
                        )
                      ),
                      this.previousPage.pipe(
                        switchMap(() =>
                          this.service.fetchPreviousPage(currentPagedResult)
                        )
                      )
                    );
                  }),
                  catchError(() => {
                    return of(cachedPagedResult);
                  }),
                  tap(pagedResult => {
                    currentPagedResult = pagedResult;
                    cachedPagedResult = pagedResult;
                    this.pageInfo$.next({
                      currentPage: pagedResult.currentPage,
                      totalCount: pagedResult.total_count,
                      totalPage: Math.ceil(pagedResult.total_count / 20)
                    });
                  }),
                  map(pagedResult => pagedResult.value)
                );
            }),
            tap(() => this.isLoading$.next(false))
          );
      })
    );
  }

  disconnect() {}
}
