import { Pipe, PipeTransform } from '@angular/core';
import { User } from '../../ecad-angular-auth-admin/interfaces/user.i';

@Pipe({
  name: 'tenants'
})
export class TenantsPipe implements PipeTransform {
  transform(value: User, args?: any): any {
    return value.membership
      .map(x => x.tenant_name)
      .filter(x => x)
      .join(', ');
  }
}
