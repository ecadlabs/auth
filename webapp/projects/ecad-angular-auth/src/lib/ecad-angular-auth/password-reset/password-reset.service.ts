import { Injectable, Inject, Optional } from '@angular/core';
import { AuthConfig } from '../interfaces/auth-config.i';
import { IPasswordReset } from '../interfaces/password-reset.i';
import { PasswordResetResult } from '../interfaces/password-reset-result.i';
import { HttpClient } from '@angular/common/http';
import { AUTH_CONFIG } from '../tokens';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { PasswordResetEmailResult } from '../interfaces/password-reset-email-result.i';

@Injectable({
  providedIn: 'root'
})
export class PasswordResetService implements IPasswordReset {

  constructor(
    @Optional()
    @Inject(AUTH_CONFIG)
    private config: AuthConfig,
    private httpClient: HttpClient,
  ) {

  }

  resetPassword(token: string, password: string): Observable<PasswordResetResult> {
    return this.httpClient.post(this.config.passwordResetUrl, { token, password }).pipe(
      map(() => ({ success: true }))
    );
  }

  sendResetEmail(email: string): Observable<PasswordResetEmailResult> {
    return this.httpClient.post(this.config.sendResetEmailUrl, { email }).pipe(map(() => ({
      success: true
    })));
  }

}
