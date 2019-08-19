import { Injectable, Optional, Inject } from '@angular/core';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';

import { map, catchError, tap, filter, switchMap } from 'rxjs/operators';
import {
  of as observableOf,
  Observable,
  Observer,
  BehaviorSubject,
  interval,
  MonoTypeOperatorFunction,
  throwError
} from 'rxjs';
import { AUTH_CONFIG } from '../tokens';
import { ILoginService } from '../interfaces/login-service.i';
import { UserToken } from '../interfaces/user-token.i';
import { AuthConfig } from '../interfaces/auth-config.i';
import { Credentials } from '../interfaces/credentials.i';
import { LoginResult } from '../interfaces/loginResult.i';
import { JwtHelperService } from '../jwt-helper.service';

@Injectable({
  providedIn: 'root'
})
export class StandardLoginService implements ILoginService {
  static readonly DEFAULT_PREFIX = 'com.ecadlabs.auth';
  private readonly AUTO_REFRESH_INTERVAL =
    (this.config && this.config.autoRefreshInterval) || 60000;
  public user: BehaviorSubject<UserToken> = new BehaviorSubject(this.token);

  public isLoggedIn: Observable<Boolean> = this.user.pipe(
    map(() => {
      const rawToken = this.config.tokenGetter() || null;
      return !!(rawToken && !this.jwtHelper.isTokenExpired(rawToken));
    })
  );

  private readonly postLoginOperations: MonoTypeOperatorFunction<
    LoginResult
  > = (obserbable: Observable<LoginResult>) => {
    return obserbable.pipe(
      tap(result => this.config.tokenSetter(result.token)),
      tap(result => localStorage.setItem('refreshTokenUrl', result.refresh)),
      tap(() => this.user.next(this.getTokenAndCheckExp()))
    );
  }

  constructor(
    @Optional()
    @Inject(AUTH_CONFIG)
    private config: AuthConfig,
    private httpClient: HttpClient,
    private jwtHelper: JwtHelperService
  ) {
    this.logoutIfExpired();
    this.initAutoRefresh();
  }

  private initAutoRefresh() {
    this.isLoggedIn
      .pipe(
        filter(isLoggedIn => !!isLoggedIn),
        switchMap(() => this.user),
        switchMap(user => {
          return interval(this.AUTO_REFRESH_INTERVAL).pipe(
            switchMap(() => {
              return this.refreshToken().pipe(
                catchError(err => {
                  // If we get a 401 from the refresh endpoint it means that the user or tenant no longer exsits
                  // We logout in order to force the user to reauthenticate
                  if (err instanceof HttpErrorResponse && err.status === 401) {
                    this.logout().subscribe();
                    return throwError(err);
                  } else {
                    return observableOf(false);
                  }
                })
              );
            })
          );
        }),
        tap(() => this.user.next(this.getTokenAndCheckExp()))
      )
      .subscribe();
  }

  private createRequestOptions(credential: Credentials) {
    const credentialString = btoa(
      `${credential.username}:${credential.password}`
    );
    return {
      headers: {
        Authorization: `Basic ${credentialString}`,
        ['Content-Type']: 'application/x-www-form-urlencoded'
      }
    };
  }

  private getPrefixed(token: any, propName: string, prefix: string) {
    return token[`${prefix}.${propName}`];
  }

  private logoutIfExpired() {
    const token = this.config.tokenGetter();
    if (token && this.jwtHelper.isTokenExpired(token)) {
      return this.logout().subscribe();
    }
  }

  private get token(): UserToken {
    const token = this.config.tokenGetter();
    if (token) {
      const decodedToken = this.jwtHelper.decodeToken(token);
      const email = this.getPrefixed(
        decodedToken,
        'email',
        this.config.tokenPropertyPrefix
      );
      const name = this.getPrefixed(
        decodedToken,
        'name',
        this.config.tokenPropertyPrefix
      );
      const tenant = this.getPrefixed(
        decodedToken,
        'tenant',
        this.config.tokenPropertyPrefix
      );
      const member = this.getPrefixed(
        decodedToken,
        'member',
        this.config.tokenPropertyPrefix
      );
      const roles = this.getPrefixed(
        decodedToken,
        'roles',
        this.config.tokenPropertyPrefix
      );
      const permissions = this.getPrefixed(
        decodedToken,
        'permissions',
        this.config.tokenPropertyPrefix
      );
      return {
        email,
        name,
        roles,
        permissions,
        tenant,
        member,
        ...decodedToken
      };
    } else {
      return null;
    }
  }

  private getTokenAndCheckExp(): UserToken {
    this.logoutIfExpired();
    return this.token;
  }

  /**
   * Authenticate the user with basic auth using loginUrl provided in authConfig
   */
  public login(credential?: Credentials): Observable<LoginResult> {
    const requestOptions = credential
      ? this.createRequestOptions(credential)
      : {};
    return this.httpClient
      .get<LoginResult>(this.config.loginUrl, requestOptions)
      .pipe(this.postLoginOperations);
  }

  /*
   * Check if the ip is whitelisted by querying the whiteListUrl provided in authConfig
   */
  public get isIpWhiteListed(): Observable<Boolean> {
    if (!this.config.whiteListUrl) {
      throw new Error('Please configure whiteListUrl to enable this feature');
    }

    return this.httpClient
      .get(this.config.whiteListUrl, { observe: 'response' })
      .pipe(
        map(response => String(response.status) === '200'),
        catchError((err, response) => observableOf(false))
      );
  }

  public updateEmail(id: string, email: string) {
    return this.httpClient.post<void>(this.config.emailUpdateUrl, {
      id,
      email
    });
  }

  public validateEmailChange(token: string) {
    return this.httpClient.post<void>(this.config.emailChangeValidationUrl, {
      token
    });
  }

  public hasPermissions(permissions: string[]): Observable<boolean> {
    return this.user.pipe(
      map(user => {
        if (!user) {
          return new Set();
        }

        return (user.roles || []).reduce((prevSet: Set<string>, role) => {
          (this.config.rolesPermissionsMapping[role] || []).forEach(
            permission => prevSet.add(permission)
          );
          return prevSet;
        }, new Set<string>(user.permissions));
      }),
      map(permissionsSet =>
        permissions.every(permission => permissionsSet.has(permission))
      )
    );
  }

  public hasOneOfRoles(allowedRoles: string[]): Observable<boolean> {
    return this.user.pipe(
      map(user => {
        const userRoles =
          user && user.roles && Array.isArray(user.roles) ? user.roles : [];
        return userRoles.some(userRole => allowedRoles.includes(userRole));
      })
    );
  }

  /*
   * Logout the user by removing his JWT from local storage
   */
  public logout(): Observable<Boolean> {
    return Observable.create((observer: Observer<Boolean>) => {
      this.config.tokenSetter('');
      localStorage.removeItem('refreshTokenUrl');
      this.user.next(this.token);
      observer.next(true);
    });
  }

  private getRefreshUrl() {
    return this.config.refreshUrl || localStorage.getItem('refreshTokenUrl');
  }

  /*
   * Refresh the JWT by querying the refresh url provided in previous login response
   */
  public refreshToken(): Observable<boolean> {
    const headers = {
      headers: {
        Authorization: `Bearer ${this.config.tokenGetter()}`
      }
    };
    return this.httpClient.get(this.getRefreshUrl(), { ...headers }).pipe(
      this.postLoginOperations,
      map(() => true)
    );
  }
}
