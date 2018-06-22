import { Injectable } from '@angular/core';
import { ResourcesService, PagedResult, FilterCondition } from './resources.service';

export interface CreateUser {
  email: string;
  password: string;
  name?: string;
  roles?: string[];
}

export interface User {
  id: string;
  email: string;
  name: string;
  added: string;
  modified: string;
  email_verified: boolean;
  roles: string[];
}

@Injectable({
  providedIn: 'root'
})
export class UsersService {

  constructor(
    private resourcesService: ResourcesService<User, CreateUser>
  ) { }

  private get apiEndpoint() {
    return '/api/v1/users';
  }

  create(payload: CreateUser) {
    return this.resourcesService.create(this.apiEndpoint, payload);
  }

  delete(id: string) {
    return this.resourcesService.delete(this.apiEndpoint, id);
  }

  fetch(filter: FilterCondition<User>[] = [], sortBy: keyof User = 'added', orderBy: 'asc' | 'desc' = 'desc') {
    return this.resourcesService.fetch(this.apiEndpoint, filter, sortBy, orderBy);
  }

  find(id: string) {
    return this.resourcesService.find(this.apiEndpoint, id);
  }

  fetchNextPage(result: PagedResult<User>)  {
    return this.resourcesService.fetchNextPage(result);
  }

  fetchPreviousPage(result: PagedResult<User>) {
    return this.resourcesService.fetchPreviousPage(result);
  }
}
