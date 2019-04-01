import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RolesPipe } from './roles.pipe';

@NgModule({
  declarations: [RolesPipe],
  imports: [CommonModule],
  exports: [RolesPipe]
})
export class AuthAdminComponentsUtilsModule {}
