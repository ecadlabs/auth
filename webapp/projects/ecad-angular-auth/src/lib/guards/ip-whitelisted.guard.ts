
import {map, tap, switchMap} from 'rxjs/operators';
import { Injectable, Inject, Optional } from '@angular/core';
import { CanActivate, ActivatedRouteSnapshot, RouterStateSnapshot } from '@angular/router';
import { Observable, of } from 'rxjs';
import { Router } from '@angular/router';
import { LoggedinGuard } from './loggedin.guard';
import { LoginService, authConfig } from '../tokens';
import { ILoginService, AuthConfig } from '../interfaces';


@Injectable()
export class IpWhiteListedGuard extends LoggedinGuard {

  constructor(
    @Inject(LoginService)
    protected loginService: ILoginService,
    protected router: Router,
    @Optional()
    @Inject(authConfig)
    protected config: AuthConfig,
  ) {
    super(loginService, config, router);
  }

  canActivate(
    next: ActivatedRouteSnapshot,
    state: RouterStateSnapshot): Observable<boolean> {
      return super.canActivate(next, state, false).pipe(switchMap((isLoggedIn) => {
        if (isLoggedIn as boolean) {
          return of(true);
        } else {
          return this.loginService.isIpWhiteListed.pipe(map((value) => {
            if (!value) {
              this.redirectOnNotAuthorized();
            }
            return !!value;
          }));
        }
      }));
  }
}
