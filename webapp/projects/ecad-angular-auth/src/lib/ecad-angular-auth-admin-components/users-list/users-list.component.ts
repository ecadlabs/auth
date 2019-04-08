import {
  Component,
  OnInit,
  ViewChild,
  Inject,
  EventEmitter,
  Output
} from '@angular/core';
import { FilteredDatasource } from '../../filterable-datasource/filtered-datasource';
import { Subject, of } from 'rxjs';
import { UserEditFormComponent } from '../user-edit-form/user-edit-form.component';
import { IPasswordReset } from '../../ecad-angular-auth/interfaces/password-reset.i';
import { PASSWORD_RESET } from '../../ecad-angular-auth/tokens';
import { User } from '../../ecad-angular-auth-admin/interfaces/user.i';
import { MatSort, MatDialog, MatSnackBar } from '@angular/material';
import { IUsersService } from '../../ecad-angular-auth-admin/interfaces/user-service.i';
import { USERS_SERVICE } from '../../ecad-angular-auth-admin/tokens';
import { first } from 'rxjs/operators';
import { ConfirmDialogService } from '../../confirm-dialog/confirm-dialog.service';
import { FilterCondition } from '../../resource-util/resources.service';

@Component({
  selector: 'auth-users-list',
  templateUrl: './users-list.component.html',
  styleUrls: ['./users-list.component.scss']
})
export class UsersListComponent implements OnInit {
  @Output()
  userClicked: EventEmitter<User> = new EventEmitter();

  @ViewChild(MatSort) sort: MatSort;
  public dataSource: FilteredDatasource<User>;
  private nextPage$ = new Subject<void>();
  private prevousPage$ = new Subject<void>();

  displayedColumns = [
    'email',
    'name',
    'added',
    'modified',
    'email_verified',
    'actions'
  ];

  constructor(
    @Inject(USERS_SERVICE)
    private userService: IUsersService,
    private dialog: MatDialog,
    @Inject(PASSWORD_RESET)
    private passwordReset: IPasswordReset,
    private snackBar: MatSnackBar,
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

    this.dataSource.addFilterConditionObservable(
      of([
        {
          operation: 'eq',
          field: 'account_type',
          value: 'regular'
        }
      ] as FilterCondition<User>[])
    );
  }

  selectUser(user: User) {
    this.userClicked.next(user);
  }

  async resetPassword($event: Event, user: User) {
    $event.stopPropagation();
    this.confirmDialog
      .confirm(
        'You are about to reset this user password. Do you wish to continue?'
      )
      .subscribe(async confirmed => {
        if (confirmed) {
          await this.passwordReset.sendResetEmail(user.email).toPromise();
          this.snackBar.open('Reset password email sent', undefined, {
            duration: 2000,
            horizontalPosition: 'end'
          });
        }
      });
  }

  updateUser($event: Event, user: User) {
    $event.stopPropagation();
    this.dialog
      .open(UserEditFormComponent, { data: user, width: '500px' })
      .afterClosed()
      .subscribe(() => {
        this.dataSource.refresh();
      });
  }

  addUser() {
    this.dialog
      .open(UserEditFormComponent, { width: '500px' })
      .afterClosed()
      .subscribe(() => {
        this.dataSource.refresh();
      });
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
