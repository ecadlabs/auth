import { Observable } from 'rxjs';
import { PasswordResetResult } from './password-reset-result.i';
import { PasswordResetEmailResult } from './password-reset-email-result.i';

export interface IPasswordReset {
    resetPassword(resetToken: string, password: string): Observable<PasswordResetResult>;
    sendResetEmail(email: string): Observable<PasswordResetEmailResult>;
}
