import { Component, OnInit, ViewChild } from '@angular/core';
import { FilteredDatasource } from '../filteredDatasource';
import { User, UsersService } from '../../ecad-angular-auth-admin/users/users.service';
import { Subject } from 'rxjs';
import { MatSort, MatDialog } from '@angular/material';
import { UserEditFormComponent } from '../user-edit-form/user-edit-form.component';

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
    'added',
    'modified',
    'email_verified',
    'roles'
  ];

  constructor(
    private userService: UsersService,
    private dialog: MatDialog
  ) { }

  getRoles(user: User) {
    return Object.keys((user.roles || {} as Object));
  }

  changePage($event) {
    this.dataSource.pageInfo$.subscribe(({currentPage}) => {
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

  updateUser(user: User) {
    this.dialog.open(UserEditFormComponent, {data: user});
  }

  addUser() {
    this.dialog.open(UserEditFormComponent);
  }
}
