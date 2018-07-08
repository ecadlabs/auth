import { Component, OnInit, Inject, ViewChild, Input, EventEmitter, Output } from '@angular/core';
import { forkJoin, of, Observable, Subject } from 'rxjs';
import { UserLogEntry } from '../../ecad-angular-auth-admin/interfaces/user-log-entry.i';
import { catchError, first, map, switchMap, tap } from 'rxjs/operators';
import { IUsersService } from '../../ecad-angular-auth-admin/interfaces/user-service.i';
import { IUserLogService } from '../../ecad-angular-auth-admin/interfaces/user-log-service.i';
import { USER_LOG_SERVICE, USERS_SERVICE } from '../../ecad-angular-auth-admin/tokens';
import { MatSort } from '@angular/material';
import { User } from '../../ecad-angular-auth-admin/interfaces/user.i';
import { FilteredDatasource } from '../../filterable-datasource/filtered-datasource';
import { FilterCondition } from '../../resource-util/resources.service';
import { style, trigger, state, transition, animate } from '@angular/animations';

@Component({
  selector: 'auth-user-logs',
  templateUrl: './user-logs.component.html',
  styleUrls: ['./user-logs.component.scss'],
  animations: [
    trigger('detailExpand', [
      state('collapsed', style({ height: '0px', minHeight: '0', display: 'none' })),
      state('expanded', style({ height: '*' })),
      transition('expanded <=> collapsed', animate('225ms cubic-bezier(0.4, 0.0, 0.2, 1)')),
    ]),
  ],
})
export class UserLogsComponent implements OnInit {
  public expandedElement;
  public expandedTarget;
  public expandedUser;

  @Input()
  public userId$: Observable<string>;

  @Output()
  userClicked: EventEmitter<User> = new EventEmitter();

  @ViewChild(MatSort) sort: MatSort;
  public dataSource: FilteredDatasource<UserLogEntry>;
  private nextPage$ = new Subject<void>();
  private prevousPage$ = new Subject<void>();

  public displayedColumns = [
    'event',
    'ts',
    'user_id',
    'target_id',
    'addr',
  ];

  constructor(
    @Inject(USERS_SERVICE)
    private userService: IUsersService,
    @Inject(USER_LOG_SERVICE)
    private logService: IUserLogService
  ) { }

  selectUser(user: User) {
    this.userClicked.next(user);
  }

  ngOnInit() {
    if (this.userId$) {
      this.dataSource = new FilteredDatasource(
        this.logService,
        this.sort.sortChange,
        this.nextPage$,
        this.prevousPage$
      );

      const createFilterCondition: (userId: string) => FilterCondition<UserLogEntry> = (uId) => {
        return {
          operation: 'eq',
          field: 'user_id',
          value: uId,
        };
      };

      this.dataSource.addFilterConditionObservable(
        this.userId$.pipe(switchMap((userId) => of([createFilterCondition(userId)]))),
      );
    } else {
      this.dataSource = new FilteredDatasource(
        this.logService,
        this.sort.sortChange,
        this.nextPage$,
        this.prevousPage$
      );
    }
  }

  changePage($event) {
    this.dataSource.pageInfo$
      .pipe(first())
      .subscribe(({ currentPage }) => {
        if (currentPage > $event.pageIndex) {
          this.prevousPage$.next();
        } else {
          this.nextPage$.next();
        }
      });
  }

  expand(log: UserLogEntry) {
    this.expandedElement = log;
    forkJoin(
      this.userService.find(log.target_id).pipe(catchError(() => of(null))),
      this.userService.find(log.user_id).pipe(catchError(() => of(null)))
    ).pipe(
      first(),
      map(([target, user]) => {
        this.expandedTarget = target;
        this.expandedUser = user;
      })).subscribe();
  }
}
