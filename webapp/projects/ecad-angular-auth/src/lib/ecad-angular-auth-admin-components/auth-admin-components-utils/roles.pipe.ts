import { Inject, Pipe, PipeTransform } from '@angular/core';
import { IUsersService } from '../../ecad-angular-auth-admin/interfaces/user-service.i';
import { USERS_SERVICE } from '../../ecad-angular-auth-admin/tokens';

@Pipe({
  name: 'roles'
})
export class RolesPipe implements PipeTransform {
  constructor(
    @Inject(USERS_SERVICE)
    private userService: IUsersService
  ) {}

  private getRoleDisplayValue(role: string) {
    return (
      this.userService.getRoles().find(r => r.value === role) || {
        displayValue: 'Unknown'
      }
    ).displayValue;
  }

  transform(value: { roles: { [key: string]: any } }, args?: any): any {
    return Object.keys(value.roles || [])
      .map(role => this.getRoleDisplayValue(role))
      .join(', ');
  }
}
