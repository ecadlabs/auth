import { Injectable, Inject } from '@angular/core';
import {
  HttpRequest,
  HttpHandler,
  HttpEvent,
  HttpInterceptor
} from '@angular/common/http';
import { JwtHelperService } from './jwt-helper.service';
import { mergeMap } from 'rxjs/operators';
import { AUTH_CONFIG } from './tokens';
import { Observable, from } from 'rxjs';
import { AuthConfig } from './interfaces/auth-config.i';

@Injectable()
export class JwtInterceptor implements HttpInterceptor {
  tokenGetter: () => string | null | Promise<string | null>;
  headerName: string;
  authScheme: string;
  whitelistedDomains: Array<string | RegExp>;
  blacklistedRoutes: Array<string | RegExp>;
  throwNoTokenError: boolean;
  skipWhenExpired: boolean;

  constructor(
    @Inject(AUTH_CONFIG) config: AuthConfig,
    public jwtHelper: JwtHelperService
  ) {
    this.tokenGetter = config.tokenGetter;
    this.headerName = 'Authorization';
    this.authScheme = 'Bearer ';
    this.skipWhenExpired = true;
    this.blacklistedRoutes = [config.loginUrl, config.passwordResetUrl];
  }

  isHostMatching(request: HttpRequest<any>) {
    return (
      request.url
        .replace(`${location.protocol}//${location.host}`, '')
        .indexOf('/') === 0
    );
  }

  isBlacklistedRoute(request: HttpRequest<any>): boolean {
    const url = request.url;

    return (
      this.blacklistedRoutes.findIndex(route =>
        typeof route === 'string'
          ? route === url
          : route instanceof RegExp
          ? route.test(url)
          : false
      ) > -1
    );
  }

  handleInterception(
    token: string | null,
    request: HttpRequest<any>,
    next: HttpHandler
  ) {
    let tokenIsExpired = false;

    if (!token && this.throwNoTokenError) {
      throw new Error('Could not get token from tokenGetter function.');
    }

    if (this.skipWhenExpired) {
      tokenIsExpired = token ? this.jwtHelper.isTokenExpired(token) : true;
    }

    if (token && tokenIsExpired && this.skipWhenExpired) {
      request = request.clone();
    } else if (
      token &&
      !this.isBlacklistedRoute(request) &&
      this.isHostMatching(request)
    ) {
      request = request.clone({
        setHeaders: {
          [this.headerName]: `${this.authScheme}${token}`
        }
      });
    }
    return next.handle(request);
  }

  intercept(
    request: HttpRequest<any>,
    next: HttpHandler
  ): Observable<HttpEvent<any>> {
    const token = this.tokenGetter();
    if (token instanceof Promise) {
      return from(token).pipe(
        mergeMap((asyncToken: string | null) => {
          return this.handleInterception(asyncToken, request, next);
        })
      );
    } else {
      return this.handleInterception(token, request, next);
    }
  }
}
