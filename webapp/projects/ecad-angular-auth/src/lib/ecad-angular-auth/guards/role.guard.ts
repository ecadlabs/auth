import { Injectable, Inject, Optional } from '@angular/core';
import { CanActivate, ActivatedRouteSnapshot, RouterStateSnapshot } from '@angular/router';
import { Observable } from 'rxjs';
import { Router } from '@angular/router';
import { ILoginService, AuthConfig } from '../interfaces';
import { LOGIN_SERVICE, AUTH_CONFIG } from '../tokens';
import { map, tap } from 'rxjs/operators';

@Injectable()
export class RoleGuard implements CanActivate {

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
      const roles = (next.data && next.data.roles && Array.isArray(next.data.roles)) ? next.data.roles : [];
      return this.loginService.user
      .pipe(map((user) => {
        if (!user) {
          return false;
        } else {
          let userRoles = user['http://localhost:8000/roles'];
          userRoles = userRoles && Array.isArray(userRoles) ? userRoles : [];
          return roles.every((role) => userRoles.includes(role));
        }
      }), tap((result) => {
        if (!result) {
          this.redirectOnNotAuthorized();
        }
      }));
    }

  protected redirectOnNotAuthorized() {
    return this.router.navigateByUrl(this.config.loginPageUrl);
  }
}
