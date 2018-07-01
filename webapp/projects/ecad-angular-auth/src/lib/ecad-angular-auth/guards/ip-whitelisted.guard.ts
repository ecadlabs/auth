
import { map, switchMap } from 'rxjs/operators';
import { Injectable, Inject, Optional } from '@angular/core';
import { ActivatedRouteSnapshot, RouterStateSnapshot } from '@angular/router';
import { Observable, of } from 'rxjs';
import { Router } from '@angular/router';
import { LoggedinGuard } from './loggedin.guard';
import { LOGIN_SERVICE, AUTH_CONFIG } from '../tokens';
import { ILoginService } from '../interfaces/login-service.i';
import { AuthConfig } from '../interfaces/auth-config.i';


@Injectable()
export class IpWhiteListedGuard extends LoggedinGuard {

  constructor(
    @Inject(LOGIN_SERVICE)
    protected loginService: ILoginService,
    protected router: Router,
    @Optional()
    @Inject(AUTH_CONFIG)
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
