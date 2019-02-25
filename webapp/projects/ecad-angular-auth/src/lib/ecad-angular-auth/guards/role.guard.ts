import { Injectable, Inject, Optional } from '@angular/core';
import { CanActivate, ActivatedRouteSnapshot, RouterStateSnapshot } from '@angular/router';
import { Observable } from 'rxjs';
import { Router } from '@angular/router';
import { ILoginService } from '../interfaces/login-service.i';
import { AuthConfig } from '../interfaces/auth-config.i';
import { LOGIN_SERVICE, AUTH_CONFIG } from '../tokens';
import { tap } from 'rxjs/operators';

@Injectable()
export class RoleGuard implements CanActivate {

  constructor(
    @Inject(LOGIN_SERVICE) protected loginService: ILoginService,
    @Optional()
    @Inject(AUTH_CONFIG)
    protected config: AuthConfig,
    protected router: Router
  ) {}

  canActivate(
    next: ActivatedRouteSnapshot,
    state: RouterStateSnapshot): Observable<boolean> {
    const allowedRoles = (next.data && next.data.roles && Array.isArray(next.data.roles)) ? next.data.roles : [];
    const redirectToUrl = next.data && next.data.redirectUrl || this.config.roleGuardRedirectUrl;

    return this.loginService.hasOneOfRoles(allowedRoles).pipe(
      tap(hasRole => {
        if (!hasRole && !!redirectToUrl) {
          return this.router.navigateByUrl(redirectToUrl);
        }
      })
    );
  }
}
