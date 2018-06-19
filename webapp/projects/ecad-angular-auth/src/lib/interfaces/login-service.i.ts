import { LoginResult } from './loginResult.i';
import { Observable } from 'rxjs';
import { Credentials } from './credentials.i';

export interface ILoginService {
    username: string;
    isIpWhiteListed: Observable<Boolean>;
    isLoggedIn(): Boolean;
    login(credentials: Credentials): Observable<LoginResult>;
    logout(): Observable<Boolean>;
}
