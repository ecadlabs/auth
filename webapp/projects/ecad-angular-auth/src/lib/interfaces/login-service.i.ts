import { LoginResult } from './loginResult.i';
import { Observable } from 'rxjs';
import { Credentials } from './credentials.i';
import { User } from './user.i';

export interface ILoginService<T = {}> {
    user: Observable<User & T>;
    isIpWhiteListed: Observable<Boolean>;
    isLoggedIn: Observable<Boolean>;
    login(credentials: Credentials): Observable<LoginResult>;
    logout(): Observable<Boolean>;
}
