import { Injectable, Optional, Inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { JwtHelperService } from '@auth0/angular-jwt';

import { map, catchError, tap, filter, switchMap, distinctUntilChanged } from 'rxjs/operators';
import { of as observableOf, Observable, Observer, BehaviorSubject, interval } from 'rxjs';
import { authConfig } from '../tokens';
import { ILoginService, Credentials, AuthConfig, LoginResult, User } from '../interfaces';

@Injectable({
  providedIn: 'root'
})
export class StandardLoginService implements ILoginService {

  private readonly AUTO_REFRESH_INTERVAL = (this.config && this.config.autoRefreshInterval) || 60000;

  constructor(
    @Optional()
    @Inject(authConfig)
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
        }))
      }),
      tap(() => this.user.next(this.token))
    )
    .subscribe()
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

  private get token() {
    const token = localStorage.getItem(this.config.tokenName);
    if (token) {
      return this.jwtHelper.decodeToken(token);
    } else {
      return null;
    }
  }

  public user: BehaviorSubject<User> = new BehaviorSubject(this.token);

  public login(credential: Credentials): Observable<LoginResult> {
    const requestOptions = this.createRequestOptions(credential);
    return this.httpClient.get<LoginResult>(this.config.loginUrl, requestOptions).pipe(
      tap((result) => localStorage.setItem(this.config.tokenName, result.token)),
      tap((result) => localStorage.setItem('refreshTokenUrl', result.refresh)),
      tap(() => this.user.next(this.token))
    );
  }

  public get isIpWhiteListed(): Observable<Boolean> {
    return this.httpClient.get(this.config.whiteListUrl, { observe: 'response' }).pipe(
      map((response) => String(response.status) === '200'),
      catchError((err, response) => observableOf(false)));
  }

  public logout(): Observable<Boolean> {
    return Observable.create((observer: Observer<Boolean>) => {
      localStorage.setItem(this.config.tokenName, '');
      this.user.next(this.token);
      observer.next(true);
    });
  }

  public refreshToken(): Observable<boolean> {
    return this.httpClient.get(localStorage.getItem('refreshTokenUrl')).pipe(map(() => true));
  }

  public isLoggedIn: Observable<Boolean> = this.user.pipe(
    map(() => {
      const rawToken = localStorage.getItem(this.config.tokenName) || null;
      return !!(rawToken && !this.jwtHelper.isTokenExpired(rawToken));
    }),
    distinctUntilChanged()
  );
}
