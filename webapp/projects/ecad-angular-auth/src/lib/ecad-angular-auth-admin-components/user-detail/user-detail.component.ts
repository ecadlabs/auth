import {
  Component,
  EventEmitter,
  Inject,
  Input,
  OnInit,
  Output
} from '@angular/core';
import { Observable, BehaviorSubject, of, timer } from 'rxjs';
import { switchMap, map, first, tap, catchError } from 'rxjs/operators';
import { IUsersService } from '../../ecad-angular-auth-admin/interfaces/user-service.i';
import { User } from '../../ecad-angular-auth-admin/interfaces/user.i';
import { USERS_SERVICE } from '../../ecad-angular-auth-admin/tokens';
import { HttpErrorResponse } from '@angular/common/http';

enum IPError {
  NO_ERROR,
  CONFLICT_ERROR,
  UNKOWN_ERROR
}

@Component({
  selector: 'auth-user-detail',
  templateUrl: './user-detail.component.html',
  styleUrls: ['./user-detail.component.scss']
})
export class UserDetailComponent implements OnInit {
  public expandedElement;
  public expandedTarget;
  public expandedUser;

  public readonly IPError = IPError;

  public userId$ = new BehaviorSubject('');
  private _refresh$ = new BehaviorSubject(true);
  public ipLoading$ = new BehaviorSubject(false);
  public ipErrors$ = new BehaviorSubject(IPError.NO_ERROR);

  @Input()
  get userId() {
    return this.userId$.value;
  }

  set userId(id: string) {
    this.userId$.next(id);
  }

  @Output()
  userClicked: EventEmitter<User> = new EventEmitter();

  public user$: Observable<User> = this.userId$.pipe(
    switchMap(userId => {
      return this._refresh$.pipe(
        switchMap(() => {
          return this.userService.find(userId, false);
        })
      );
    })
  );

  public ips$: Observable<string[]> = this.user$.pipe(
    map(user => {
      return user.address_whitelist ? Object.keys(user.address_whitelist) : [];
    }),
    tap(() => {
      this.ipLoading$.next(false);
    })
  );

  public isServiceAccount$ = this.user$.pipe(
    map(user => {
      return user.account_type === 'service';
    })
  );

  constructor(
    @Inject(USERS_SERVICE)
    private userService: IUsersService
  ) {}

  selectUser(user: User) {
    this.userClicked.next(user);
  }

  newIp(ip: string) {
    this.userId$
      .pipe(
        first(),
        switchMap(id => {
          return this.userService.update({
            id,
            address_whitelist: { [ip]: true }
          });
        }),
        tap(() => this.refresh()),
        catchError(err => {
          return this.handleIpError(err);
        })
      )
      .subscribe();
  }

  removeIp(ip: string) {
    this.userId$
      .pipe(
        first(),
        switchMap(id => {
          return this.userService.update({
            id,
            address_whitelist: { [ip]: false }
          });
        }),
        tap(() => this.refresh()),
        catchError(err => {
          return this.handleIpError(err);
        })
      )
      .subscribe();
  }

  private handleIpError(err: any) {
    if (err instanceof HttpErrorResponse && err.status === 409) {
      this.ipErrors$.next(IPError.CONFLICT_ERROR);
    } else {
      this.ipErrors$.next(IPError.UNKOWN_ERROR);
    }

    return timer(5000).pipe(
      tap(() => {
        this.ipErrors$.next(IPError.NO_ERROR);
      })
    );
  }

  private refresh() {
    this._refresh$.next(true);
    this.ipLoading$.next(true);
  }

  ngOnInit() {}
}
