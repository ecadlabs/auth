import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import {
  MatButtonModule,
  MatIconModule,
  MatPaginatorModule,
  MatSortModule,
  MatTableModule,
  MatTooltipModule
} from '@angular/material';
import { AuthAdminComponentsUtilsModule } from '../auth-admin-components-utils/auth-admin-components-utils.module';
import { ServiceAccountEditFormModule } from '../service-account-edit-form/service-account-edit-form.module';
import { ServiceAccountListComponent } from './service-account-list.component';

@NgModule({
  declarations: [ServiceAccountListComponent],
  imports: [
    CommonModule,
    MatIconModule,
    MatTableModule,
    MatPaginatorModule,
    MatSortModule,
    MatButtonModule,
    MatTooltipModule,
    AuthAdminComponentsUtilsModule,
    ServiceAccountEditFormModule
  ],
  exports: [ServiceAccountListComponent]
})
export class ServiceAccountListModule {}
