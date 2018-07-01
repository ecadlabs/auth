import { NgModule, ModuleWithProviders } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AuthAdminConfig } from './interfaces/auth-admin-config.i';
import { AUTH_ADMIN_CONFIG, USERS_SERVICE } from './tokens';
import { UsersService } from './users/users.service';
import { ResourceUtilModule } from '../resource-util/resource-util.module';

@NgModule({
  imports: [
    CommonModule,
    ResourceUtilModule
  ],
  declarations: []
})
export class EcadAngularAuthAdminModule {
  public static forRoot(config: AuthAdminConfig): ModuleWithProviders {
    return {
      ngModule: EcadAngularAuthAdminModule,
      providers: [
        { provide: USERS_SERVICE, useClass: UsersService },
        { provide: AUTH_ADMIN_CONFIG, useValue: config },
      ]
    };
  }
}
