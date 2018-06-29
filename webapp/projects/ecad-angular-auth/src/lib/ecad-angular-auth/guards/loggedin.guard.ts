import { Injectable, Inject, Optional } from '@angular/core';
import { CanActivate, ActivatedRouteSnapshot, RouterStateSnapshot } from '@angular/router';
import { Observable } from 'rxjs';
import { Router } from '@angular/router';
import { ILoginService, AuthConfig } from '../interfaces';
import { LOGIN_SERVICE, AUTH_CONFIG } from '../tokens';
import { map } from 'rxjs/operators';

@Injectable()
export class LoggedinGuard implements CanActivate {

  constructor(
    @Inject(LOGIN_SERVICE) protected loginService: ILoginService,
    @Optional()
    @Inject(AUTH_CONFIG)
    protected config: AuthConfig,
    protected router: Router
  ) { }

  canActivate(
    next: ActivatedRouteSnapshot,
    state: RouterStateSnapshot,
    redirect = true): Observable<boolean> {
    return this.loginService.isLoggedIn.pipe(map((isLoggedIn) => {
      if (!isLoggedIn) {
        if (redirect) {
          this.redirectOnNotAuthorized();
        }
        return false;
      } else {
        return true;
      }
    }));
  }

  protected redirectOnNotAuthorized() {
    return this.router.navigateByUrl(this.config.loginPageUrl);
  }
}
