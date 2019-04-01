import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
  name: 'roles'
})
export class RolesPipe implements PipeTransform {
  transform(value: { roles: { [key: string]: any } }, args?: any): any {
    return Object.keys(value.roles || [])
      .map(str => {
        return `${str[0].toUpperCase()}${str.substr(1)}`;
      })
      .join(', ');
  }
}
