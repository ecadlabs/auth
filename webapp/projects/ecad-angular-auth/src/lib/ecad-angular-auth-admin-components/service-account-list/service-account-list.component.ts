import {
  Component,
  EventEmitter,
  Inject,
  OnInit,
  Output,
  ViewChild
} from '@angular/core';
import { MatDialog, MatSort } from '@angular/material';
import { of, Subject } from 'rxjs';
import { first, map, pluck, startWith } from 'rxjs/operators';
import { ConfirmDialogService } from '../../confirm-dialog/confirm-dialog.service';
import { IUsersService } from '../../ecad-angular-auth-admin/interfaces/user-service.i';
import { User } from '../../ecad-angular-auth-admin/interfaces/user.i';
import { USERS_SERVICE } from '../../ecad-angular-auth-admin/tokens';
import { FilteredDatasource } from '../../filterable-datasource/filtered-datasource';
import { FilterCondition } from '../../resource-util/resources.service';

@Component({
  selector: 'auth-service-account-list',
  templateUrl: './service-account-list.component.html',
  styleUrls: ['./service-account-list.component.scss']
})
export class ServiceAccountListComponent implements OnInit {
  @Output()
  userClicked: EventEmitter<User> = new EventEmitter();

  @ViewChild(MatSort) sort: MatSort;
  public dataSource: FilteredDatasource<User>;
  private nextPage$ = new Subject<void>();
  private prevousPage$ = new Subject<void>();

  displayedColumns = ['id', 'added', 'modified', 'actions'];

  public pageIndex$;

  constructor(
    @Inject(USERS_SERVICE)
    private userService: IUsersService,
    private dialog: MatDialog,
    private confirmDialog: ConfirmDialogService
  ) {}

  changePage($event) {
    this.dataSource.pageInfo$.pipe(first()).subscribe(({ currentPage }) => {
      if (currentPage > $event.pageIndex) {
        this.prevousPage$.next();
      } else {
        this.nextPage$.next();
      }
    });
  }

  ngOnInit() {
    this.dataSource = new FilteredDatasource<User>(
      this.userService,
      this.sort.sortChange,
      this.nextPage$,
      this.prevousPage$
    );

    this.pageIndex$ = this.dataSource.pageInfo$.pipe(
      startWith(1),
      pluck('currentPage'),
      map(x => Number(x) - 1)
    );

    this.dataSource.addFilterConditionObservable(
      of([
        {
          operation: 'eq',
          field: 'account_type',
          value: 'service'
        }
      ] as FilterCondition<User>[])
    );
  }

  selectUser(user: User) {
    this.userClicked.next(user);
  }

  delete($event: Event, user: User) {
    $event.stopPropagation();
    this.confirmDialog
      .confirm(
        'This will delete the user permanently. Do you wish to continue?'
      )
      .subscribe(confirmed => {
        if (confirmed) {
          this.userService.delete(user.id).subscribe(() => {
            this.dataSource.refresh();
          });
        }
      });
  }
}
