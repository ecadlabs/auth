import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ServiceAccountListComponent } from './service-account-list.component';
import {
  MatIconModule,
  MatTableModule,
  MatPaginatorModule,
  MatSortModule,
  MatButtonModule,
  MatTooltipModule
} from '@angular/material';
import { AuthAdminComponentsUtilsModule } from '../auth-admin-components-utils/auth-admin-components-utils.module';

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
    AuthAdminComponentsUtilsModule
  ],
  exports: [ServiceAccountListComponent]
})
export class ServiceAccountListModule {}
