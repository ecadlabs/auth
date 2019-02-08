import {
  Component,
  OnInit,
  Input,
  Inject,
  ViewChild,
  Output,
  EventEmitter
} from '@angular/core';
import {
  USERS_SERVICE,
  USER_LOG_SERVICE
} from '../../ecad-angular-auth-admin/tokens';
import { IUsersService } from '../../ecad-angular-auth-admin/interfaces/user-service.i';
import { shareReplay, first, map, catchError, switchMap } from 'rxjs/operators';
import { Observable, Subject, forkJoin, of } from 'rxjs';
import { User } from '../../ecad-angular-auth-admin/interfaces/user.i';
import { IUserLogService } from '../../ecad-angular-auth-admin/interfaces/user-log-service.i';
import { UserLogEntry } from '../../ecad-angular-auth-admin/interfaces/user-log-entry.i';
import { FilteredDatasource } from '../../filterable-datasource/filtered-datasource';
import { MatSort } from '@angular/material';
import {
  trigger,
  state,
  transition,
  animate,
  style
} from '@angular/animations';

@Component({
  selector: 'auth-user-detail',
  templateUrl: './user-detail.component.html',
  styleUrls: ['./user-detail.component.scss']
})
export class UserDetailComponent implements OnInit {
  public expandedElement;
  public expandedTarget;
  public expandedUser;

  @Input()
  public userId$: Observable<string>;

  @Output()
  userClicked: EventEmitter<User> = new EventEmitter();

  public user$: Observable<User>;

  public logs$: Observable<UserLogEntry[]>;

  constructor(
    @Inject(USERS_SERVICE)
    private userService: IUsersService
  ) {}

  selectUser(user: User) {
    this.userClicked.next(user);
  }

  ngOnInit() {
    this.user$ = this.userId$.pipe(
      switchMap(userId => this.userService.find(userId))
    );
  }
}
