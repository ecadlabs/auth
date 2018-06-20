import { Injectable, Inject, Optional } from '@angular/core';
import { CanActivate, ActivatedRouteSnapshot, RouterStateSnapshot } from '@angular/router';
import { Observable } from 'rxjs';
import { Router } from '@angular/router';
import { ILoginService, AuthConfig } from '../interfaces';
import { LoginService, authConfig } from '../tokens';

@Injectable()
export class LoggedinGuard implements CanActivate {

  constructor(
    @Inject(LoginService) protected loginService: ILoginService,
    @Optional()
    @Inject(authConfig)
    protected config: AuthConfig,
    protected router: Router
  ) {}

  canActivate(
    next: ActivatedRouteSnapshot,
    state: RouterStateSnapshot,
    redirect = true): Observable<boolean> | Promise<boolean> | boolean {
      if (!this.loginService.isLoggedIn()) {
        if (redirect) {
          this.redirectOnNotAuthorized();
        }
        return false;
      } else {
        return true;
      }
  }

  protected redirectOnNotAuthorized() {
    return this.router.navigateByUrl(this.config.loginPageUrl);
  }
}
