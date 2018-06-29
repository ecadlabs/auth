import { Injectable, Inject, Optional } from '@angular/core';
import { CanActivate, ActivatedRouteSnapshot, RouterStateSnapshot } from '@angular/router';
import { Observable } from 'rxjs';
import { Router } from '@angular/router';
import { ILoginService, AuthConfig } from '../interfaces';
import { LOGIN_SERVICE, AUTH_CONFIG } from '../tokens';
import { map, tap } from 'rxjs/operators';

@Injectable()
export class PermissionsGuard implements CanActivate {

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
        const permissions = (next.data && next.data.permissions && Array.isArray(next.data.permissions)) ? next.data.permissions : [];
        const permissionsObservable = this.loginService.hasPermissions(permissions);
        return permissionsObservable.pipe(tap((result) => {
            if (!result) {
                this.redirectOnNotAuthorized();
            }
        }));
    }

    protected redirectOnNotAuthorized() {
        return this.router.navigateByUrl(this.config.loginPageUrl);
    }
}
