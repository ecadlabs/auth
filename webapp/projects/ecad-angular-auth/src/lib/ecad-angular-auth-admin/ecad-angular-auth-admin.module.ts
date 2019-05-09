import { NgModule, ModuleWithProviders } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AuthAdminConfig } from './interfaces/auth-admin-config.i';
import {
  AUTH_ADMIN_CONFIG,
  USERS_SERVICE,
  USER_LOG_SERVICE,
  USER_MEMBERSHIPS_FACTORY,
  TENANTS_SERVICE
} from './tokens';
import { UsersService } from './users/users.service';
import { ResourceUtilModule } from '../resource-util/resource-util.module';
import { LogsService } from './logs/logs.service';
import { UserMembershipsService } from './members/members.service';
import { ResourcesService } from '../resource-util/resources.service';
import { TenantsService } from './tenants/tenants.service';

@NgModule({
  imports: [CommonModule, ResourceUtilModule],
  declarations: []
})
export class EcadAngularAuthAdminModule {
  public static forRoot(config: AuthAdminConfig): ModuleWithProviders {
    return {
      ngModule: EcadAngularAuthAdminModule,
      providers: [
        { provide: USERS_SERVICE, useClass: UsersService },
        { provide: TENANTS_SERVICE, useClass: TenantsService },
        { provide: USER_LOG_SERVICE, useClass: LogsService },
        { provide: AUTH_ADMIN_CONFIG, useValue: config },
        {
          provide: USER_MEMBERSHIPS_FACTORY,
          useFactory: userMembershipsFactory,
          deps: [ResourcesService, AUTH_ADMIN_CONFIG]
        }
      ]
    };
  }
}

export function userMembershipsFactory(
  resourceService: ResourcesService<any, any>,
  authConfig: AuthAdminConfig
) {
  class UserMembershipsServiceCreate {
    static create(userId: string) {
      const result = new UserMembershipsService(
        userId,
        resourceService,
        authConfig
      );
      return result;
    }
  }
  return UserMembershipsServiceCreate.create;
}
