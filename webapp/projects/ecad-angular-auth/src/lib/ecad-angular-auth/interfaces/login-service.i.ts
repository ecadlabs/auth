import { LoginResult } from './loginResult.i';
import { Observable } from 'rxjs';
import { Credentials } from './credentials.i';
import { UserToken } from './user-token.i';

export interface ILoginService<T = {}> {
    user: Observable<UserToken & T>;
    isIpWhiteListed: Observable<Boolean>;
    isLoggedIn: Observable<Boolean>;
    login(credentials: Credentials): Observable<LoginResult>;
    logout(): Observable<Boolean>;
    hasPermissions(permissions: string[]): Observable<boolean>;
    updateEmail(id: string, email: string): Observable<void>;
    validateEmailChange(token: string): Observable<void>;
    hasOneOfRoles(allowedRoles: string[]): Observable<boolean>;
}
