import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { ServiceAccountListModule } from './service-account-list/service-account-list.module';
import { UserDetailModule } from './user-detail/user-detail.module';
import { UserLogsModule } from './user-logs/user-logs.module';
import { UsersListModule } from './users-list/users-list.module';
@NgModule({
  imports: [
    CommonModule,
    UsersListModule,
    ServiceAccountListModule,
    UserDetailModule,
    UserLogsModule
  ],
  declarations: [],
  exports: [
    UsersListModule,
    ServiceAccountListModule,
    UserDetailModule,
    UserLogsModule
  ]
})
export class EcadAngularAuthAdminComponentsModule {}
