import { PagedResult, FilterCondition } from '../../resource-util/resources.service';
import { User } from '../interfaces/user.i';
import { CreateUser } from '../interfaces/create-user.i';
import { UpdateUser } from '../interfaces/update-user.i';
import { Observable } from 'rxjs';

export interface IUsersService {
    getRoles(): {
        value: string;
        displayValue: string;
    }[];
    create(payload: CreateUser): Observable<User>;
    updateEmail(id: string, email: string): Observable<void>;

    update(payload: UpdateUser, addedRoles?: string[], deletedRoles?: string[]): Observable<User>;

    delete(id: string): Observable<boolean>;

    fetch(
        filter?: FilterCondition<User>[],
        sortBy?: keyof User,
        orderBy?: 'asc' | 'desc'
    ): Observable<PagedResult<User>>;

    find(id: string): Observable<User>;
    fetchNextPage(result: PagedResult<User>): Observable<PagedResult<User>>;
    fetchPreviousPage(result: PagedResult<User>): Observable<PagedResult<User>>;
}
