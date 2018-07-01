import { Component, OnInit, ViewChild, Inject } from '@angular/core';
import { FilteredDatasource } from '../../filterable-datasource/filtered-datasource';
import { Subject } from 'rxjs';
import { UserEditFormComponent } from '../user-edit-form/user-edit-form.component';
import { IPasswordReset } from '../../ecad-angular-auth/interfaces/password-reset.i';
import { PASSWORD_RESET } from '../../ecad-angular-auth/tokens';
import { User } from '../../ecad-angular-auth-admin/interfaces/user.i';
import { MatSort, MatDialog, MatSnackBar } from '@angular/material';
import { IUsersService } from '../../ecad-angular-auth-admin/interfaces/user-service.i';
import { USERS_SERVICE } from '../../ecad-angular-auth-admin/tokens';

@Component({
  selector: 'auth-users-list',
  templateUrl: './users-list.component.html',
  styleUrls: ['./users-list.component.scss']
})
export class UsersListComponent implements OnInit {

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
    'roles',
    'edit',
    'delete',
    'reset-password'
  ];

  constructor(
    @Inject(USERS_SERVICE)
    private userService: IUsersService,
    private dialog: MatDialog,
    @Inject(PASSWORD_RESET)
    private passwordReset: IPasswordReset,
    private snackBar: MatSnackBar
  ) { }

  getRoles(user: User) {
    return Object.keys((user.roles || {} as Object));
  }

  changePage($event) {
    this.dataSource.pageInfo$.subscribe(({ currentPage }) => {
      if (currentPage > $event.value) {
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
  }

  getDisplayRoles(user: User) {
    return this.userService.getRoles()
      .filter(({ value }) => Object.keys(user.roles).includes(value))
      .map(({ displayValue }) => displayValue);
  }
  async resetPassword(user: User) {
    await this.passwordReset.sendResetEmail(user.email).toPromise();
    this.snackBar.open('Reset password email sent', undefined, { duration: 2000, horizontalPosition: 'end' });
  }

  updateUser(user: User) {
    this.dialog.open(UserEditFormComponent, { data: user, width: '500px' })
      .afterClosed()
      .subscribe(() => {
        this.dataSource.refresh();
      });
  }

  addUser() {
    this.dialog.open(UserEditFormComponent, { width: '500px' })
      .afterClosed()
      .subscribe(() => {
        this.dataSource.refresh();
      });
  }

  delete(user: User) {
    this.userService.delete(user.id)
      .subscribe(() => {
        this.dataSource.refresh();
      });
  }
}
