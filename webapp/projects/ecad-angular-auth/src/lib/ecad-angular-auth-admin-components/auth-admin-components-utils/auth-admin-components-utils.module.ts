import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RolesPipe } from './roles.pipe';
import { TenantsPipe } from './tenants.pipe';

@NgModule({
  declarations: [RolesPipe, TenantsPipe],
  imports: [CommonModule],
  exports: [RolesPipe, TenantsPipe]
})
export class AuthAdminComponentsUtilsModule {}
