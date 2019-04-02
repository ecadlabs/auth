import {
  Component,
  EventEmitter,
  Inject,
  Input,
  OnInit,
  Output
} from '@angular/core';
import { Observable } from 'rxjs';
import { switchMap } from 'rxjs/operators';
import { UserLogEntry } from '../../ecad-angular-auth-admin/interfaces/user-log-entry.i';
import { IUsersService } from '../../ecad-angular-auth-admin/interfaces/user-service.i';
import { User } from '../../ecad-angular-auth-admin/interfaces/user.i';
import { USERS_SERVICE } from '../../ecad-angular-auth-admin/tokens';

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
