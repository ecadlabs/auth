import { Injectable, Optional, Inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { JwtHelperService } from '@auth0/angular-jwt';

import { map, catchError, tap, filter, switchMap, distinctUntilChanged } from 'rxjs/operators';
import { of as observableOf, Observable, Observer, BehaviorSubject, interval } from 'rxjs';
import { AUTH_CONFIG } from '../tokens';
import { ILoginService } from '../interfaces/login-service.i';
import { UserToken } from '../interfaces/user-token.i';
import { AuthConfig } from '../interfaces/auth-config.i';
import { Credentials } from '../interfaces/credentials.i';
import { LoginResult } from '../interfaces/loginResult.i';

@Injectable({
  providedIn: 'root'
})
export class StandardLoginService implements ILoginService {

  private readonly AUTO_REFRESH_INTERVAL = (this.config && this.config.autoRefreshInterval) || 60000;
  public user: BehaviorSubject<UserToken> = new BehaviorSubject(this.token);

  private readonly postLoginOperations = [
    tap((result: { token: string }) => this.config.tokenSetter(result.token)),
    tap((result: { refresh: string }) => localStorage.setItem('refreshTokenUrl', result.refresh)),
    tap(() => this.user.next(this.token))
  ];

  public isLoggedIn: Observable<Boolean> = this.user.pipe(
    map(() => {
      const rawToken = this.config.tokenGetter() || null;
      return !!(rawToken && !this.jwtHelper.isTokenExpired(rawToken));
    }),
  );

  private readonly DEFAULT_PREFIX = 'com.ecadlabs.auth';

  constructor(
    @Optional()
    @Inject(AUTH_CONFIG)
    private config: AuthConfig,
    private httpClient: HttpClient,
    private jwtHelper: JwtHelperService
  ) {
    this.initAutoRefresh();
  }

  private initAutoRefresh() {
    this.isLoggedIn.pipe(
      filter(isLoggedIn => !!isLoggedIn),
      switchMap(() => this.user),
      switchMap((user) => {
        return interval(this.AUTO_REFRESH_INTERVAL)
          .pipe(switchMap(() => {
            return this.refreshToken().pipe(catchError(() => observableOf(false)));
          }));
      }),
      tap(() => this.user.next(this.token))
    )
      .subscribe();
  }

  private createRequestOptions(credential: Credentials) {
    const credentialString = btoa(`${credential.username}:${credential.password}`);
    return {
      headers: {
        'Authorization': `Basic ${credentialString}`,
        ['Content-Type']: 'application/x-www-form-urlencoded',
      }
    };
  }

  private getPrefixed(token: any, propName: string) {
    return token[`${this.config.tokenPropertyPrefix || this.DEFAULT_PREFIX}.${propName}`];
  }

  private get token(): UserToken {
    const token = this.config.tokenGetter();
    if (token) {
      const decodedToken = this.jwtHelper.decodeToken(token);
      const email = this.getPrefixed(decodedToken, 'email');
      const name = this.getPrefixed(decodedToken, 'name');
      const roles = this.getPrefixed(decodedToken, 'roles');
      return { email, name, roles, ...decodedToken };
    } else {
      return null;
    }
  }

  /**
  * Authenticate the user with basic auth using loginUrl provided in authConfig
  */
  public login(credential: Credentials): Observable<LoginResult> {
    const requestOptions = this.createRequestOptions(credential);
    return this.httpClient.get<LoginResult>(this.config.loginUrl, requestOptions).pipe(
      ...this.postLoginOperations
    );
  }

  /*
  * Check if the ip is whitelisted by querying the whiteListUrl provided in authConfig
  */
  public get isIpWhiteListed(): Observable<Boolean> {
    if (!this.config.whiteListUrl) {
      throw new Error('Please configure whiteListUrl to enable this feature');
    }

    return this.httpClient.get(this.config.whiteListUrl, { observe: 'response' }).pipe(
      map((response) => String(response.status) === '200'),
      catchError((err, response) => observableOf(false)));
  }

  public hasPermissions(permissions: string[]): Observable<boolean> {
    return this.user.pipe(
      map((user) => {
        if (!user) {
          return new Set();
        }

        return user.roles.reduce((prevSet: Set<string>, role) => {
          this.config.rolesPermissionsMapping[role].forEach((permission) => prevSet.add(permission));
          return prevSet;
        }, new Set<string>());
      }),
      map((permissionsSet) => permissions.every(permission => permissionsSet.has(permission)))
    );
  }

  /*
  * Logout the user by removing his JWT from local storage
  */
  public logout(): Observable<Boolean> {
    return Observable.create((observer: Observer<Boolean>) => {
      this.config.tokenSetter('');
      this.user.next(this.token);
      observer.next(true);
    });
  }

  /*
  * Refresh the JWT by querying the refresh url provided in previous login response
  */
  public refreshToken(): Observable<boolean> {
    return this.httpClient
      .get(localStorage.getItem('refreshTokenUrl'))
      .pipe(
        ...this.postLoginOperations,
        map(() => true)
      );
  }
}
