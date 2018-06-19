import { Injectable, Inject, Optional } from '@angular/core';
import { IPasswordReset, AuthConfig, PasswordResetResult } from '../interfaces';
import { HttpClient } from '@angular/common/http';
import { authConfig } from '../tokens';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class PasswordResetService implements IPasswordReset {

  constructor(
    @Optional()
    @Inject(authConfig)
    private config: AuthConfig,
    private httpClient: HttpClient,
  ) {

  }

  resetPassword(password: string): Observable<PasswordResetResult> {
    return this.httpClient.post(this.config.passwordResetUrl, {password}).pipe(
      map(() => ({ success: true}))
    );
  }

}
